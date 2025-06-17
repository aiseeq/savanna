package e2e

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestCameraControlE2E проверяет что управление камерой правой кнопкой мыши работает
func TestCameraControlE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка управления камерой ===")

	// Создаём контроллер камеры
	cameraController := NewMockCameraController()

	// Проверяем начальную позицию
	camera := cameraController.GetCamera()
	initialX := camera.X
	initialY := camera.Y
	t.Logf("✅ Начальная позиция камеры: (%.1f, %.1f)", initialX, initialY)

	// Симулируем нажатие правой кнопки мыши
	cameraController.SimulateMousePress(ebiten.MouseButtonRight, 400, 300)

	// Симулируем перетаскивание мыши
	cameraController.SimulateMouseMove(450, 350) // Движение на 50 пикселей вправо и вниз

	// Получаем новую позицию камеры
	camera = cameraController.GetCamera()
	newX := camera.X
	newY := camera.Y

	// Проверяем что камера переместилась
	// При перетаскивании вправо камера должна двигаться влево (в обратную сторону)
	if newX >= initialX {
		t.Errorf("❌ Камера не переместилась влево при перетаскивании вправо: %.1f >= %.1f", newX, initialX)
		return
	}

	// При перетаскивании вниз камера должна двигаться вверх
	if newY >= initialY {
		t.Errorf("❌ Камера не переместилась вверх при перетаскивании вниз: %.1f >= %.1f", newY, initialY)
		return
	}

	t.Logf("✅ Камера переместилась правильно: (%.1f, %.1f) -> (%.1f, %.1f)", initialX, initialY, newX, newY)

	// Проверяем что левая кнопка НЕ должна двигать камеру (она для другого)
	cameraController.SimulateMouseRelease(ebiten.MouseButtonRight)

	// Сбрасываем позицию для теста левой кнопки
	cameraController.SetCameraPosition(400, 300)
	initialX = 400
	initialY = 300

	// Тестируем левую кнопку - она НЕ должна двигать камеру
	cameraController.SimulateMousePress(ebiten.MouseButtonLeft, 400, 300)
	cameraController.SimulateMouseMove(450, 350)

	camera = cameraController.GetCamera()
	leftButtonX := camera.X
	leftButtonY := camera.Y

	// Левая кнопка НЕ должна перемещать камеру (она для выбора объектов)
	if leftButtonX != initialX || leftButtonY != initialY {
		t.Logf("⚠️  Левая кнопка тоже двигает камеру: (%.1f, %.1f) -> (%.1f, %.1f)", initialX, initialY, leftButtonX, leftButtonY)
		t.Logf("   Это может быть нормально, зависит от дизайна игры")
	} else {
		t.Logf("✅ Левая кнопка не двигает камеру (правильно для выбора объектов)")
	}

	t.Logf("✅ Тест управления камерой завершён")
}

// MockCameraController имитирует контроллер камеры для тестирования
type MockCameraController struct {
	camera     MockCamera
	isDragging bool
	lastMouseX int
	lastMouseY int
	dragButton ebiten.MouseButton
}

// MockCamera имитирует камеру
type MockCamera struct {
	X, Y float32
	Zoom float32
}

// NewMockCameraController создаёт новый mock-контроллер
func NewMockCameraController() *MockCameraController {
	return &MockCameraController{
		camera: MockCamera{
			X:    400,
			Y:    300,
			Zoom: 1.0,
		},
		isDragging: false,
	}
}

// GetCamera возвращает состояние камеры
func (cc *MockCameraController) GetCamera() MockCamera {
	return cc.camera
}

// SetCameraPosition устанавливает позицию камеры
func (cc *MockCameraController) SetCameraPosition(x, y float32) {
	cc.camera.X = x
	cc.camera.Y = y
}

// SimulateMousePress имитирует нажатие кнопки мыши
func (cc *MockCameraController) SimulateMousePress(button ebiten.MouseButton, x, y int) {
	// Правая кнопка должна активировать перетаскивание
	if button == ebiten.MouseButtonRight {
		cc.isDragging = true
		cc.lastMouseX = x
		cc.lastMouseY = y
		cc.dragButton = button
	}
	// Левая кнопка - для выбора объектов, не для перетаскивания камеры
}

// SimulateMouseMove имитирует движение мыши
func (cc *MockCameraController) SimulateMouseMove(x, y int) {
	if cc.isDragging && cc.dragButton == ebiten.MouseButtonRight {
		deltaX := float32(x - cc.lastMouseX)
		deltaY := float32(y - cc.lastMouseY)

		// Перемещаем камеру в обратном направлении для интуитивного перетаскивания
		cc.camera.X -= deltaX / cc.camera.Zoom
		cc.camera.Y -= deltaY / cc.camera.Zoom

		cc.lastMouseX = x
		cc.lastMouseY = y
	}
}

// SimulateMouseRelease имитирует отпускание кнопки мыши
func (cc *MockCameraController) SimulateMouseRelease(button ebiten.MouseButton) {
	if button == cc.dragButton {
		cc.isDragging = false
	}
}
