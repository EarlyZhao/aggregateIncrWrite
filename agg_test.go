package aggregateIncrWrite

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
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

	Convey("test incr", t, func() {
		times := 300
		var total int64
		agg := New(
			&Config{concurrencyBuffer: 10, saveConcurrency: 10},
			SetOptionStore(NewLocalStore()),
			SetOptionSaveHandler(func(id string, aggIncr int64) error {
				fmt.Printf("id: %s, val: %d\n", id, aggIncr)
				atomic.AddInt64(&total, aggIncr)
				return nil
			}),
			SetOptionFailHandler(nil))

		for i := 0; i < times; i ++ {
			agg.Incr(ctx, fmt.Sprintf("%d", i / 10), 1)
			time.Sleep(10*time.Millisecond)
		}

		done := make(chan bool)
		go func() {
			for i := 0; i < times; i ++ {
				agg.Incr(ctx, fmt.Sprintf("%d", i / 10), 1)
			}
			done <- true
		}()

		agg.Stop(ctx)
		<- done
		So(total, ShouldEqual, times + times)
	})

}
