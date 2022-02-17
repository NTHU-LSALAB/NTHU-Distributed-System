package dao

import (
	"context"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	"github.com/go-redis/cache/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type videoRedisDAO struct {
	cache   *cache.Cache
	baseDAO VideoDAO
}

var _ VideoDAO = (*videoRedisDAO)(nil)

const (
	videoDAOLocalCacheSize     = 1024
	videoDAOLocalCacheDuration = 1 * time.Minute
	videoDAORedisCacheDuration = 3 * time.Minute
)

func NewVideoRedisDAO(client *rediskit.RedisClient, baseDAO VideoDAO) *videoRedisDAO {
	return &videoRedisDAO{
		cache: cache.New(&cache.Options{
			Redis:      client,
			LocalCache: cache.NewTinyLFU(videoDAOLocalCacheSize, videoDAOLocalCacheDuration),
		}),
		baseDAO: baseDAO,
	}
}

func (dao *videoRedisDAO) Get(ctx context.Context, id primitive.ObjectID) (*Video, error) {
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

func (dao *videoRedisDAO) List(ctx context.Context, limit, skip int64) ([]*Video, error) {
	var video []*Video

	if err := dao.cache.Once(&cache.Item{
		Key:   listVideoKey(limit, skip),
		Value: &video,
		TTL:   videoDAORedisCacheDuration,
		Do: func(*cache.Item) (interface{}, error) {
			return dao.baseDAO.List(ctx, limit, skip)
		},
	}); err != nil {
		return nil, err
	}

	return video, nil
}

// The following operations are not cachable, just pass down to baseDAO.

func (dao *videoRedisDAO) Create(ctx context.Context, video *Video) error {
	return dao.baseDAO.Create(ctx, video)
}

func (dao *videoRedisDAO) Update(ctx context.Context, video *Video) error {
	return dao.baseDAO.Update(ctx, video)
}

func (dao *videoRedisDAO) Delete(ctx context.Context, id primitive.ObjectID) error {
	return dao.baseDAO.Delete(ctx, id)
}
