package unit

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProjectStructure проверяет что все необходимые папки и файлы существуют
//
//nolint:gocognit // Комплексный unit тест структуры проекта
func TestProjectStructure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		isDir    bool
		required bool
	}{
		// Основные файлы
		{"go.mod", "go.mod", false, true},
		{"Makefile", "Makefile", false, true},
		{"README.md", "README.md", false, true},
		{"CLAUDE.md", "CLAUDE.md", false, true},
		{".gitignore", ".gitignore", false, true},

		// Структура папок
		{"config dir", "config", true, true},
		{"cmd dir", "cmd", true, true},
		{"cmd/game dir", "cmd/game", true, true},
		{"cmd/headless dir", "cmd/headless", true, true},
		{"internal dir", "internal", true, true},
		{"internal/core dir", "internal/core", true, true},
		{"internal/simulation dir", "internal/simulation", true, true},
		{"internal/physics dir", "internal/physics", true, true},
		{"internal/rendering dir", "internal/rendering", true, true},
		{"internal/generator dir", "internal/generator", true, true},
		{"assets dir", "assets", true, true},
		{"assets/sprites dir", "assets/sprites", true, true},
		{"assets/sprites/rabbit dir", "assets/sprites/rabbit", true, true},
		{"assets/sprites/wolf dir", "assets/sprites/wolf", true, true},
		{"assets/terrain dir", "assets/terrain", true, true},
		{"tests dir", "tests", true, true},
		{"tests/unit dir", "tests/unit", true, true},
		{"tests/integration dir", "tests/integration", true, true},
		{"tests/performance dir", "tests/performance", true, true},
		{"tests/fixtures dir", "tests/fixtures", true, true},
		{"scripts dir", "scripts", true, true},
		{"docs dir", "docs", true, true},

		// Main файлы
		{"game main.go", "cmd/game/main.go", false, true},
		{"headless main.go", "cmd/headless/main.go", false, true},

		// Собранные бинари (могут отсутствовать)
		{"game binary", "bin/savanna-game", false, false},
		{"headless binary", "bin/savanna-headless", false, false},
	}

	// Получаем корневую папку проекта (поднимаемся на 2 уровня вверх)
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fullPath := filepath.Join(projectRoot, tt.path)
			info, err := os.Stat(fullPath)

			if tt.required {
				if os.IsNotExist(err) {
					t.Errorf("Обязательный файл/папка не найдены: %s", tt.path)
					return
				}
				if err != nil {
					t.Errorf("Ошибка при проверке %s: %v", tt.path, err)
					return
				}

				if tt.isDir && !info.IsDir() {
					t.Errorf("%s должна быть папкой, но это файл", tt.path)
				}
				if !tt.isDir && info.IsDir() {
					t.Errorf("%s должен быть файлом, но это папка", tt.path)
				}
			} else if os.IsNotExist(err) {
				// Для необязательных файлов просто логируем их отсутствие
				t.Logf("Необязательный файл отсутствует: %s", tt.path)
			}
		})
	}
}

// TestGoMod проверяет содержимое go.mod
func TestGoMod(t *testing.T) {
	t.Parallel()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Не удалось определить корневую папку проекта: %v", err)
	}

	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Не удалось прочитать go.mod: %v", err)
	}

	goModContent := string(content)

	// Проверяем обязательные элементы
	tests := []struct {
		name     string
		contains string
	}{
		{"module name", "github.com/aiseeq/savanna"},
		{"go version", "go 1."},
		{"ebiten dependency", "github.com/hajimehoshi/ebiten/v2"},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !contains(goModContent, tt.contains) {
				t.Errorf("go.mod должен содержать '%s'", tt.contains)
			}
		})
	}
}

// contains проверяет содержит ли строка подстроку (helper функция)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (substr == "" || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
