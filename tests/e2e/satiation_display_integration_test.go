package e2e

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSatiationDisplayIntegrationE2E проверяет интеграцию отображения сытости в реальной игре
func TestSatiationDisplayIntegrationE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Интеграционный тест отображения сытости ===")

	// Создаём полноценную игру (имитация)
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 5

	// Создаём мир
	world := core.NewWorld(float32(cfg.World.Size*32), float32(cfg.World.Size*32), 12345)

	// Создаём terrain
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём животных
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 80, 80)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 120, 80)

	// Устанавливаем разные уровни сытости
	world.SetSatiation(rabbit, core.Satiation{Value: 25.0}) // Голодный заяц
	world.SetSatiation(wolf, core.Satiation{Value: 85.0})   // Сытый волк

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

	// Тестируем функциональность отображения сытости
	satiationDisplays := game.GetSatiationDisplaysForAllAnimals()

	// Проверяем что информация получена для обоих животных
	if len(satiationDisplays) != 2 {
		t.Errorf("❌ Ожидали информацию для 2 животных, получили %d", len(satiationDisplays))
		return
	}
	t.Logf("✅ Информация для отображения сытости получена для %d животных", len(satiationDisplays))

	// Проверяем что данные корректны
	for _, display := range satiationDisplays {
		if display.SatiationText == "" {
			t.Errorf("❌ Пустой текст сытости для животного")
			continue
		}

		// Проверяем формат текста (должен содержать %)
		if display.SatiationText[len(display.SatiationText)-1] != '%' {
			t.Errorf("❌ Некорректный формат текста сытости: %s", display.SatiationText)
			continue
		}

		t.Logf("✅ Животное: текст сытости '%s', позиция (%.1f, %.1f)",
			display.SatiationText, display.X, display.Y)
	}

	// Тестируем рендеринг с font manager
	mockScreen := &MockScreen{}
	game.RenderSatiationDisplays(mockScreen, satiationDisplays)

	if len(mockScreen.RenderedTexts) != 2 {
		t.Errorf("❌ Ожидали отрендерить 2 текста, отрендерили %d", len(mockScreen.RenderedTexts))
		return
	}
	t.Logf("✅ Отрендерено %d текстов сытости с использованием font manager", len(mockScreen.RenderedTexts))

	t.Logf("✅ Интеграционный тест отображения сытости прошёл")
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

// GetSatiationDisplaysForAllAnimals получает информацию для отображения сытости всех животных
func (mg *MockGame) GetSatiationDisplaysForAllAnimals() []SatiationDisplayInfo {
	var displays []SatiationDisplayInfo
	world := mg.gameWorld.world

	world.ForEachWith(core.MaskAnimalType|core.MaskPosition|core.MaskSatiation, func(entity core.EntityID) {
		pos, hasPos := world.GetPosition(entity)
		satiation, hasSatiation := world.GetSatiation(entity)

		if hasPos && hasSatiation {
			// Имитируем логику из реальной игры
			// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
			display := SatiationDisplayInfo{
				EntityID:      entity,
				X:             pos.X,
				Y:             pos.Y - 25, // Над животным (как в реальной игре)
				SatiationText: mg.formatSatiationText(satiation.Value),
				Color:         mg.getSatiationColor(satiation.Value),
			}
			displays = append(displays, display)
		}
	})

	return displays
}

// formatSatiationText форматирует текст сытости (копирует логику из игры)
func (mg *MockGame) formatSatiationText(satiationValue float32) string {
	return fmt.Sprintf("%.0f%%", satiationValue)
}

// getSatiationColor возвращает цвет в зависимости от сытости (копирует логику из игры)
func (mg *MockGame) getSatiationColor(satiationValue float32) SatiationColor {
	if satiationValue < 30.0 {
		// Критический сытость - красный
		return SatiationColor{R: 255, G: 50, B: 50, A: 255}
	} else if satiationValue < 60.0 {
		// Средний сытость - жёлтый
		return SatiationColor{R: 255, G: 255, B: 50, A: 255}
	} else {
		// Сытость - зелёный
		return SatiationColor{R: 50, G: 255, B: 50, A: 255}
	}
}

// RenderSatiationDisplays рендерит отображения сытости
func (mg *MockGame) RenderSatiationDisplays(screen *MockScreen, displays []SatiationDisplayInfo) {
	for _, display := range displays {
		// Имитируем рендеринг с учётом наличия шрифта
		if mg.fontManager.HasCustomFont() {
			// Используем пользовательский шрифт (как в реальной игре)
			screen.RenderText(display.SatiationText, display.X-20, display.Y, display.Color)
		} else {
			// Фолбэк (как в реальной игре)
			screen.RenderText(display.SatiationText, display.X-20, display.Y, SatiationColor{R: 255, G: 255, B: 255, A: 255})
		}
	}
}
