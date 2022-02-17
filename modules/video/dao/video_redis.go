package dao

import (
	"context"
	"fmt"
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

func NewVideoRedisDAO(client *rediskit.RedisClient, baseDAO VideoDAO) *videoRedisDAO {
	size := 1000
	redisCache := cache.New(&cache.Options{
		Redis:      client.Client,
		LocalCache: cache.NewTinyLFU(size, time.Minute),
	})
	return &videoRedisDAO{
		cache:   redisCache,
		baseDAO: baseDAO,
	}
}

func (dao *videoRedisDAO) Get(ctx context.Context, id primitive.ObjectID) (*Video, error) {
	var video Video
	if err := dao.cache.Once(&cache.Item{
		Key:   id.Hex(),
		Value: &video,
		Do: func(*cache.Item) (interface{}, error) {
			dbVideo, dberr := dao.baseDAO.Get(ctx, id)
			return dbVideo, dberr
		},
	}); err != nil {
		return nil, err
	}

	return &video, nil
}

func (dao *videoRedisDAO) List(ctx context.Context, limit, skip int64) ([]*Video, error) {
	var video []*Video
	id := fmt.Sprintf("%d_%d", limit, skip)
	if err := dao.cache.Once(&cache.Item{
		Key:   id,
		Value: &video,
		Do: func(*cache.Item) (interface{}, error) {
			dbVideo, dberr := dao.baseDAO.List(ctx, limit, skip)
			return dbVideo, dberr
		},
	}); err != nil {
		return nil, err
	}

	return video, nil
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
