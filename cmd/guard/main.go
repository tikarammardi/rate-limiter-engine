package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"rate-limiter-engine/internal/limiter"

	"syscall"

	"google.golang.org/grpc"
	"rate-limiter-engine/github.com/tikarammardi/rate-limiter-engine/proto"
)

type server struct {
	proto.UnimplementedRateLimiterServer
	guard *limiter.Guard
}

func (s *server) Check(ctx context.Context, req *proto.LimitRequest) (*proto.LimitResponse, error) {
	allowed := s.guard.Allow(ctx, req.UserId)

	return &proto.LimitResponse{
		Allowed: allowed,
	}, nil
}

func main() {
	store := limiter.NewMemoryStore(10, 50) // 10tokens/sec, capacity of 50 for brusts
	g := limiter.NewGuard(store)

	grpcServer := grpc.NewServer()
	proto.RegisterRateLimiterServer(grpcServer, &server{
		guard: g,
	})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server stopped serving: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down gracefully...")
	grpcServer.GracefulStop()
}
