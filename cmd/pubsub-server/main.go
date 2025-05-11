package main

import (
	"VKIntern/grpc/pubsub"
	"VKIntern/grpc/server"
	"VKIntern/subpub"
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "50051", "gRPC server port")
	flag.Parse()

	logger := log.New(os.Stdout, "[pubsub] ", log.LstdFlags)

	sp := subpub.NewSubPub()

	srv := server.NewPubSubServer(sp, logger)
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pubsub.RegisterPubSubServer(grpcServer, srv)

	go func() {
		logger.Printf("serving gRPC on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("Serve error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Println("shutting down gRPC server...")

	go grpcServer.GracefulStop()

	select {
	case <-time.After(5 * time.Second):
		logger.Println("timeout exceeded, forcing Stop()")
		grpcServer.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sp.Close(ctx); err != nil {
		logger.Printf("subpub close error: %v", err)
	}
	logger.Println("shutdown complete")
}
