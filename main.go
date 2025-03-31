package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	errorLogFilePath string
	infoLogFilePath  string
	serverPort       string
	logFilePaths     []string
)

func main() {
	flag.StringVar(&errorLogFilePath, "error-log", "./error.log", "Path to the error log file")
	flag.StringVar(&infoLogFilePath, "info-log", "./info.log", "Path to the info log file")
	flag.StringVar(&serverPort, "port", "8080", "The port to run the server on")
	flag.Parse()

	// append to the log file paths
	logFilePaths = []string{errorLogFilePath, infoLogFilePath}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	r.Get("/grep", grepLogHandler)

	fmt.Printf("Starting server on port %s...\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), r); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func grepLogHandler(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	if pattern == "" {
		http.Error(w, "Pattern query parameter is required", http.StatusBadRequest)
		return
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		http.Error(w, "Invalid regex pattern", http.StatusBadRequest)
		return
	}

	// loop through the log file paths
	var matches []string
	for _, logFilePath := range logFilePaths {
		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Log file %s does not exist", logFilePath), http.StatusInternalServerError)
			return
		}

		file, err := os.Open(logFilePath)
		if err != nil {
			http.Error(w, "Failed to open log file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			if re.MatchString(line) {
				matches = append(matches, line)
			}
		}

		if err := scanner.Err(); err != nil {
			http.Error(w, "Error reading log file", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, match := range matches {
		fmt.Fprintln(w, match)
	}
}
