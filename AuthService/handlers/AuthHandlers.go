package handlers

import (
	"AuthService/data"
	"AuthService/grpc"
	service "AuthService/service"
	"AuthService/util"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type AuthController struct {
	userService        *service.UserService
	emailService       *service.EmailService
	tokenServiceClient *util.TokenServiceClient
	auditClient        *grpc.AuditClient
}

func NewAuthController(userService *service.UserService, emailService *service.EmailService, tokenServiceClient *util.TokenServiceClient, auditClient *grpc.AuditClient) *AuthController {
	return &AuthController{
		userService:        userService,
		emailService:       emailService,
		tokenServiceClient: tokenServiceClient,
		auditClient:        auditClient,
	}
}

func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var newUser data.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	user, err := ac.userService.Register(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ac.logAuditEvent("REGISTER", user.Username, time.Now().String(), "User registered successfully", ipAddress); err != nil {
		log.Printf("Error logging audit event for user %s: %v", user.Username, err)
	}
	err = json.NewEncoder(w).Encode(user.Username)
	if err != nil {
		return
	}
}

func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var user data.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	userResponse, err := ac.userService.Login(user)
	if err != nil {
		ac.handleLoginFailure(w, user.Username, err, ipAddress)
		return
	}

	token, err := ac.tokenServiceClient.GenerateToken(userResponse.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ac.logAuditEvent("LOGIN", userResponse.Username, time.Now().String(), "User logged in successfully from IP: "+ipAddress, ipAddress); err != nil {
		log.Printf("Error logging audit event for user %s: %v", userResponse.Username, err)
	}

	if err := ac.emailService.SendLoginEmail(userResponse.Email, ipAddress); err != nil {
		log.Printf("Error sending login email for user %s: %v", userResponse.Username, err)
	}
	w.Write([]byte(token))
}

func (ac *AuthController) handleLoginFailure(w http.ResponseWriter, username string, err error, ipAddress string) {
	logErr := ac.auditClient.LogEvent("LOGIN_FAILED", username, time.Now().String(), err.Error()+" from IP: "+ipAddress, ipAddress)
	if logErr != nil {
		log.Printf("Error logging login failure for user %s: %v", username, logErr)
	}
	http.Error(w, "Login failed: "+err.Error(), http.StatusUnauthorized)
}

func (ac *AuthController) Verify2FA(w http.ResponseWriter, r *http.Request) {
	var request data.VerificationRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	verificationResult, err := ac.userService.Verify2FA(request.Username, request.Verificationcode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !verificationResult {
		http.Error(w, "Invalid 2FA code", http.StatusUnauthorized)
		return
	}

	token, err := ac.tokenServiceClient.GenerateToken(request.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(token))
}

func (ac *AuthController) Enable2FA(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)          // Obtiene los parámetros de la ruta
	username := vars["username"] // Extrae el username

	qrCode, err := ac.userService.Enable2FAForUser(username)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error enabling 2FA: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(qrCode)
}

func (ac *AuthController) logAuditEvent(eventType, username, timestamp, message, ipAddress string) error {
	return ac.auditClient.LogEvent(eventType, username, timestamp, message, ipAddress)
}
