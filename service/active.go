package service

import (
	"context"
	"github.com/webook-project-go/webook-active/domain"
	"github.com/webook-project-go/webook-active/repository/redis"
)

type Service interface {
	MarkActive(ctx context.Context, users []domain.User) error
	IsActive(ctx context.Context, uid int64) (bool, error)
	GetLastActiveAt(ctx context.Context, uid int64) (int64, error)
	GetActiveUsers(ctx context.Context, since, limit int64) ([]int64, error)
}

type service struct {
	client redis.Client
}

func New(client redis.Client) Service {
	return &service{client: client}
}
func (s *service) MarkActive(ctx context.Context, users []domain.User) error {
	return s.client.SetActive(ctx, users)
}

func (s *service) IsActive(ctx context.Context, uid int64) (bool, error) {
	return s.client.JudgeActive(ctx, uid)
}

func (s *service) GetLastActiveAt(ctx context.Context, uid int64) (int64, error) {
	return s.client.GetLastActiveAt(ctx, uid)
}

func (s *service) GetActiveUsers(ctx context.Context, since, limit int64) ([]int64, error) {
	return s.client.GetActiveUsers(ctx, since, limit)
}
