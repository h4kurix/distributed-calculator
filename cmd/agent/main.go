package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID            string  `json:"id"`
	ExpressionID  string  `json:"expression_id"`
	Arg1          string  `json:"arg1"`
	Arg2          string  `json:"arg2"`
	Operator      string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
	Result        float64 `json:"result,omitempty"`
}

type TaskResponse struct {
	Task *Task `json:"task"`
}

const (
	maxRetries     = 5
	baseRetryDelay = 2 * time.Second
)

func main() {
	orchestratorHost := os.Getenv("ORCHESTRATOR_HOST")
	if orchestratorHost == "" {
		orchestratorHost = "localhost"
	}

	computingPower := getEnvAsInt("COMPUTING_POWER", 10)
	maxWorkers = computingPower
	log.Printf("Starting agent with %d workers", computingPower)

	for i := 0; i < computingPower; i++ {
		go worker(i+1, orchestratorHost)
	}
	select {}
}

func worker(workerID int, orchestratorHost string) {
	for {
		task, ok := fetchTask(orchestratorHost)
		if !ok {
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("Worker %d: Processing task %s (%s %s %s)",
			workerID, task.ID, task.Arg1, task.Operator, task.Arg2)

		result, err := processTask(task)
		if err != nil {
			log.Printf("Worker %d: Task %s failed: %v", workerID, task.ID, err)
			// Уменьшаем счетчик при ошибке обработки
			taskMutex.Lock()
			activeWorkers--
			taskMutex.Unlock()
			continue
		}

		if err := sendResult(orchestratorHost, task.ID, result); err != nil {
			log.Printf("Worker %d: Failed to send result: %v", workerID, err)
		} else {
			log.Printf("Worker %d: Task %s result %.2f sent", workerID, task.ID, result)
		}

		// Уменьшаем счетчик после успешной обработки
		taskMutex.Lock()
		activeWorkers--
		taskMutex.Unlock()
	}
}

var (
	taskMutex     sync.Mutex
	activeWorkers int
	maxWorkers    int
)

func fetchTask(orchestratorHost string) (*Task, bool) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	// Проверяем количество активных воркеров
	if activeWorkers >= maxWorkers {
		return nil, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("http://%s:8080/internal/task", orchestratorHost)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Network error: %v", err)
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Server error: %d", resp.StatusCode)
		return nil, false
	}

	var response TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Decoding error: %v", err)
		return nil, false
	}
	activeWorkers++
	return response.Task, response.Task != nil
}

func processTask(task *Task) (float64, error) {
	orchestratorHost := os.Getenv("ORCHESTRATOR_HOST")
	if orchestratorHost == "" {
		orchestratorHost = "orchestrator"
	}

	arg1, err := resolveArgument(orchestratorHost, task.Arg1)
	if err != nil {
		return 0, fmt.Errorf("argument 1: %w", err)
	}
	arg2, err := resolveArgument(orchestratorHost, task.Arg2)
	if err != nil {
		return 0, fmt.Errorf("argument 2: %w", err)
	}
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	return calculate(task.Operator, arg1, arg2)
}

func resolveArgument(orchestratorHost, arg string) (float64, error) {
	if strings.HasPrefix(arg, "task:") {
		taskID := strings.TrimPrefix(arg, "task:")
		return fetchTaskResultWithRetry(orchestratorHost, taskID)
	}
	return strconv.ParseFloat(arg, 64)
}

func fetchTaskResult(orchestratorHost, taskID string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("http://%s:8080/api/v1/tasks/%s", orchestratorHost, taskID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}
	log.Printf("Task %s result response: %s", taskID, string(body))
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		Result float64 `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("decoding error: %w, body: %s", err, string(body))
	}
	return result.Result, nil
}

func fetchTaskResultWithRetry(orchestratorHost, taskID string) (float64, error) {
	for i := 0; i < maxRetries; i++ {
		result, err := fetchTaskResult(orchestratorHost, taskID)
		if err == nil {
			return result, nil
		}
		delay := time.Duration(i+1) * baseRetryDelay
		log.Printf("Retry %d/%d for task %s (delay: %v)", i+1, maxRetries, taskID, delay)
		time.Sleep(delay)
	}
	return 0, fmt.Errorf("max retries exceeded for task %s", taskID)
}

func calculate(operator string, a, b float64) (float64, error) {
	switch operator {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", operator)
	}
}

func sendResult(orchestratorHost string, taskID string, result float64) error {
	payload := struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}{
		ID:     taskID,
		Result: result,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	url := fmt.Sprintf("http://%s:8080/internal/task", orchestratorHost)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("POST error: %v", err)
		return fmt.Errorf("post error: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("Response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
func getEnvAsInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Invalid value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	return result
}
