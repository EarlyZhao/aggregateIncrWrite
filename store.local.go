package aggregateIncrWrite

import "context"

func NewLocalStore() AggregateStoreInterface{
	return &storeLocal{}
}

type storeLocal struct {

}

func(a *storeLocal) incr(ctx context.Context, id string, val int64)( err error) {
	return
}

func(a *storeLocal) stop(ctx context.Context) (err error) {
	return
}

func(a *storeLocal) start(c *Config){
	return
}

func (a *storeLocal) batchAgg() chan aggItem{
	return nil
}