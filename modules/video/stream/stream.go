package stream

import (
	"context"
	"strconv"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/kafkakit"
	"github.com/justin0u0/protoc-gen-grpc-sarama/pkg/saramakit"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
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
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, &saramakit.HandlerError{Retry: false, Err: err}
	}

	if req.GetScale() != 0 {
		variant := strconv.Itoa(int(req.GetScale()))

		// [Kafka TODO]
		// [Describe] Transcode video if get message with scale != 0, you can handle error occurance like above primitive.ObjectIDFromHex(req.GetId()).

		return &emptypb.Empty{}, nil
	}

	// [Kafka TODO]
	// [Describe] Fanout create events to each variant [1080, 720, 480, 320], you can handle error occurance like above primitive.ObjectIDFromHex(req.GetId()).

	return &emptypb.Empty{}, nil
}

func (s *stream) handleVideoWithVariant(ctx context.Context, id primitive.ObjectID, variant string, url string) error {
	// we mock the video transcoding only
	time.Sleep(3 * time.Second)

	if err := s.videoDAO.UpdateVariant(ctx, id, variant, url); err != nil {
		return err
	}

	return nil
}

func (s *stream) produceVideoCreatedWithScaleEvent(req *pb.HandleVideoCreatedRequest) error {
	valueBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	msgs := []*kafkakit.ProducerMessage{
		{Value: valueBytes},
	}

	// [Kafka TODO]
	// [Describe] Send message to kafka

	return nil
}
