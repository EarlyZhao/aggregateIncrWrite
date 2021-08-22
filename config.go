package aggregateIncrWrite

import (
	"time"
)

// 配置只能更新一次
type Config struct {
	// 聚合时间窗口 窗口到期后触发写操作
	interval time.Duration
	// 并发队列数量
	concurrency int
	// 单队列buffer大小
	buffer int
}

func (c *Config) getInterval() time.Duration {
	if c.interval > 0 {
		return c.interval
	}
	return time.Second
}

func (c *Config) getConcurrency() int {
	if c.concurrency > 0 {
		return c.concurrency
	}

	return 1
}

func (c *Config) getBufferNum() int {
	if c.buffer > 0 {
		return c.buffer
	}

	return 2048
}

func emptyConf() *Config{
	return &Config{}
}