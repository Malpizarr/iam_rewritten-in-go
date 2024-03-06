package handlers

import (
	cf "AuthService/config"
	"AuthService/grpc"
	service "AuthService/service"
	"AuthService/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	HandleOAuthCallback(w, r, cf.OAuth2ConfigGithub, "github")
}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := cf.OAuth2ConfigGoogle.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	HandleOAuthCallback(w, r, cf.OAuth2ConfigGoogle, "google")
}

func HandleOAuthCallback(w http.ResponseWriter, r *http.Request, oauthConfig *oauth2.Config, provider string) {
	tokenValid := util.NewTokenServiceClient()
	code := r.FormValue("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
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

	service := service.NewCustomOAuth2UserService(oauthConfig, *userClient, *auditClient)

	user, err := service.ProcessUserDetails(token.AccessToken, provider, r.RemoteAddr)
	if err != nil {
		var userErr *util.UserError
		if errors.As(err, &userErr) && strings.Contains(userErr.Error(), "2FA verification required") {
			encodedUsername := url.QueryEscape(userErr.Username)
			twoFaPageUrl := "http://localhost:3000/path-to-2fa-page.html?username=" + encodedUsername

			http.Redirect(w, r, twoFaPageUrl, http.StatusSeeOther)
			return
		}

		log.Fatalf("Error processing user details: %v", err)
	}

	jwtJSON, err := tokenValid.GenerateToken(user.Username)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var jwtResponse struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(jwtJSON), &jwtResponse); err != nil {
		http.Error(w, "Failed to parse JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	setCookie(w, jwtResponse.Token)

	targetURL := "http://localhost:3000/2FATEST.html?username=" + user.Username
	http.Redirect(w, r, targetURL, http.StatusSeeOther)
}

func setCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "AUTH_TOKEN",
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}
