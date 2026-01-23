package config

import (
	_ "embed"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DBUser        string `env:"POSTGRES_USER"`
	DBName        string `env:"POSTGRES_DB"`
	DBPass        string `env:"POSTGRES_PASSWORD"`
	DBHost        string `env:"DB_HOST" env-default:"localhost"`
	DBPort        string `env:"DB_PORT" env-default:"5432"`
	CSRFKey       string `env:"CSRF_KEY"`
	SessionSecret string `env:"SESSION_SECRET"`
}

//go:embed .env
var envFile string

var Conf = func() *Config {
	m, err := godotenv.Unmarshal(envFile)
	if err != nil {
		panic(err)
	}

	for k, v := range m {
		os.Setenv(k, v)
	}

	var c Config
	err = cleanenv.ReadEnv(&c)
	if err != nil {
		panic(err)
	}
	return &c
}()
