package dao

import (
	"context"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	"github.com/go-redis/cache/v8"
	"github.com/google/uuid"
)

type redisCommentDAO struct {
	cache   *cache.Cache
	baseDAO CommentDAO
}

var _ CommentDAO = (*redisCommentDAO)(nil)

const (
	commentDAOLocalCacheSize     = 1024
	commentDAOLocalCacheDuration = 1 * time.Minute
	commentDAORedisCacheDuration = 3 * time.Minute
)

func NewRedisCommentDAO(client *rediskit.RedisClient, baseDAO CommentDAO) *redisCommentDAO {
	return &redisCommentDAO{
		cache: cache.New(&cache.Options{
			Redis:      client,
			LocalCache: cache.NewTinyLFU(commentDAOLocalCacheSize, commentDAOLocalCacheDuration),
		}),
		baseDAO: baseDAO,
	}
}

func (dao *redisCommentDAO) ListByVideoID(ctx context.Context, videoID string, limit, offset int) ([]*Comment, error) {
	var comment []*Comment

	if err := dao.cache.Once(&cache.Item{
		Key:   listCommentKey(videoID, limit, offset),
		Value: &comment,
		TTL:   commentDAORedisCacheDuration,
		Do: func(*cache.Item) (interface{}, error) {
			return dao.baseDAO.ListByVideoID(ctx, videoID, limit, offset)
		},
	}); err != nil {
		return nil, err
	}
	return comment, nil
}

// The following operations are not cachable, just pass down to baseDAO

func (dao *redisCommentDAO) Create(ctx context.Context, comment *Comment) (uuid.UUID, error) {
	return dao.baseDAO.Create(ctx, comment)
}

func (dao *redisCommentDAO) Update(ctx context.Context, comment *Comment) error {
	return dao.baseDAO.Update(ctx, comment)
}

func (dao *redisCommentDAO) Delete(ctx context.Context, id uuid.UUID) error {
	return dao.baseDAO.Delete(ctx, id)
}

func (dao *redisCommentDAO) DeleteByVideoID(ctx context.Context, videoID string) error {
	return dao.baseDAO.DeleteByVideoID(ctx, videoID)
}
