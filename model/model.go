package model

import (
	"context"
	"crypto/tls"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	uri = "mongodb://gaido:xW3g3JAhsZjE-@72.60.250.80/storage?authSource=admin"
)

var (
	RedisCore *redis.Client
	RedisEdge *redis.Client
	Ctx       context.Context
	client    *mongo.Client
	db        *mongo.Database
)

func init() {
	var (
		err error
	)
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
	client, err = mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	db = client.Database("redistic")
}
