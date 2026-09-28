package minioclient

import (
	"context"

	"github.com/minio/minio-go/v7"
)

// Deprecated: use S3Client interface instead.
type Remover interface {
	RemoveObject(
		ctx context.Context,
		bucketName, objectName string,
		opts minio.RemoveObjectOptions,
	) error
}
