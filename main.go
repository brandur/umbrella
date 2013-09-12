package main

import "database/sql"
import "fmt"
import "github.com/bmizerany/pq"
import "net/http"
import "net/http/httputil"
import "net/url"

type User struct {
	email string
	id string
}

var (
	db *sql.DB
	proxy *httputil.ReverseProxy
)

func init() {
		db = openDB()
	url, err := url.Parse(RequireEnv("PROXY_URL"))
	if err != nil {
		panic(err)
	}
	proxy = httputil.NewSingleHostReverseProxy(url)
}

func handler(res http.ResponseWriter, req *http.Request) {
	user, err := Authorize(db, req.Header.Get("Authorization"))
	if err != nil {
		panic(err)
	}
	if user != nil {
		fmt.Printf("authenticated user=%s\n", user.email)
		req.Header.Set("X-Heroku-User-Email", user.email)

        sudo, err := AuthorizeSudo(db, user, req.Header.Get("X-Heroku-Sudo"))
		if err != nil {
			panic(err)
		}
		if sudo {
			req.Header.Set("X-Heroku-Sudo", "true")

			email := req.Header.Get("X-Heroku-Sudo-User")
		    fmt.Printf("sudo=true email=%s\n", email)
			if email != "" {
				req.Header.Set("X-Heroku-User-Email", email)
			}
		} else {
		    fmt.Printf("sudo=false\n")
		}
	} else {
		fmt.Printf("unauthenticated\n")
	}

	// scrub the authorization header
	req.Header.Set("Authorization", "")

	proxy.ServeHTTP(res, req)
}

func openDB() *sql.DB {
	conf, err := pq.ParseURL(RequireEnv("CORE_DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	db, err := sql.Open("postgres", conf)
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":5100", nil)
}
