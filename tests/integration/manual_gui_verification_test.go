package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
)

// TestGUILogicAutomated автоматически проверяет логику которая должна работать в GUI
func TestGUILogicAutomated(t *testing.T) {
	t.Parallel()
	t.Logf("=== Автоматическая проверка GUI логики ===")

	// Проверяем что можем загрузить конфигурацию
	cfg := config.LoadDefaultConfig()
	if cfg == nil {
		t.Fatal("Не удалось загрузить конфигурацию")
	}

	// Проверяем разумные значения конфигурации
	if cfg.World.Size <= 0 || cfg.World.Size > 200 {
		t.Errorf("Неразумный размер мира: %d (ожидается 1-200)", cfg.World.Size)
	}

	if cfg.Population.Rabbits <= 0 || cfg.Population.Rabbits > 1000 {
		t.Errorf("Неразумное количество зайцев: %d (ожидается 1-1000)", cfg.Population.Rabbits)
	}

	if cfg.Population.Wolves < 0 || cfg.Population.Wolves > 100 {
		t.Errorf("Неразумное количество волков: %d (ожидается 0-100)", cfg.Population.Wolves)
	}

	t.Logf("✅ Конфигурация валидна:")
	t.Logf("  Размер мира: %dx%d тайлов", cfg.World.Size, cfg.World.Size)
	t.Logf("  Количество зайцев: %d", cfg.Population.Rabbits)
	t.Logf("  Количество волков: %d", cfg.Population.Wolves)

	// TODO: Добавить проверку что GUI системы инициализируются без паники
	// TODO: Добавить проверку что животные создаются в границах мира
	// TODO: Добавить проверку что анимационные системы работают
}
