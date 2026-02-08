package rdb

import (
	"errors"
	"runtime"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func New(conf Config) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
		PoolSize: 10 * runtime.GOMAXPROCS(0),
	})
	return &Client{rdb}
}

var (
	ErrNotFound = errors.New("key not found")
	ErrExpired  = errors.New("key expired")
)
