package aggregateIncrWrite

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"sync"
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
	return &storeRedis{cache: cache, notChangeBussinessId: notChangeBussinessId, wait: &sync.WaitGroup{}}
}

const key = "aggIncrWrite"

type storeRedis struct {
	Config
	notChangeBussinessId string

	cache *redis.Client

	skipListKeys []string
	hashMapKeys []string

	wait *sync.WaitGroup

}

func(a *storeRedis) incr(ctx context.Context, id string, val int64)( err error) {

	return
}

func(a *storeRedis) stop(ctx context.Context) (err error) {
	return
}

func(a *storeRedis) start(c *Config)  {
	for i := 0; i < a.Config.getSaveConcurrency(); i++ {
		zset := a.skipKey(i)
		a.skipListKeys = append(a.skipListKeys, zset)
		hash := a.hashKey(i)
		a.hashMapKeys = append(a.hashMapKeys, hash)
		a.wait.Add(1)
		go a.aggregating(zset, hash)
	}
	return
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

}