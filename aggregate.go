package aggregateIncrWrite

import (
	"context"
	"time"
	"sync"
)

type Option func(a *Aggregate)

type Aggregate struct {
	config *Config

	store AggregateStoreInterface

	// 聚合保存回调
	saveHandler func(id string, aggIncr int64) error
	// 回调错误时触发
	failureHandler func(id string, aggIncr int64)

	stoped chan bool
	wait sync.WaitGroup
}

func (c *Aggregate) SetLogger(logger Logger) {
	c.config.logger = logger
}

func (c *Aggregate) SetMetric(m Metric) {
	c.config.metric = m
}

func (a *Aggregate) Stop(ctx context.Context) error {
	a.store.stop(ctx)
	for i := 0; i < a.config.saveConcurrency; i++ {
		a.stoped <- true
	}
	a.wait.Wait()

	return nil
}

func (a *Aggregate) Incr(ctx context.Context, id string, delta int64) (err error) {
	err = a.store.incr(ctx, id, delta)
	if err != nil {
		// todo metric
		err = a.saveHandler(id, delta)
	}
	return
}

func New(conf *Config, options ...Option) *Aggregate {
	if conf == nil {
		conf = emptyConf()
	}
	agg := &Aggregate{
		store: NewLocalStore(), // default
		config: conf,
		stoped: make(chan bool),
	}

	for _, op := range options{
		op(agg)
	}

	if agg.saveHandler == nil {
		panic("saveHandler must exist, use SetOptionSaveHandler")
	}

	agg.store.start(agg.config)

	for i := 0; i < agg.config.saveConcurrency; i++ {
		agg.wait.Add(1)
		go agg.saveWorker()
	}

	return agg
}

func SetOptionStore(store AggregateStoreInterface) Option{
	return func(a *Aggregate){
		a.store = store
	}
}

func SetOptionSaveHandler(save func(id string, aggIncr int64) error) Option{
	return func(a *Aggregate){
		a.saveHandler = save
	}
}

func SetOptionFailHandler(fail func(id string, aggIncr int64)) Option{
	return func(a *Aggregate){
		a.failureHandler = fail
	}
}

func SetOptionIncrTimeout(t time.Duration) Option{
	return func(a *Aggregate){
		a.config.incrTimeout = t
	}
}