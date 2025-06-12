package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGUIAppCompilation проверяет что GUI приложение компилируется
func TestGUIAppCompilation(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	// Проверяем что можем собрать GUI версию
	buildCmd := exec.Command("go", "build", "-o", "bin/savanna-game-test", "./cmd/game")
	buildCmd.Dir = projectRoot

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("GUI приложение не компилируется: %v\nВывод: %s", err, output)
	}

	// Проверяем что бинарь создался
	testBinaryPath := filepath.Join(projectRoot, "bin", "savanna-game-test")
	if _, err := os.Stat(testBinaryPath); os.IsNotExist(err) {
		t.Errorf("Тестовый бинарь GUI приложения не создан")
	} else {
		t.Logf("GUI приложение успешно скомпилировано")
		// Удаляем тестовый бинарь
		os.Remove(testBinaryPath)
	}
}

// Функции запуска GUI убраны из автоматических тестов
// Для ручного тестирования используйте: make run
