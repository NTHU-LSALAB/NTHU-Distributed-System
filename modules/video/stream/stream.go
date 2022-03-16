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

	producer kafkakit.Producer
	videoDAO dao.VideoDAO
}

func NewStream(videoDAO dao.VideoDAO, producer kafkakit.Producer) *stream {
	// pass producer and db
	return &stream{
		videoDAO: videoDAO,
		producer: producer,
	}
}

func (s *stream) HandleVideoCreated(ctx context.Context, req *pb.HandleVideoCreatedRequest) (*emptypb.Empty, error) {
	// FIXME: implement me
	if req.Scale == 0 {
		reqHighRes := &pb.HandleVideoCreatedRequest{
			Id:    req.Id,
			Url:   req.Url,
			Scale: 1080,
		}
		reqLowRes := &pb.HandleVideoCreatedRequest{
			Id:    req.Id,
			Url:   req.Url,
			Scale: 720,
		}
		if err := s.uploadVideoHandle(ctx, reqHighRes); err != nil {
			return &emptypb.Empty{}, err
		}
		if err := s.uploadVideoHandle(ctx, reqLowRes); err != nil {
			return &emptypb.Empty{}, err
		}
	} else {
		time.Sleep(time.Second * 3)
		if err := s.updateMongoVideo(ctx, req); err != nil {
			return &emptypb.Empty{}, err
		}
	}
	if req.Id == "" {
		return nil, &saramakit.HandlerError{Retry: false, Err: errors.New("video ID is required")}
	}

	return &emptypb.Empty{}, nil
}

func (s *stream) uploadVideoHandle(ctx context.Context, req *pb.HandleVideoCreatedRequest) error {

	valueBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	msg := make([]*kafkakit.ProducerMessage, 0, 1)

	msg = append(msg, &kafkakit.ProducerMessage{
		Value: valueBytes,
	})

	if err := s.producer.SendMessages(msg); err != nil {
		return err
	}

	return nil
}

func (s *stream) updateMongoVideo(ctx context.Context, req *pb.HandleVideoCreatedRequest) error {
	id, _ := primitive.ObjectIDFromHex(req.Id)
	// Here has an error, I'm not sure whether is that go compile just detect there is no "UpdateVariant" in videoDAO interface.
	// Will this error be remove after we pass mongoDAO here?
	if err := s.videoDAO.UpdateVariant(ctx, id, string(req.Scale), req.Url); err != nil {
		return err
	}
	return nil
}
