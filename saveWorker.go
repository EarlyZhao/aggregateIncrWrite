package aggregateIncrWrite

type incrItem struct {
	id string
	delta int64
}

type aggItem map[string]int64


func (a *Aggregate ) saveWorker() {
	defer a.wait.Done()

	for {
		select {
		case item, ok := <- a.store.batchAgg():
			if ok {
				a.saveBatch(item)
			}
		case <- a.stoped:
			// 打扫战场
			select {
			case item, ok := <- a.store.batchAgg():
				if !ok{
					return
				}
				a.saveBatch(item)
			default:
				return
			}
		}
	}
}

func (a *Aggregate) saveBatch(batch aggItem) {
	for id, delta := range batch{
		if err := a.saveHandler(id, delta); err != nil {
			a.failureHandler(id, delta)
		}
	}
}