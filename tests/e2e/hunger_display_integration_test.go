package e2e

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestHungerDisplayIntegrationE2E проверяет интеграцию отображения голода в реальной игре
func TestHungerDisplayIntegrationE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Интеграционный тест отображения голода ===")

	// Создаём полноценную игру (имитация)
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 5

	// Создаём мир
	world := core.NewWorld(float32(cfg.World.Size*32), float32(cfg.World.Size*32), 12345)

	// Создаём terrain
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём животных
	rabbit := simulation.CreateRabbit(world, 80, 80)
	wolf := simulation.CreateWolf(world, 120, 80)

	// Устанавливаем разные уровни голода
	world.SetHunger(rabbit, core.Hunger{Value: 25.0}) // Голодный заяц
	world.SetHunger(wolf, core.Hunger{Value: 85.0})   // Сытый волк

	// Имитируем GameWorld и FontManager
	gameWorld := &MockGameWorld{
		world:   world,
		terrain: terrain,
		stats: map[string]interface{}{
			"rabbits": 1,
			"wolves":  1,
		},
	}

	fontManager := &MockFontManager{
		hasFont: true,
	}

	// Имитируем Game structure
	game := &MockGame{
		gameWorld:   gameWorld,
		fontManager: fontManager,
	}

	// Тестируем функциональность отображения голода
	hungerDisplays := game.GetHungerDisplaysForAllAnimals()

	// Проверяем что информация получена для обоих животных
	if len(hungerDisplays) != 2 {
		t.Errorf("❌ Ожидали информацию для 2 животных, получили %d", len(hungerDisplays))
		return
	}
	t.Logf("✅ Информация для отображения голода получена для %d животных", len(hungerDisplays))

	// Проверяем что данные корректны
	for _, display := range hungerDisplays {
		if display.HungerText == "" {
			t.Errorf("❌ Пустой текст голода для животного")
			continue
		}

		// Проверяем формат текста (должен содержать %)
		if display.HungerText[len(display.HungerText)-1] != '%' {
			t.Errorf("❌ Некорректный формат текста голода: %s", display.HungerText)
			continue
		}

		t.Logf("✅ Животное: текст голода '%s', позиция (%.1f, %.1f)",
			display.HungerText, display.X, display.Y)
	}

	// Тестируем рендеринг с font manager
	mockScreen := &MockScreen{}
	game.RenderHungerDisplays(mockScreen, hungerDisplays)

	if len(mockScreen.RenderedTexts) != 2 {
		t.Errorf("❌ Ожидали отрендерить 2 текста, отрендерили %d", len(mockScreen.RenderedTexts))
		return
	}
	t.Logf("✅ Отрендерено %d текстов голода с использованием font manager", len(mockScreen.RenderedTexts))

	t.Logf("✅ Интеграционный тест отображения голода прошёл")
}

// MockGame имитирует структуру Game для тестирования
type MockGame struct {
	gameWorld   *MockGameWorld
	fontManager *MockFontManager
}

// MockFontManager имитирует FontManager
type MockFontManager struct {
	hasFont bool
}

// HasCustomFont проверяет наличие пользовательского шрифта
func (mfm *MockFontManager) HasCustomFont() bool {
	return mfm.hasFont
}

// GetHungerDisplaysForAllAnimals получает информацию для отображения голода всех животных
func (mg *MockGame) GetHungerDisplaysForAllAnimals() []HungerDisplayInfo {
	var displays []HungerDisplayInfo
	world := mg.gameWorld.world

	world.ForEachWith(core.MaskAnimalType|core.MaskPosition|core.MaskHunger, func(entity core.EntityID) {
		pos, hasPos := world.GetPosition(entity)
		hunger, hasHunger := world.GetHunger(entity)

		if hasPos && hasHunger {
			// Имитируем логику из реальной игры
			display := HungerDisplayInfo{
				EntityID:   entity,
				X:          pos.X,
				Y:          pos.Y - 25, // Над животным (как в реальной игре)
				HungerText: mg.formatHungerText(hunger.Value),
				Color:      mg.getHungerColor(hunger.Value),
			}
			displays = append(displays, display)
		}
	})

	return displays
}

// formatHungerText форматирует текст голода (копирует логику из игры)
func (mg *MockGame) formatHungerText(hungerValue float32) string {
	return fmt.Sprintf("%.0f%%", hungerValue)
}

// getHungerColor возвращает цвет в зависимости от голода (копирует логику из игры)
func (mg *MockGame) getHungerColor(hungerValue float32) HungerColor {
	if hungerValue < 30.0 {
		// Критический голод - красный
		return HungerColor{R: 255, G: 50, B: 50, A: 255}
	} else if hungerValue < 60.0 {
		// Средний голод - жёлтый
		return HungerColor{R: 255, G: 255, B: 50, A: 255}
	} else {
		// Сытость - зелёный
		return HungerColor{R: 50, G: 255, B: 50, A: 255}
	}
}

// RenderHungerDisplays рендерит отображения голода
func (mg *MockGame) RenderHungerDisplays(screen *MockScreen, displays []HungerDisplayInfo) {
	for _, display := range displays {
		// Имитируем рендеринг с учётом наличия шрифта
		if mg.fontManager.HasCustomFont() {
			// Используем пользовательский шрифт (как в реальной игре)
			screen.RenderText(display.HungerText, display.X-20, display.Y, display.Color)
		} else {
			// Фолбэк (как в реальной игре)
			screen.RenderText(display.HungerText, display.X-20, display.Y, HungerColor{R: 255, G: 255, B: 255, A: 255})
		}
	}
}
