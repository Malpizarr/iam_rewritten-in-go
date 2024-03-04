package Service

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type TokenValidationService struct {
	AuthServiceUrl string
}

type TokenValidationRequest struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

func NewTokenValidationService() *TokenValidationService {
	return &TokenValidationService{
		AuthServiceUrl: "http://localhost:8082/token/validate",
	}
}

func (s *TokenValidationService) ValidateToken(token string, username string) bool {
	requestBody := TokenValidationRequest{
		Token:    token,
		Username: username,
	}
	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return false
	}

	resp, err := http.Post(s.AuthServiceUrl, "application/json", bytes.NewBuffer(jsonRequestBody))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var validationResponse bool
	err = json.NewDecoder(resp.Body).Decode(&validationResponse)
	if err != nil {
		return false
	}

	return validationResponse
}
