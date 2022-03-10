package stream

import (
	"context"
	"errors"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/justin0u0/protoc-gen-grpc-sarama/pkg/saramakit"
	"google.golang.org/protobuf/types/known/emptypb"
)

type stream struct {
	pb.UnimplementedVideoStreamServer
}

func NewStream() *stream {
	return &stream{}
}

func (s *stream) HandleVideoCreated(ctx context.Context, req *pb.HandleVideoCreatedRequest) (*emptypb.Empty, error) {
	// FIXME: implement me

	if req.Id == "" {
		return nil, &saramakit.HandlerError{Retry: false, Err: errors.New("video ID is required")}
	}

	return &emptypb.Empty{}, nil
}
