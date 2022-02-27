package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/mock/daomock"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"github.com/golang/mock/gomock"
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
			req  *pb.ListCommentRequest
			resp *pb.ListCommentResponse
			err  error
		)

		JustBeforeEach(func() {
			resp, err = svc.ListComment(ctx, req)
		})

		When("postgres error", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().List(ctx, req.GetVideoId(), req.GetLimit(), req.GetSkip()).Return(nil, errPGUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errPGUnknown))
			})
		})

		When("success", func() {
			var comments []*dao.Comment

			BeforeEach(func() {
				comments = []*dao.Comment{dao.NewFakeComment(), dao.NewFakeComment()}
				commentDAO.EXPECT().List(ctx, req.GetVideoId(), req.GetLimit(), req.GetSkip()).Return(comments, nil)
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

	Describe("UpdateComment", func() {
		var (
			req  *pb.UpdateCommentRequest
			resp *pb.UpdateCommentResponse
			err  error
		)

		JustBeforeEach(func() {
			resp, err = svc.UpdateComment(ctx, req)
		})

		When("postgres error", func() {
			var comment *dao.Comment

			BeforeEach(func() {
				comment = dao.NewFakeComment()
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
				comment = dao.NewFakeComment()
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
				comment = dao.NewFakeComment()
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
			id   string
			err  error
		)

		JustBeforeEach(func() {
			resp, err = svc.DeleteComment(ctx, req)
		})

		When("postgres error", func() {

			BeforeEach(func() {
				commentDAO.EXPECT().Delete(ctx, comment).Return(errPGUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errPGUnknown))
			})
		})

		When("comment not found", func() {
			var comment *dao.Comment

			BeforeEach(func() {
				comment = dao.NewFakeComment()
				commentDAO.EXPECT().Delete(ctx, comment).Return(ErrCommentNotFound)
			})

			It("return comment not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			var comment *dao.Comment

			BeforeEach(func() {
				comment = dao.NewFakeComment()
				commentDAO.EXPECT().Update(ctx, comment).Return(nil)
			})

			It("returns without any error", func() {
				Expect(resp).To(Equal(&pb.UpdateCommentResponse{}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
