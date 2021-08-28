# Aggregate increase Write
## 背景
数据库中有大量的增量写操作，常见于点赞类incr数值操作，SQL语句类似：
```SQL
update xxtable set xxcolums = xxcolumn + 1 where id = xxx ;

```
## 问题
此等写操作偶尔会很频繁，如果id分散，则写操作的压力，一方面异步化，另一方面可以分库解决。

但是当在类似点赞、增加积分等业务场景下，可能会集中给某个id执行incr操作。且在业务上来源可能分散，难以在
业务源头对incr类操作进行聚合。

此时上面的集中的incr操作，很容易因为写冲突出现性能瓶颈，且该压力难以分散。往往可能阻碍业务发展，且大量浪费数据库资源。

## 解决方案
由于incr操作的独特性，如果能有种方式将某个id的incr操作聚合起来，以一定时间窗口或数量做合并更新，则写冲突和性能压力都会引刃而解。

## 实现
本项目就是上面解决方案的实现。 暂时提供两种机制：
- 本地聚合式写聚合，基于本地内存
- 分布式写聚合，基于分布式存储

可以根据实际场景选择合适的方案。

## 使用
#### 本地聚合
```golang
    // 程序启动： 初始化聚合对象
    agg := New(
	    &Config{ConcurrencyBuffer: 10, SaveConcurrency: 10},
            SetOptionStore(NewLocalStore()),
            SetOptionSaveHandler(func(id string, aggIncr int64) error {
                // 回调函数： 将聚合后的值写入数据库
                return nil
            }),
            SetOptionFailHandler(func(id string, aggIncr int64) error {
                // 失败回调函数
                return nil
        }))
        
	// incr操作
    agg.Incr(ctx, xxid, 2)
    
    // 程序退出
    agg.Stop(ctx)
```

#### redis聚合
```golang
    // 程序启动： 初始化聚合对象
    agg := New(
		&Config{ConcurrencyBuffer: 10, SaveConcurrency: 100},
		    SetOptionStore(NewRedisStore("someIdNotChange", redis.NewClient(&redis.Options{Addr:"127.0.0.1:6379"}))),
		    SetOptionSaveHandler(func(id string, aggIncr int64) error {
                // 回调函数
				return nil
			}),
			SetOptionFailHandler(nil),
		)
    
    // incr操作
    agg.Incr(ctx, xxid, 1)
    
    // 程序退出
    agg.Stop(ctx)
```