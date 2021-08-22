package aggregateIncrWrite

import "testing"

func Test_NewAgg(t *testing.T) {
	New(&Config{}, SetOptionStore(NewLocalStore()), SetOptionSaveHandler(func(id string, aggIncr int64) error {
		return nil
	}), SetOptionFailHandler(nil))
}
