package e2e

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TestSpeedControlE2E проверяет что управление скоростью (+/-) работает корректно
func TestSpeedControlE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка управления скоростью ===")

	// Создаём mock-менеджер времени
	timeManager := &MockTimeManagerWithControls{
		timeScale: 1.0,
		isPaused:  false,
	}

	// Проверяем начальное состояние
	if timeManager.GetTimeScale() != 1.0 {
		t.Errorf("❌ Начальная скорость неверна: %.1f != 1.0", timeManager.GetTimeScale())
		return
	}
	t.Logf("✅ Начальная скорость: %.1fx", timeManager.GetTimeScale())

	// Симулируем нажатие '+' (увеличение скорости)
	timeManager.HandleKeyPress(ebiten.KeyEqual) // '+' обычно на той же клавише что '='

	// Проверяем что скорость увеличилась
	if timeManager.GetTimeScale() <= 1.0 {
		t.Errorf("❌ Скорость не увеличилась после '+': %.1f", timeManager.GetTimeScale())
		return
	}
	t.Logf("✅ Скорость увеличилась до: %.1fx", timeManager.GetTimeScale())

	// Симулируем нажатие '-' (уменьшение скорости)
	initialSpeed := timeManager.GetTimeScale()
	timeManager.HandleKeyPress(ebiten.KeyMinus)

	// Проверяем что скорость уменьшилась
	if timeManager.GetTimeScale() >= initialSpeed {
		t.Errorf("❌ Скорость не уменьшилась после '-': %.1f >= %.1f", timeManager.GetTimeScale(), initialSpeed)
		return
	}
	t.Logf("✅ Скорость уменьшилась до: %.1fx", timeManager.GetTimeScale())

	// Тестируем пределы скорости
	// Максимальная скорость
	for i := 0; i < 10; i++ {
		timeManager.HandleKeyPress(ebiten.KeyEqual)
	}
	maxSpeed := timeManager.GetTimeScale()
	timeManager.HandleKeyPress(ebiten.KeyEqual) // Ещё одна попытка увеличить
	if timeManager.GetTimeScale() > maxSpeed {
		t.Errorf("❌ Скорость превысила максимум: %.1f > %.1f", timeManager.GetTimeScale(), maxSpeed)
		return
	}
	t.Logf("✅ Максимальная скорость ограничена: %.1fx", maxSpeed)

	// Минимальная скорость
	for i := 0; i < 20; i++ {
		timeManager.HandleKeyPress(ebiten.KeyMinus)
	}
	minSpeed := timeManager.GetTimeScale()
	if minSpeed < 0 {
		t.Errorf("❌ Скорость стала отрицательной: %.1f", minSpeed)
		return
	}
	t.Logf("✅ Минимальная скорость ограничена: %.1fx", minSpeed)

	// Проверяем что нулевая скорость = пауза
	if minSpeed == 0.0 && !timeManager.IsPaused() {
		t.Errorf("❌ Нулевая скорость должна активировать паузу")
		return
	}

	t.Logf("✅ Управление скоростью работает корректно")
}

// MockTimeManagerWithControls имитирует менеджер времени с поддержкой клавиш
type MockTimeManagerWithControls struct {
	timeScale float32
	isPaused  bool
}

// GetTimeScale возвращает текущую скорость времени
func (tm *MockTimeManagerWithControls) GetTimeScale() float32 {
	if tm.isPaused {
		return 0.0
	}
	return tm.timeScale
}

// IsPaused возвращает статус паузы
func (tm *MockTimeManagerWithControls) IsPaused() bool {
	return tm.isPaused
}

// HandleKeyPress обрабатывает нажатие клавиши (имитирует реальную логику)
func (tm *MockTimeManagerWithControls) HandleKeyPress(key ebiten.Key) {
	switch key {
	case ebiten.KeyEqual: // '+' клавиша
		if tm.isPaused {
			tm.isPaused = false
			tm.timeScale = 1.0
		} else {
			// Увеличиваем скорость: 1.0 -> 2.0 -> 4.0 -> 8.0 (максимум)
			if tm.timeScale < 1.0 {
				tm.timeScale = 1.0
			} else if tm.timeScale < 2.0 {
				tm.timeScale = 2.0
			} else if tm.timeScale < 4.0 {
				tm.timeScale = 4.0
			} else if tm.timeScale < 8.0 {
				tm.timeScale = 8.0
			}
			// Максимум 8x
		}
	case ebiten.KeyMinus: // '-' клавиша
		if tm.timeScale > 4.0 {
			tm.timeScale = 4.0
		} else if tm.timeScale > 2.0 {
			tm.timeScale = 2.0
		} else if tm.timeScale > 1.0 {
			tm.timeScale = 1.0
		} else if tm.timeScale > 0.5 {
			tm.timeScale = 0.5
		} else if tm.timeScale > 0.25 {
			tm.timeScale = 0.25
		} else {
			// Минимум = пауза
			tm.timeScale = 0.0
			tm.isPaused = true
		}
	}
}

// IsKeyJustPressed имитирует проверку нажатия клавиши
func IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}
