package providers

import "context"

type Client interface {
	SearchProduct(ctx context.Context, group, query string) ([]Product, error)
}
