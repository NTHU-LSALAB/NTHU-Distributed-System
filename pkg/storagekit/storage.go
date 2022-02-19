package storagekit

import (
	"context"
	"io"
)

type PutObjectOptions struct {
	ContentType string
}

// Provide a simplifier interface to upload file
type Storage interface {
	// Endpoint returns the endpoint of the object storage
	Endpoint() string
	// Bucket returns the bucket name in the object storage
	Bucket() string

	// PutObject add an object into the storage bucket
	PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, opts PutObjectOptions) error
}
