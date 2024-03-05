package main

import (
	"AuthService/handlers"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", handlers.HandleMain)
	http.HandleFunc("/login/github", handlers.HandleGitHubLogin)
	http.HandleFunc("/callback/github", handlers.HandleGitHubCallback)
	http.HandleFunc("/login/google", handlers.HandleGoogleLogin)
	http.HandleFunc("/callback/google", handlers.HandleGoogleCallback)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
