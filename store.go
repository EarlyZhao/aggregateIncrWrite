package aggregateIncrWrite

import(
	"context"
)

type AggregateStoreInterface interface {
	incr(ctx context.Context, id string, delta int64)( err error)
	stop(ctx context.Context) (err error)
	start(config *Config)
	batchAgg() chan aggItem
}