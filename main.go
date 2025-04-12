package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	errorLogFilePath         string
	infoLogFilePath          string
	serverPort               string
	authorizationHeaderToken string
	logFilePaths             map[string]string
)

func main() {
	flag.StringVar(&errorLogFilePath, "error-log", "./error.log", "Path to the error log file")
	flag.StringVar(&infoLogFilePath, "info-log", "./info.log", "Path to the info log file")
	flag.StringVar(&serverPort, "port", "8080", "The port to run the server on")
	flag.StringVar(&authorizationHeaderToken, "auth-header-token", "", "Authorization header token for the server")

	listCustomLogFiles := make(stringMapFlag)
	flag.Var(&listCustomLogFiles, "log", "Set key=value (can be used multiple times)")

	flag.Parse()

	// append to the log file paths
	logFilePaths = make(map[string]string)
	logFilePaths["error"] = errorLogFilePath
	logFilePaths["info"] = infoLogFilePath
	for logType, logFilePath := range listCustomLogFiles {
		logFilePaths[strings.ToLower(logType)] = logFilePath
	}
	fmt.Println(logFilePaths)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Minute)) // 1 min timeout
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/grep", grepLogHandler)
	})

	fmt.Printf("Starting server on port %s...\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), r); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authorizationHeaderToken != "" {
			tokenHeaderVal := r.Header.Get("Authorization")

			// Extract token from "Bearer <token>"
			tokenStr := strings.TrimPrefix(tokenHeaderVal, "Bearer ")
			if tokenStr == "" {
				http.Error(w, "invalid token format", http.StatusForbidden)
				return
			}

			// Check if the token matches the expected token
			if tokenStr != authorizationHeaderToken {
				http.Error(w, "invalid token", http.StatusForbidden)
				return
			}
		}

		// Token is valid, proceed with the request
		next.ServeHTTP(w, r)
	})
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
	matches := make(map[string][]string)
	for logType, logFilePath := range logFilePaths {
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

		var tempMatcher []string
		for scanner.Scan() {
			line := scanner.Text()
			if re.MatchString(line) {
				tempMatcher = append(tempMatcher, line)
			}
		}

		if err := scanner.Err(); err != nil {
			http.Error(w, "Error reading log file", http.StatusInternalServerError)
			return
		}

		matches[logType] = tempMatcher
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Found %d matches:\n", len(matches))
	if len(matches) > 0 {
		for logType, logMatches := range matches {
			fmt.Fprintln(w, logType+":")
			for _, match := range logMatches {
				fmt.Fprintln(w, match)
			}
		}
	}
}

// Custom map type
type stringMapFlag map[string]string

// Implement the flag.Value interface
func (m *stringMapFlag) String() string {
	var pairs []string
	for k, v := range *m {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}

func (m *stringMapFlag) Set(input string) error {
	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: expected key=value, got %s", input)
	}
	key := parts[0]
	value := parts[1]
	(*m)[key] = value
	return nil
}
