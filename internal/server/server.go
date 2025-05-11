package server

import (
	"VKIntern/grpc/pubsub"
	"VKIntern/subpub"
	"context"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PubSubServer struct {
	pubsub.UnimplementedPubSubServer
	sp     subpub.SubPub
	logger *log.Logger
}

func NewPubSubServer(sp subpub.SubPub, logger *log.Logger) *PubSubServer {
	return &PubSubServer{sp: sp, logger: logger}
}

func (s *PubSubServer) Subscribe(req *pubsub.SubscribeRequest, stream pubsub.PubSub_SubscribeServer) error {
	key := req.GetKey()
	s.logger.Printf("Subscribe: key=%s", key)

	ch := make(chan interface{}, 16)
	sub, err := s.sp.Subscribe(key, func(msg interface{}) {
		select {
		case ch <- msg:
		default:
			s.logger.Printf("dropping msg for key %s", key)
		}
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg := <-ch:
			data, ok := msg.(string)
			if !ok {
				s.logger.Printf("invalid message type: %T", msg)
				continue
			}
			if err := stream.Send(&pubsub.Event{Data: data}); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

func (s *PubSubServer) Publish(ctx context.Context, req *pubsub.PublishRequest) (*empty.Empty, error) {
	if err := s.sp.Publish(req.GetKey(), req.GetData()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}
