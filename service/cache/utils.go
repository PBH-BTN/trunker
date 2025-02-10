package cache

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/shamaton/msgpack/v2"
)

func Get[T any](ctx context.Context, key string) (*T, bool) {
	str := Client.Get(ctx, key).Val()
	if str == "" {
		return nil, false
	}
	var value T
	err := msgpack.Unmarshal([]byte(str), &value)
	if err != nil {
		hlog.CtxWarnf(ctx, "failed to unmarshal value, key: %s, value: %s , error:%s", key, str, err)
		return nil, false
	}
	hlog.CtxInfof(ctx, "get value from redis, key: "+key)
	return &value, true
}

func GetList[T any](ctx context.Context, key string) ([]*T, bool) {
	str := Client.Get(ctx, key).Val()
	if str == "" {
		return nil, false
	}
	var value []*T
	err := msgpack.Unmarshal([]byte(str), &value)
	if err != nil {
		hlog.CtxWarnf(ctx, "failed to unmarshal value, key: %s, value: %s , error:%s", key, str, err)
		return nil, false
	}
	hlog.CtxInfof(ctx, "get value from redis, key: "+key)
	return value, true
}

func Del(ctx context.Context, key string) error {
	hlog.CtxInfof(ctx, "delete key from redis, key: "+key)
	return Client.Del(ctx, key).Err()
}

func Set[T any](ctx context.Context, key string, value T, expired time.Duration) error {
	str, _ := msgpack.Marshal(value)
	return Client.Set(ctx, key, str, expired).Err()
}
