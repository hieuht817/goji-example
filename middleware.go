package main

import (
	"net/http"
	"github.com/zenazn/goji/web"
	"strings"
	"encoding/base64"
)

// PlainText sets the content-type of response to text/plain.
func PlainText(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		h.ServeHTTP(w, r)
	})
}

// Nobody will ever guess this!
const Password = "admin:admin"

// SuperSecure is HTTP Basic Auth middleware for super-secret admin page. Shhhh!
func SuperSecure(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			pleaseAuth(w)
			return
		}

		password, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil || string(password) != Password {
			pleaseAuth(w)
			return
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func pleaseAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Gritter"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Go away!\n"))
}