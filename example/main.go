package main

import (
	"log"

	"github.com/sfperusacdev/logviewer"
)

func main() {
	memory, err := logviewer.NewMemoryLogger(logviewer.WithLogFile("./logfile.txt"))
	if err != nil {
		log.Fatalln(err)
	}
	defer memory.Close()
	var logger = logviewer.NewSlog(memory)

	logger.Info("Starting the application...")
	logger.Info("Initializing configuration", "module", "config", "status", "pending")
	logger.Info("Configuration loaded", "module", "config", "status", "success")
	logger.Warn("Low disk space detected", "module", "system", "availableMB", 512)
	logger.Error("Failed to connect to database", "module", "database", "error", "timeout")
	logger.Info("Retrying database connection", "module", "database", "attempt", 2)
	logger.Info("Application started successfully")
	logger.Warn("Remember to monitor system health")

	logviewer.StartServer(memory, ":14009")
}
