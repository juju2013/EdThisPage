package main

import (
	_ "fmt"
	_ "github.com/dgrijalva/jwt-go"
	"github.com/goware/jwtauth"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"net/http"
	"os"
)

func main() {
	r := chi.NewRouter()
	//r.Use(DefaultHtml)
	//r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	//r.Use(middleware.RedirectSlashes)

	r.Post("/api/login", FakeLogin)
	r.Get("/api/login", FakeLogin)
	r.Post("/api/logout", Noop)

	d, _ := os.Getwd()
	StaticServer(r, "/", http.Dir(d))

	// Protected API routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(TokenAuth.Verifier)
		r.Use(jwtauth.Authenticator)

		r.Post("/api/text", SaveText)
	})
	http.ListenAndServe(":3000", r)
}
