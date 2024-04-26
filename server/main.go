package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	pb "task2/grpc"
)

type server struct {
	pb.UnimplementedUrlsServer
}

// mock function to returning suspicious url
func loadUrl() string {
	urls := []string{
		"https://go0gle-support.com",
		"http://paypal-login-info.com",
		"http://accounts_login.cz/google.com",
	}
	i := rand.Int() % len(urls)
	return urls[i]
}

func (s *server) GetSuspiciousUrl(ctx context.Context, in *pb.GetUrlRequest) (*pb.GetUrlResponse, error) {
	return &pb.GetUrlResponse{Url: loadUrl()}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Fatalf("failed to listen on port 50051: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUrlsServer(s, &server{})
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
