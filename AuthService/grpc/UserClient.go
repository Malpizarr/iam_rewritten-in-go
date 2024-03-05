package grpc

import (
	user2 "AuthService/proto/user"
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"time"

	"google.golang.org/grpc"
)

type UserClient struct {
	conn    *grpc.ClientConn
	service user2.UserServiceClient
}

func NewUserClient(host string, port int) (*UserClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	service := user2.NewUserServiceClient(conn)

	return &UserClient{conn: conn, service: service}, nil
}

func (c *UserClient) CreateUser(username, email, password string) (*user2.UserProto, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.service.CreateUser(ctx, &user2.UserRequest{
		User: &user2.UserProto{
			Username: username,
			Email:    email,
			Password: password,
		},
	})
	if err != nil {
		return nil, err
	}

	return r.User, nil
}

func (c *UserClient) GetUser(username, password string) (*user2.UserProto, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.service.GetUser(ctx, &user2.UserRequest{
		User: &user2.UserProto{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return nil, err
	}

	return r.User, nil
}

func (c *UserClient) VerifyUser(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.service.VerifyUser(ctx, &user2.VerifyUserRequest{
		Id: id,
	})
	if err != nil {
		return "", err
	}

	return r.Message, nil
}

func (c *UserClient) Enable2FAForUser(username string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.service.Enable2FA(ctx, &user2.Enable2FARequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}

	return r.QrCode, nil
}

func (c *UserClient) Verify2FAForUser(username, code string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.service.Verify2FA(ctx, &user2.Verify2FARequest{
		Username:         username,
		VerificationCode: code,
	})
	if err != nil {
		return false, err
	}

	return r.Success, nil
}

func (c *UserClient) ProcessOAuthUser(ctx context.Context, email, sub, provider, name string) (*user2.UserProto, error) {
	r, err := c.service.ProcessOAuthUser(ctx, &user2.ProcessOAuthUserRequest{
		UserDetails: map[string]string{
			"email":    email,
			"name":     name,
			"provider": provider,
			"sub":      sub,
		},
	})
	if err != nil {
		return nil, err
	}

	return &user2.UserProto{
		Email:    r.OauthDTO.Email,
		Username: r.OauthDTO.Name,
	}, nil
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}
