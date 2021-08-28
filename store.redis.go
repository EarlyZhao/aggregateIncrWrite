package aggregateIncrWrite

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"hash/crc32"
	"math/rand"
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

const key = "aggIncrWrite"

type storeRedis struct {
	Config
	notChangeBussinessId string

	cache *redis.Client

	skipListKeys []string
	hashMapKeys []string

	wait *sync.WaitGroup

	batchAggChan chan aggItem

	stopChan chan bool

}

func(a *storeRedis) incr(ctx context.Context, id string, val int64)( err error) {
	item := &incrItem{id: id, delta: val}
	zsetKey, hashKey := a.dispatch(ctx, item)
	// 先 incr
	if err = a.incrHash(ctx, hashKey, item); err != nil {
		return
	}
	// 再zadd
	if err = a.recordInList(ctx, zsetKey, item);err != nil {
		return
	}

	return
}

func (a *storeRedis) recordInList(ctx context.Context, key string, item *incrItem) (err error) {
	err = a.cache.ZAddNX(key, redis.Z{Score:float64( time.Now().Unix()), Member:key}).Err()
	return
}

func (a *storeRedis) incrHash(ctx context.Context, key string, item *incrItem) (err error) {
	err = a.cache.HIncrBy(key, item.id, item.delta).Err()
	return
}

func (a *storeRedis) zrangebyScore(key string, score int64, limit int64) []string{
	return a.cache.ZRangeByScore(key, redis.ZRangeBy{Min:"0", Max: fmt.Sprintf("%d", score),Count: limit}).Val()
}

func (a *storeRedis) zrem(key string, id string) int64{
	return a.cache.ZRem(key, id).Val()
}

func (a *storeRedis) hdel(key string, id string) int64{
	return a.cache.HDel(key, id).Val()
}

func(a *storeRedis) stop(ctx context.Context) (err error) {
	close(a.stopChan)
	a.wait.Wait()
	return
}

func(a *storeRedis) start(c *Config)  {
	a.batchAggChan = make(chan aggItem, 100 *a.Config.getSaveConcurrency())
	for i := 0; i < a.Config.getSaveConcurrency(); i++ {
		zset := a.skipKey(i)
		a.skipListKeys = append(a.skipListKeys, zset)
		hash := a.hashKey(i)
		a.hashMapKeys = append(a.hashMapKeys, hash)
		a.wait.Add(1)
		go a.aggregating(zset, hash)
	}

	// todo: 启动时应该有对应的任务清理旧数据。应对初始化参数中队列数量缩减的情况。
	return
}

func (a *storeRedis) dispatch(ctx context.Context, item *incrItem) (zsetKey, hashKey string){
	sum := crc32.ChecksumIEEE([]byte(item.id))
	index := sum % uint32(len(a.hashMapKeys))
	return a.skipListKeys[index], a.hashMapKeys[index]
}

func (a *storeRedis) batchAgg() chan aggItem{
	return nil
}

func (a *storeRedis) skipKey(index int) string {
	return fmt.Sprintf("%s:zset:%s:%v", key, a.notChangeBussinessId, index)
}

func (a *storeRedis) hashKey(index int) string {
	return fmt.Sprintf("%s:hash:%s:%v", key, a.notChangeBussinessId, index)
}

func (a *storeRedis) aggregating(zset, hash string) {
	defer a.wait.Done()
	// 让tick的启动时间分散
	time.Sleep(time.Duration(rand.Intn(int(a.Config.getInterval()))))
	tiker := time.Tick(a.Config.getInterval())
	for {
		select{
		case <- a.stopChan:
			a.clear()
			return
		case <- tiker:
			ids := a.zrangebyScore(zset, time.Now().Unix() - 1, 400)
			aggs := make(aggItem)
			for _, key := range ids{
				if a.zrem(zset, key) > 0 {
					if delta := a.hdel(hash, key); delta != 0 {
						aggs[key] = delta
					}
				}
			}
			a.batchAggChan <- aggs
		}
	}
}

func (a *storeRedis) clear() {
	// todo: 此处需要记录下当前有哪些key，防治有数据忘记落在了 hashMapKeys里面
}