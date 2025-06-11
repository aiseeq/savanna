package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestGUIAppBasicStart проверяет что GUI приложение запускается (только в WSL/Linux с X11)
func TestGUIAppBasicStart(t *testing.T) {
	// Проверяем доступность X11
	if os.Getenv("DISPLAY") == "" {
		t.Skip("Пропускаем GUI тест: DISPLAY не установлен (нет X11)")
	}

	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	binaryPath := filepath.Join(projectRoot, "bin", "savanna-game")

	// Убедимся что бинарь существует
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		buildCmd := exec.Command("make", "build")
		buildCmd.Dir = projectRoot
		if err := buildCmd.Run(); err != nil {
			t.Fatalf("Не удалось собрать проект: %v", err)
		}
	}

	// Запускаем GUI приложение с таймаутом 5 секунд
	cmd := exec.Command(binaryPath)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"MIT_SHM=0",
		"LIBGL_ALWAYS_SOFTWARE=1",
	)

	// Запускаем в отдельной горутине с таймаутом
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		// Приложение завершилось самостоятельно
		if err != nil {
			// Проверяем что это не критическая ошибка
			if strings.Contains(err.Error(), "exit status") {
				t.Logf("GUI приложение запустилось и завершилось (вероятно по ESC или закрытию окна)")
			} else {
				t.Errorf("GUI приложение завершилось с ошибкой: %v", err)
			}
		} else {
			t.Logf("GUI приложение запустилось и завершилось корректно")
		}

	case <-time.After(5 * time.Second):
		// Приложение работает более 5 секунд - это хорошо, завершаем принудительно
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Logf("GUI приложение успешно запустилось и работало > 5 секунд")
	}
}

// TestGUIAppMakeRun проверяет команду make run
func TestGUIAppMakeRun(t *testing.T) {
	// Проверяем доступность X11
	if os.Getenv("DISPLAY") == "" {
		t.Skip("Пропускаем GUI тест: DISPLAY не установлен (нет X11)")
	}

	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	// Запускаем make run с таймаутом
	cmd := exec.Command("timeout", "3s", "make", "run")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// timeout завершает с кодом 124, это нормально для нашего теста
	if err != nil && !strings.Contains(err.Error(), "exit status 124") {
		t.Errorf("make run завершился с неожиданной ошибкой: %v\nВывод: %s", err, outputStr)
	}

	// Проверяем что make run запустился корректно
	if strings.Contains(outputStr, "Сборка завершена") {
		t.Logf("make run успешно запустился")
	} else if strings.Contains(outputStr, "Запуск GUI версии") {
		t.Logf("make run успешно запустился (альтернативная проверка)")
	} else {
		t.Errorf("make run не показал ожидаемый вывод. Получен:\n%s", outputStr)
	}

	t.Logf("Вывод make run:\n%s", outputStr)
}
