package grpc

import (
	Audit2 "AuthService/proto/audit"
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"time"

	"google.golang.org/grpc"
)

type AuditClient struct {
	conn    *grpc.ClientConn
	service Audit2.AuditServiceClient
}

func NewAuditClient(host string, port int) (*AuditClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	service := Audit2.NewAuditServiceClient(conn)

	return &AuditClient{conn: conn, service: service}, nil
}

func (c *AuditClient) LogEvent(eventType, username, eventDateTime, details, ipAddress string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.service.LogEvent(ctx, &Audit2.AuditEvent{
		EventType:     eventType,
		Username:      username,
		EventDateTime: eventDateTime,
		Details:       details,
		IpAddress:     ipAddress,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *AuditClient) Close() {
	err := c.conn.Close()
	if err != nil {
		return
	}
}
