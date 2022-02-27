package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/mock/daomock"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Service")
}

var (
	errPGUnknown = errors.New("unknown postgres error")
)

var _ = Describe("Service", func() {
	var (
		controller *gomock.Controller
		commentDAO *daomock.MockCommentDAO
		svc        *service
		ctx        context.Context
	)

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		commentDAO = daomock.NewMockCommentDAO(controller)
		svc = NewService(commentDAO)
		ctx = context.Background()
	})

	AfterEach(func() {
		controller.Finish()
	})

	Describe("ListComment", func() {
		var (
			req     *pb.ListCommentRequest
			videoID string
			limit   int32
			skip    int32
			resp    *pb.ListCommentResponse
			err     error
		)

		BeforeEach(func() {
			videoID = "fake id"
			limit = 10
			skip = 0
			req = &pb.ListCommentRequest{VideoId: videoID, Limit: limit, Skip: skip}
		})

		JustBeforeEach(func() {
			resp, err = svc.ListComment(ctx, req)
		})

		When("postgres error", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().ListByVideoID(ctx, req.GetVideoId(), int(req.GetLimit()), int(req.GetSkip())).Return(nil, errPGUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errPGUnknown))
			})
		})

		When("success", func() {
			var comments []*dao.Comment

			BeforeEach(func() {
				comments = []*dao.Comment{dao.NewFakeComment(""), dao.NewFakeComment("")}
				commentDAO.EXPECT().ListByVideoID(ctx, req.GetVideoId(), int(req.GetLimit()), int(req.GetSkip())).Return(comments, nil)
			})

			It("returns comments with no error", func() {
				Expect(resp).To(Equal(&pb.ListCommentResponse{
					Comments: []*pb.CommentInfo{
						comments[0].ToProto(),
						comments[1].ToProto(),
					},
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CreateComment", func() {
		var (
			req     *pb.CreateCommentRequest
			videoID string
			content string
			resp    *pb.CreateCommentResponse
			err     error
		)

		BeforeEach(func() {
			videoID = "fake id"
			content = "fake content"
			req = &pb.CreateCommentRequest{
				VideoId: videoID,
				Content: content,
			}
		})

		JustBeforeEach(func() {
			resp, err = svc.CreateComment(ctx, req)
		})

		When("postgres error", func() {
			var comment *dao.Comment

			BeforeEach(func() {
				comment = &dao.Comment{
					VideoID: videoID,
					Content: content,
				}

				commentDAO.EXPECT().Create(ctx, comment).Return(uuid.Nil, errPGUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errPGUnknown))
			})
		})

		When("success", func() {
			var comment *dao.Comment
			var id uuid.UUID

			BeforeEach(func() {
				comment = &dao.Comment{
					VideoID: videoID,
					Content: content,
				}
				id = uuid.New()
				commentDAO.EXPECT().Create(ctx, comment).Return(id, nil)
			})

			It("returns no error", func() {
				Expect(resp).To(Equal(&pb.CreateCommentResponse{
					Id: id.String(),
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("UpdateComment", func() {
		var (
			req     *pb.UpdateCommentRequest
			id      uuid.UUID
			content string
			resp    *pb.UpdateCommentResponse
			err     error
		)

		BeforeEach(func() {
			id = uuid.New()
			content = "fake content"
			req = &pb.UpdateCommentRequest{
				Id:      id.String(),
				Content: content,
			}
		})

		JustBeforeEach(func() {
			resp, err = svc.UpdateComment(ctx, req)
		})

		When("postgres error", func() {
			var comment *dao.Comment

			BeforeEach(func() {
				comment = &dao.Comment{
					ID:      id,
					Content: content,
				}
				commentDAO.EXPECT().Update(ctx, comment).Return(errPGUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errPGUnknown))
			})
		})

		When("comment not found", func() {
			var comment *dao.Comment

			BeforeEach(func() {
				comment = &dao.Comment{
					ID:      id,
					Content: content,
				}
				commentDAO.EXPECT().Update(ctx, comment).Return(ErrCommentNotFound)
			})

			It("return comment not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			var comment *dao.Comment

			BeforeEach(func() {
				comment = &dao.Comment{
					ID:      id,
					Content: content,
				}
				commentDAO.EXPECT().Update(ctx, comment).Return(nil)
			})

			It("returns without any error", func() {
				Expect(resp).To(Equal(&pb.UpdateCommentResponse{}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeleteComment", func() {
		var (
			req  *pb.DeleteCommentRequest
			resp *pb.DeleteCommentResponse
			id   uuid.UUID
			err  error
		)

		BeforeEach(func() {
			id = uuid.New()
			req = &pb.DeleteCommentRequest{Id: id.String()}
		})

		JustBeforeEach(func() {
			resp, err = svc.DeleteComment(ctx, req)
		})

		When("postgres error", func() {

			BeforeEach(func() {
				commentDAO.EXPECT().Delete(ctx, id).Return(errPGUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errPGUnknown))
			})
		})

		When("comment not found", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Delete(ctx, id).Return(ErrCommentNotFound)
			})

			It("return comment not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Delete(ctx, id).Return(nil)
			})

			It("returns without any error", func() {
				Expect(resp).To(Equal(&pb.DeleteCommentResponse{}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
