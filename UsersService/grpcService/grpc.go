package grpcService

import (
	"UsersService/Service"
	"UsersService/model"
	pb "UsersService/proto"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type UserServiceImpl struct {
	pb.UnimplementedUserServiceServer
	userService *Service.UserService
}

func NewUserServiceImpl(userService *Service.UserService) *UserServiceImpl {
	return &UserServiceImpl{
		userService: userService,
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
