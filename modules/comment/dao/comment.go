package dao

import (
	"context"
	"errors"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Comment struct {
	ID        uuid.UUID
	VideoID   string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Comment) ToProto() *pb.CommentInfo {
	return &pb.CommentInfo{
		Id:        c.ID.String(),
		VideoId:   c.VideoID,
		Content:   c.Content,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

type CommentDAO interface {
	ListByVideoID(ctx context.Context, videoID string, limit, offset int) ([]*Comment, error)
	Create(ctx context.Context, comment *Comment) (uuid.UUID, error)
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByVideoID(ctx context.Context, videoID string) error
}

var (
	ErrCommentNotFound = errors.New("comment not found")
)

// key format: "listComment:{videoID}:{limit}:{offset}"
func listCommentKey(videoID string, limit, offset int) string {
	// Redis TODO
}

func NewFakeComment(videoID string) *Comment {
	if videoID == "" {
		videoID = primitive.NewObjectID().Hex()
	}

	return &Comment{
		ID:      uuid.New(),
		VideoID: videoID,
		Content: "comment test",
	}
}
