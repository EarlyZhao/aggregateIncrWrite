package aggregateIncrWrite

import (
	"time"
)

// 配置只能更新一次
type Config struct {
	// 聚合时间窗口 窗口到期后触发写操作
	interval time.Duration
	// 并发队列数量 独立channel
	concurrencyBuffer int
	// 单队列buffer大小
	buffer int
	// 写操作并行度 执行save操作gourouting数量
	saveConcurrency int

	incrTimeout time.Duration

	// 自定义日志
	logger Logger
	// 自定义埋点
	metric Metric
}

func (c *Config) getInterval() time.Duration {
	if c.interval > 0 {
		return c.interval
	}
	return time.Second
}

func (c *Config) getConcurrency() int {
	if c.concurrencyBuffer > 0 {
		return c.concurrencyBuffer
	}

	return 1
}

func (c *Config) getBufferNum() int {
	if c.buffer > 0 {
		return c.buffer
	}

	return 2048
}

func (c *Config) getSaveConcurrency() int {
	if c.saveConcurrency > 0 {
		return c.saveConcurrency
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

