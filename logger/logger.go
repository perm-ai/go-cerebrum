package logger

import (
	"fmt"
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
