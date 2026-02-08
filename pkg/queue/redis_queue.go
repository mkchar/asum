package queue

import (
	"asum/pkg/rdb"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrQueueTimeout = errors.New("queue pop timeout")
	ErrQueueEmpty   = errors.New("queue is empty")
)

type RedisQueue[T any] struct {
	client *rdb.Client
	key    string
}

func NewRedisQueue[T any](client *rdb.Client, key string) *RedisQueue[T] {
	return &RedisQueue[T]{
		client: client,
		key:    key,
	}
}

func (q *RedisQueue[T]) Push(ctx context.Context, item T) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, q.key, data).Err()
}

func (q *RedisQueue[T]) Pop(ctx context.Context, timeout time.Duration) (T, error) {
	var result T
	val, err := q.client.BLPop(ctx, timeout, q.key).Result()
	if err != nil {
		if err == redis.Nil {
			return result, ErrQueueTimeout
		}
		return result, err
	}

	if err := json.Unmarshal([]byte(val[1]), &result); err != nil {
		return result, err
	}

	return result, nil
}

func (q *RedisQueue[T]) Len(ctx context.Context) (int64, error) {
	return q.client.LLen(ctx, q.key).Result()
}

func (q *RedisQueue[T]) Clear(ctx context.Context) error {
	return q.client.Del(ctx, q.key).Err()
}
