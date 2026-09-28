package minioclient

import (
	"context"

	"github.com/minio/minio-go/v7"
)

// Deprecated: use S3Client interface instead.
type Getter interface {
	GetObject(
		ctx context.Context,
		bucketName, objectName string,
		opts minio.GetObjectOptions,
	) (*minio.Object, error)
}
