package calculator

import (
	"calc-service/internal/store"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"
)

// Глобальные переменные и генерация ID задачи
var taskCounter uint64

func generateTaskID() string {
	id := atomic.AddUint64(&taskCounter, 1)
	return "task-" + strconv.FormatUint(id, 10)
}

// Лексический анализ: токенизация
type token struct {
	value string
	type_ tokenType
}

type tokenType int

const (
	number tokenType = iota
	operator
	leftParen
	rightParen
)

func tokenize(expression string) ([]token, error) {
	var tokens []token
	var current strings.Builder
	parenCount := 0

	for i, ch := range expression {
		switch {
		case unicode.IsDigit(ch) || ch == '.':
			current.WriteRune(ch)
			if i == len(expression)-1 || !(unicode.IsDigit(rune(expression[i+1])) || expression[i+1] == '.') {
				tokens = append(tokens, token{current.String(), number})
				current.Reset()
			}
		case ch == '+' || ch == '-' || ch == '*' || ch == '/':
			if current.Len() > 0 {
				tokens = append(tokens, token{current.String(), number})
				current.Reset()
			}
			tokens = append(tokens, token{string(ch), operator})
		case ch == '(':
			tokens = append(tokens, token{"(", leftParen})
			parenCount++
		case ch == ')':
			if current.Len() > 0 {
				tokens = append(tokens, token{current.String(), number})
				current.Reset()
			}
			tokens = append(tokens, token{")", rightParen})
			parenCount--
			if parenCount < 0 {
				return nil, fmt.Errorf("unbalanced parentheses")
			}
		}
	}
	if parenCount != 0 {
		return nil, fmt.Errorf("unbalanced parentheses")
	}
	return tokens, nil
}

// Синтаксический анализ: построение дерева выражения
type Node struct {
	Value    string
	Left     *Node
	Right    *Node
	TaskID   string
	Priority int
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

func buildExpressionTree(tokens []token) (*Node, error) {
	var outputQueue []*Node
	var operatorStack []string

	for _, t := range tokens {
		switch t.type_ {
		case number:
			outputQueue = append(outputQueue, &Node{Value: t.value})
		case operator:
			for len(operatorStack) > 0 &&
				precedence(operatorStack[len(operatorStack)-1]) >= precedence(t.value) &&
				operatorStack[len(operatorStack)-1] != "(" {
				op := operatorStack[len(operatorStack)-1]
				operatorStack = operatorStack[:len(operatorStack)-1]
				if len(outputQueue) < 2 {
					return nil, fmt.Errorf("invalid expression")
				}
				right := outputQueue[len(outputQueue)-1]
				left := outputQueue[len(outputQueue)-2]
				outputQueue = outputQueue[:len(outputQueue)-2]
				outputQueue = append(outputQueue, &Node{
					Value:    op,
					Left:     left,
					Right:    right,
					Priority: precedence(op),
				})
			}
			operatorStack = append(operatorStack, t.value)
		case leftParen:
			operatorStack = append(operatorStack, t.value)
		case rightParen:
			for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {
				op := operatorStack[len(operatorStack)-1]
				operatorStack = operatorStack[:len(operatorStack)-1]
				if len(outputQueue) < 2 {
					return nil, fmt.Errorf("invalid expression")
				}
				right := outputQueue[len(outputQueue)-1]
				left := outputQueue[len(outputQueue)-2]
				outputQueue = outputQueue[:len(outputQueue)-2]
				outputQueue = append(outputQueue, &Node{
					Value:    op,
					Left:     left,
					Right:    right,
					Priority: precedence(op),
				})
			}
			if len(operatorStack) == 0 {
				return nil, fmt.Errorf("unbalanced parentheses")
			}
			operatorStack = operatorStack[:len(operatorStack)-1]
		}
	}

	for len(operatorStack) > 0 {
		op := operatorStack[len(operatorStack)-1]
		operatorStack = operatorStack[:len(operatorStack)-1]
		if len(outputQueue) < 2 {
			return nil, fmt.Errorf("invalid expression")
		}
		right := outputQueue[len(outputQueue)-1]
		left := outputQueue[len(outputQueue)-2]
		outputQueue = outputQueue[:len(outputQueue)-2]
		outputQueue = append(outputQueue, &Node{
			Value:    op,
			Left:     left,
			Right:    right,
			Priority: precedence(op),
		})
	}

	if len(outputQueue) != 1 {
		return nil, fmt.Errorf("invalid expression")
	}
	return outputQueue[0], nil
}

// Генерация задач на основе дерева выражения
func isOperator(op string) bool {
	return op == "+" || op == "-" || op == "*" || op == "/"
}

func getNodeReference(n *Node) string {
	if n == nil {
		return ""
	}
	if isOperator(n.Value) {
		return "task:" + n.TaskID
	}
	return n.Value
}

func getOperationTime(op string) int {
	var envVar string
	switch op {
	case "+":
		envVar = os.Getenv("TIME_ADDITION_MS")
	case "-":
		envVar = os.Getenv("TIME_SUBTRACTION_MS")
	case "*":
		envVar = os.Getenv("TIME_MULTIPLICATIONS_MS")
	case "/":
		envVar = os.Getenv("TIME_DIVISIONS_MS")
	default:
		return 0
	}
	t, err := strconv.Atoi(envVar)
	if err != nil {
		return 100
	}
	return t
}

func createTasksFromTree(exprID string, node *Node) []*store.Task {
	var tasks []*store.Task
	if node == nil {
		return tasks
	}
	// Обход в пост-ордера
	if node.Left != nil {
		tasks = append(tasks, createTasksFromTree(exprID, node.Left)...)
	}
	if node.Right != nil {
		tasks = append(tasks, createTasksFromTree(exprID, node.Right)...)
	}
	if isOperator(node.Value) {
		taskID := generateTaskID()
		node.TaskID = taskID
		arg1 := getNodeReference(node.Left)
		arg2 := getNodeReference(node.Right)
		task := &store.Task{
			ID:            taskID,
			ExpressionID:  exprID,
			Arg1:          arg1,
			Arg2:          arg2,
			Operator:      node.Value,
			OperationTime: getOperationTime(node.Value),
			Ready:         false,
			InProgress:    false,
			Completed:     false,
		}
		tasks = append(tasks, task)
	}
	return tasks
}

// Валидация выражения и основной процессинг
func ValidateExpression(expr string) error {
	valid := "0123456789.+-*/() "
	balance := 0
	for _, ch := range expr {
		if !strings.ContainsRune(valid, ch) && !unicode.IsSpace(ch) {
			return fmt.Errorf("invalid symbol in expression")
		}
		if ch == '(' {
			balance++
		}
		if ch == ')' {
			balance--
			if balance < 0 {
				return fmt.Errorf("unbalanced parentheses")
			}
		}
	}
	if balance != 0 {
		return fmt.Errorf("unbalanced parentheses")
	}
	return nil
}

func ProcessExpression(exprStr string) (*store.Expression, error) {
	// Очистка строки от пробелов
	exprStr = strings.ReplaceAll(exprStr, " ", "")
	if err := ValidateExpression(exprStr); err != nil {
		return nil, err
	}
	tokens, err := tokenize(exprStr)
	if err != nil {
		return nil, err
	}
	tree, err := buildExpressionTree(tokens)
	if err != nil {
		return nil, err
	}
	expr := store.NewExpression(exprStr)
	tasks := createTasksFromTree(expr.ID, tree)
	store.RegisterTasks(expr.ID, tasks)
	store.UpdateTasksReadiness(expr.ID)
	return expr, nil
}
