package hello

import (
	"context"
	"fmt"
	"go-misc/internal/cache"
	e "go-misc/internal/errors"
	"go-misc/internal/validator"

	"log/slog"
)

type HelloRepository interface {
	Get(address string) (Message, error)
}

type service struct {
	l *slog.Logger
	v *validator.Validation
	c *cache.Cache
	r HelloRepository
}

func NewService(l *slog.Logger, v *validator.Validation, c *cache.Cache, r HelloRepository) *service {
	return &service{l: l, v: v, c: c, r: r}
}

func (s *service) Say(ctx context.Context, id string) (string, error) {
	err := s.v.Struct(struct {
		Id string `validate:"required"`
	}{
		Id: id,
	})
	if err != nil {
		return "", fmt.Errorf("get message: %w", e.Wrap(e.CodeInvalidArgument, err))
	}

	msg, err := s.r.Get(id)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s, %s, %s", msg.English, msg.French, msg.Malagasy), nil
}
