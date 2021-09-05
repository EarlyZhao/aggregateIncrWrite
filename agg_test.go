package aggregateIncrWrite

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func Test_NewAgg(t *testing.T) {
	New(&Config{}, SetOptionStore(NewLocalStore()), SetOptionSaveHandler(func(id string, aggIncr int64) error {
		return nil
	}), SetOptionFailHandler(nil))
}

func Test_AggLocalIncr(t *testing.T) {

	ctx := context.TODO()

	Convey("test local incr", t, func() {
		times := 300000
		var total int64
		agg := New(
			&Config{ConcurrencyBuffer: 10, SaveConcurrency: 10},
			SetOptionStore(NewLocalStore()),
			SetOptionSaveHandler(func(id string, aggIncr int64) error {
				time.Sleep(time.Millisecond)
				fmt.Printf("local: id: %s, val: %d\n", id, aggIncr)
				atomic.AddInt64(&total, aggIncr)
				return nil
			}),
			SetOptionFailHandler(nil))

		for i := 0; i < times; i ++ {
			agg.Incr(ctx, fmt.Sprintf("%d", rand.Intn(times)), 1)
			time.Sleep(5000*time.Nanosecond)
		}

		done := make(chan bool)
		go func() {
			for i := 0; i < times; i ++ {
				agg.Incr(ctx, fmt.Sprintf("%d", rand.Intn(1000)), 1)
			}
			done <- true
			for i := 0; i < times/100; i ++ {
				agg.Incr(ctx, fmt.Sprintf("%d", rand.Intn(1000)), 1)
			}
		}()
		<- done
		agg.Stop(ctx)
		So(total, ShouldEqual, times + times + times/100)
	})

}


func Test_AggRedisIncr(t *testing.T) {

	ctx := context.TODO()

	Convey("test redis incr", t, func() {
		times := 300000
		var total int64
		agg := New(
			&Config{ConcurrencyBuffer: 10, SaveConcurrency: 100},
			SetOptionStore(NewRedisStore("test", redis.NewClient(&redis.Options{Addr:"127.0.0.1:6379"}))),
			SetOptionSaveHandler(func(id string, aggIncr int64) error {
				fmt.Printf("redis: id: %s, val: %d\n", id, aggIncr)
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&total, aggIncr)
				return nil
			}),
			SetOptionFailHandler(nil))

		for i := 0; i < times; i ++ {
			agg.Incr(ctx, fmt.Sprintf("%d", rand.Intn(times)), 1)
			//time.Sleep(55*time.Millisecond)
		}
		done := make(chan bool)
		go func() {
			for i := 0; i < times; i ++ {
				if i % 50 == 0 {
					go agg.Incr(ctx, fmt.Sprintf("%d", rand.Intn(1000)), 1)
					continue
				}
				agg.Incr(ctx, fmt.Sprintf("%d", rand.Intn(1000)), 1)
			}
			done <- true
		}()

		<- done
		time.Sleep(2*time.Second)
		agg.Stop(ctx)

		So(total, ShouldEqual, times + times)
	})

}