package dao

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	"github.com/go-redis/cache/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ = Describe("VideoRedisDAO", func() {
	var conf = rediskit.RedisConfig{
		Addr: "redis:6379",
	}
	var videoRedisDAO *videoRedisDAO
	var videoMongoDAO *videoMongoDAO
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		videoMongoDAO = NewVideoMongoDAO(mongoClient.Database().Collection("videos"))
		videoRedisDAO = NewVideoRedisDAO(rediskit.NewRedisClient(ctx, &conf), videoMongoDAO)
	})

	Describe("Get", func() {
		var (
			video *Video
			resp  *Video
			id    primitive.ObjectID

			err error
		)

		BeforeEach(func() {
			video = NewFakeVideo()
			id = video.ID
		})

		JustBeforeEach(func() {
			resp, err = videoRedisDAO.Get(ctx, id)
		})

		Context("cache hit", func() {
			BeforeEach(func() {
				insertVideoInRedis(ctx, videoRedisDAO, video)
			})

			AfterEach(func() {
				deleteVideoInRedis(ctx, videoRedisDAO, getVideoKey(video.ID))
			})

			When("video not found", func() {
				BeforeEach(func() { id = primitive.NewObjectID() })

				It("returns video not found error", func() {
					Expect(resp).To(BeNil())
					Expect(err).To(MatchError(ErrVideoNotFound))
				})
			})

			When("success", func() {
				BeforeEach(func() { id = video.ID })

				It("returns the video with no error", func() {
					matchVideo(resp, video)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("cache miss", func() {
			BeforeEach(func() {
				insertVideo(ctx, videoMongoDAO, video)
			})

			AfterEach(func() {
				deleteVideo(ctx, videoMongoDAO, video.ID)
				deleteVideoInRedis(ctx, videoRedisDAO, getVideoKey(video.ID))
			})

			When("video not found", func() {
				BeforeEach(func() { id = primitive.NewObjectID() })

				It("returns video not found error", func() {
					Expect(resp).To(BeNil())
					Expect(err).To(MatchError(ErrVideoNotFound))
				})
			})

			When("success", func() {
				BeforeEach(func() { id = video.ID })

				It("returns the video with no error", func() {
					matchVideo(resp, video)
					Expect(err).NotTo(HaveOccurred())
				})

				It("insert the video to cache", func() {
					var getVideo Video
					Expect(videoRedisDAO.cache.Get(ctx, getVideoKey(id), &getVideo)).NotTo(HaveOccurred())
					matchVideo(&getVideo, video)
				})
			})
		})
	})

	Describe("List", func() {
		var (
			videos []*Video
			resp   []*Video
			limit  int64
			skip   int64

			err error
		)

		BeforeEach(func() {
			videos = []*Video{NewFakeVideo(), NewFakeVideo(), NewFakeVideo()}
		})

		JustBeforeEach(func() {
			resp, err = videoRedisDAO.List(ctx, limit, skip)
		})

		Context("cache hit", func() {
			BeforeEach(func() {
				limit, skip = 3, 0
				insertVideosInRedis(ctx, videoRedisDAO, videos, limit, skip)
			})

			AfterEach(func() {
				deleteVideosInRedis(ctx, videoRedisDAO, limit, skip)
			})

			When("videos not found", func() {
				BeforeEach(func() { limit = 1 })

				It("returns empty video with no error", func() {
					Expect(resp).To(HaveLen(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("success", func() {
				It("returns the videos with no error", func() {
					for i := range resp {
						matchVideo(resp[i], videos[i])
					}
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("cache miss", func() {
			BeforeEach(func() {
				for _, video := range videos {
					insertVideo(ctx, videoMongoDAO, video)
				}
			})

			AfterEach(func() {
				for _, video := range videos {
					deleteVideo(ctx, videoMongoDAO, video.ID)
				}
				deleteVideoInRedis(ctx, videoRedisDAO, listVideoKey(limit, skip))
			})

			When("success", func() {
				It("returns the videos with no error", func() {
					for i := range resp {
						matchVideo(resp[i], videos[i])
					}
					Expect(err).NotTo(HaveOccurred())
				})

				It("insert the videos to cache", func() {
					var getVideos []*Video
					Expect(videoRedisDAO.cache.Get(ctx, listVideoKey(limit, skip), &getVideos)).NotTo(HaveOccurred())
					for i := range getVideos {
						matchVideo(getVideos[i], videos[i])
					}
				})
			})
		})
	})

	Describe("Create", func() {
		var (
			video *Video

			err error
		)

		BeforeEach(func() {
			video = NewFakeVideo()
			video.ID = primitive.NilObjectID
		})

		AfterEach(func() {
			deleteVideo(ctx, videoMongoDAO, video.ID)
		})

		JustBeforeEach(func() {
			err = videoRedisDAO.baseDAO.Create(ctx, video)
		})

		When("success", func() {
			It("returns the new video ID with no error", func() {
				Expect(video.ID).NotTo(Equal(primitive.NilObjectID))
				Expect(err).NotTo(HaveOccurred())
			})

			It("inserts the document", func() {
				var getVideo Video

				Expect(
					videoMongoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo),
				).NotTo(HaveOccurred())

				Expect(&getVideo).To(Equal(video))
			})
		})
	})

	Describe("Update", func() {
		var (
			video *Video
			id    primitive.ObjectID

			err error
		)

		BeforeEach(func() {
			video = NewFakeVideo()
			id = video.ID

			insertVideo(ctx, videoMongoDAO, video)
		})

		AfterEach(func() {
			deleteVideo(ctx, videoMongoDAO, id)
		})

		JustBeforeEach(func() {
			err = videoMongoDAO.Update(ctx, video)
		})

		When("video not found", func() {
			BeforeEach(func() { video.ID = primitive.NewObjectID() })

			It("returns video not found error", func() {
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		When("success", func() {
			var size uint64

			BeforeEach(func() {
				size = 1234
				video.Size = size
			})

			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the document", func() {
				var getVideo Video

				Expect(videoMongoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo)).NotTo(HaveOccurred())

				Expect(getVideo.Size).To(Equal(size))
				Expect(&getVideo).To(Equal(video))
			})
		})
	})

	Describe("Delete", func() {
		var (
			video *Video
			id    primitive.ObjectID

			err error
		)

		BeforeEach(func() {
			video = NewFakeVideo()
			id = video.ID

			insertVideo(ctx, videoMongoDAO, video)
		})

		JustBeforeEach(func() {
			err = videoMongoDAO.Delete(ctx, video.ID)
		})

		When("video not found", func() {
			BeforeEach(func() {
				video.ID = primitive.NewObjectID()
			})

			AfterEach(func() {
				deleteVideo(ctx, videoMongoDAO, id)
			})

			It("returns video not found error", func() {
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		When("success", func() {
			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes the document", func() {
				var getVideo Video

				Expect(
					videoMongoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo),
				).To(Equal(mongo.ErrNoDocuments))
			})
		})
	})
})

func insertVideoInRedis(ctx context.Context, videoDAO *videoRedisDAO, video *Video) {
	Expect(videoDAO.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   getVideoKey(video.ID),
		Value: video,
		TTL:   videoDAORedisCacheDuration,
	})).NotTo(HaveOccurred())
}

func deleteVideoInRedis(ctx context.Context, videoDAO *videoRedisDAO, key string) {
	Expect(videoDAO.cache.Delete(ctx, key)).NotTo(HaveOccurred())
}

func insertVideosInRedis(ctx context.Context, videoDAO *videoRedisDAO, videos []*Video, limit int64, skip int64) {
	Expect(videoDAO.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   listVideoKey(limit, skip),
		Value: videos,
		TTL:   videoDAORedisCacheDuration,
	})).NotTo(HaveOccurred())
}

func deleteVideosInRedis(ctx context.Context, videoDAO *videoRedisDAO, limit int64, skip int64) {
	Expect(videoDAO.cache.Delete(ctx, listVideoKey(limit, skip))).NotTo(HaveOccurred())
}

func matchVideo(actual *Video, expect *Video) {
	Expect(*actual).To(MatchFields(IgnoreExtras, Fields{
		"ID":       Equal(expect.ID),
		"Width":    Equal(expect.Width),
		"Height":   Equal(expect.Height),
		"Size":     Equal(expect.Size),
		"Duration": Equal(expect.Duration),
		"URL":      Equal(expect.URL),
		"Status":   Equal(expect.Status),
		"Variants": Equal(expect.Variants),
	}))
}
