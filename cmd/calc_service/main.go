package main

import (
	"calc-service/internal/handler"
	"calc-service/pkg/logger"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize logger
	initLogger()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		logger.Info("Warning: .env file not found or error loading: %v", err)
	}

	// Set default operation times if not in env
	setDefaultEnvIfEmpty("TIME_ADDITION_MS", "100")
	setDefaultEnvIfEmpty("TIME_SUBTRACTION_MS", "100")
	setDefaultEnvIfEmpty("TIME_MULTIPLICATIONS_MS", "200")
	setDefaultEnvIfEmpty("TIME_DIVISIONS_MS", "300")

	// API for user
	http.HandleFunc("/api/v1/calculate", handler.HandleCalculate)
	http.HandleFunc("/api/v1/expressions", handler.HandleExpressions)
	http.HandleFunc("/api/v1/expressions/", handler.HandleExpressionByID)

	// Internal API for agents
	http.HandleFunc("/internal/task", handler.TaskHandler)
	http.HandleFunc("/api/v1/tasks/", handler.HandleTaskByID)

	// Frontend
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Read port from env, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start background task to update task readiness periodically
	go updateTasksReadinessPeriodically()

	logger.Info("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Initialize logger or use fallback if unavailable
func initLogger() {
	// Dummy initialization to handle the case if logger.Init is not defined
	defer func() {
		if r := recover(); r != nil {
			// If logger.Init panics, fall back to default logger
			log.Println("Warning: Logger initialization failed, using default logger")
		}
	}()

	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logger.Init(logLevel)
}

func setDefaultEnvIfEmpty(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

func updateTasksReadinessPeriodically() {
	// Periodically check task readiness
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update task readiness for all expressions
			handler.UpdateAllTasksReadiness()
		}
	}
}
