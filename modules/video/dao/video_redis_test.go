package dao

import (
	"context"

	"github.com/go-redis/cache/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ = Describe("VideoRedisDAO", func() {
	var videoRedisDAO *videoRedisDAO
	var videoMongoDAO *videoMongoDAO
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		videoMongoDAO = NewVideoMongoDAO(mongoClient.Database().Collection("videos"))
		videoRedisDAO = NewVideoRedisDAO(redisClient, videoMongoDAO)
	})

	Describe("Get", func() {
		var (
			video *Video
			id    primitive.ObjectID

			resp *Video
			err  error
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
					Expect(resp).To(matchVideo(video))
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
					Expect(resp).To(matchVideo(video))
					Expect(err).NotTo(HaveOccurred())
				})

				It("insert the video to cache", func() {
					var getVideo Video

					Expect(
						videoRedisDAO.cache.Get(ctx, getVideoKey(id), &getVideo),
					).NotTo(HaveOccurred())
					Expect(resp).To(matchVideo(video))
				})
			})
		})
	})

	Describe("List", func() {
		var (
			videos []*Video
			limit  int64
			skip   int64

			resp []*Video
			err  error
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

				It("returns empty list with no error", func() {
					Expect(resp).To(HaveLen(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("success", func() {
				It("returns the videos with no error", func() {
					for i := range resp {
						Expect(resp[i]).To(matchVideo(videos[i]))
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
						Expect(resp[i]).To(matchVideo(videos[i]))
					}
					Expect(err).NotTo(HaveOccurred())
				})

				It("insert the videos to cache", func() {
					var getVideos []*Video
					Expect(videoRedisDAO.cache.Get(ctx, listVideoKey(limit, skip), &getVideos)).NotTo(HaveOccurred())
					for i := range getVideos {
						Expect(getVideos[i]).To(matchVideo(videos[i]))
					}
				})
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

func matchVideo(video *Video) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"ID":       Equal(video.ID),
		"Width":    Equal(video.Width),
		"Height":   Equal(video.Height),
		"Size":     Equal(video.Size),
		"Duration": Equal(video.Duration),
		"URL":      Equal(video.URL),
		"Status":   Equal(video.Status),
		"Variants": Equal(video.Variants),
	}))
}
