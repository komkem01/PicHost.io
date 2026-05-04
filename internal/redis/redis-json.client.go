package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var Nil = redis.Nil

type JSONClient struct {
	*redis.Client
}

func newJSONClient(cli *redis.Client) *JSONClient {
	return &JSONClient{cli}
}

func (rjs *JSONClient) GetJSON(ctx context.Context, key string, value any) error {
	b, err := rjs.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, value)
}

func (rjs *JSONClient) SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rjs.Set(ctx, key, b, expiration).Err()
}
