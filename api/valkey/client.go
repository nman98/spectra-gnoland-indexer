package valkey

import (
	"context"
	"fmt"
	"time"

	valkey "github.com/valkey-io/valkey-go"
)

type ValkeyClient struct {
	client valkey.Client
}

func NewValkeyClient(host string, port int) (*ValkeyClient, error) {
	timeout := 5 * time.Second
	cfg := valkey.ClientOption{
		InitAddress:      []string{fmt.Sprintf("%s:%d", host, port)},
		ConnWriteTimeout: timeout,
		ConnLifetime:     timeout,
	}
	client, err := valkey.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ValkeyClient{
		client: client,
	}, nil
}

func (c *ValkeyClient) Close() {
	c.client.Close()
}

func (c *ValkeyClient) Increment(key string, ctx context.Context) (int64, error) {
	return c.client.Do(ctx, c.client.B().Incr().Key(key).Build()).AsInt64()
}

func (c *ValkeyClient) Expirer(key string, ctx context.Context, expiration time.Duration) (bool, error) {
	return c.client.Do(ctx, c.client.B().Expire().Key(key).Seconds(int64(expiration.Seconds())).Build()).AsBool()
}

// ExpireNX sets the TTL on key only when the key currently has no expiry
// (EXPIRE key seconds NX). This is idempotent and safe to call on every
// request: it will not slide an already-running window, but it will
// self-heal any key that was incremented without ever receiving a TTL.
func (c *ValkeyClient) ExpireNX(key string, ctx context.Context, expiration time.Duration) (bool, error) {
	return c.client.Do(ctx, c.client.B().Expire().Key(key).Seconds(int64(expiration.Seconds())).Nx().Build()).AsBool()
}
