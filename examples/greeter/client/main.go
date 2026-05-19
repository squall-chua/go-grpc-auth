package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/squall-chua/go-grpc-auth/api/v1/auth"
	"github.com/squall-chua/go-grpc-auth/api/v1/greeter"
	interceptorclient "github.com/squall-chua/go-grpc-auth/pkg/interceptor/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// staticTokenProvider holds a fixed token. For production, implement
// client.TokenProvider with refresh logic since access tokens expire.
type staticTokenProvider struct {
	token string
}

func (p *staticTokenProvider) GetToken(ctx context.Context) (string, error) {
	return p.token, nil
}

func main() {
	authURI := flag.String("auth", "", "auth server URI (grpc://user:pass@host:port)")
	greeterURI := flag.String("greeter", "", "greeter server URI (grpc://host:port)")
	flag.Parse()

	if *authURI == "" || *greeterURI == "" {
		log.Fatal("-auth and -greeter flags are required")
	}

	// Parse auth URI
	authURL, err := url.Parse(*authURI)
	if err != nil {
		log.Fatalf("invalid auth URI: %v", err)
	}
	authHost := authURL.Host
	username := authURL.User.Username()
	password, _ := authURL.User.Password()

	// Login to auth server
	authConn, err := grpc.NewClient(authHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth server: %v", err)
	}
	defer authConn.Close()

	authClient := auth.NewAuthServiceClient(authConn)
	tokenPair, err := authClient.Login(context.Background(), &auth.LoginRequest{
		Login:    username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("failed to login: %v", err)
	}
	fmt.Printf("Logged in as %s\n\n", username)

	// Parse greeter URI
	greeterURL, err := url.Parse(*greeterURI)
	if err != nil {
		log.Fatalf("invalid greeter URI: %v", err)
	}

	// Connect to greeter server with auth interceptor
	provider := &staticTokenProvider{token: tokenPair.AccessToken}
	greeterConn, err := grpc.NewClient(greeterURL.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(interceptorclient.UnaryAuthInterceptor(provider)),
	)
	if err != nil {
		log.Fatalf("failed to connect to greeter server: %v", err)
	}
	defer greeterConn.Close()

	client := greeter.NewGreeterServiceClient(greeterConn)
	ctx := context.Background()
	req := &greeter.HelloRequest{Name: "World"}

	// Call all 4 RPCs
	resp, err := client.SayHello(ctx, req)
	if err != nil {
		fmt.Printf("SayHello: %v\n", err)
	} else {
		fmt.Printf("SayHello: %s\n", resp.Message)
	}

	resp, err = client.SayHelloAuthenticated(ctx, req)
	if err != nil {
		fmt.Printf("SayHelloAuthenticated: %v\n", err)
	} else {
		fmt.Printf("SayHelloAuthenticated: %s\n", resp.Message)
	}

	resp, err = client.SayHelloAdmin(ctx, req)
	if err != nil {
		fmt.Printf("SayHelloAdmin: %v\n", err)
	} else {
		fmt.Printf("SayHelloAdmin: %s\n", resp.Message)
	}

	resp, err = client.SayHelloEditor(ctx, req)
	if err != nil {
		fmt.Printf("SayHelloEditor: %v\n", err)
	} else {
		fmt.Printf("SayHelloEditor: %s\n", resp.Message)
	}
}
