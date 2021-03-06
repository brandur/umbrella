package main

import "crypto/hmac"
import "crypto/sha256"
import "database/sql"
import "encoding/base64"
import "encoding/hex"
import "errors"
import "fmt"
import "net/http"
import "strings"

var (
	key string
)

func init() {
	key = buildTokenKey()
}

func Authorize(db *sql.DB, authorization string) (*User, error) {
	fmt.Printf("authorization=%s\n", authorization)
	parts := strings.Fields(authorization)
	if len(parts) == 2 {
		if parts[0] == "Basic" {
			bytes, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return nil, err
			}
			pair := strings.SplitN(string(bytes), ":", 2)
			if len(pair) != 2 {
				return nil, nil
			}
			return authorizeWithToken(db, pair[1])
		} else if parts[0] == "Bearer" {
			return authorizeWithToken(db, parts[1])
		}
	} else {
		return nil, nil
	}
	return nil, nil
}

func AuthorizeSudo(db *sql.DB, user *User, header *http.Header) (*User, error) {
 	if header.Get("X-Heroku-Sudo") != "true" {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT up.feature
		FROM user_passes up INNER JOIN users u ON up.user_id = u.id
		WHERE u.email = $1
		AND up.feature = 'sudoer'`, user.email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("Not authorized for sudo.")
	}

    var sudoUser *User
	email := header.Get("X-Heroku-Sudo-User")
	if email != "" {
	    rows, err := db.Query(`
			SELECT u.email, u.uuid
			FROM users u
			WHERE email = $1`, email)

		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if !rows.Next() {
			return nil, errors.New(fmt.Sprintf("Sudo user \"%s\" does not exist.", email))
		}
		sudoUser = new(User)
		rows.Scan(&sudoUser.email, &sudoUser.id)
	} else {
	    sudoUser = user
	}
	return sudoUser, nil
}

func authorizeWithToken(db *sql.DB, token string) (*User, error) {
	fmt.Printf("token=%s\n", token)
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(token))
	mac := hex.EncodeToString(hash.Sum(nil))
	fmt.Printf("mac=%s\n", mac)

	rows, err := db.Query(`
		SELECT u.email, u.uuid
		FROM access_tokens at
		INNER JOIN authorizations a ON a.id = at.authorization_id
		INNER JOIN users u ON u.id = a.user_id
		WHERE token_hash = $1`, mac)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	user := new(User)
	rows.Scan(&user.email, &user.id)
	return user, nil
}

func buildTokenKey() string {
	hash := hmac.New(sha256.New, []byte(RequireEnv("TOKEN_ENV_KEY")))
	hash.Write([]byte(RequireEnv("TOKEN_STATIC_KEY")))
	return hex.EncodeToString(hash.Sum(nil))
}
