package s3client

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type PutterGetter interface {
	GetRemover
	PutRemover
}

type GetRemover interface {
	Getter
	Remover
}

type PutRemover interface {
	Putter
	Remover
	Stat
}

func New(conf *Config) *s3.Client {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(conf.Region),
		config.WithHTTPClient(httpClient()),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(conf.AccessKey, conf.SecretKey, ""),
		),
	)
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(conf.Endpoint)
		o.UsePathStyle = true
	})

	_, err = client.HeadBucket(
		ctx,
		&s3.HeadBucketInput{Bucket: aws.String(conf.BucketName)},
	)
	if err != nil && !isNotFound(err) {
		panic(err)
	}

	if err != nil {
		if _, err := client.CreateBucket(
			ctx,
			&s3.CreateBucketInput{Bucket: aws.String(conf.BucketName)},
		); err != nil {
			panic(err)
		}
	}

	return client
}

func httpClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone() //nolint:errcheck
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
	}

	return &http.Client{Transport: tr}
}

// isNotFound reports whether the error means that the bucket or object does not exist.
func isNotFound(err error) bool {
	if _, ok := errors.AsType[*types.NotFound](err); ok {
		return true
	}

	var responseErr *awshttp.ResponseError

	return errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == http.StatusNotFound
}
