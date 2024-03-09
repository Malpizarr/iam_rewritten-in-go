package main

import (
	"AuthService/Middleware"
	"AuthService/config"
	"AuthService/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()

	authController := handlers.NewAuthController(
		config.UserService,
		config.EmailService,
		config.TokenServiceClient,
		config.AuditClient,
	)

	router.HandleFunc("/", handlers.HandleMain)
	router.HandleFunc("/login/github", handlers.HandleGitHubLogin)
	router.HandleFunc("/callback/github", handlers.HandleGitHubCallback)
	router.HandleFunc("/login/google", handlers.HandleGoogleLogin)
	router.HandleFunc("/callback/google", handlers.HandleGoogleCallback)
	router.HandleFunc("/login/google-cli", handlers.HandleGoogleLoginCLI)
	router.HandleFunc("/oauth/exchange", handlers.HandleGoogleCallbackCLI)
	router.HandleFunc("/auth/register", authController.Register)
	router.HandleFunc("/auth/login", authController.Login)
	router.HandleFunc("/auth/verify-2fa", authController.Verify2FA)
	router.HandleFunc("/auth/{username}/2fa/enable", func(w http.ResponseWriter, r *http.Request) {
		Middleware.AuthMiddleware(http.HandlerFunc(authController.Enable2FA)).ServeHTTP(w, r)
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
