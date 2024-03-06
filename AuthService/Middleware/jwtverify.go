package Middleware

import (
	"AuthService/util"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["username"]

		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token == "" {
			if cookie, err := r.Cookie("AUTH_TOKEN"); err == nil {
				token = cookie.Value
			}
		}

		if token == "" {
			http.Error(w, "Authorization token is missing", http.StatusUnauthorized)
			return
		}

		tokenServiceClient := util.NewTokenServiceClient()
		isValid, err := tokenServiceClient.ValidateToken(token, username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !isValid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
