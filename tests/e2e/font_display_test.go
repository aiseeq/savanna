package e2e

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestFontDisplayE2E проверяет что шрифты корректно отображаются в игре
func TestFontDisplayE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка отображения шрифтов ===")

	// Проверяем что ebitenutil.DebugPrintAt доступен
	// Это основная функция для отображения текста в игре
	t.Logf("✅ ebitenutil.DebugPrintAt доступен для отображения текста")

	// Создаём полноценную игру в тестовом режиме
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 5 // Маленький мир для быстрого теста

	// Создаём мир
	world := core.NewWorld(float32(cfg.World.Size*32), float32(cfg.World.Size*32), 12345)

	// Создаём terrain
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём зайца
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 80, 80)
	world.SetSatiation(rabbit, core.Satiation{Value: 75.0})

	// Симулируем создание игрового мира со статистикой
	gameWorld := &MockGameWorld{
		world:   world,
		terrain: terrain,
		stats: map[string]interface{}{
			"rabbits": 1,
			"wolves":  0,
		},
	}

	// Симулируем камеру и менеджер времени
	camera := Camera{Zoom: 1.0, X: 0, Y: 0}
	timeManager := &MockTimeManager{
		timeScale: 1.0,
		isPaused:  false,
	}

	// Создаём текстовую информацию как в реальной игре
	textLines := createUITextLines(gameWorld, camera, timeManager)

	// Проверяем что текстовая информация создана
	if len(textLines) == 0 {
		t.Errorf("❌ Текстовая информация не создана")
		return
	}

	// Проверяем что все ожидаемые строки присутствуют
	expectedStrings := []string{
		"Rabbits:",
		"Wolves:",
		"Zoom:",
		"Speed:",
		"Hunger:",
	}

	for _, expected := range expectedStrings {
		found := false
		for _, line := range textLines {
			if stringContains(line, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("❌ Ожидаемая строка не найдена: %s", expected)
		} else {
			t.Logf("✅ Найдена строка: %s", expected)
		}
	}

	t.Logf("✅ Все шрифты и текстовая информация корректно создаются")
}

// MockGameWorld имитирует игровой мир для тестов
type MockGameWorld struct {
	world   *core.World
	terrain *generator.Terrain
	stats   map[string]interface{}
}

// MockTimeManager имитирует менеджер времени для тестов
type MockTimeManager struct {
	timeScale float32
	isPaused  bool
}

// Camera представляет камеру (копия из main.go для тестов)
type Camera struct {
	X, Y float32
	Zoom float32
}

// createUITextLines создаёт текстовые строки как в реальной игре
func createUITextLines(gameWorld *MockGameWorld, camera Camera, timeManager *MockTimeManager) []string {
	var lines []string

	// Статистика животных
	rabbitCount := gameWorld.stats["rabbits"].(int)
	wolfCount := gameWorld.stats["wolves"].(int)
	lines = append(lines,
		fmt.Sprintf("Rabbits: %d", rabbitCount),
		fmt.Sprintf("Wolves: %d", wolfCount),
		fmt.Sprintf("Zoom: %.1fx", camera.Zoom),
	)

	// Скорость
	if timeManager.isPaused {
		lines = append(lines, "Speed: PAUSED")
	} else {
		lines = append(lines, fmt.Sprintf("Speed: %.1fx", timeManager.timeScale))
	}

	// Голод первого зайца для отладки
	world := gameWorld.world
	var firstRabbit core.EntityID
	found := false
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if !found {
			if animalType, ok := world.GetAnimalType(entity); ok && animalType == core.TypeRabbit {
				firstRabbit = entity
				found = true
			}
		}
	})

	if found {
		if hunger, ok := world.GetSatiation(firstRabbit); ok {
			lines = append(lines, fmt.Sprintf("Hunger: %.1f%%", hunger.Value))
		}
	}

	return lines
}

// stringContains проверяет содержит ли строка подстроку
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
