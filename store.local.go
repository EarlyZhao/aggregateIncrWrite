package aggregateIncrWrite

import (
	"context"
	"errors"
	"hash/crc32"
	"sync"
	"time"
)

const incrBufferFull = "buffer full"

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

	select {
		case <- ctx.Done():
			return ctx.Err()
		case buffer <- item:
			return nil
		default:
			return errors.New(incrBufferFull)
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
	a.batchAggChan = make(chan aggItem, a.Config.getBufferNum() * 100)

	interval := int(a.Config.getInterval())
	intervalPice := interval/a.Config.getConcurrency()

	for i := 0; i < a.Config.getConcurrency(); i ++ {
		buffer := make(chan *incrItem, a.Config.getBufferNum())
		a.buffers[i] = buffer
		a.wait.Add(1)

		delayInterval := interval
		go func() {
			time.Sleep(time.Duration(delayInterval)) // 均匀打散
			a.aggregating(buffer)
		}()

		interval -= intervalPice
	}

	return
}

func (a *storeLocal) batchAgg() chan aggItem {
	return a.batchAggChan
}

func (a *storeLocal) aggregating(buffer chan *incrItem) {
	pool := make(aggItem)
	// 让tick的启动时间分散
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
				a.Config.getLogger().Info("lcoal run save")
				a.batchAggChan <- pool
				a.Config.getMetric().MetricBatchCount(int64(len(pool)))
				pool = make(aggItem)
		}
	}
}
