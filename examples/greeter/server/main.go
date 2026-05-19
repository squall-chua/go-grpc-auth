package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/api/v1/greeter"
	interceptorclient "github.com/squall-chua/go-grpc-auth/pkg/interceptor/client"
	interceptorserver "github.com/squall-chua/go-grpc-auth/pkg/interceptor/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// staticTokenProvider holds a fixed token. For production, implement
// client.TokenProvider with refresh logic since access tokens expire.
type staticTokenProvider struct {
	token string
}

func (p *staticTokenProvider) GetToken(ctx context.Context) (string, error) {
	return p.token, nil
}

type greeterServer struct {
	greeter.UnimplementedGreeterServiceServer
}

func (s *greeterServer) SayHello(ctx context.Context, req *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}

func (s *greeterServer) SayHelloAuthenticated(ctx context.Context, req *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	p := interceptorserver.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Internal, "principal missing from context")
	}
	return &greeter.HelloResponse{
		Message: fmt.Sprintf("Hello, %s! (user: %s)", req.Name, p.UserId),
	}, nil
}

func (s *greeterServer) SayHelloAdmin(ctx context.Context, req *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	p := interceptorserver.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Internal, "principal missing from context")
	}
	return &greeter.HelloResponse{
		Message: fmt.Sprintf("Hello, %s! (admin user: %s)", req.Name, p.UserId),
	}, nil
}

func (s *greeterServer) SayHelloEditor(ctx context.Context, req *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	p := interceptorserver.GetPrincipal(ctx)
	if p == nil {
		return nil, status.Error(codes.Internal, "principal missing from context")
	}
	return &greeter.HelloResponse{
		Message: fmt.Sprintf("Hello, %s! (editor user: %s)", req.Name, p.UserId),
	}, nil
}

func main() {
	addr := flag.String("addr", ":9090", "greeter server listen address")
	authURI := flag.String("auth", "", "auth server URI (grpc://user:pass@host:port)")
	flag.Parse()

	if *authURI == "" {
		log.Fatal("-auth flag is required")
	}

	// Parse auth URI
	u, err := url.Parse(*authURI)
	if err != nil {
		log.Fatalf("invalid auth URI: %v", err)
	}
	authHost := u.Host
	username := u.User.Username()
	password, _ := u.User.Password()

	// Login to auth server to get a service token
	bootstrapConn, err := grpc.NewClient(authHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth server: %v", err)
	}
	defer bootstrapConn.Close()

	authClient := auth.NewAuthServiceClient(bootstrapConn)
	tokenPair, err := authClient.Login(context.Background(), &auth.LoginRequest{
		Login:    username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("failed to login to auth server: %v", err)
	}
	log.Printf("authenticated with auth server as %s", username)

	// Create authenticated connection to auth server for token validation
	provider := &staticTokenProvider{token: tokenPair.AccessToken}
	authenticatedConn, err := grpc.NewClient(authHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(interceptorclient.UnaryAuthInterceptor(provider)),
	)
	if err != nil {
		log.Fatalf("failed to create authenticated connection: %v", err)
	}
	defer authenticatedConn.Close()

	// Set up auth interceptor using authenticated auth client
	authInterceptor := interceptorserver.NewAuthInterceptor(auth.NewAuthServiceClient(authenticatedConn))

	// Create and start gRPC server
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(authInterceptor.Unary()))
	greeter.RegisterGreeterServiceServer(srv, &greeterServer{})

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("shutting down...")
		srv.GracefulStop()
	}()

	log.Printf("greeter server listening on %s", *addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
