package main

import (
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"pubsubservice/grpc/pubsub"
)

type server struct {
	pubsub.UnimplementedPubSubServer
	subpub SubPub
}

func (s *server) Publish(ctx context.Context, req *pubsub.PublishRequest) (*pubsub.PublishResponse, error) {
	err := s.subpub.Publish(req.Subject, req.Message)
	if err != nil {
		return &pubsub.PublishResponse{Success: false}, err
	}
	return &pubsub.PublishResponse{Success: true}, nil
}

func (s *server) Subscribe(req *pubsub.SubscribeRequest, stream pubsub.PubSub_SubscribeServer) error {
	sub, err := s.subpub.Subscribe(req.Subject, func(msg interface{}) {
		if strMsg, ok := msg.(string); ok {
			stream.Send(&pubsub.SubscribeResponse{Message: strMsg})
		}
	})
	if err != nil {
		return err
	}

	<-stream.Context().Done()
	sub.Unsubscribe()
	return nil
}

func StartGRPCServer(subpub SubPub, wg *sync.WaitGroup) {
	defer wg.Done()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pubsub.RegisterPubSubServer(grpcServer, &server{subpub: subpub})
	reflection.Register(grpcServer)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
