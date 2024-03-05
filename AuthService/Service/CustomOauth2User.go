package service

import (
	"AuthService/data"
	"AuthService/grpc"
	user2 "AuthService/proto/user"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"time"
)

type CustomOAuth2UserService struct {
	userClient   *grpc.UserClient
	auditClient  *grpc.AuditClient
	oauth2Config *oauth2.Config
}

func NewCustomOAuth2UserService(oAuth2Config *oauth2.Config, userClient grpc.UserClient, auditClient grpc.AuditClient) *CustomOAuth2UserService {
	return &CustomOAuth2UserService{
		userClient:   &userClient,
		auditClient:  &auditClient,
		oauth2Config: oAuth2Config,
	}
}

func (s *CustomOAuth2UserService) ProcessUserDetails(accessToken, providerName, ipaddress string) (*user2.UserProto, error) {
	var err error
	var user *user2.UserProto
	var email, sub, name string
	if providerName == "github" {
		email, err = s.fetchEmailFromGitHub(accessToken)
		if err != nil {
			return nil, err
		}
		sub, name, err = s.fetchSubFromGithub(accessToken)
		if err != nil {
			return nil, err
		}
	}
	if providerName == "google" {
		sub, err = s.fetchSubFromGoogle(accessToken)
		if err != nil {
			return nil, err
		}
		email, name, err = s.fetchEmailAndNameFromGoogle(accessToken)
	}
	user, err = s.userClient.ProcessOAuthUser(context.Background(), email, sub, providerName, name)
	if err != nil {
		return nil, err
	}

	_ = s.auditClient.LogEvent("User logged in", user.Username, time.DateTime, "User logged in", ipaddress)
	return user, nil
}

func (s *CustomOAuth2UserService) fetchEmailFromGitHub(accessToken string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []data.EmailInfo
	err = json.Unmarshal(body, &emails)
	if err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no primary email found")
}

type GithubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

func (s *CustomOAuth2UserService) fetchSubFromGithub(accessToken string) (string, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned non-OK status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var user GithubUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		return "", "", err
	}

	// Devuelve tanto el ID como el Login del usuario
	return fmt.Sprintf("%d", user.ID), user.Login, nil
}

func (s *CustomOAuth2UserService) fetchSubFromGoogle(accessToken string) (string, error) {
	userInfo, err := s.oauth2Config.Client(context.Background(), &oauth2.Token{AccessToken: accessToken}).Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(userInfo.Body)

	body, err := io.ReadAll(userInfo.Body)
	if err != nil {
		return "", err
	}

	var googleUser struct {
		Sub string `json:"sub"`
	}
	err = json.Unmarshal(body, &googleUser)
	if err != nil {
		return "", err
	}

	return googleUser.Sub, nil

}

func (s *CustomOAuth2UserService) fetchEmailAndNameFromGoogle(accessToken string) (string, string, error) {
	userInfo, err := s.oauth2Config.Client(context.Background(), &oauth2.Token{AccessToken: accessToken}).Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(userInfo.Body)

	body, err := io.ReadAll(userInfo.Body)
	if err != nil {
		return "", "", err
	}

	var googleUser GoogleUser
	err = json.Unmarshal(body, &googleUser)
	if err != nil {
		return "", "", err
	}

	return googleUser.Email, googleUser.Name, nil
}

type GoogleUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
