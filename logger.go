package aggregateIncrWrite

import (
	"context"
	"fmt"
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
	fmt.Println(fmt.Sprintf("info: %s", val))
}
func (l *emptyLog) Error(val string){
	fmt.Println(fmt.Sprintf("error: %s", val))
}
func (l *emptyLog) Errorc(ctx context.Context, val string){

}
