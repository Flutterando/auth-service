package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/auth_server/authjwt"

	"github.com/gorilla/mux"
)

type check struct {
	auth bool
}

func main() {
	r := mux.NewRouter()
	r.Handle("/v1/check", authjwt.JwtMiddleware.Handler(checkauth))
	r.Handle("/v1/gettoken", authjwt.GetTokenHandler)

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal("ListenAndServe:", nil)
	}
}

var checkauth = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	c := &check{
		auth: true,
	}
	json.NewEncoder(w).Encode(c)
})
