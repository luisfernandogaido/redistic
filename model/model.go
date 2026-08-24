package model

import (
	"context"
	"crypto/tls"

	"github.com/redis/go-redis/v9"
)

var (
	RedisCore *redis.Client
	RedisEdge *redis.Client
	Ctx       context.Context
)

func init() {
	RedisCore = redis.NewClient(&redis.Options{
		Addr:     "72.60.250.80:6380",
		Password: "sbCK7YLM0qcRM6c",
		DB:       0,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	RedisEdge = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	Ctx = context.Background()
}
