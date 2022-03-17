package stream

import (
	"context"
	"errors"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/kafkakit"
	"github.com/golang/protobuf/proto"
	"github.com/justin0u0/protoc-gen-grpc-sarama/pkg/saramakit"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/emptypb"
)

type stream struct {
	pb.UnimplementedVideoStreamServer

	videoDAO dao.VideoDAO
	producer kafkakit.Producer
}

func NewStream(videoDAO dao.VideoDAO, producer kafkakit.Producer) *stream {
	return &stream{
		videoDAO: videoDAO,
		producer: producer,
	}
}

func (s *stream) HandleVideoCreated(ctx context.Context, req *pb.HandleVideoCreatedRequest) (*emptypb.Empty, error) {
	if req.Id == "" {
		return nil, &saramakit.HandlerError{Retry: false, Err: errors.New("video ID is required")}
	}

	if req.Scale != 0 {
		time.Sleep(time.Second * 3)
		if err := s.updateMongoVideo(ctx, req); err != nil {
			return &emptypb.Empty{}, err
		}
		return &emptypb.Empty{}, nil
	}

	variants := []int32{1080, 720, 480, 320}

	for _, scale := range variants {
		if err := s.updateVideoHandle(ctx, &pb.HandleVideoCreatedRequest{
			Id:    req.GetId(),
			Url:   req.GetUrl(),
			Scale: scale,
		}); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *stream) updateVideoHandle(ctx context.Context, req *pb.HandleVideoCreatedRequest) error {

	valueBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	msgs := []*kafkakit.ProducerMessage{
		{Value: valueBytes},
	}

	if err := s.producer.SendMessages(msgs); err != nil {
		return err
	}

	return nil
}

func (s *stream) updateMongoVideo(ctx context.Context, req *pb.HandleVideoCreatedRequest) error {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return err
	}

	if err := s.videoDAO.UpdateVariant(ctx, id, string(req.Scale), req.Url); err != nil {
		return err
	}
	return nil
}
