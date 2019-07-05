package log

import "fmt"

func NewLogger() *Logger {
	return &Logger{}
}

type Logger struct {
	Message string
}

func (me *Logger) Debug(msg string) {
	fmt.Printf("[DEBUG] %s\n", msg)
}
