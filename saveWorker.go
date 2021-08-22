package aggregateIncrWrite

type incrItem struct {
	id string
	delta int64
}

type aggItem map[string]int64