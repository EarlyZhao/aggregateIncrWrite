package aggregateIncrWrite

import "context"

func NewLocalStore() AggregateStoreInterface{
	return &storeLocal{}
}

type storeLocal struct {
	Config
}

func(a *storeLocal) incr(ctx context.Context, id string, val int64)( err error) {
	return
}

func(a *storeLocal) stop(ctx context.Context) (err error) {
	return
}

func(a *storeLocal) start(){
	return
}
