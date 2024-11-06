package logviewer

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"io"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

type MemoryLogger interface {
	io.Writer
	ReadLines() ([]string, error)
	Close() error
}

type sqliteLogger struct {
	db          *sql.DB
	file        *os.File
	DbPath      string
	LogFilePath *string
}

type Option = func(*sqliteLogger)

func WithLogDbPath(dbpath string) Option {
	return func(o *sqliteLogger) {
		o.DbPath = dbpath
	}
}

func WithLogFile(filePath string) Option {
	return func(o *sqliteLogger) {
		o.LogFilePath = &filePath
	}
}

func NewMemoryLogger(options ...Option) (MemoryLogger, error) {
	var logger = &sqliteLogger{
		DbPath: ":memory:",
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

	db, err := sql.Open("sqlite", logger.DbPath)
	if err != nil {
		slog.Error("sqlite log db", "error", err)
		return nil, err
	}
	if _, err := db.Exec("create table logs(value text)"); err != nil {
		slog.Error("creating log table", "error", err)
		return nil, err
	}
	logger.db = db
	return logger, nil
}

func (l *sqliteLogger) Close() error {
	var err error
	if l.db != nil {
		err = errors.Join(err, l.db.Close())
	}
	if l.file != nil {
		err = errors.Join(err, l.file.Close())
	}
	return err
}

func (l *sqliteLogger) Write(p []byte) (int, error) {
	const qry = "INSERT INTO logs(value) VALUES (?)"
	_, err := l.db.Exec(qry, string(p))
	if err != nil {
		log.Println("Insert log ERROR:", err)
	}
	if l.file != nil {
		if _, fileErr := l.file.Write(p); fileErr != nil {
			log.Println("ERROR: file log write", err)
		}
	}
	return len(p), err
}

func (l *sqliteLogger) ReadLines() ([]string, error) {
	const qry = "select * from logs"
	rows, err := l.db.Query(qry)
	if err != nil {
		log.Println("select logs ERROR", err)
		return nil, err
	}
	defer rows.Close()
	var values []string
	for rows.Next() {
		var value sql.NullString
		rows.Scan(&value)
		values = append(values, value.String)
	}
	return values, nil
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
