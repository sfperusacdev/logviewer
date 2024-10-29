package logviewer

import (
	"database/sql"

	"fmt"
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
	db *sql.DB
}

func NewMemoryLogger() (MemoryLogger, error) {
	var sqlitePath = ":memory:"
	db, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("create table logs(value text)"); err != nil {
		slog.Error("creating log table", "error", err)
		return nil, err
	}

	return &sqliteLogger{db: db}, nil
}

func (l *sqliteLogger) Close() error { return l.db.Close() }

func (l *sqliteLogger) Write(p []byte) (int, error) {
	const qry = "insert into logs(value) values ('%s')"
	var _, err = l.db.Exec(fmt.Sprintf(qry, string(p)))
	return len(p), err
}

func (l *sqliteLogger) ReadLines() ([]string, error) {
	const qry = "select * from logs"
	rows, err := l.db.Query(qry)
	if err != nil {
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
