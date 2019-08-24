package authjwt

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"

	"github.com/auth_server/db"
	_ "github.com/lib/pq"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	Mail             string `json:"mail"`
	password         string
	Info_date        string         `json:"info_date"`
	Photo            string         `json:"photo,omitempty"`
	Github_user      string         `json:"github_user,omitempty"`
	photo_null       sql.NullString `json:"photo"`
	github_user_null sql.NullString `json:"github_user"`
}

type UserRegister struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Mail        string `json:"mail"`
	Password    string `json:"password"`
	Code        int    `json:"code"`
	Photo       string `json:"photo,omitempty"`
	Github_user string `json:"github_user,omitempty"`
}

func (u User) createToken() string {
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

	var u User
	err = d.QueryRow(`
			SELECT id, name, mail, info_date, photo, github_user
			FROM users
			WHERE mail = $1 AND password = $2;
		`, mail, pass).Scan(&(u.Id), &(u.Name), &(u.Mail), &(u.Info_date), &(u.photo_null), &(u.github_user_null))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(err)
		jsonReturn := make(map[string]interface{})
		jsonReturn["error"] = err

		jsError, _ := json.Marshal(jsonReturn)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsError)
		return
	}
	u.Photo = u.photo_null.String
	u.Github_user = u.github_user_null.String

	tokenString := u.createToken()

	jsonReturn := make(map[string]interface{})
	jsonReturn["user"] = u
	jsonReturn["token"] = tokenString

	js, err := json.Marshal(jsonReturn)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
})

var GetRegisterHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	d, err := db.Connect()
	if err != nil {
		fmt.Println(err)
	}
	defer d.Close()

	decoder := json.NewDecoder(r.Body)
	var register UserRegister
	err = decoder.Decode(&register)
	if err != nil {
		panic(err)
	}

	id := 0
	err = d.QueryRow(`
			INSERT INTO users (name, mail, password, photo) VALUES ($1, $2, $3, $4)
			RETURNING id`,
		register.Name, register.Mail, register.Password, register.Photo).Scan(&id)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(err)
		jsonReturn := make(map[string]interface{})
		jsonReturn["error"] = err

		jsError, _ := json.Marshal(jsonReturn)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsError)
		return
	}
	w.Write([]byte(string(id)))
})

var CheckMailHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var register UserRegister
	err := decoder.Decode(&register)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(err)
		jsonReturn := make(map[string]interface{})
		jsonReturn["error"] = err

		jsError, _ := json.Marshal(jsonReturn)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsError)
		return
	}

	if SendEmail(register.Mail, register.Name, register.Code) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(err)
		jsonReturn := make(map[string]interface{})
		jsonReturn["error"] = "Erro ao enviar email"

		jsError, _ := json.Marshal(jsonReturn)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsError)
		return
	}

	w.Write([]byte(string("Email enviado!")))
})

func SendEmail(mail string, name string, code int) bool {

	m := gomail.NewMessage()
	m.SetHeader("From", "perguntando@flutterando.com.br")
	m.SetHeader("To", mail)
	m.SetHeader("Subject", "Seu código de acesso!")
	m.SetBody("text/html", fmt.Sprintf("Opa! <i>%s</i>,<br>Seu código é: <b>%d</b>!", name, code))

	d := gomail.NewDialer("smtp.umbler.com", 587, "perguntando@flutterando.com.br", "Ja36451485")

	if err := d.DialAndSend(m); err != nil {
		return true
	} else {
		return false
	}

}

// JwtMiddleware check token
var JwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SIGNINGKEY")), nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})
