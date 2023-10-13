package grpc

import (
	"context"
	"errors"
	"io"

	e "go-misc/internal/errors"

	"google.golang.org/grpc"
)

// ErrorUnaryServerInterceptor returns a new unary server interceptor for catching errors and only sending trusted errors.
func ErrorUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err == nil || err == io.EOF {
		return resp, nil
	}

	var ierr *e.Error
	if errors.As(err, &ierr) {
		return resp, ierr
	}

	var ierrs *e.Errors
	if errors.As(err, &ierrs) {
		return resp, ierrs
	}

	return resp, err
}
