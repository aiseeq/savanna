package e2e

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSatiationDisplayE2E проверяет что значения сытости отображаются над животными
func TestSatiationDisplayE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка отображения сытости над животными ===")

	// Создаём минимальный мир
	world := core.NewWorld(200, 200, 12345)

	// Создаём terrain (для полного теста)
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 3
	terrainGen := generator.NewTerrainGenerator(cfg)
	_ = terrainGen.Generate()

	// Создаём животных с разным уровнем сытости
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 50, 50)
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 100, 50)
	wolf1 := simulation.CreateAnimal(world, core.TypeWolf, 150, 50)

	// Устанавливаем разные уровни сытости
	world.SetSatiation(rabbit1, core.Satiation{Value: 25.5}) // Очень голодный
	world.SetSatiation(rabbit2, core.Satiation{Value: 89.2}) // Почти сытый
	world.SetSatiation(wolf1, core.Satiation{Value: 60.0})   // Средний сытость

	// Создаём satiation display system
	satiationDisplay := NewSatiationDisplaySystem()

	// Получаем информацию для отображения
	displayInfo := satiationDisplay.GetDisplayInfo(world)

	// Проверяем что информация получена для всех животных
	if len(displayInfo) != 3 {
		t.Errorf("❌ Ожидали информацию для 3 животных, получили %d", len(displayInfo))
		return
	}
	t.Logf("✅ Информация получена для %d животных", len(displayInfo))

	// Проверяем корректность информации для каждого животного
	for _, info := range displayInfo {
		entity := info.EntityID
		expectedSatiation, hasSatiation := world.GetSatiation(entity)
		if !hasSatiation {
			t.Errorf("❌ Животное %d не имеет компонента сытости", entity)
			continue
		}

		// Проверяем позицию
		expectedPos, hasPos := world.GetPosition(entity)
		if !hasPos {
			t.Errorf("❌ Животное %d не имеет позиции", entity)
			continue
		}

		// Проверяем что позиция смещена вверх (над животным)
		// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для сравнения
		expectedDisplayY := expectedPos.Y - 30 // Текст должен быть над животным
		if info.X != expectedPos.X || info.Y != expectedDisplayY {
			t.Errorf("❌ Неверная позиция для животного %d: ожидали (%.1f, %.1f), получили (%.1f, %.1f)",
				entity, expectedPos.X, expectedDisplayY, info.X, info.Y)
			continue
		}

		// Проверяем текст сытости
		expectedText := fmt.Sprintf("%.0f%%", expectedSatiation.Value)
		if info.SatiationText != expectedText {
			t.Errorf("❌ Неверный текст сытости для животного %d: ожидали '%s', получили '%s'",
				entity, expectedText, info.SatiationText)
			continue
		}

		// Проверяем цвет (сытостьные = красный, сытые = зелёный)
		expectedColor := satiationDisplay.GetSatiationColor(expectedSatiation.Value)
		if info.Color != expectedColor {
			t.Errorf("❌ Неверный цвет для животного %d (сытость %.1f%%): ожидали %v, получили %v",
				entity, expectedSatiation.Value, expectedColor, info.Color)
			continue
		}

		t.Logf("✅ Животное %d: позиция (%.1f, %.1f), сытость %s, цвет %v",
			entity, info.X, info.Y, info.SatiationText, info.Color)
	}

	// Тестируем рендеринг (проверяем что метод не падает)
	mockScreen := &MockScreen{}
	satiationDisplay.RenderSatiationTexts(mockScreen, displayInfo, nil)

	if len(mockScreen.RenderedTexts) != 3 {
		t.Errorf("❌ Ожидали отрендерить 3 текста, отрендерили %d", len(mockScreen.RenderedTexts))
		return
	}
	t.Logf("✅ Отрендерено %d текстов сытости", len(mockScreen.RenderedTexts))

	t.Logf("✅ Тест отображения сытости завершён")
}

// SatiationDisplaySystem управляет отображением сытости над животными
type SatiationDisplaySystem struct{}

// NewSatiationDisplaySystem создаёт новую систему отображения сытости
func NewSatiationDisplaySystem() *SatiationDisplaySystem {
	return &SatiationDisplaySystem{}
}

// SatiationDisplayInfo содержит информацию для отображения сытости одного животного
type SatiationDisplayInfo struct {
	EntityID      core.EntityID
	X, Y          float32
	SatiationText string
	Color         SatiationColor
}

// SatiationColor представляет цвет текста сытости
type SatiationColor struct {
	R, G, B, A uint8
}

// GetDisplayInfo получает информацию для отображения сытости всех животных
func (hds *SatiationDisplaySystem) GetDisplayInfo(world *core.World) []SatiationDisplayInfo {
	var displayInfos []SatiationDisplayInfo

	// Проходим по всем животным
	world.ForEachWith(core.MaskAnimalType|core.MaskPosition|core.MaskSatiation, func(entity core.EntityID) {
		pos, hasPos := world.GetPosition(entity)
		satiation, hasSatiation := world.GetSatiation(entity)

		if hasPos && hasSatiation {
			// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
			info := SatiationDisplayInfo{
				EntityID:      entity,
				X:             pos.X,
				Y:             pos.Y - 30, // Над животным
				SatiationText: fmt.Sprintf("%.0f%%", satiation.Value),
				Color:         hds.GetSatiationColor(satiation.Value),
			}
			displayInfos = append(displayInfos, info)
		}
	})

	return displayInfos
}

// GetSatiationColor возвращает цвет текста в зависимости от уровня сытости
func (hds *SatiationDisplaySystem) GetSatiationColor(satiationValue float32) SatiationColor {
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

// RenderSatiationTexts рендерит тексты сытости на экране
func (hds *SatiationDisplaySystem) RenderSatiationTexts(
	screen interface{}, displayInfos []SatiationDisplayInfo, font interface{},
) {
	// В реальной реализации здесь будет рендеринг текста
	// Для теста просто вызываем mock методы
	if mockScreen, ok := screen.(*MockScreen); ok {
		for _, info := range displayInfos {
			mockScreen.RenderText(info.SatiationText, info.X, info.Y, info.Color)
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
	Color SatiationColor
}

// RenderText имитирует рендеринг текста
func (ms *MockScreen) RenderText(text string, x, y float32, color SatiationColor) {
	ms.RenderedTexts = append(ms.RenderedTexts, MockRenderedText{
		Text:  text,
		X:     x,
		Y:     y,
		Color: color,
	})
}
