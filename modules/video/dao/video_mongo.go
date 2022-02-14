package dao

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type videoMongoDAO struct {
	collection *mongo.Collection
}

var _ VideoDAO = (*videoMongoDAO)(nil)

func NewVideoMongoDAO(collection *mongo.Collection) *videoMongoDAO {
	return &videoMongoDAO{
		collection: collection,
	}
}

func (dao *videoMongoDAO) Get(ctx context.Context, id primitive.ObjectID) (*Video, error) {
	var video Video
	if err := dao.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&video); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	return &video, nil
}

func (dao *videoMongoDAO) List(ctx context.Context, limit, skip int64) ([]*Video, error) {
	o := options.Find().SetLimit(limit).SetSkip(skip)

	cursor, err := dao.collection.Find(ctx, bson.M{}, o)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	videos := make([]*Video, 0)
	for cursor.Next(ctx) {
		var video Video
		if err := cursor.Decode(&video); err != nil {
			return nil, err
		}

		videos = append(videos, &video)
	}

	return videos, nil
}

func (dao *videoMongoDAO) Create(ctx context.Context, video *Video) error {
	result, err := dao.collection.InsertOne(ctx, video)
	if err != nil {
		return err
	}

	video.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (dao *videoMongoDAO) Update(ctx context.Context, video *Video) error {
	if result, err := dao.collection.ReplaceOne(
		ctx,
		bson.M{"_id": video.ID},
		video,
	); err != nil {
		return err
	} else if result.ModifiedCount == 0 {
		return ErrVideoNotFound
	}

	return nil
}

func (dao *videoMongoDAO) Delete(ctx context.Context, id primitive.ObjectID) error {
	if result, err := dao.collection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return err
	} else if result.DeletedCount == 0 {
		return ErrVideoNotFound
	}

	return nil
}
