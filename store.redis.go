package aggregateIncrWrite

import "context"

func NewRedisStore() AggregateStoreInterface{
	return &storeRedis{}
}

type storeRedis struct {
	Config
}

func(a *storeRedis) incr(ctx context.Context, id string, val int64)( err error) {
	return
}

func(a *storeRedis) stop(ctx context.Context) (err error) {
	return
}

func(a *storeRedis) start(c *Config)  {
	return
}

func (a *storeRedis) batchAgg() chan aggItem{
	return nil
}