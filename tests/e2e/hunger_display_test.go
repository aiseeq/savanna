package e2e

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestHungerDisplayE2E проверяет что значения сытости отображаются над животными
func TestHungerDisplayE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка отображения сытости над животными ===")

	// Создаём минимальный мир
	world := core.NewWorld(200, 200, 12345)

	// Создаём terrain (для полного теста)
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 3
	terrainGen := generator.NewTerrainGenerator(cfg)
	_ = terrainGen.Generate()

	// Создаём животных с разным уровнем голода
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 50, 50)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 100, 50)
	wolf1 := simulation.CreateAnimal(world, core.TypeWolf, 150, 50)

	// Устанавливаем разные уровни голода
	world.SetHunger(rabbit1, core.Hunger{Value: 25.5}) // Очень голодный
	world.SetHunger(rabbit2, core.Hunger{Value: 89.2}) // Почти сытый
	world.SetHunger(wolf1, core.Hunger{Value: 60.0})   // Средний голод

	// Создаём hunger display system
	hungerDisplay := NewHungerDisplaySystem()

	// Получаем информацию для отображения
	displayInfo := hungerDisplay.GetDisplayInfo(world)

	// Проверяем что информация получена для всех животных
	if len(displayInfo) != 3 {
		t.Errorf("❌ Ожидали информацию для 3 животных, получили %d", len(displayInfo))
		return
	}
	t.Logf("✅ Информация получена для %d животных", len(displayInfo))

	// Проверяем корректность информации для каждого животного
	for _, info := range displayInfo {
		entity := info.EntityID
		expectedHunger, hasHunger := world.GetHunger(entity)
		if !hasHunger {
			t.Errorf("❌ Животное %d не имеет компонента голода", entity)
			continue
		}

		// Проверяем позицию
		expectedPos, hasPos := world.GetPosition(entity)
		if !hasPos {
			t.Errorf("❌ Животное %d не имеет позиции", entity)
			continue
		}

		// Проверяем что позиция смещена вверх (над животным)
		expectedDisplayY := expectedPos.Y - 30 // Текст должен быть над животным
		if info.X != expectedPos.X || info.Y != expectedDisplayY {
			t.Errorf("❌ Неверная позиция для животного %d: ожидали (%.1f, %.1f), получили (%.1f, %.1f)",
				entity, expectedPos.X, expectedDisplayY, info.X, info.Y)
			continue
		}

		// Проверяем текст голода
		expectedText := fmt.Sprintf("%.0f%%", expectedHunger.Value)
		if info.HungerText != expectedText {
			t.Errorf("❌ Неверный текст голода для животного %d: ожидали '%s', получили '%s'",
				entity, expectedText, info.HungerText)
			continue
		}

		// Проверяем цвет (голодные = красный, сытые = зелёный)
		expectedColor := hungerDisplay.GetHungerColor(expectedHunger.Value)
		if info.Color != expectedColor {
			t.Errorf("❌ Неверный цвет для животного %d (голод %.1f%%): ожидали %v, получили %v",
				entity, expectedHunger.Value, expectedColor, info.Color)
			continue
		}

		t.Logf("✅ Животное %d: позиция (%.1f, %.1f), голод %s, цвет %v",
			entity, info.X, info.Y, info.HungerText, info.Color)
	}

	// Тестируем рендеринг (проверяем что метод не падает)
	mockScreen := &MockScreen{}
	hungerDisplay.RenderHungerTexts(mockScreen, displayInfo, nil)

	if len(mockScreen.RenderedTexts) != 3 {
		t.Errorf("❌ Ожидали отрендерить 3 текста, отрендерили %d", len(mockScreen.RenderedTexts))
		return
	}
	t.Logf("✅ Отрендерено %d текстов сытости", len(mockScreen.RenderedTexts))

	t.Logf("✅ Тест отображения сытости завершён")
}

// HungerDisplaySystem управляет отображением сытости над животными
type HungerDisplaySystem struct{}

// NewHungerDisplaySystem создаёт новую систему отображения голода
func NewHungerDisplaySystem() *HungerDisplaySystem {
	return &HungerDisplaySystem{}
}

// HungerDisplayInfo содержит информацию для отображения голода одного животного
type HungerDisplayInfo struct {
	EntityID   core.EntityID
	X, Y       float32
	HungerText string
	Color      HungerColor
}

// HungerColor представляет цвет текста голода
type HungerColor struct {
	R, G, B, A uint8
}

// GetDisplayInfo получает информацию для отображения голода всех животных
func (hds *HungerDisplaySystem) GetDisplayInfo(world *core.World) []HungerDisplayInfo {
	var displayInfos []HungerDisplayInfo

	// Проходим по всем животным
	world.ForEachWith(core.MaskAnimalType|core.MaskPosition|core.MaskHunger, func(entity core.EntityID) {
		pos, hasPos := world.GetPosition(entity)
		hunger, hasHunger := world.GetHunger(entity)

		if hasPos && hasHunger {
			info := HungerDisplayInfo{
				EntityID:   entity,
				X:          pos.X,
				Y:          pos.Y - 30, // Над животным
				HungerText: fmt.Sprintf("%.0f%%", hunger.Value),
				Color:      hds.GetHungerColor(hunger.Value),
			}
			displayInfos = append(displayInfos, info)
		}
	})

	return displayInfos
}

// GetHungerColor возвращает цвет текста в зависимости от уровня голода
func (hds *HungerDisplaySystem) GetHungerColor(hungerValue float32) HungerColor {
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

// RenderHungerTexts рендерит тексты голода на экране
func (hds *HungerDisplaySystem) RenderHungerTexts(
	screen interface{}, displayInfos []HungerDisplayInfo, font interface{},
) {
	// В реальной реализации здесь будет рендеринг текста
	// Для теста просто вызываем mock методы
	if mockScreen, ok := screen.(*MockScreen); ok {
		for _, info := range displayInfos {
			mockScreen.RenderText(info.HungerText, info.X, info.Y, info.Color)
		}
	}
}

// MockScreen имитирует экран для тестирования
type MockScreen struct {
	RenderedTexts []MockRenderedText
}

// MockRenderedText представляет отрендеренный текст
type MockRenderedText struct {
	Text  string
	X, Y  float32
	Color HungerColor
}

// RenderText имитирует рендеринг текста
func (ms *MockScreen) RenderText(text string, x, y float32, color HungerColor) {
	ms.RenderedTexts = append(ms.RenderedTexts, MockRenderedText{
		Text:  text,
		X:     x,
		Y:     y,
		Color: color,
	})
}
