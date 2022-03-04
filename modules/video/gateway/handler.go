package gateway

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Handler interface {
	HandleUploadVideo(w http.ResponseWriter, r *http.Request, params map[string]string)
}

type handler struct {
	client pb.VideoClient
	logger *logkit.Logger
}

func NewHandler(client pb.VideoClient, logger *logkit.Logger) *handler {
	return &handler{
		client: client,
		logger: logger,
	}
}

func (h *handler) HandleUploadVideo(w http.ResponseWriter, req *http.Request, params map[string]string) {
	if err := req.ParseForm(); err != nil {
		h.encodeJSONResponse(w, NewResponseError(http.StatusBadRequest, "failed to parse form", err))
		return
	}

	f, header, err := req.FormFile("file")
	if err != nil {
		h.encodeJSONResponse(w, NewResponseError(http.StatusBadRequest, "failed to get file", err))
		return
	}
	defer f.Close()

	stream, err := h.client.UploadVideo(req.Context())
	if err != nil {
		h.encodeJSONResponse(w, NewResponseError(http.StatusInternalServerError, "failed to create stream client", err))
	}

	// 1. send file header first
	if serr := stream.Send(&pb.UploadVideoRequest{
		Data: &pb.UploadVideoRequest_Header{
			Header: &pb.VideoHeader{
				Filename: header.Filename,
				Size:     uint64(header.Size),
			},
		},
	}); serr != nil {
		h.encodeJSONResponse(w, NewResponseError(http.StatusInternalServerError, "failed to send file header", serr))
		return
	}

	// 2. send file chunk
	reader := bufio.NewReader(f)
	buffer := make([]byte, 1024)

	for {
		n, rerr := reader.Read(buffer)
		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				break
			}

			h.encodeJSONResponse(w, NewResponseError(http.StatusInternalServerError, "failed to read file into buffer", rerr))
			return
		}

		if serr := stream.Send(&pb.UploadVideoRequest{
			Data: &pb.UploadVideoRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}); serr != nil {
			h.encodeJSONResponse(w, NewResponseError(http.StatusInternalServerError, "failed to send file chuck data", serr))
			return
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		h.encodeJSONResponse(w, NewResponseError(http.StatusInternalServerError, "failed to receive upload file response", err))
		return
	}

	h.encodeJSONResponse(w, resp)
}

func (h *handler) encodeJSONResponse(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if message, ok := resp.(proto.Message); ok {
		h.encodeProtoJSONResponse(w, message)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if coder, ok := resp.(StatusCoder); ok {
		w.WriteHeader(coder.StatusCode())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) encodeProtoJSONResponse(w http.ResponseWriter, resp proto.Message) {
	o := &protojson.MarshalOptions{
		EmitUnpopulated: true,
	}

	bytes, err := o.Marshal(resp)
	if err != nil {
		h.logger.Error("failed to encode protobuf JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(bytes); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if coder, ok := resp.(StatusCoder); ok {
		w.WriteHeader(coder.StatusCode())
		return
	}

	w.WriteHeader(http.StatusOK)
}
