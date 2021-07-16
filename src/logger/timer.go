package logger

import (
	"fmt"
	"time"
)

type Timer struct {
	Name  string
	Start int
}

func GetTimeMs() int {
	return int(time.Now().UnixNano() / int64(time.Millisecond))
}

func StartTimer(name string) Timer {

	return Timer{name, GetTimeMs()}

}

func (t Timer) LogTimeTaken() {

	fmt.Printf("%s process took %d ms.\n", t.Name, GetTimeMs()-t.Start)

}
