package handler

import (
	"calc-service/internal/calculator"
	"calc-service/internal/store"
	"calc-service/pkg/logger"
	"encoding/json"
	"net/http"
	"strings"
)

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ID string `json:"id"`
}

type ExpressionsResponse struct {
	Expressions []ExpressionResponse `json:"expressions"`
}

type ExpressionResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result,omitempty"`
}

type ExpressionDetailResponse struct {
	Expression ExpressionResponse `json:"expression"`
}

func HandleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	expr, err := calculator.ProcessExpression(req.Expression)
	if err != nil {
		logger.Error("Expression processing error: %v", err)
		http.Error(w, "Invalid expression", http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CalculateResponse{ID: expr.ID})
}

func HandleExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expressions := store.ListExpressions()
	response := make([]ExpressionResponse, 0, len(expressions))

	for _, expr := range expressions {
		response = append(response, ExpressionResponse{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ExpressionsResponse{Expressions: response})
}

func HandleExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	expr, exists := store.GetExpression(id)
	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	response := ExpressionResponse{
		ID:     expr.ID,
		Status: expr.Status,
		Result: expr.Result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ExpressionDetailResponse{Expression: response})
}

func UpdateAllTasksReadiness() {
	expressions := store.ListExpressions()
	for _, expr := range expressions {
		if expr.Status == "pending" {
			store.UpdateTasksReadiness(expr.ID)
		}
	}
}
