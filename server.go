package logviewer

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func getLogType(logLine string) string {
	if strings.Contains(logLine, "INFO") {
		return "INFO"
	} else if strings.Contains(logLine, "ERROR") {
		return "ERROR"
	} else if strings.Contains(logLine, "DEBUG") {
		return "DEBUG"
	} else if strings.Contains(logLine, "WARN") {
		return "WARN"
	}
	return "OTHER"
}

//go:embed index.html
var index string
var tmpl = template.Must(template.New("logs").Parse(index))

func StartServer(memory MemoryLogger, address string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lines, err := memory.ReadLines()
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}
		var logEntries []Log
		for _, logLine := range lines {
			logEntries = append(logEntries, Log{
				Type:    getLogType(logLine),
				Message: logLine,
			})
		}
		w.WriteHeader(http.StatusOK)
		tmpl.Execute(w, logEntries)
	})
	fmt.Println("log server is running:", address)
	return http.ListenAndServe(address, nil)
}
