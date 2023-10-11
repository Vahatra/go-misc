package grpc

import (
	"context"
	"log/slog"
	"runtime/debug"

	e "go-misc/internal/errors"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func RecoverUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
	defer func() {
		if re := recover(); re != nil {
			stack := string(debug.Stack())
			LogEntryAttr(
				ctx,
				slog.String("stack", stack),
			)
			err = e.Newf(e.CodeInternal, "panic caught: %v", re)
		}
	}()

	return handler(ctx, req)
}

// // StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
// func RecoverStreamServerInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
// 	defer func() {
// 		if re := recover(); re != nil {
// 			stack := string(debug.Stack())
// 			LogEntryAttr(
// 				ctx,
// 				slog.String("panic", fmt.Sprintf("panic caught: %v (internal)", re)),
// 				slog.String("stack", stack),
// 			)
// 		}
// 	}()

// 	return handler(srv, stream)
// }
