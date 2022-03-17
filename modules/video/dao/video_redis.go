package dao

import (
	"context"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	"github.com/go-redis/cache/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type redisVideoDAO struct {
	cache   *cache.Cache
	baseDAO VideoDAO
}

var _ VideoDAO = (*redisVideoDAO)(nil)

const (
	videoDAOLocalCacheSize     = 1024
	videoDAOLocalCacheDuration = 1 * time.Minute
	videoDAORedisCacheDuration = 3 * time.Minute
)

func NewRedisVideoDAO(client *rediskit.RedisClient, baseDAO VideoDAO) *redisVideoDAO {
	return &redisVideoDAO{
		cache: cache.New(&cache.Options{
			Redis:      client,
			LocalCache: cache.NewTinyLFU(videoDAOLocalCacheSize, videoDAOLocalCacheDuration),
		}),
		baseDAO: baseDAO,
	}
}

func (dao *redisVideoDAO) Get(ctx context.Context, id primitive.ObjectID) (*Video, error) {
	var video Video

	if err := dao.cache.Once(&cache.Item{
		Key:   getVideoKey(id),
		Value: &video,
		TTL:   videoDAORedisCacheDuration,
		Do: func(*cache.Item) (interface{}, error) {
			return dao.baseDAO.Get(ctx, id)
		},
	}); err != nil {
		return nil, err
	}

	return &video, nil
}

func (dao *redisVideoDAO) List(ctx context.Context, limit, skip int64) ([]*Video, error) {
	var videos []*Video

	if err := dao.cache.Once(&cache.Item{
		Key:   listVideoKey(limit, skip),
		Value: &videos,
		TTL:   videoDAORedisCacheDuration,
		Do: func(*cache.Item) (interface{}, error) {
			return dao.baseDAO.List(ctx, limit, skip)
		},
	}); err != nil {
		return nil, err
	}

	return videos, nil
}

// The following operations are not cachable, just pass down to baseDAO.

func (dao *redisVideoDAO) Create(ctx context.Context, video *Video) error {
	return dao.baseDAO.Create(ctx, video)
}

func (dao *redisVideoDAO) Update(ctx context.Context, video *Video) error {
	return dao.baseDAO.Update(ctx, video)
}

func (dao *redisVideoDAO) UpdateVariant(ctx context.Context, id primitive.ObjectID, variant string, url string) error {
	return dao.baseDAO.UpdateVariant(ctx, id, variant, url)
}

func (dao *redisVideoDAO) Delete(ctx context.Context, id primitive.ObjectID) error {
	return dao.baseDAO.Delete(ctx, id)
}
