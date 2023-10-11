package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type contextKey int

const (
	// generated uuid.
	ContextKeyRequestID contextKey = iota

	// log entry.
	ContextKeyLogEntry
)

func RequestIDUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	id := uuid.New()
	ctx = context.WithValue(ctx, ContextKeyRequestID, id.String())

	return handler(ctx, req)
}
