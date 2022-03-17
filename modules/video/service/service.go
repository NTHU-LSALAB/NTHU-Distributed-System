package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"path"

	commentpb "github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/kafkakit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/storagekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
)

type service struct {
	pb.UnimplementedVideoServer

	videoDAO      dao.VideoDAO
	storage       storagekit.Storage
	commentClient commentpb.CommentClient
	producer      kafkakit.Producer
}

func NewService(videoDAO dao.VideoDAO, storage storagekit.Storage, commentClient commentpb.CommentClient, producer kafkakit.Producer) *service {
	return &service{
		videoDAO:      videoDAO,
		storage:       storage,
		commentClient: commentClient,
		producer:      producer,
	}
}

func (s *service) Healthz(ctx context.Context, req *pb.HealthzRequest) (*pb.HealthzResponse, error) {
	return &pb.HealthzResponse{Status: "ok"}, nil
}

func (s *service) GetVideo(ctx context.Context, req *pb.GetVideoRequest) (*pb.GetVideoResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, ErrInvalidObjectID
	}

	video, err := s.videoDAO.Get(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrVideoNotFound) {
			return nil, ErrVideoNotFound
		}

		return nil, err
	}

	return &pb.GetVideoResponse{Video: video.ToProto()}, nil
}

func (s *service) ListVideo(ctx context.Context, req *pb.ListVideoRequest) (*pb.ListVideoResponse, error) {
	videos, err := s.videoDAO.List(ctx, req.GetLimit(), req.GetSkip())
	if err != nil {
		return nil, err
	}

	pbVideos := make([]*pb.VideoInfo, 0, len(videos))
	for _, video := range videos {
		pbVideos = append(pbVideos, video.ToProto())
	}

	return &pb.ListVideoResponse{Videos: pbVideos}, nil
}

func (s *service) UploadVideo(stream pb.Video_UploadVideoServer) error {
	ctx := stream.Context()

	req, err := stream.Recv()
	if err != nil {
		return err
	}

	filename := req.GetHeader().GetFilename()
	size := req.GetHeader().GetSize()

	var buf bytes.Buffer

	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		chunk := req.GetChunkData()
		if _, err := buf.Write(chunk); err != nil {
			return err
		}
	}

	id := primitive.NewObjectID()
	objectName := id.Hex() + "-" + filename

	if err := s.storage.PutObject(ctx, objectName, bufio.NewReader(&buf), int64(size), storagekit.PutObjectOptions{
		ContentType: "application/octet-stream",
	}); err != nil {
		return err
	}

	video := &dao.Video{
		ID:     id,
		Size:   size,
		URL:    path.Join(s.storage.Endpoint(), s.storage.Bucket(), objectName),
		Status: dao.VideoStatusUploaded,
	}

	if err := s.videoDAO.Create(ctx, video); err != nil {
		return err
	}

	if err := s.produceVideoCreatedEvent(&pb.HandleVideoCreatedRequest{
		Id:  id.Hex(),
		Url: path.Join(s.storage.Endpoint(), s.storage.Bucket(), objectName),
	}); err != nil {
		return err
	}

	if err := stream.SendAndClose(&pb.UploadVideoResponse{
		Id: id.Hex(),
	}); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteVideo(ctx context.Context, req *pb.DeleteVideoRequest) (*pb.DeleteVideoResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, ErrInvalidObjectID
	}

	if err := s.videoDAO.Delete(ctx, id); err != nil {
		if errors.Is(err, dao.ErrVideoNotFound) {
			return nil, ErrVideoNotFound
		}

		return nil, err
	}

	if _, err := s.commentClient.DeleteCommentByVideoID(ctx, &commentpb.DeleteCommentByVideoIDRequest{
		VideoId: id.Hex(),
	}); err != nil {
		return nil, err
	}

	return &pb.DeleteVideoResponse{}, nil
}

func (s *service) produceVideoCreatedEvent(req *pb.HandleVideoCreatedRequest) error {
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
