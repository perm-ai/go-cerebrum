package logger

import (
	"fmt"
	"sync"
	"time"
)

type Logger struct {
	enabled bool
}

func NewLogger(enabled bool) Logger {
	return Logger{enabled: enabled}
}

func (l Logger) Log(s string) {
	if l.enabled {
		fmt.Print(time.Now().Format("2006-01-02T15:04:05-0700"))
		fmt.Print("\t" + s + "\n")
	}
}

type OperationsCounter struct {
	name 	string
	current int
	total   int
	mu		sync.Mutex
}

func NewOperationsCounter(name string, total int) OperationsCounter {

	return OperationsCounter{name: name, current: 0, total: total}

}

func (o *OperationsCounter) Increment(){
	o.mu.Lock()
	o.current++
	o.mu.Unlock()
	fmt.Printf("\r%s %d/%d", o.name, o.current, o.total)
	if o.current == o.total{
		fmt.Printf("\n")
	}
}