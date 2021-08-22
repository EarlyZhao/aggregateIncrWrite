package aggregateIncrWrite

import "context"

type Option func(a *Aggregate)

type Aggregate struct {
	config *Config

	store AggregateStoreInterface

	// 自定义日志
	logger Logger
	// 自定义埋点
	metric Metric

	// 聚合保存回调
	saveHandler func(id string, aggIncr int64) error
	// 回调错误时触发
	failureHandler func(id string, aggIncr int64)
}

func (c *Aggregate) SetLogger(logger Logger) {
	c.logger = logger
}

func (c *Aggregate) SetMetric(m Metric) {
	c.metric = m
}

func (c *Aggregate) Stop(ctx context.Context) error {
	return nil
}

func New(conf *Config, options ...Option) *Aggregate {
	agg := &Aggregate{
		store: NewLocalStore(), // default
		config: conf,
	}

	for _, op := range options{
		op(agg)
	}

	if agg.saveHandler == nil {
		panic("saveHandler must exist, use SetOptionSaveHandler")
	}

	agg.store.start(agg.config)
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
