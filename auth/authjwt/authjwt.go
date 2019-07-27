package authjwt

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/auth_server/db"
	_ "github.com/lib/pq"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
)

type user struct {
	id       int
	name     string
	mail     string
	password string
}

func (u user) createToken() string {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["admin"] = true
	claims["sub"] = strconv.Itoa(u.id)
	claims["name"] = u.name
	claims["mail"] = u.mail
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	hasura := make(map[string]interface{})
	hasura["x-hasura-user-id"] = strconv.Itoa(u.id)
	hasura["x-hasura-default-role"] = "user"
	hasura["x-hasura-allowed-roles"] = []string{"editor", "user", "mod", "admin"}

	claims["https://hasura.io/jwt/claims"] = hasura
	tokenString, _ := token.SignedString([]byte(os.Getenv("SIGNINGKEY")))
	return tokenString
}

// GetTokenHandler get token
var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	d, err := db.Connect()
	if err != nil {
		fmt.Println(err)
	}
	defer d.Close()

	mail, pass, _ := r.BasicAuth()

	var u user
	err = d.QueryRow(`
			SELECT *
			FROM users
			WHERE mail = $1 AND password = $2;
		`, mail, pass).Scan(&(u.id), &(u.name), &(u.mail), &(u.password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized\n"))
		return
	}
	tokenString := u.createToken()
	w.Write([]byte(tokenString))
})

// JwtMiddleware check token
var JwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SIGNINGKEY")), nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})
