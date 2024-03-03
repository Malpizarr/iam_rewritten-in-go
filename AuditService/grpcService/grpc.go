package grpcService

import (
	model "AuditService/Model"
	repository "AuditService/Repository"
	audit "AuditService/proto"

	"context"
	"log"
	"time"
)

func NewAuditServiceImpl(repo repository.AuditEventRepository) *AuditServiceImpl {
	return &AuditServiceImpl{
		Repo: repo,
	}
}

type AuditServiceImpl struct {
	audit.UnimplementedAuditServiceServer
	Repo repository.AuditEventRepository
}

func (s *AuditServiceImpl) LogEvent(ctx context.Context, req *audit.AuditEvent) (*audit.LogResponse, error) {
	event := model.AuditEvent{
		EventType:     req.GetEventType(),
		Username:      req.GetUsername(),
		EventDateTime: time.Now(),
		Details:       req.GetDetails(),
		IpAddress:     req.GetIpAddress(),
	}

	if err := s.Repo.Save(&event); err != nil {
		log.Fatalf("Error logging event: %v", err)
		return nil, err
	}

	return &audit.LogResponse{Success: true}, nil
}
