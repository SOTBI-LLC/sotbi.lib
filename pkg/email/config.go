package email

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	User   string `env:"USER"            envDefault:"bh.dev@grp.loc"`
	Server string `env:"SERVER,notEmpty"`
	// 0 - вообще без любого TLS, 1 - используем StartTLS, 2 - используем tls.Dial
	UseTLS int8   `env:"USE_TLS"         envDefault:"1"`
	Sender string `env:"SENDER"          envDefault:"bh.dev@grp.loc"`
}

func (r *Config) Load() error {
	return env.Parse(&r)
}
