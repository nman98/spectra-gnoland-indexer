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
	return c.client.Do(ctx, c.client.B().Expire().Key(key).Seconds(int64(expiration)).Build()).AsBool()
}
