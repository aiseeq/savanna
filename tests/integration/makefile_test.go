package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMakefileCommands проверяет основные команды Makefile
func TestMakefileCommands(t *testing.T) {
	// Получаем корневую папку проекта
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	// Переходим в корневую папку
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Не удалось получить текущую папку: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Не удалось вернуться в исходную папку: %v", err)
		}
	}()

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Не удалось перейти в корневую папку: %v", err)
	}

	tests := []struct {
		name        string
		command     []string
		expectError bool
		timeout     time.Duration
	}{
		{"make help", []string{"make", "help"}, false, 10 * time.Second},
		{"make clean", []string{"make", "clean"}, false, 10 * time.Second},
		{"make fmt", []string{"make", "fmt"}, false, 30 * time.Second},
		{"make build", []string{"make", "build"}, false, 60 * time.Second},
		{"make test-unit", []string{"make", "test-unit"}, false, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command[0], tt.command[1:]...)
			cmd.Dir = projectRoot

			// Устанавливаем timeout
			if tt.timeout > 0 {
				go func() {
					time.Sleep(tt.timeout)
					if cmd.Process != nil {
						cmd.Process.Kill()
					}
				}()
			}

			output, err := cmd.CombinedOutput()

			if tt.expectError && err == nil {
				t.Errorf("Ожидалась ошибка для команды %v, но команда выполнилась успешно", tt.command)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Команда %v не удалась: %v\nВывод: %s", tt.command, err, output)
			}

			t.Logf("Вывод команды %v:\n%s", tt.command, output)
		})
	}
}

// TestMakeBuild проверяет что make build создает бинари
func TestMakeBuild(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Не удалось получить текущую папку: %v", err)
	}
	defer func() {
		os.Chdir(originalDir)
	}()

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Не удалось перейти в корневую папку: %v", err)
	}

	// Очищаем перед сборкой
	exec.Command("make", "clean").Run()

	// Собираем
	cmd := exec.Command("make", "build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make build не удался: %v\nВывод: %s", err, output)
	}

	// Проверяем что бинари созданы
	binaries := []string{
		"bin/savanna-game",
		"bin/savanna-headless",
	}

	for _, binary := range binaries {
		if _, err := os.Stat(binary); os.IsNotExist(err) {
			t.Errorf("Бинарь не создан: %s", binary)
		} else {
			t.Logf("Бинарь успешно создан: %s", binary)
		}
	}
}

// TestMakeHelp проверяет что make help показывает все команды
func TestMakeHelp(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	cmd := exec.Command("make", "help")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make help не удался: %v", err)
	}

	helpOutput := string(output)

	// Проверяем что все основные команды присутствуют в help
	expectedCommands := []string{
		"build",
		"run",
		"run-headless",
		"test",
		"fmt",
		"clean",
		"help",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(helpOutput, cmd) {
			t.Errorf("Команда '%s' не найдена в выводе make help", cmd)
		}
	}

	t.Logf("Вывод make help:\n%s", helpOutput)
}
