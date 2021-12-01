package redis

import (
	"ServerServing/config"
	"context"
	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func GetCli() *redis.Client {
	return rdb
}

func InitRedis() {
	conf := config.GetConfig()
	rdb = redis.NewClient(&redis.Options{
		Addr:     conf.RedisConfig.Addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cmd := rdb.Ping(context.Background())
	_, err := cmd.Result()
	if err != nil {
		panic(err)
	}
}
