package main

import "database/sql"
import "fmt"
import "github.com/bmizerany/pq"
import "net/http"

type user struct {
	email string
}

var (
	db *sql.DB
)

func init() {
	conf, err := pq.ParseURL(RequireEnv("CORE_DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	db, err = sql.Open("postgres", conf)
	if err != nil {
		panic(err)
	}
}

func handler(res http.ResponseWriter, req *http.Request) {
	user, err := Authorize(db, req.Header.Get("Authorization"))
	if err != nil {
		panic(err)
	}
	if user != nil {
		fmt.Printf("authenticated user=%s\n", user.email)
	} else {
		fmt.Printf("unauthenticated\n")
	}
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("Hello web\n"))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":5000", nil)
}
