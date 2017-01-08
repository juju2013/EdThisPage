package main

import (
	_ "fmt"
	_ "github.com/dgrijalva/jwt-go"
	"github.com/goware/jwtauth"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	//r.Use(DefaultHtml)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/api/login", FakeLogin)
	r.Get("/api/login", FakeLogin)
	r.Post("/api/logout", Noop)
	r.FileServer("/page", http.Dir("."))

	// Protected API routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(TokenAuth.Verifier)
		r.Use(jwtauth.Authenticator)

		r.Post("/api/text", SaveText)
		/*		r.Post("/api", func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				token := ctx.Value("jwt").(*jwt.Token)
				claims := token.Claims
				w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user"])))
			})*/
	})
	http.ListenAndServe(":3000", r)
}
