package s3client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Remover interface {
	DeleteObject(
		ctx context.Context,
		params *s3.DeleteObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.DeleteObjectOutput, error)
}
