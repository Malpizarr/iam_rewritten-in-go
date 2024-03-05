package grpcService

import (
	"UsersService/Service"
	"UsersService/model"
	pb "UsersService/proto"
	"context"
	"encoding/base32"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type UserServiceImpl struct {
	pb.UnimplementedUserServiceServer
	userService            *Service.UserService
	tokenValidationService *Service.TwoFactorAuthenticationService
}

func NewUserServiceImpl(userService *Service.UserService, tokenValidationService *Service.TwoFactorAuthenticationService) *UserServiceImpl {
	return &UserServiceImpl{
		userService:            userService,
		tokenValidationService: tokenValidationService,
	}
}

func ConvertProtoUserToDomainUser(protoUser *pb.UserProto) (*model.GORMUser, error) {
	totpSecret := protoUser.GetTotpSecret()
	var totpSecretPtr *string
	if totpSecret != "" {
		totpSecretPtr = &totpSecret
	}

	return &model.GORMUser{
		ID:              protoUser.GetId(),
		Username:        protoUser.GetUsername(),
		Email:           protoUser.GetEmail(),
		Password:        protoUser.GetPassword(),
		TotpSecret:      totpSecretPtr,
		IsTwoFaEnabled:  protoUser.GetIsTwoFaEnabled(),
		IsEmailVerified: protoUser.GetIsEmailVerified(),
	}, nil
}

func ConvertDomainUserToProtoUser(user *model.GORMUser) *pb.UserProto {
	var totpSecret string
	if user.TotpSecret != nil {
		totpSecret = *user.TotpSecret
	}

	return &pb.UserProto{
		Id:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		Password:        user.Password,
		TotpSecret:      totpSecret,
		IsTwoFaEnabled:  user.IsTwoFaEnabled,
		IsEmailVerified: user.IsEmailVerified,
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	user, err := ConvertProtoUserToDomainUser(req.User)
	if err != nil {
		log.Printf("Error al convertir ProtoUser a User: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Error al convertir el usuario: %v", err)
	}

	createdUser, err := s.userService.CreateUser(user)
	if err != nil {
		log.Printf("Error al crear el usuario: %v", err)
		return nil, status.Errorf(codes.Internal, "Error al crear el usuario: %v", err)
	}

	return &pb.UserResponse{
		User: ConvertDomainUserToProtoUser(createdUser),
	}, nil
}

func (s *UserServiceImpl) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	user, err := ConvertProtoUserToDomainUser(req.GetUser())
	if err != nil {
		log.Printf("Error al convertir ProtoUser a User: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Error al convertir el usuario: %v", err)
	}

	userResponse, err := s.userService.GetUser(user)

	response := &pb.UserResponse{
		User: ConvertDomainUserToProtoUser(userResponse),
	}

	return response, nil
}

func (s *UserServiceImpl) ProcessOAuthUser(ctx context.Context, req *pb.ProcessOAuthUserRequest) (*pb.ProcessOAuthUserResponse, error) {
	userDetails := req.GetUserDetails()
	email := userDetails["email"]
	name := userDetails["name"]
	provider := userDetails["provider"]
	sub := userDetails["sub"]

	user, err := s.userService.FindOrCreateUser(email, name, provider, sub)
	if err != nil {
		log.Printf("Error finding or creating user: %v", err)
		return nil, status.Errorf(codes.Internal, "Error finding or creating user: %v", err)
	}

	oauthDTO := s.userService.ConvertToDto(user)

	providerFound := false
	for _, oauthProvider := range user.OAuthProvider {
		if oauthProvider.ProviderName == provider {
			oauthDTO.Email = user.Email
			oauthDTO.Name = user.Username
			oauthDTO.Provider = oauthProvider.ProviderName
			oauthDTO.Sub = oauthProvider.ProviderID
			providerFound = true
			break
		}
	}

	if !providerFound {
		log.Printf("OAuth provider information not found for user: %s", email)
		return nil, status.Errorf(codes.Internal, "OAuth provider information not found")
	}

	grpcUserDto := &pb.OauthDTO{
		Email:    oauthDTO.Email,
		Name:     oauthDTO.Name,
		Provider: oauthDTO.Provider,
		Sub:      oauthDTO.Sub,
	}

	response := &pb.ProcessOAuthUserResponse{
		OauthDTO: grpcUserDto,
	}

	return response, nil
}

func (s *UserServiceImpl) Verify2FA(ctx context.Context, req *pb.Verify2FARequest) (*pb.Verify2FAResponse, error) {
	username := req.GetUsername()
	verificationCode := req.GetVerificationCode()

	if len(verificationCode) != 6 {
		return nil, status.Errorf(codes.InvalidArgument, "Verification code must be a 6-digit number")
	}

	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	_, err = base32.StdEncoding.DecodeString(*user.TotpSecret)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid secret key")
	}

	isCodeValid, err := s.tokenValidationService.VerifyCode(verificationCode, *user.TotpSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to verify code: %v", err)
	}

	if !isCodeValid {
		return &pb.Verify2FAResponse{
			Success: false,
		}, nil
	}

	return &pb.Verify2FAResponse{
		Success: true,
	}, nil
}

func (s *UserServiceImpl) Enable2FA(ctx context.Context, req *pb.Enable2FARequest) (*pb.Enable2FAResponse, error) {
	username := req.GetUsername()

	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	if user.IsTwoFaEnabled {
		return nil, status.Errorf(codes.AlreadyExists, "2FA is already enabled")
	}

	secretKey := s.tokenValidationService.GenerateSecretKey(user.Email)
	user.TotpSecret = &secretKey
	user.IsTwoFaEnabled = true

	otpAuthUrl := s.tokenValidationService.GenerateTotpUrl(secretKey, "IAM", user.Email)

	qrCode, err := s.tokenValidationService.GenerateQrCode(otpAuthUrl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	_, err = s.userService.UpdateUser(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.Enable2FAResponse{
		QrCode: qrCode,
	}, nil
}
