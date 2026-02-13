package main

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"rate-limiter-engine/github.com/tikarammardi/rate-limiter-engine/proto"
)

func main() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	client := proto.NewRateLimiterClient(conn)

	var wg sync.WaitGroup
	numRequests := 100
	userID := "stress_user"

	log.Printf("Starting stress test with %d concurrent requests...", numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			resp, err := client.Check(ctx, &proto.LimitRequest{UserId: userID})
			if err != nil {
				log.Printf("Worker %d failed: %v", id, err)
				return
			}

			if resp.Allowed {
				log.Printf("Worker %d: ALLOWED", id)
			} else {
				log.Printf("Worker %d: BLOCKED", id)
			}
		}(i)
	}

	wg.Wait()
	log.Println("Stress test complete.")
}
