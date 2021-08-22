package aggregateIncrWrite

var emptyMetricer = &emptyMetric{}

type Metric interface {
	MetricIncrCount(delta int64)
	MetricBatchCount(delta int64)
}

type emptyMetric struct {

}

func (m *emptyMetric) MetricIncrCount(delta int64) {

}

func (m *emptyMetric) MetricBatchCount(delta int64) {

}
