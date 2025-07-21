package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DBUser string `env:"POSTGRES_USER"`
	DBName string `env:"POSTGRES_DB"`
	DBPass string `env:"POSTGRES_PASSWORD"`
	DBHost string `env:"DB_HOST" env-default:"localhost"`
	DBPort string `env:"DB_PORT" env-default:"5432"`
}

var Conf = func() *Config {
	if err := godotenv.Load(); err != nil {
        panic(err)
    }
    
	var c Config
	err := cleanenv.ReadEnv(&c)
	if err != nil {
		panic(err)
	}
	return &c
}()
