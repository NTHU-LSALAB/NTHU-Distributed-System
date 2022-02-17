package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type videoRedisDAO struct {
	cache   *cache.Cache
	baseDAO VideoDAO
}

// var _ VideoDAO = (*videoRedisDAO)(nil)
var ErrCacheMiss = errors.New("cache miss")
var ErrSingleFlight = errors.New("singleflight error")

func NewVideoRedisDAO(client *rediskit.RedisClient, baseDAO VideoDAO) *videoRedisDAO {
	cache := cache.New(&cache.Options{
		Redis:      client.Client,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return &videoRedisDAO{
		cache:   cache,
		baseDAO: baseDAO,
	}
}

func (dao *videoRedisDAO) Get(ctx context.Context, id primitive.ObjectID) (*Video, error) {
	var video Video
	err := dao.cache.Get(ctx, id.Hex(), &video)
	if err == redis.Nil { // cache miss
		err = dao.cache.Once(&cache.Item{
			Key:   id.Hex(),
			Value: &video,
			Do: func(*cache.Item) (interface{}, error) {
				dbVideo, err := dao.baseDAO.Get(ctx, id)
				return dbVideo, err
			},
		})
		if err != nil {
			return nil, err
		} else {
			return &video, nil
		}
	} else if err != nil {
		return nil, err
	} else { // cache hit
		return &video, nil
	}
}

func (dao *videoRedisDAO) List(ctx context.Context, limit, skip int64) ([]*Video, error) {
	var video []*Video
	id := fmt.Sprintf("%d_%d", limit, skip)
	err := dao.cache.Get(ctx, id, &video)
	if err == redis.Nil { // cache miss
		err = dao.cache.Once(&cache.Item{
			Key:   id,
			Value: &video,
			Do: func(*cache.Item) (interface{}, error) {
				dbVideo, err := dao.baseDAO.List(ctx, limit, skip)
				return dbVideo, err
			},
		})
		if err != nil {
			return nil, err
		} else {
			return video, nil
		}
	} else if err != nil {
		return nil, err
	} else { // cache hit
		return video, nil
	}
}

func (dao *videoRedisDAO) Create(ctx context.Context, video *Video) error {
	return dao.baseDAO.Create(ctx, video)
}

func (dao *videoRedisDAO) Update(ctx context.Context, video *Video) error {
	return dao.baseDAO.Update(ctx, video)
}

func (dao *videoRedisDAO) Delete(ctx context.Context, id primitive.ObjectID) error {
	return dao.baseDAO.Delete(ctx, id)
}
