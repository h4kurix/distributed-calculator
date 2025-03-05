package store

import (
	"testing"
)

func TestNewExpression(t *testing.T) {
	expr := NewExpression("3 + 5")

	if expr.Expression != "3 + 5" {
		t.Errorf("ожидалось '3 + 5', получено %s", expr.Expression)
	}

	if expr.Status != "pending" {
		t.Errorf("ожидалось 'pending', получено %s", expr.Status)
	}
}

func TestGetExpression(t *testing.T) {
	expr := NewExpression("2 * 2")
	res, found := GetExpression(expr.ID)

	if !found {
		t.Errorf("выражение не найдено")
	}

	if res.ID != expr.ID {
		t.Errorf("ID не совпадает")
	}
}

func TestRegisterTasksAndGetTask(t *testing.T) {
	expr := NewExpression("1 + 1")
	task := &Task{
		ID:           "task-1",
		ExpressionID: expr.ID,
		Arg1:         "1",
		Arg2:         "1",
		Operator:     "+",
	}

	RegisterTasks(expr.ID, []*Task{task})
	res, found := GetTask("task-1")

	if !found || res.ID != "task-1" {
		t.Errorf("задача не зарегистрирована")
	}
}

func TestUpdateTasksReadiness(t *testing.T) {
	expr := NewExpression("1 + 1")
	task := &Task{
		ID:           "task-1",
		ExpressionID: expr.ID,
		Arg1:         "1",
		Arg2:         "1",
		Operator:     "+",
	}

	RegisterTasks(expr.ID, []*Task{task})
	UpdateTasksReadiness(expr.ID)

	if !task.Ready {
		t.Errorf("задача должна быть готова")
	}
}

func TestCompleteTask(t *testing.T) {
	expr := NewExpression("3 + 2")
	task := &Task{
		ID:           "task-1",
		ExpressionID: expr.ID,
		Arg1:         "3",
		Arg2:         "2",
		Operator:     "+",
	}

	RegisterTasks(expr.ID, []*Task{task})
	err := CompleteTask("task-1", 5)

	if err != nil {
		t.Errorf("ошибка завершения задачи: %v", err)
	}

	if !task.Completed || task.Result != 5 {
		t.Errorf("задача не завершена корректно")
	}
}
