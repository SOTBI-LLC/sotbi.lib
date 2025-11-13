package minioclient

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type Remover interface {
	RemoveObject(
		ctx context.Context,
		bucketName, objectName string,
		opts minio.RemoveObjectOptions,
	) error
}
