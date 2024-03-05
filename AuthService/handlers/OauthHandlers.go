package handlers

import (
	service "AuthService/Service"
	cf "AuthService/config"
	"AuthService/grpc"
	"AuthService/util"
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

func HandleMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<a href="/login/github">Login with GitHub</a>`)
	fmt.Fprintf(w, `<a href="/login/google">Login with Google</a>`)
}

func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := cf.OAuth2ConfigGithub.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	tokenValid := util.NewTokenServiceClient()
	code := r.FormValue("code")
	token, err := cf.OAuth2ConfigGithub.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userClient, err := grpc.NewUserClient("localhost", 9091)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}
	defer func(userClient *grpc.UserClient) {
		err := userClient.Close()
		if err != nil {

		}
	}(userClient)

	auditClient, err := grpc.NewAuditClient("localhost", 50052)
	if err != nil {
		log.Fatalf("Failed to create audit client: %v", err)
	}
	defer auditClient.Close()

	service := service.NewCustomOAuth2UserService(cf.OAuth2ConfigGithub, *userClient, *auditClient)

	user, err := service.ProcessUserDetails(token.AccessToken, "github", r.RemoteAddr)
	if err != nil {
		log.Fatalf("Error processing user details: %v", err)
	}

	JWT, _ := tokenValid.GenerateToken(user.Username)

	fmt.Fprintf(w, "Hello, %s!", user.Username+" JWT: "+JWT)

}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := cf.OAuth2ConfigGoogle.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	tokenValid := util.NewTokenServiceClient()
	code := r.FormValue("code")
	token, err := cf.OAuth2ConfigGoogle.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userClient, err := grpc.NewUserClient("localhost", 9091)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}
	defer func(userClient *grpc.UserClient) {
		err := userClient.Close()
		if err != nil {

		}
	}(userClient)

	auditClient, err := grpc.NewAuditClient("localhost", 50052)
	if err != nil {
		log.Fatalf("Failed to create audit client: %v", err)
	}
	defer auditClient.Close()

	service := service.NewCustomOAuth2UserService(cf.OAuth2ConfigGoogle, *userClient, *auditClient)

	user, err := service.ProcessUserDetails(token.AccessToken, "google", r.RemoteAddr)
	if err != nil {
		log.Fatalf("Error processing user details: %v", err)
	}

	JWT, _ := tokenValid.GenerateToken(user.Username)

	fmt.Fprintf(w, "Hello, %s!", user.Username+" JWT: "+JWT)

}
