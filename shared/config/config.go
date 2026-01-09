package config

import (
	"github.com/burkeclove/shared/db/sqlc"
	"time"
	"log"
	"os"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DatabaseURL  string
}

func Load() Config {
	cfg := Config{
		DatabaseURL:  mustEnv("DATABASE_URL"),
	}
	return cfg
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env var %s", k)
	}
	return v
}

func (c *Config) CreatePool() *sqlc.Queries {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, c.DatabaseURL)
	if err != nil {
		log.Panicf("could not establish connection to db: ", err.Error())
	}
	q := sqlc.New(pool)
	return q
}
