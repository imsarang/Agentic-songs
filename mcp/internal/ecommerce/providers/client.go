package providers

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	SearchProduct(ctx context.Context, group, query string, rdb *redis.Client) ([]Product, error)
}
