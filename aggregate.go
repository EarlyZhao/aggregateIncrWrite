package aggregateIncrWrite

import "context"

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

func New(store AggregateStoreInterface, conf *Config,
		 saveHandler func(id string, aggIncr int64) error,
		 failureHandler func(id string, aggIncr int64)) *Aggregate {

	agg := &Aggregate{
		store: store,
		config: conf,
		saveHandler: saveHandler,
		failureHandler: failureHandler,
	}

	agg.store.start()
	return agg
}

