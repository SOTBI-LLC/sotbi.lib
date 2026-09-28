package s3client

type Config struct {
	Endpoint   string `env:"ENDPOINT,notEmpty"`
	AccessKey  string `env:"ACCESS_KEY,notEmpty"`
	SecretKey  string `env:"SECRET_KEY,notEmpty"`
	Region     string `env:"REGION"              envDefault:"us-east-1"`
	BucketName string `env:"BUCKET_NAME"         envDefault:"attachments"`
}
