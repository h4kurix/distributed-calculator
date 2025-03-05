package calculator

import (
	"os"
	"testing"
)

func TestGenerateTaskID(t *testing.T) {
	id1 := generateTaskID()
	id2 := generateTaskID()
	if id1 == id2 {
		t.Errorf("expected different IDs, got same: %s", id1)
	}
}

func TestValidateExpression(t *testing.T) {
	valid := []string{
		"1+2",
		"3.14*2",
		"(1+2)*3",
	}
	for _, expr := range valid {
		if err := ValidateExpression(expr); err != nil {
			t.Errorf("expected valid expression, got error: %v for expr: %s", err, expr)
		}
	}

	invalid := []string{
		"1+2a",
		"1+2$",
		"1+(2*3",
	}
	for _, expr := range invalid {
		if err := ValidateExpression(expr); err == nil {
			t.Errorf("expected error for invalid expr: %s", expr)
		}
	}
}

func TestProcessExpression_Valid(t *testing.T) {
	// Устанавливаем переменные окружения для времени операций
	os.Setenv("TIME_ADDITION_MS", "120")
	os.Setenv("TIME_SUBTRACTION_MS", "130")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "140")
	os.Setenv("TIME_DIVISIONS_MS", "150")

	expr, err := ProcessExpression("1+2")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if expr == nil {
		t.Error("expected non-nil expression")
	}
}

func TestProcessExpression_Invalid(t *testing.T) {
	_, err := ProcessExpression("1+2a")
	if err == nil {
		t.Error("expected error for invalid expression")
	}
}

func TestGetOperationTime(t *testing.T) {
	os.Setenv("TIME_ADDITION_MS", "200")
	if opTime := getOperationTime("+"); opTime != 200 {
		t.Errorf("expected 200, got %d", opTime)
	}

	os.Setenv("TIME_SUBTRACTION_MS", "210")
	if opTime := getOperationTime("-"); opTime != 210 {
		t.Errorf("expected 210, got %d", opTime)
	}

	os.Setenv("TIME_MULTIPLICATIONS_MS", "220")
	if opTime := getOperationTime("*"); opTime != 220 {
		t.Errorf("expected 220, got %d", opTime)
	}

	os.Setenv("TIME_DIVISIONS_MS", "230")
	if opTime := getOperationTime("/"); opTime != 230 {
		t.Errorf("expected 230, got %d", opTime)
	}
}
