package minioclient

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type Stat interface {
	StatObject(
		ctx context.Context,
		bucketName, objectName string,
		opts minio.StatObjectOptions,
	) (minio.ObjectInfo, error)
}
