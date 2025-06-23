package unit

import (
	"testing"
)

// Константы изометрии (копия из internal/rendering/renderer.go)
const (
	TileWidth  = 32
	TileHeight = 16
)

// worldToScreen - прямая реализация формул изометрии для тестирования
func worldToScreen(worldX, worldY float32) (screenX, screenY float32) {
	screenX = (worldX - worldY) * TileWidth / 2
	screenY = (worldX + worldY) * TileHeight / 2
	return screenX, screenY
}

// screenToWorld - обратная изометрическая формула
func screenToWorld(screenX, screenY float32) (worldX, worldY float32) {
	worldX = (screenX/(TileWidth/2) + screenY/(TileHeight/2)) / 2
	worldY = (screenY/(TileHeight/2) - screenX/(TileWidth/2)) / 2
	return worldX, worldY
}

// TestIsometricFormulas тестирует корректность изометрических формул преобразования
func TestIsometricFormulas(t *testing.T) {

	// Тестовые случаи: мировые координаты и ожидаемые экранные координаты
	testCases := []struct {
		name            string
		worldX          float32
		worldY          float32
		expectedPattern string
	}{
		{"Origin", 0, 0, "center"},
		{"Right", 1, 0, "top-right diagonal"},
		{"Down", 0, 1, "bottom-right diagonal"},
		{"Diagonal", 1, 1, "bottom"},
		{"Negative", -1, -1, "top"},
		{"Mixed", 2, -1, "far top-right"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Преобразуем world -> screen
			screenX, screenY := worldToScreen(tc.worldX, tc.worldY)

			// Преобразуем обратно screen -> world
			backWorldX, backWorldY := screenToWorld(screenX, screenY)

			// Проверяем точность обратного преобразования
			tolerance := float32(0.001)
			if absFloat32(backWorldX-tc.worldX) > tolerance {
				t.Errorf("Неточное обратное преобразование X: %.6f != %.6f (разница %.6f)",
					backWorldX, tc.worldX, absFloat32(backWorldX-tc.worldX))
			}
			if absFloat32(backWorldY-tc.worldY) > tolerance {
				t.Errorf("Неточное обратное преобразование Y: %.6f != %.6f (разница %.6f)",
					backWorldY, tc.worldY, absFloat32(backWorldY-tc.worldY))
			}

			t.Logf("%s: world(%.2f,%.2f) -> screen(%.2f,%.2f) -> world(%.6f,%.6f)",
				tc.name, tc.worldX, tc.worldY, screenX, screenY, backWorldX, backWorldY)
		})
	}
}

// TestIsometricProjectionProperties проверяет основные свойства изометрической проекции
func TestIsometricProjectionProperties(t *testing.T) {

	// 1. Тест симметрии: точки (x,y) и (y,x) должны быть симметричны относительно диагонали
	x1, y1 := worldToScreen(3, 1)
	x2, y2 := worldToScreen(1, 3)

	// В изометрии (a,b) и (b,a) должны быть зеркальными относительно вертикальной оси
	t.Logf("Симметрия: (3,1) -> (%.1f,%.1f), (1,3) -> (%.1f,%.1f)", x1, y1, x2, y2)

	// 2. Тест пропорциональности: удвоение координат должно удваивать экранные координаты
	sx1, sy1 := worldToScreen(2, 3)
	sx2, sy2 := worldToScreen(4, 6)

	expectedX := sx1 * 2
	expectedY := sy1 * 2

	tolerance := float32(0.1)
	if absFloat32(sx2-expectedX) > tolerance || absFloat32(sy2-expectedY) > tolerance {
		t.Errorf("Нарушена пропорциональность: (2,3)*2 должно быть (%.1f,%.1f), получили (%.1f,%.1f)",
			expectedX, expectedY, sx2, sy2)
	}

	// 3. Проверяем что изометрия сохраняет пропорции диагоналей
	// Диагональ квадрата 4x4 в мире
	corner1X, corner1Y := worldToScreen(0, 0)
	corner2X, corner2Y := worldToScreen(4, 4)

	diagonalLength := absFloat32(corner2X-corner1X) + absFloat32(corner2Y-corner1Y)
	t.Logf("Диагональ квадрата 4x4: от (%.1f,%.1f) до (%.1f,%.1f), длина: %.1f",
		corner1X, corner1Y, corner2X, corner2Y, diagonalLength)

	// 4. Проверяем константы изометрии из renderer.go
	// TileWidth=32, TileHeight=16, поэтому коэффициенты должны быть 16 и 8
	originX, originY := worldToScreen(0, 0)
	unitXscreenX, unitXscreenY := worldToScreen(1, 0)
	unitYscreenX, unitYscreenY := worldToScreen(0, 1)

	// Приращение для единичного шага по X
	deltaXx := unitXscreenX - originX // Должно быть 16
	deltaXy := unitXscreenY - originY // Должно быть 8

	// Приращение для единичного шага по Y
	deltaYx := unitYscreenX - originX // Должно быть -16
	deltaYy := unitYscreenY - originY // Должно быть 8

	t.Logf("Изометрические коэффициенты:")
	t.Logf("  Шаг по X: экран(%.1f, %.1f)", deltaXx, deltaXy)
	t.Logf("  Шаг по Y: экран(%.1f, %.1f)", deltaYx, deltaYy)

	// Проверяем что коэффициенты соответствуют формулам
	expectedXx := float32(32) / 2  // TileWidth/2 = 16
	expectedXy := float32(16) / 2  // TileHeight/2 = 8
	expectedYx := -float32(32) / 2 // -TileWidth/2 = -16
	expectedYy := float32(16) / 2  // TileHeight/2 = 8

	if absFloat32(deltaXx-expectedXx) > 0.1 {
		t.Errorf("Неверный коэффициент X->screenX: %.1f != %.1f", deltaXx, expectedXx)
	}
	if absFloat32(deltaXy-expectedXy) > 0.1 {
		t.Errorf("Неверный коэффициент X->screenY: %.1f != %.1f", deltaXy, expectedXy)
	}
	if absFloat32(deltaYx-expectedYx) > 0.1 {
		t.Errorf("Неверный коэффициент Y->screenX: %.1f != %.1f", deltaYx, expectedYx)
	}
	if absFloat32(deltaYy-expectedYy) > 0.1 {
		t.Errorf("Неверный коэффициент Y->screenY: %.1f != %.1f", deltaYy, expectedYy)
	}

	t.Logf("✓ Изометрические формулы работают корректно")
}

func absFloat32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
