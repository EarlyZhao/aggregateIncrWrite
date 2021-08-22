package aggregateIncrWrite

import (
	"context"
)

var emptyLogger = &emptyLog{}

type Logger interface {
	Infoc(ctx context.Context, val string)
	Info(val string)
	Error(val string)
	Errorc(ctx context.Context, val string)
}

type emptyLog struct {

}

func (l *emptyLog) Infoc(ctx context.Context, val string){

}
func (l *emptyLog) Info(val string){

}
func (l *emptyLog) Error(val string){

}
func (l *emptyLog) Errorc(ctx context.Context, val string){

}
