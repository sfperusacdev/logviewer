package logviewer

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"io"
	"log/slog"
	"os"
)

type MemoryLogger interface {
	io.Writer
	ReadLines() ([]string, error)
	Close() error
}

type logStorage struct {
	sync.RWMutex
	logLines []string
}

func (l *logStorage) AppendLine(line string) {
	l.Lock()
	defer l.Unlock()
	l.logLines = append(l.logLines, line)
}

func (l *logStorage) GetLines() []string {
	l.RLock()
	defer l.RUnlock()
	var dst = make([]string, len(l.logLines))
	copy(dst, l.logLines)
	return dst
}

type sqliteLogger struct {
	file   *os.File
	memory *logStorage

	LogFilePath *string
}

type Option = func(*sqliteLogger)

func WithLogFile(filePath string) Option {
	return func(o *sqliteLogger) {
		o.LogFilePath = &filePath
	}
}

func NewMemoryLogger(options ...Option) (MemoryLogger, error) {
	var logger = &sqliteLogger{
		memory: &logStorage{logLines: make([]string, 0, 1000)},
	}
	for _, opt := range options {
		opt(logger)
	}

	if logger.LogFilePath != nil {
		file, err := os.OpenFile(*logger.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Println("ERROR: opening log file", logger.LogFilePath)
			return nil, err
		}
		var builder strings.Builder
		builder.WriteString(
			fmt.Sprintln(
				"-------------------------------",
				time.Now().Format(time.DateTime),
			),
		)
		file.WriteString(builder.String())
		logger.file = file
	}
	return logger, nil
}

func (l *sqliteLogger) Close() error {
	var err error
	if l.file != nil {
		err = errors.Join(err, l.file.Close())
	}
	return err
}

func (l *sqliteLogger) Write(p []byte) (int, error) {
	l.memory.AppendLine(string(p))
	if l.file != nil {
		if _, fileErr := l.file.Write(p); fileErr != nil {
			log.Println("ERROR: file log write", fileErr)
			return len(p), fileErr
		}
	}
	return len(p), nil
}

func (l *sqliteLogger) ReadLines() ([]string, error) {
	return l.memory.GetLines(), nil
}

func NewSlog(memory MemoryLogger) *slog.Logger {
	var writer = io.MultiWriter(memory, os.Stdout)
	var handler = slog.NewTextHandler(writer, nil)
	var logger = slog.New(handler)
	return logger
}

type Log struct {
	Type    string
	Message string
}
