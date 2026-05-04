package redis

import (
	"context"

	"pichost.io/internal/provider"

	dto "pichost.io/internal/redis/dto"
)

type RedisModule struct {
	Svc *RedisService
}

var _ provider.Close = (*RedisModule)(nil)

func New(appEnv string, opts map[string]*dto.Option) *RedisModule {
	svc := newService(appEnv, opts)
	return &RedisModule{
		Svc: svc,
	}
}

func (db *RedisModule) Close(ctx context.Context) error {
	return db.Svc.close(ctx)
}
