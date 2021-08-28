package aggregateIncrWrite

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"hash/crc32"
	"strconv"
	"sync"
	"time"
)

/*
方案概述：

1. 核心数据结构：
	- SkipList 按照首次写入时间(score)， id为member 维护一个跳表。保存(触发回调)时优先取score小的。
	- HashMap id为hash 的filed， value的累加通过 hincrby累计

2. 取数据
	- zset zrem可以取一个id出来
    - 在hash中删除这个field --- hdel
  均是线程安全的。

3. 并发能力
 	通过saveConcurrency 扩展队列的数量(skiplist 和 hashmap的数量)

4. 风险
   1. 数据安全。在 hdel 获取数据后，调用saveHandler失败，此时有丢数据的风险，可以通过实现failureHandler来尽量规避。
   2. 热key风险。每个队列会写同样的key，此风险可以通过增加saveConcurrency来规避

5. 数据积压
   redis是一个天然的暂存队列，当程序退出后，队列中暂存的数据，可以待下次启动再处理。
   当队列数量增加或扩展时，需要注意老队列被遗忘的情况。(可能老队列还有数据)
*/

func NewRedisStore(notChangeBussinessId string, cache *redis.Client) AggregateStoreInterface{
	return &storeRedis{cache: cache,
			notChangeBussinessId: notChangeBussinessId,
			wait: &sync.WaitGroup{},
			stopChan: make(chan bool),
	}
}

const _redisstorekey = "aIw"

type storeRedis struct {
	*Config
	notChangeBussinessId string

	cache *redis.Client

	skipListKeys []string

	wait *sync.WaitGroup

	batchAggChan chan aggItem

	stopChan chan bool

}

func(a *storeRedis) incr(ctx context.Context, id string, val int64)( err error) {
	item := &incrItem{id: id, delta: val}
	zsetKey := a.dispatch(ctx, item)
	incrKey := a.incrKey(item.id)
	// 先 incr
	if err = a.incrByKey(ctx, incrKey, item); err != nil {
		return
	}
	// 再zadd
	if err = a.recordInList(ctx, zsetKey, item.id);err != nil {
		return
	}

	return
}

func (a *storeRedis) incrKey(id string) string {
	return _redisstorekey + a.notChangeBussinessId + id
}

func (a *storeRedis) recordInList(ctx context.Context, key string, member string) (err error) {
	err = a.cache.ZAddNX(key, redis.Z{Score:float64( time.Now().Unix()), Member: member}).Err()
	return
}

func (a *storeRedis) incrByKey(ctx context.Context, key string, item *incrItem) (err error) {
	err = a.cache.IncrBy(key, item.delta).Err()
	a.cache.Expire(key, time.Hour * 12)
	return
}

func (a *storeRedis) zrangebyScore(key string, score int64, limit int64) []string{
	return a.cache.ZRangeByScore(key, redis.ZRangeBy{Min:"0", Max: fmt.Sprintf("%d", score),Count: limit}).Val()
}

func (a *storeRedis) zrem(key string, id string) int64{
	return a.cache.ZRem(key, id).Val()
}

func (a *storeRedis) getset(key string, set int) int64{
	ret := a.cache.GetSet(key, 0).Val()
	intVal, _ := strconv.ParseInt(ret, 10, 0)
	return intVal
}

func(a *storeRedis) stop(ctx context.Context) (err error) {
	close(a.stopChan)
	a.wait.Wait()
	return
}

func(a *storeRedis) start(c *Config)  {
	a.Config = c
	a.batchAggChan = make(chan aggItem, 100 *a.Config.getSaveConcurrency())

	interval := int(a.Config.getInterval())
	intervalPice := interval/a.Config.getConcurrency()
	for i := 0; i < a.Config.getSaveConcurrency(); i++ {
		zset := a.skipKey(i)
		a.skipListKeys = append(a.skipListKeys, zset)
		a.wait.Add(1)

		delayInterval := interval
		go func() {
			time.Sleep(time.Duration(delayInterval))
			a.aggregating(zset)
		}()

		interval -= intervalPice
	}

	// todo: 启动时应该有对应的任务清理旧数据。应对初始化参数中队列数量缩减的情况。
	return
}

func (a *storeRedis) dispatch(ctx context.Context, item *incrItem) (zsetKey string){
	sum := crc32.ChecksumIEEE([]byte(item.id))
	index := sum % uint32(len(a.skipListKeys))
	return a.skipListKeys[index]
}

func (a *storeRedis) batchAgg() chan aggItem{
	return a.batchAggChan
}

func (a *storeRedis) skipKey(index int) string {
	return fmt.Sprintf("%s:zset:%s:%v", _redisstorekey, a.notChangeBussinessId, index)
}

func (a *storeRedis) aggregating(zset string) {
	defer a.wait.Done()

	tiker := time.Tick(a.Config.getInterval())
	for {
		select{
		case <- a.stopChan:
			a.clear()
			return
		case <- tiker:
			ids := a.zrangebyScore(zset, time.Now().Unix() - 1, 500)
			aggs := make(aggItem)
			for _, key := range ids{
				if a.zrem(zset, key) > 0 {
					if delta := a.getset(a.incrKey(key), 0); delta != 0 {
						aggs[key] = delta
					}
				}
			}
			a.batchAggChan <- aggs
			// todo add metric
		}
	}
}

func (a *storeRedis) clear() {
	// todo: 此处需要记录下当前有哪些key，防治有数据忘记落在了 hashMapKeys里面
}