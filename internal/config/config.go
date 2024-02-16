package config

import "github.com/caarlos0/env/v10"

type Environment struct {
	PostgresURI string `env:"POSTGRES_DB_URI" envDefault:"postgres://echopguser:pgpw4echo@localhost:6462/echodb?sslmode=disable" validate:"uri"`
	RpcURI      string `env:"RPC_URI" envDefault:"localhost:9091" validate:"uri"`
}

func New() (conf Environment, err error) {
	conf = Environment{}
	err = env.Parse(&conf)
	return
}
