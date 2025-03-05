package handler

import (
	"calc-service/internal/store"
	"calc-service/pkg/logger"
	"encoding/json"
	"net/http"
	"strings"
)

type TaskResponse struct {
	Task *store.Task `json:"task"`
}

type TaskResultRequest struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetTask(w, r)
	case http.MethodPost:
		handlePostTaskResult(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	task, found := store.GetReadyTask()
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := TaskResponse{Task: task}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode task: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func handlePostTaskResult(w http.ResponseWriter, r *http.Request) {
	var req TaskResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode task result: %v", err)
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	if _, exists := store.GetTask(req.ID); !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if err := store.CompleteTask(req.ID, req.Result); err != nil {
		logger.Error("Failed to complete task: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	task, exists := store.GetTask(id)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if !task.Completed {
		http.Error(w, "Task not completed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Result float64 `json:"result"`
	}{Result: task.Result})
}