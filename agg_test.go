package aggregateIncrWrite

import "testing"

func Test_NewAgg(t *testing.T) {
	New(NewLocalStore(), &Config{}, nil, nil)
}
