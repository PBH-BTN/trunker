package cache

import (
	"context"
	"errors"
	"time"

	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/redis/go-redis/v9"
	"github.com/shamaton/msgpack/v2"
)

func Get[T any](ctx context.Context, key string) (*T, bool) {
	str, err := Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get value from redis, key: %s, error: %s", key, err)
		return nil, false
	}
	if str == "" {
		return nil, false
	}
	var value T
	err = msgpack.Unmarshal(conv.UnsafeStringToBytes(str), &value)
	if err != nil {
		hlog.CtxWarnf(ctx, "failed to unmarshal value, key: %s, value: %s , error:%s", key, str, err)
		return nil, false
	}
	return &value, true
}

func GetList[T any](ctx context.Context, key string) ([]*T, bool) {
	str := Client.Get(ctx, key).Val()
	if str == "" {
		return nil, false
	}
	var value []*T
	err := msgpack.Unmarshal(conv.UnsafeStringToBytes(str), &value)
	if err != nil {
		hlog.CtxWarnf(ctx, "failed to unmarshal value, key: %s, value: %s , error:%s", key, str, err)
		return nil, false
	}
	return value, true
}

func Del(ctx context.Context, key string) error {
	hlog.CtxInfof(ctx, "delete key from redis, key: "+key)
	return Client.Del(ctx, key).Err()
}

func Set[T any](ctx context.Context, key string, value T, expired time.Duration) error {
	str, err := msgpack.Marshal(value)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to marshal value, key: %s, value: %v, error: %s", key, value, err)
		return err
	}
	err = Client.Set(ctx, key, str, expired).Err()
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to set value to redis, key: %s, value: %v, error: %s", key, value, err)
	}
	return err
}
