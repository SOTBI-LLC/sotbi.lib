package minioclient

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

type Putter interface {
	PutObject(
		ctx context.Context,
		bucketName, objectName string,
		reader io.Reader,
		objectSize int64,
		opts minio.PutObjectOptions,
	) (minio.UploadInfo, error)
}
