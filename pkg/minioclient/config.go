package minioclient

type Config struct {
	Endpoint   string `env:"ENDPOINT,notEmpty"`
	AccessKey  string `env:"ACCESS_KEY,notEmpty"`
	SecretKey  string `env:"SECRET_KEY,notEmpty"`
	UseSSL     bool   `env:"USE_SSL"     envDefault:"false"`
	BucketName string `env:"BUCKET_NAME" envDefault:"attachments"`
}
