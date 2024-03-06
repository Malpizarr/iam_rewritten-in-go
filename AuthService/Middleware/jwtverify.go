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

		// Intenta obtener el token del encabezado de autorización
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Si el token no está en el encabezado de autorización, intenta obtenerlo de las cookies
		if token == "" {
			if cookie, err := r.Cookie("AUTH_TOKEN"); err == nil {
				token = cookie.Value
			}
		}

		// Si aún no hay token, devuelve un error
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

		// Llama al siguiente handler en la cadena
		next.ServeHTTP(w, r)
	})
}
