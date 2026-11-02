package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func unauthorized(w http.ResponseWriter) {
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must contain at least 32 characters")
	}

	backendURL, err := url.Parse("http://httpbin:8080")
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("backend proxy error: %v", err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			unauthorized(w)
			return
		}

		rawToken := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if rawToken == "" {
			unauthorized(w)
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(
			rawToken,
			claims,
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(30*time.Second),
		)
		if err != nil || token == nil || !token.Valid {
			unauthorized(w)
			return
		}

		// Avoid forwarding the client's bearer credential to the legacy service.
		r.Header.Del("Authorization")

		// The proxy must not trust client-supplied forwarding metadata.
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("X-Forwarded-Host")
		r.Header.Del("X-Forwarded-Proto")

		proxy.ServeHTTP(w, r)
	})

	addr := ":8080"
	log.Printf("JWT reverse proxy listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
