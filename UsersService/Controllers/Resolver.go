package Controllers

import (
	"UsersService/Service"
	"UsersService/graph"
	"UsersService/model"
	"context"
	"errors"
	"strings"
)

type Resolver struct {
	userService  *Service.UserService
	TokenService *Service.TokenValidationService
}

func (r *Resolver) GORMRole() graph.GORMRoleResolver {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) GORMUser() graph.GORMUserResolver {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) Mutation() graph.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() graph.QueryResolver {
	return &queryResolver{r}
}

func NewResolver(userService *Service.UserService, TokenService *Service.TokenValidationService) *Resolver {
	return &Resolver{userService: userService, TokenService: TokenService}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Users(ctx context.Context) ([]*model.GORMUser, error) {
	return r.userService.GetAllUsers()
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.GORMUser, error) {
	return r.userService.GetUserById(id)
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(ctx context.Context, username string, email string, password string) (*model.GORMUser, error) {
	//TODO implement me
	panic("implement me")
}

func (r *mutationResolver) UpdateUser(ctx context.Context, id string, username *string, email *string, password *string) (*model.GORMUser, error) {

	authHeader, ok := ctx.Value("Authorization").(string)
	if !ok || authHeader == "" {
		return nil, errors.New("unauthorized")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("No token provided")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	if username == nil {
		return nil, errors.New("username cannot be nil")
	}

	if r.TokenService == nil {
		return nil, errors.New("TokenService is not initialized")
	}

	if !r.TokenService.ValidateToken(token, *username) {
		return nil, errors.New("unauthorized")
	}

	if r.userService == nil {
		return nil, errors.New("userService is not initialized")
	}

	user, err := r.userService.GetUserById(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if username != nil {
		user.Username = *username
	}
	if email != nil {
		user.Email = *email
	}
	if password != nil {
		user.Password = *password
	}

	updatedUser, err := r.userService.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (*bool, error) {
	userToDelete := &model.GORMUser{}
	userToDelete.ID = id
	err := r.userService.DeleteUser(userToDelete)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
