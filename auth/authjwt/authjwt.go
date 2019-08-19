package authjwt

import (
	"encoding/json"
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
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Mail        string `json:"mail"`
	Password    string `json:"password"`
	Info_date   string `json:"info_date"`
	Photo       string `json:"photo"`
	Github_user string `json:"github_user"`
}

func (u user) createToken() string {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["admin"] = true
	claims["sub"] = strconv.Itoa(u.Id)
	claims["name"] = u.Name
	claims["mail"] = u.Mail
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	hasura := make(map[string]interface{})
	hasura["x-hasura-user-id"] = strconv.Itoa(u.Id)
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
		`, mail, pass).Scan(&(u.Id), &(u.Name), &(u.Mail), &(u.Password), &(u.Info_date), &(u.Photo), &(u.Github_user))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(err)
		w.Write([]byte("Unauthorized\n"))
		return
	}
	tokenString := u.createToken()

	jsonReturn := make(map[string]interface{})
	jsonReturn["user"] = u
	jsonReturn["token"] = tokenString

	js, err := json.Marshal(jsonReturn)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
})

// JwtMiddleware check token
var JwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SIGNINGKEY")), nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})
