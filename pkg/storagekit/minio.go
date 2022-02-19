package storagekit

import (
	"context"
	"io"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type MinIOConfig struct {
	Endpoint string `long:"endpoint" env:"ENDPOINT" description:"the endpoint of MinIO server" required:"true"`
	Bucket   string `long:"bucket" env:"BUCKET" description:"the bucket name" required:"true"`
	Username string `long:"username" env:"USERNAME" description:"the access key id (username) to the MinIO server" required:"true"`
	Password string `long:"password" env:"PASSWORD" description:"the secret access key (password) to the MinIO server" required:"true"`
	Insecure bool   `long:"insecure" env:"INSECURE" description:"disable HTTPS or not"`
}

type MinIOClient struct {
	*minio.Client
	bucketName string
}

var _ Storage = (*MinIOClient)(nil)

func NewMinIOClient(ctx context.Context, conf *MinIOConfig) *MinIOClient {
	logger := logkit.FromContext(ctx).
		With(zap.String("endpoint", conf.Endpoint)).
		With(zap.String("bucket", conf.Bucket))

	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.Username, conf.Password, ""),
		Secure: !conf.Insecure,
	})
	if err != nil {
		logger.Fatal("failed to create MinIO client", zap.Error(err))
	}

	if conf.Bucket != "" {
		ok, err := client.BucketExists(ctx, conf.Bucket)
		if err != nil {
			logger.Fatal("failed to check bucket existence", zap.Error(err))
		}

		if !ok {
			if err := client.MakeBucket(ctx, conf.Bucket, minio.MakeBucketOptions{}); err != nil {
				logger.Fatal("failed to create bucket", zap.Error(err))
			}
		}
	}

	logger.Info("create MinIO client successfully")

	return &MinIOClient{
		Client:     client,
		bucketName: conf.Bucket,
	}
}

func (c *MinIOClient) Endpoint() string {
	return c.Client.EndpointURL().Path
}

func (c *MinIOClient) Bucket() string {
	return c.bucketName
}

func (c *MinIOClient) PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, opts PutObjectOptions) error {
	if _, err := c.Client.PutObject(ctx, c.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: opts.ContentType,
	}); err != nil {
		return err
	}

	return nil
}
