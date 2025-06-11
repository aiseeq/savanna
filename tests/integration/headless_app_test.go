package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHeadlessAppBasic проверяет базовую функциональность headless приложения
func TestHeadlessAppBasic(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	// Убедимся что бинарь существует
	binaryPath := filepath.Join(projectRoot, "bin", "savanna-headless")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Если бинаря нет, соберем его
		buildCmd := exec.Command("make", "build")
		buildCmd.Dir = projectRoot
		if err := buildCmd.Run(); err != nil {
			t.Fatalf("Не удалось собрать проект: %v", err)
		}
	}

	// Тест с коротким запуском (2 секунды)
	cmd := exec.Command(binaryPath, "-duration=2s", "-verbose=true")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Headless приложение завершилось с ошибкой: %v\nВывод: %s", err, output)
	}

	outputStr := string(output)
	t.Logf("Вывод headless приложения:\n%s", outputStr)

	// Проверяем ключевые элементы вывода
	expectedStrings := []string{
		"Запуск headless симуляции",
		"Параметры:",
		"Время | Зайцы | Волки",
		"Симуляция завершена",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Ожидаемая строка не найдена в выводе: '%s'", expected)
		}
	}
}

// TestHeadlessAppFlags проверяет различные флаги командной строки
func TestHeadlessAppFlags(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	binaryPath := filepath.Join(projectRoot, "bin", "savanna-headless")

	tests := []struct {
		name string
		args []string
	}{
		{"default params", []string{"-duration=1s"}},
		{"custom rabbits", []string{"-duration=1s", "-rabbits=50"}},
		{"custom wolves", []string{"-duration=1s", "-wolves=5"}},
		{"custom seed", []string{"-duration=1s", "-seed=123"}},
		{"verbose mode", []string{"-duration=1s", "-verbose=true"}},
		{"all custom", []string{"-duration=1s", "-rabbits=30", "-wolves=4", "-seed=42", "-verbose=true"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = projectRoot

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Команда с флагами %v завершилась с ошибкой: %v\nВывод: %s", tt.args, err, output)
			}

			outputStr := string(output)

			// Проверяем что приложение запустилось и завершилось корректно
			if !strings.Contains(outputStr, "Запуск headless симуляции") {
				t.Errorf("Приложение не запустилось корректно с флагами %v", tt.args)
			}

			if !strings.Contains(outputStr, "Симуляция завершена") {
				t.Errorf("Приложение не завершилось корректно с флагами %v", tt.args)
			}

			t.Logf("Тест '%s' прошел успешно", tt.name)
		})
	}
}

// TestHeadlessAppTimeout проверяет что приложение корректно завершается по таймауту
func TestHeadlessAppTimeout(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	binaryPath := filepath.Join(projectRoot, "bin", "savanna-headless")

	// Запускаем на 3 секунды и проверяем что оно завершится примерно через это время
	start := time.Now()

	cmd := exec.Command(binaryPath, "-duration=3s")
	cmd.Dir = projectRoot

	err = cmd.Run()
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Приложение завершилось с ошибкой: %v", err)
	}

	// Проверяем что время выполнения близко к ожидаемому (3 секунды ± 1 секунда)
	expectedDuration := 3 * time.Second
	tolerance := 2 * time.Second

	if elapsed < expectedDuration-tolerance || elapsed > expectedDuration+tolerance {
		t.Errorf("Время выполнения %v не соответствует ожидаемому %v (±%v)", elapsed, expectedDuration, tolerance)
	}

	t.Logf("Приложение корректно завершилось через %v", elapsed)
}
