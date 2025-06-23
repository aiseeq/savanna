package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/rendering"
)

// TestIsometricCoordinateMapping проверяет что изометрическая проекция
// корректно преобразует прямоугольную карту в ромбовидную проекцию на экране
func TestIsometricCoordinateMapping(t *testing.T) {
	// Используем изометрический рендерер (для проверки функций преобразования)
	// renderer := rendering.NewIsometricRenderer() // НЕ используется напрямую

	// Создаем камеру (без GUI)
	camera := rendering.NewCamera(10, 10)

	// Проверяем что координаты углов карты образуют ромб на экране
	// Углы карты в мировых координатах: (0,0), (9,0), (0,9), (9,9)

	// Преобразуем углы карты в экранные координаты
	topLeftX, topLeftY := camera.WorldToScreen(0, 0)
	topRightX, topRightY := camera.WorldToScreen(9, 0)
	bottomLeftX, bottomLeftY := camera.WorldToScreen(0, 9)
	bottomRightX, bottomRightY := camera.WorldToScreen(9, 9)

	t.Logf("Углы карты на экране:")
	t.Logf("  Верх-лево: (%.1f, %.1f)", topLeftX, topLeftY)
	t.Logf("  Верх-право: (%.1f, %.1f)", topRightX, topRightY)
	t.Logf("  Низ-лево: (%.1f, %.1f)", bottomLeftX, bottomLeftY)
	t.Logf("  Низ-право: (%.1f, %.1f)", bottomRightX, bottomRightY)

	// Проверяем базовые свойства изометрической проекции

	// 1. Карта НЕ должна быть квадратной в экранных координатах
	screenWidth := absFloat32(topRightX - bottomLeftX)
	screenHeight := absFloat32(bottomLeftY - topRightY)

	if screenWidth == screenHeight {
		t.Errorf("Карта отображается квадратом на экране (%.1f x %.1f), "+
			"но должна быть ромбовидной для изометрии", screenWidth, screenHeight)
	}

	// 2. В правильной изометрии диагональ ромба по X больше чем по Y
	// (потому что TileWidth > TileHeight в константах)
	if screenWidth <= screenHeight {
		t.Errorf("Карта не выглядит как изометрическая проекция: "+
			"ширина (%.1f) должна быть больше высоты (%.1f)", screenWidth, screenHeight)
	}

	// 3. Проверяем что углы образуют правильный ромб
	// В изометрии центр ромба должен быть в средней точке диагоналей
	centerX := (topLeftX + bottomRightX) / 2
	centerY := (topLeftY + bottomRightY) / 2

	// Альтернативный центр через другую диагональ
	altCenterX := (topRightX + bottomLeftX) / 2
	altCenterY := (topRightY + bottomLeftY) / 2

	tolerance := float32(0.1)
	if absFloat32(centerX-altCenterX) > tolerance || absFloat32(centerY-altCenterY) > tolerance {
		t.Errorf("Диагонали ромба не пересекаются в центре: "+
			"центр1(%.1f,%.1f) != центр2(%.1f,%.1f)", centerX, centerY, altCenterX, altCenterY)
	}

	t.Logf("✓ Карта корректно отображается в изометрической проекции")
	t.Logf("  Размер ромба: %.1f x %.1f", screenWidth, screenHeight)
	t.Logf("  Центр ромба: (%.1f, %.1f)", centerX, centerY)
	t.Logf("  Отношение ширины к высоте: %.2f", screenWidth/screenHeight)
}

// TestWorldToScreenTransformation тестирует корректность преобразования координат
func TestWorldToScreenTransformation(t *testing.T) {
	renderer := rendering.NewIsometricRenderer()

	// Тестируем стандартные случаи
	testCases := []struct {
		worldX, worldY  float32
		expectedPattern string
	}{
		{0, 0, "origin"},
		{1, 0, "right"},
		{0, 1, "down"},
		{1, 1, "diagonal"},
		{5, 5, "center"},
	}

	for _, tc := range testCases {
		screenX, screenY := renderer.WorldToScreen(tc.worldX, tc.worldY)

		// Проверяем что преобразование обратимо
		backWorldX, backWorldY := renderer.ScreenToWorld(screenX, screenY)

		tolerance := float32(0.001)
		if absFloat32(backWorldX-tc.worldX) > tolerance || absFloat32(backWorldY-tc.worldY) > tolerance {
			t.Errorf("Преобразование координат не обратимо для %s: "+
				"world(%.3f,%.3f) -> screen(%.3f,%.3f) -> world(%.3f,%.3f)",
				tc.expectedPattern, tc.worldX, tc.worldY, screenX, screenY, backWorldX, backWorldY)
		}
	}

	t.Logf("✓ Преобразования координат работают корректно")
}

func absFloat32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
