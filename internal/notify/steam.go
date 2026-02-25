package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"asum/pkg/rdb"
	"asum/pkg/utils"
	"asum/pkg/wshub"

	"github.com/redis/go-redis/v9"
)

const (
	StreamKeyDefault = "stream:notify"
	GroupDefault     = "notify-group"
)

type Event struct {
	Kind string `json:"kind"`
	UID  uint64 `json:"uid,omitempty"`
	Data any    `json:"data,omitempty"`
	Ts   int64  `json:"ts"`
}

func unreadKey(uid uint64) string {
	return fmt.Sprintf("notify:unread:%d", uid)
}

func EnsureGroup(ctx context.Context, redisDB *rdb.Client, streamKey, group string) error {
	err := redisDB.XGroupCreateMkStream(ctx, streamKey, group, "$").Err()
	if err == nil {
		return nil
	}
	if utils.StringsContains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func PublishUser(ctx context.Context, redisDB *rdb.Client, streamKey string, uid uint64, data any) error {
	ev := Event{Kind: "user", UID: uid, Data: data, Ts: time.Now().Unix()}
	b, _ := json.Marshal(ev)
	return redisDB.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]any{"event": string(b)},
	}).Err()
}

func PublishBroadcast(ctx context.Context, redisDB *rdb.Client, streamKey string, data any) error {
	ev := Event{Kind: "broadcast", Data: data, Ts: time.Now().Unix()}
	b, _ := json.Marshal(ev)
	return redisDB.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]any{"event": string(b)},
	}).Err()
}

func GetUnread(ctx context.Context, redisDB *rdb.Client, uid uint64) (int64, error) {
	val, err := redisDB.Get(ctx, unreadKey(uid)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, nil
	}
	return n, nil
}

func IncUnread(ctx context.Context, redisDB *rdb.Client, uid uint64, delta int64) (int64, error) {
	return redisDB.IncrBy(ctx, unreadKey(uid), delta).Result()
}

func ClearUnread(ctx context.Context, redisDB *rdb.Client, uid uint64) error {
	return redisDB.Set(ctx, unreadKey(uid), 0, 0).Err()
}

func RunConsumer(ctx context.Context, redisDB *rdb.Client, hub *wshub.Hub, streamKey, group, consumerName string) error {
	if err := EnsureGroup(ctx, redisDB, streamKey, group); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		res, err := redisDB.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumerName,
			Streams:  []string{streamKey, ">"},
			Count:    50,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {
				raw, ok := msg.Values["event"]
				if !ok {
					_ = redisDB.XAck(ctx, streamKey, group, msg.ID).Err()
					continue
				}

				evStr := fmt.Sprintf("%v", raw)
				var ev Event
				if err := json.Unmarshal([]byte(evStr), &ev); err != nil {
					_ = redisDB.XAck(ctx, streamKey, group, msg.ID).Err()
					continue
				}

				// 推送逻辑
				switch ev.Kind {
				case "user":
					if hub.Online(ev.UID) {
						hub.PushToUID(ev.UID, wshub.MustJSON(Out{Type: "notify", Data: ev.Data, Ts: ev.Ts}))
						if n, err := GetUnread(ctx, redisDB, ev.UID); err == nil {
							hub.PushToUID(ev.UID, wshub.MustJSON(Out{Type: "badge", Unread: n, Ts: ev.Ts}))
						}
					}
				case "broadcast":
					hub.Broadcast(wshub.MustJSON(Out{Type: "notify", Data: ev.Data, Ts: ev.Ts}))
					for _, uid := range hub.OnlineUIDs() {
						if n, err := GetUnread(ctx, redisDB, uid); err == nil {
							hub.PushToUID(uid, wshub.MustJSON(Out{Type: "badge", Unread: n, Ts: ev.Ts}))
						}
					}
				default:
				}

				_ = redisDB.XAck(ctx, streamKey, group, msg.ID).Err()
			}
		}
	}
}
