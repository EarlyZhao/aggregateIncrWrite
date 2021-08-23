package aggregateIncrWrite

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"sync"
	"time"
)

const incrTimeoutStr = "incr timeout"

func NewLocalStore() AggregateStoreInterface{
	return &storeLocal{stopChan: make(chan bool), wait: &sync.WaitGroup{}}
}

type storeLocal struct {
	*Config
	buffers []chan *incrItem
	stopChan chan bool
	batchAggChan chan aggItem
	stoped bool
	wait *sync.WaitGroup
}

func(a *storeLocal) incr(ctx context.Context, id string, val int64)( err error) {
	if a.stoped {
		return errors.New("has closesed, incr failed")
	}

	item := &incrItem{id: id, delta: val}
	buffer := a.dispatch(ctx, item)

	var incrTimeout  <-chan time.Time

	// 防御性超时
	if a.Config.incrTimeout > 0 {
		incrTimeout = time.After(time.Duration(a.Config.incrTimeout))
	}
	select {
		case <- ctx.Done():
			return ctx.Err()
		case buffer <- item:
			return nil
		case <-incrTimeout:
			return errors.New(fmt.Sprintf("%s after: %v", incrTimeoutStr, a.Config.incrTimeout))
	}
}

func (a *storeLocal) dispatch(ctx context.Context, item *incrItem) chan *incrItem{
	sum := crc32.ChecksumIEEE([]byte(item.id))
	index := sum % uint32(len(a.buffers))
	return a.buffers[index]
}

func(a *storeLocal) stop(ctx context.Context) (err error) {
	a.stoped = true

	close(a.stopChan)
	a.wait.Wait()

	return
}

func(a *storeLocal) start(c *Config){
	a.Config = c
	a.buffers = make([]chan *incrItem, a.Config.getConcurrency())
	a.batchAggChan = make(chan aggItem, a.Config.getBufferNum())

	for i := 0; i < a.Config.getConcurrency(); i ++ {
		buffer := make(chan *incrItem, a.Config.getBufferNum())
		a.buffers[i] = buffer
		a.wait.Add(1)
		go a.aggregating(buffer)
	}

	return
}

func (a *storeLocal) batchAgg() chan aggItem {
	return a.batchAggChan
}

func (a *storeLocal) aggregating(buffer chan *incrItem) {
	pool := make(aggItem)
	ticker := time.Tick(a.Config.getInterval())
	defer a.wait.Done()

	for {
		select {
			case <- a.stopChan:
				a.stoped = true

				for {
					select {
					case item := <- buffer:
						pool[item.id] += item.delta
					default:
						a.batchAggChan <- pool
						a.Config.getLogger().Info("aggregating exit")
						return
					}
				}
			case item, ok := <- buffer:
				if !ok {
					break
				}

				pool[item.id] += item.delta
				a.Config.getMetric().MetricIncrCount(item.delta)
			case <-ticker:
				a.Config.getLogger().Info("run save")
				a.batchAggChan <- pool
				a.Config.getMetric().MetricBatchCount(int64(len(pool)))
				pool = make(aggItem)
		}
	}
}
