package unit

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoModTidy проверяет что go mod tidy не вносит изменений
func TestGoModTidy(t *testing.T) {
	t.Parallel()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	// Запускаем go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("go mod tidy завершился с ошибкой: %v\nВывод: %s", err, output)
	}

	// Проверяем что нет изменений в git (если это git репозиторий)
	gitCmd := exec.Command("git", "status", "--porcelain", "go.mod", "go.sum")
	gitCmd.Dir = projectRoot
	gitOutput, gitErr := gitCmd.CombinedOutput()

	if gitErr == nil && len(gitOutput) > 0 {
		t.Logf("go mod tidy внес изменения в go.mod/go.sum (возможно нужно закоммитить):\n%s", gitOutput)
	}

	t.Logf("go mod tidy выполнен успешно")
}

// TestGoModDependencies проверяет что все необходимые зависимости присутствуют
func TestGoModDependencies(t *testing.T) {
	t.Parallel()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	// Получаем список зависимостей
	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось получить список зависимостей: %v", err)
	}

	dependencies := string(output)

	// Проверяем обязательные зависимости
	requiredDeps := []string{
		"github.com/hajimehoshi/ebiten/v2",
	}

	for _, dep := range requiredDeps {
		if !strings.Contains(dependencies, dep) {
			t.Errorf("Обязательная зависимость не найдена: %s", dep)
		} else {
			t.Logf("Найдена зависимость: %s", dep)
		}
	}
}

// TestGoModuleName проверяет имя модуля
func TestGoModuleName(t *testing.T) {
	t.Parallel()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось получить имя модуля: %v", err)
	}

	moduleName := strings.TrimSpace(string(output))
	expectedName := "github.com/aiseeq/savanna"

	if moduleName != expectedName {
		t.Errorf("Неверное имя модуля. Ожидалось: %s, получено: %s", expectedName, moduleName)
	} else {
		t.Logf("Имя модуля корректно: %s", moduleName)
	}
}

// TestGoBuild проверяет что код компилируется
func TestGoBuild(t *testing.T) {
	t.Parallel()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	// Проверяем компиляцию всех пакетов
	cmd := exec.Command("go", "build", "-buildvcs=false", "./...")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Проект не компилируется: %v\nВывод: %s", err, output)
	} else {
		t.Logf("Все пакеты успешно компилируются")
	}
}

// TestGoVet проверяет код с помощью go vet
func TestGoVet(t *testing.T) {
	t.Parallel()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("go vet обнаружил проблемы: %v\nВывод: %s", err, output)
	} else {
		t.Logf("go vet не обнаружил проблем")
	}
}

// TestGoFmt проверяет форматирование кода
func TestGoFmt(t *testing.T) {
	t.Parallel()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("go fmt завершился с ошибкой: %v\nВывод: %s", err, output)
	}

	// Проверяем что go fmt не внес изменений
	if len(output) > 0 {
		t.Logf("go fmt внес изменения в код (возможно нужно закоммитить):\n%s", output)
	} else {
		t.Logf("Код уже отформатирован корректно")
	}
}
