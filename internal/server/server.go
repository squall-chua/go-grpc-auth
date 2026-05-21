package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	bufLis     *bufconn.Listener
	port       string
}

func NewServer(port string, bufLis *bufconn.Listener, grpcServer *grpc.Server, httpServer *http.Server) *Server {
	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		bufLis:     bufLis,
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

	// Handle graceful shutdown in a separate goroutine
	go func() {
		<-ctx.Done()
		zap.L().Info("Gracefully shutting down servers...")

		// Shutdown HTTP server with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("HTTP server shutdown error", zap.Error(err))
		}

		// GracefulStop gRPC server (drains both the cmux listener and the bufconn listener)
		s.grpcServer.GracefulStop()

		// Close the bufconn listener and the main TCP listener to stop cmux
		s.bufLis.Close()
		lis.Close()
	}()

	go func() {
		if err := s.grpcServer.Serve(grpcLis); err != nil && err != cmux.ErrListenerClosed && err != net.ErrClosed && err != grpc.ErrServerStopped {
			zap.L().Error("gRPC server error", zap.Error(err))
		}
	}()

	// Serve gRPC over an in-memory bufconn listener so the grpc-gateway
	// can dial the gRPC server without a TCP loopback round-trip.
	go func() {
		if err := s.grpcServer.Serve(s.bufLis); err != nil && err != cmux.ErrListenerClosed && err != net.ErrClosed && err != grpc.ErrServerStopped {
			zap.L().Error("gRPC bufconn server error", zap.Error(err))
		}
	}()

	go func() {
		if err := s.httpServer.Serve(httpLis); err != nil && err != http.ErrServerClosed && err != cmux.ErrListenerClosed && err != net.ErrClosed {
			zap.L().Error("HTTP server error", zap.Error(err))
		}
	}()

	zap.L().Info("Multiplexed server starting", zap.String("port", s.port))
	err = m.Serve()
	if err != nil && (err == cmux.ErrListenerClosed || err == net.ErrClosed || strings.Contains(err.Error(), "use of closed network connection")) {
		return nil
	}
	return err
}
