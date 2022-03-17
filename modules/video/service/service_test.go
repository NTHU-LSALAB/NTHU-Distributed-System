package service

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	commentpbmock "github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/mock/pbmock"
	commentpb "github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/mock/daomock"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/mock/pbmock"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/kafkakit/mock/kafkamock"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/storagekit/mock/storagemock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Service")
}

var (
	errDAOUnknown = errors.New("unknown DAO error")
)

var _ = Describe("Service", func() {
	var (
		controller    *gomock.Controller
		videoDAO      *daomock.MockVideoDAO
		storage       *storagemock.MockStorage
		commentClient *commentpbmock.MockCommentClient
		producer      *kafkamock.MockProducer
		svc           *service
		ctx           context.Context
	)

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		videoDAO = daomock.NewMockVideoDAO(controller)
		storage = storagemock.NewMockStorage(controller)
		commentClient = commentpbmock.NewMockCommentClient(controller)
		producer = kafkamock.NewMockProducer(controller)
		svc = NewService(videoDAO, storage, commentClient, producer)
		ctx = context.Background()
	})

	AfterEach(func() {
		controller.Finish()
	})

	Describe("GetVideo", func() {
		var (
			req  *pb.GetVideoRequest
			id   primitive.ObjectID
			resp *pb.GetVideoResponse
			err  error
		)

		BeforeEach(func() {
			id = primitive.NewObjectID()
			req = &pb.GetVideoRequest{Id: id.Hex()}
		})

		JustBeforeEach(func() {
			resp, err = svc.GetVideo(ctx, req)
		})

		When("DAO error", func() {
			BeforeEach(func() {
				videoDAO.EXPECT().Get(ctx, id).Return(nil, errDAOUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errDAOUnknown))
			})
		})

		When("video not found", func() {
			BeforeEach(func() {
				videoDAO.EXPECT().Get(ctx, id).Return(nil, dao.ErrVideoNotFound)
			})

			It("returns video not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		When("success", func() {
			var video *dao.Video

			BeforeEach(func() {
				video = dao.NewFakeVideo()
				videoDAO.EXPECT().Get(ctx, id).Return(video, nil)
			})

			It("returns the video with no error", func() {
				Expect(resp).To(Equal(&pb.GetVideoResponse{
					Video: video.ToProto(),
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("ListVideo", func() {
		var (
			req  *pb.ListVideoRequest
			resp *pb.ListVideoResponse
			err  error
		)

		BeforeEach(func() {
			req = &pb.ListVideoRequest{Limit: 10, Skip: 0}
		})

		JustBeforeEach(func() {
			resp, err = svc.ListVideo(ctx, req)
		})

		When("DAO error", func() {
			BeforeEach(func() {
				videoDAO.EXPECT().List(ctx, req.GetLimit(), req.GetSkip()).Return(nil, errDAOUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errDAOUnknown))
			})
		})

		When("success", func() {
			var videos []*dao.Video

			BeforeEach(func() {
				videos = []*dao.Video{dao.NewFakeVideo(), dao.NewFakeVideo()}
				videoDAO.EXPECT().List(ctx, req.GetLimit(), req.GetSkip()).Return(videos, nil)
			})

			It("returns videos with no error", func() {
				Expect(resp).To(Equal(&pb.ListVideoResponse{
					Videos: []*pb.VideoInfo{
						videos[0].ToProto(),
						videos[1].ToProto(),
					},
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("UploadVideo", func() {
		var stream *pbmock.MockVideo_UploadVideoServer
		var err error

		BeforeEach(func() {
			stream = pbmock.NewMockVideo_UploadVideoServer(controller)
			stream.EXPECT().Context().Return(ctx)
		})

		JustBeforeEach(func() {
			err = svc.UploadVideo(stream)
		})

		When("success", func() {
			BeforeEach(func() {
				file, rerr := os.ReadFile("./fixtures/big_buck_bunny_240p_1mb.mp4")
				Expect(rerr).NotTo(HaveOccurred())

				requests := []*pb.UploadVideoRequest{
					{
						Data: &pb.UploadVideoRequest_Header{
							Header: &pb.VideoHeader{
								Filename: "big_buck_bunny_240p_1mb.mp4",
								Size:     1053651,
							},
						},
					},
					{
						Data: &pb.UploadVideoRequest_ChunkData{
							ChunkData: file,
						},
					},
				}

				for _, req := range requests {
					stream.EXPECT().Recv().Return(req, nil)
				}
				stream.EXPECT().Recv().Return(nil, io.EOF)

				storage.EXPECT().PutObject(
					ctx,
					gomock.Any(),
					gomock.Any(),
					int64(requests[0].GetHeader().GetSize()),
					gomock.Any(),
				).Return(nil)

				storage.EXPECT().Endpoint().AnyTimes().Return("https://play.min.io")
				storage.EXPECT().Bucket().AnyTimes().Return("videos")

				videoDAO.EXPECT().Create(ctx, gomock.Any()).Return(nil)

				producer.EXPECT().SendMessages(gomock.Any()).Return(nil)

				stream.EXPECT().SendAndClose(gomock.Any()).Return(nil)
			})

			It("returns no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeleteVideo", func() {
		var (
			req  *pb.DeleteVideoRequest
			id   primitive.ObjectID
			resp *pb.DeleteVideoResponse
			err  error
		)

		BeforeEach(func() {
			id = primitive.NewObjectID()
			req = &pb.DeleteVideoRequest{Id: id.Hex()}
		})

		JustBeforeEach(func() {
			resp, err = svc.DeleteVideo(ctx, req)
		})

		When("DAO error", func() {
			BeforeEach(func() {
				videoDAO.EXPECT().Delete(ctx, id).Return(errDAOUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errDAOUnknown))
			})
		})

		When("video not found", func() {
			BeforeEach(func() {
				videoDAO.EXPECT().Delete(ctx, id).Return(dao.ErrVideoNotFound)
			})

			It("returns video not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrVideoNotFound))
			})
		})

		When("success", func() {
			BeforeEach(func() {
				videoDAO.EXPECT().Delete(ctx, id).Return(nil)
				commentClient.EXPECT().DeleteCommentByVideoID(ctx, &commentpb.DeleteCommentByVideoIDRequest{
					VideoId: id.Hex(),
				})
			})

			It("returns no error", func() {
				Expect(resp).To(Equal(&pb.DeleteVideoResponse{}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
