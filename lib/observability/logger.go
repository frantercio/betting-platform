package observability

import (
	"context"
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func New() *Logger {
	return &Logger{log.New(os.Stdout, "", log.LstdFlags)}
}

func (l *Logger) Info(ctx context.Context, msg string, kvs ...any) {
	l.Printf("[INFO] %s %v", msg, kvs)
}

func (l *Logger) Error(ctx context.Context, msg string, kvs ...any) {
	l.Printf("[ERROR] %s %v", msg, kvs)
}
