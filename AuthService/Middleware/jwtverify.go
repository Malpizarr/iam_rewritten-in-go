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
		authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenServiceClient := util.NewTokenServiceClient()
		isValid, err := tokenServiceClient.ValidateToken(authHeader, username)
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
