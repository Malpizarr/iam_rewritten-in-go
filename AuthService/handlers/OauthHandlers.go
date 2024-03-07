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
	_ "golang.org/x/oauth2/google"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type OAuthExchangeRequest struct {
	Code string `json:"code"`
}

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

func HandleGoogleLoginCLI(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	codeChallenge := r.URL.Query().Get("code_challenge")

	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?response_type=code&client_id=472975436513-qtrfpb07j2ngdbgf79vnufl0mbaegsum.apps.googleusercontent.com&redirect_uri=http://localhost:8080/oauth/exchange&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
		url.QueryEscape("https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email"), state, codeChallenge)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func HandleGoogleCallbackCLI(w http.ResponseWriter, r *http.Request) {
	OAuthExchange(w, r)
}

func OAuthExchange(w http.ResponseWriter, r *http.Request) {
	userClient, err := grpc.NewUserClient("localhost", 9091)
	auditClient, err := grpc.NewAuditClient("localhost", 9090)
	tokenValid := util.NewTokenServiceClient()
	code := r.URL.Query().Get("code")
	codeVerifier := r.URL.Query().Get("code_verifier")

	token, err := exchangeCodeForToken(code, codeVerifier)
	if err != nil {
		http.Error(w, fmt.Sprintf("No se pudo intercambiar el código por un token: %v", err), http.StatusInternalServerError)
		return
	}

	service := service.NewCustomOAuth2UserService(cf.OAuth2ConfigGoogleCLI, *userClient, *auditClient)

	user, err := service.ProcessUserDetails(token.AccessToken, "google", r.RemoteAddr)

	jwtToken, err := tokenValid.GenerateToken(user.Username)
	if err != nil {
		http.Error(w, fmt.Sprintf("No se pudo generar el JWT: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "http://localhost:8081/callback?jwt="+url.QueryEscape(jwtToken), http.StatusFound)

}

func exchangeCodeForToken(code, codeVerifier string) (*oauth2.Token, error) {

	ctx := context.Background()

	token, err := cf.OAuth2ConfigGoogleCLI.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("no se pudo intercambiar el código por un token: %w", err)
	}

	return token, nil
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

	auditClient, err := grpc.NewAuditClient("localhost", 9090)
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
