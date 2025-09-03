package grpc

import (
	"context"
	"github.com/webook-project-go/webook-active/service"
	"github.com/webook-project-go/webook-apis/gen/go/apis/active/v1"
)

type Service struct {
	svc service.Service
	v1.UnimplementedActiveServiceServer
}

func New(svc service.Service) *Service {
	return &Service{
		svc: svc,
	}
}
func (s *Service) IsActive(ctx context.Context, request *v1.IsActiveRequest) (*v1.IsActiveResponse, error) {
	res, err := s.svc.IsActive(ctx, request.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.IsActiveResponse{Active: res}, nil
}

func (s *Service) GetLastActiveAt(ctx context.Context, request *v1.GetLastActiveAtRequest) (*v1.GetLastActiveAtResponse, error) {
	res, err := s.svc.GetLastActiveAt(ctx, request.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetLastActiveAtResponse{LastActiveAt: res}, nil
}

func (s *Service) GetActiveUsers(ctx context.Context, request *v1.GetActiveUsersRequest) (*v1.GetActiveUsersResponse, error) {
	res, err := s.svc.GetActiveUsers(ctx, request.GetSince(), request.GetLimit())
	if err != nil {
		return nil, err
	}
	return &v1.GetActiveUsersResponse{Uids: res}, nil
}
