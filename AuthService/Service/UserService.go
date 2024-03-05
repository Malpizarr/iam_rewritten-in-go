package service

import (
	"AuthService/data"
	"AuthService/email"
	"AuthService/grpc"
	"AuthService/proto/user"
	"AuthService/repository"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	PasswordEncoder             func(password string) (string, error)
	EmailService                email.EmailService
	UserClientService           grpc.UserClient
	VerificationTokenRepository repository.VerificationTokenRepository
	UserClient                  grpc.UserClient
}

func NewUserService(emailService email.EmailService, userClientService grpc.UserClient, verificationTokenRepository repository.VerificationTokenRepository, userClient grpc.UserClient) *UserService {
	return &UserService{
		PasswordEncoder: func(password string) (string, error) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return "", err
			}
			return string(hash), nil
		},
		EmailService:                emailService,
		UserClientService:           userClientService,
		VerificationTokenRepository: verificationTokenRepository,
		UserClient:                  userClient,
	}
}

func (s *UserService) Register(newUser *user.UserProto) (*user.UserProto, error) {
	user, err := s.UserClient.CreateUser(newUser.Username, newUser.Email, newUser.Password)
	if err != nil {
		return nil, err
	}

	token := generateToken()
	verificationToken := &data.VerificationToken{
		Token:  token,
		UserID: user.Id,
	}
	err = s.VerificationTokenRepository.Save(verificationToken)
	if err != nil {
		return nil, err
	}

	verificationUrl := "http://localhost:8080/verification/verify?token=" + token
	err = s.EmailService.SendVerificationEmail(newUser.Email, verificationUrl)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(user *user.UserProto) (*user.UserProto, error) {
	userResponse, err := s.UserClient.GetUser(user.Username, user.Password)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(userResponse.Password), []byte(user.Password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return userResponse, nil
}

func (s *UserService) Enable2FAForUser(username string) ([]byte, error) {
	return s.UserClient.Enable2FAForUser(username)
}

func (s *UserService) Verify2FA(username, verificationCode string) (bool, error) {
	return s.UserClient.Verify2FAForUser(username, verificationCode)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
