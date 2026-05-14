package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	port       string
}

func NewServer(port string, grpcServer *grpc.Server, httpServer *http.Server) *Server {
	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		port:       port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	// Create cmux
	m := cmux.New(lis)

	// Match gRPC
	grpcLis := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	// Match HTTP
	httpLis := m.Match(cmux.Any())

	go func() {
		if err := s.grpcServer.Serve(grpcLis); err != nil && err != cmux.ErrListenerClosed {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	go func() {
		if err := s.httpServer.Serve(httpLis); err != nil && err != http.ErrServerClosed && err != cmux.ErrListenerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	fmt.Printf("Multiplexed server starting on port %s\n", s.port)
	return m.Serve()
}
