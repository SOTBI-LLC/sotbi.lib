package minioclient

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type Getter interface {
	GetObject(
		ctx context.Context,
		bucketName, objectName string,
		opts minio.GetObjectOptions,
	) (*minio.Object, error)
}
