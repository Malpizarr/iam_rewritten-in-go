package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type TokenServiceClient struct {
	Client *http.Client
}

func NewTokenServiceClient() *TokenServiceClient {
	return &TokenServiceClient{
		Client: &http.Client{},
	}
}

type TokenRequest struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

func (t *TokenServiceClient) ValidateToken(token, username string) (bool, error) {
	body := &TokenRequest{
		Token:    token,
		Username: username,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", "http://localhost:8082/token/validate", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}

func (t *TokenServiceClient) GenerateToken(username string) (string, error) {
	body := &TokenRequest{
		Username: username,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:8082/token/generate", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(bodyBytes), nil
	}
	return "", nil
}
