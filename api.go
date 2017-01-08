package main

import (
	"fmt"
	_ "github.com/dgrijalva/jwt-go"
	"github.com/goware/jwtauth"
	"net/http"
)

var TokenAuth *jwtauth.JwtAuth

func init() {
	TokenAuth = jwtauth.New("HS256", []byte("XXX:Change me"), nil)

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	//_, tokenString, _ := TokenAuth.Encode(jwtauth.Claims{"user_id": 123})
}

func newJwtToken(user string) string {
	_, tokenString, _ := TokenAuth.Encode(jwtauth.Claims{"user": user})
	return tokenString
}

// fakeLogin : simulate a fake login, every thing will be mapped to user "fake"
func FakeLogin(w http.ResponseWriter, r *http.Request) {
	jwtToken := newJwtToken("fake")
	fmt.Printf("FakeLogin:%v\n", jwtToken)
	w.Write([]byte(jwtToken))
}

// SaveText : save content as text
func SaveText(w http.ResponseWriter, r *http.Request) {
	// always OK
	fmt.Println("Text saved.")
}

// Noop : No action
func Noop(w http.ResponseWriter, r *http.Request) {
	// always OK
	fmt.Println("Noop.")
}
