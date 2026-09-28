package minioclient

import (
	"context"

	"github.com/minio/minio-go/v7"
)

// Deprecated: use S3Client interface instead.
type Stat interface {
	StatObject(
		ctx context.Context,
		bucketName, objectName string,
		opts minio.StatObjectOptions,
	) (minio.ObjectInfo, error)
}
