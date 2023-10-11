package grpc

import (
	"context"
	"log/slog"

	"go-misc/internal/grpc/pb"
)

type helloService interface {
	Say(ctx context.Context, id string) (string, error)
}

type helloServer struct {
	pb.UnimplementedHelloServiceServer
	l *slog.Logger
	s helloService
}

func NewServer(l *slog.Logger, s helloService) *helloServer {
	return &helloServer{l: l, s: s}
}

func (s *helloServer) Say(ctx context.Context, req *pb.SayRequest) (*pb.SayResponse, error) {
	msg, err := s.s.Say(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	resp := &pb.SayResponse{
		Message: msg,
	}
	return resp, nil
}
