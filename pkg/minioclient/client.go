package minioclient

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

func New(conf *Config) *minio.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone() //nolint:errcheck
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
	}

	minioClient, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure:    conf.UseSSL,
		Transport: tr,
	})
	if err != nil {
		panic(err)
	}

	exists, err := minioClient.BucketExists(
		context.Background(),
		conf.BucketName,
	)
	if err != nil {
		panic(err)
	}

	if !exists {
		if err := minioClient.MakeBucket(
			context.Background(),
			conf.BucketName,
			minio.MakeBucketOptions{},
		); err != nil {
			panic(err)
		}
	}

	return minioClient
}
