package dao

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ = Describe("mongoVideoDAO", func() {
	var videoDAO *mongoVideoDAO
	var ctx context.Context

	BeforeEach(func() {
		videoDAO = NewMongoVideoDAO(mongoClient.Database().Collection("videos"))
		ctx = context.Background()
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

			insertVideo(ctx, videoDAO, video)
		})

		AfterEach(func() {
			deleteVideo(ctx, videoDAO, video.ID)
		})

		JustBeforeEach(func() {
			resp, err = videoDAO.Get(ctx, id)
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
				Expect(resp).To(Equal(video))
				Expect(err).NotTo(HaveOccurred())
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

			for _, video := range videos {
				insertVideo(ctx, videoDAO, video)
			}
		})

		AfterEach(func() {
			for _, video := range videos {
				deleteVideo(ctx, videoDAO, video.ID)
			}
		})

		JustBeforeEach(func() {
			resp, err = videoDAO.List(ctx, limit, skip)
		})

		Context("success", func() {
			When("no limit and offset", func() {
				It("returns videos with no error", func() {
					Expect(resp).To(Equal(videos))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1", func() {
				BeforeEach(func() { limit = 1 })

				It("returns the first video with no error", func() {
					Expect(resp).To(Equal(videos[:1]))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("limit = 1 and skip = 1", func() {
				BeforeEach(func() { limit, skip = 1, 1 })

				It("returns the second video with no error", func() {
					Expect(resp).To(Equal(videos[1:2]))
					Expect(err).NotTo(HaveOccurred())
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
			deleteVideo(ctx, videoDAO, video.ID)
		})

		JustBeforeEach(func() {
			err = videoDAO.Create(ctx, video)
		})

		When("success", func() {
			It("returns the new video ID with no error", func() {
				Expect(video.ID).NotTo(Equal(primitive.NilObjectID))
				Expect(err).NotTo(HaveOccurred())
			})

			It("inserts the document", func() {
				var getVideo Video

				Expect(
					videoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo),
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

			insertVideo(ctx, videoDAO, video)
		})

		AfterEach(func() {
			deleteVideo(ctx, videoDAO, id)
		})

		JustBeforeEach(func() {
			err = videoDAO.Update(ctx, video)
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

				Expect(
					videoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo),
				).NotTo(HaveOccurred())

				Expect(getVideo.Size).To(Equal(size))
				Expect(&getVideo).To(Equal(video))
			})
		})
	})

	Describe("UpdateVariant", func() {
		var (
			video   *Video
			id      primitive.ObjectID
			url     string
			variant string

			err error
		)

		BeforeEach(func() {
			video = NewFakeVideo()
			id = video.ID
			variant = "720p"
			url = video.URL

			insertVideo(ctx, videoDAO, video)
		})

		AfterEach(func() {
			deleteVideo(ctx, videoDAO, id)
		})

		JustBeforeEach(func() {
			err = videoDAO.UpdateVariant(ctx, video.ID, variant, url)
		})

		When("video not found", func() {
			BeforeEach(func() { video.ID = primitive.NewObjectID() })

			It("returns video not found error", func() {
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		When("success", func() {
			BeforeEach(func() {
				video.Variants[variant] = url
			})

			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the variant", func() {
				var getVideo Video

				Expect(
					videoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo),
				).NotTo(HaveOccurred())

				Expect(getVideo.Variants[variant]).To(Equal(url))
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

			insertVideo(ctx, videoDAO, video)
		})

		JustBeforeEach(func() {
			err = videoDAO.Delete(ctx, video.ID)
		})

		When("video not found", func() {
			BeforeEach(func() { video.ID = primitive.NewObjectID() })

			AfterEach(func() {
				deleteVideo(ctx, videoDAO, id)
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
					videoDAO.collection.FindOne(ctx, bson.M{"_id": video.ID}).Decode(&getVideo),
				).To(Equal(mongo.ErrNoDocuments))
			})
		})
	})
})

// useful methods for testing

func insertVideo(ctx context.Context, videoDAO *mongoVideoDAO, video *Video) {
	Expect(videoDAO.collection.InsertOne(ctx, video)).
		To(Equal(&mongo.InsertOneResult{InsertedID: video.ID}))
}

func deleteVideo(ctx context.Context, videoDAO *mongoVideoDAO, id primitive.ObjectID) {
	Expect(videoDAO.collection.DeleteOne(ctx, bson.M{"_id": id})).
		To(Equal(&mongo.DeleteResult{DeletedCount: 1}))
}
