package internal

import (
	"io"
	"log"
	"os"
)

// SetupLogging configures the server to write logs to both the terminal and a file.
// It returns the log file pointer so the caller can close it when the server shuts down.
func SetupLogging() *os.File {
	logFile, err := os.OpenFile("server_logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return logFile
}