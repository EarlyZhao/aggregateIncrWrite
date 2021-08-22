package aggregateIncrWrite


type Metric interface {
	MetricIncr(delta int)
}

type emptyMetric struct {

}

func (m *emptyMetric) MetricIncr(delta int) {

}


func defaultMetric() Metric{
	return &emptyMetric{}
}