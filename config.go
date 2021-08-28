package aggregateIncrWrite

import (
	"time"
)

// 配置只能更新一次
type Config struct {
	// 聚合时间窗口 窗口到期后触发写操作
	Interval time.Duration
	// 并发队列数量 独立channel
	ConcurrencyBuffer int
	// 单队列buffer大小
	Buffer int
	// 写操作并行度 执行save操作gourouting数量
	SaveConcurrency int

	IncrTimeout time.Duration

	// 自定义日志
	logger Logger
	// 自定义埋点
	metric Metric
}

func (c *Config) getInterval() time.Duration {
	if c.Interval > 0 {
		return c.Interval
	}
	return time.Second
}

func (c *Config) getConcurrency() int {
	if c.ConcurrencyBuffer > 0 {
		return c.ConcurrencyBuffer
	}

	return 1
}

func (c *Config) getBufferNum() int {
	if c.Buffer > 0 {
		return c.Buffer
	}

	return 2048
}

func (c *Config) getSaveConcurrency() int {
	if c.SaveConcurrency > 0 {
		return c.SaveConcurrency
	}

	return 1
}


func emptyConf() *Config{
	return &Config{}
}

func (c *Config) getLogger() Logger {
	if c.logger != nil {
		return c.logger
	}

	return emptyLogger
}

func (c *Config) getMetric() Metric {
	if c.metric != nil {
		return c.metric
	}

	return emptyMetricer
}

