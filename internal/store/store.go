package store

import (
	"fmt"
	"sync"
	"time"
)

var (
	exprMutex sync.Mutex
	taskMutex sync.Mutex

	// Maps to store expressions and tasks
	expressions = make(map[string]*Expression)
	tasks       = make(map[string]*Task)

	// Map to store tasks by expression ID
	exprTasks = make(map[string][]*Task)
)

// Expression represents a mathematical expression
type Expression struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Status     string  `json:"status"`
	Result     float64 `json:"result,omitempty"`
	CreatedAt  time.Time
}

// Task represents an atomic calculation operation
type Task struct {
	ID            string  `json:"id"`
	ExpressionID  string  `json:"expression_id"`
	Arg1          string  `json:"arg1"`
	Arg2          string  `json:"arg2"`
	Operator      string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
	Result        float64 `json:"result,omitempty"`
	Ready         bool
	InProgress    bool
	Completed     bool
}

// NewExpression creates a new expression record
func NewExpression(exprText string) *Expression {
	exprMutex.Lock()
	defer exprMutex.Unlock()

	id := fmt.Sprintf("expr-%d", time.Now().UnixNano())

	expr := &Expression{
		ID:         id,
		Expression: exprText,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	expressions[id] = expr
	return expr
}

// GetExpression retrieves an expression by ID
func GetExpression(id string) (*Expression, bool) {
	exprMutex.Lock()
	defer exprMutex.Unlock()

	expr, found := expressions[id]
	return expr, found
}

// ListExpressions returns all expressions
func ListExpressions() []*Expression {
	exprMutex.Lock()
	defer exprMutex.Unlock()

	result := make([]*Expression, 0, len(expressions))
	for _, expr := range expressions {
		result = append(result, expr)
	}
	return result
}

// RegisterTasks associates tasks with an expression
func RegisterTasks(exprID string, tasksList []*Task) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	exprTasks[exprID] = tasksList

	for _, task := range tasksList {
		tasks[task.ID] = task
	}
}

// UpdateTasksReadiness updates the ready status of tasks
func UpdateTasksReadiness(exprID string) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	taskList, ok := exprTasks[exprID]
	if !ok {
		return
	}

	for _, task := range taskList {
		if !task.Completed && !task.InProgress {
			// Check if dependencies are resolved
			arg1Ready := !isTaskReference(task.Arg1) || isTaskCompleted(task.Arg1[5:])
			arg2Ready := !isTaskReference(task.Arg2) || isTaskCompleted(task.Arg2[5:])

			task.Ready = arg1Ready && arg2Ready
		}
	}
}

// GetReadyTask returns a task that is ready to be processed
func GetReadyTask() (*Task, bool) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	for _, task := range tasks {
		if task.Ready && !task.InProgress && !task.Completed {
			task.InProgress = true
			task.Ready = false
			return task, true
		}
	}

	return nil, false
}

// GetTask retrieves a task by ID
func GetTask(taskID string) (*Task, bool) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	task, exists := tasks[taskID]
	return task, exists
}

// CompleteTask marks a task as completed and updates dependent tasks
func CompleteTask(taskID string, result float64) error {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	// Find the task
	task, exists := tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Update task status
	task.Completed = true
	task.InProgress = false
	task.Result = result

	// Get expression ID and tasks
	exprID := task.ExpressionID
	taskList, ok := exprTasks[exprID]
	if !ok {
		return fmt.Errorf("expression tasks not found: %s", exprID)
	}

	// Check if all tasks are completed
	allCompleted := true
	var lastTask *Task

	for _, t := range taskList {
		if !t.Completed {
			allCompleted = false
			break
		}
		// Identify the last task (root of expression tree)
		if lastTask == nil || t.ID > lastTask.ID {
			lastTask = t
		}
	}

	// Update expression status if all tasks are completed
	exprMutex.Lock()
	defer exprMutex.Unlock()

	if expr, found := expressions[exprID]; found {
		if allCompleted && lastTask != nil {
			expr.Status = "done"
			expr.Result = lastTask.Result
		} else {
			// Update readiness of dependent tasks
			for _, t := range taskList {
				if !t.Completed && !t.InProgress {
					arg1Ready := !isTaskReference(t.Arg1) || isTaskCompleted(t.Arg1[5:])
					arg2Ready := !isTaskReference(t.Arg2) || isTaskCompleted(t.Arg2[5:])
					t.Ready = arg1Ready && arg2Ready
				}
			}
		}
	}

	return nil
}

// Helper functions
func isTaskReference(arg string) bool {
	return len(arg) > 5 && arg[:5] == "task:"
}

func isTaskCompleted(taskID string) bool {
	task, exists := tasks[taskID]
	return exists && task.Completed
}
