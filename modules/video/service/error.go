package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidObjectID = status.Errorf(codes.InvalidArgument, "invalid objectID")
	ErrVideoNotFound   = status.Errorf(codes.NotFound, "video not found")
)
