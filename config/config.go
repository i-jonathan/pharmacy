package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	DBUser string `env:"POSTGRES_USER"`
	DBName string `env:"POSTGRES_DB"`
	DBPass string `env:"POSTGRES_PASSWORD"`
	DBHost string `env:"DB_HOST" env-default:"localhost"`
	DBPort string `env:"DB_PORT" env-default:"5432"`
}

var c *Config

func LoadConfig() error {
	var conf Config
	err := cleanenv.ReadEnv(&conf)
	if err != nil {
		return err
	}
	c = &conf
	return nil
}

func GetConfig() *Config {
	return c
}
