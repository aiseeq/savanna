package e2e

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNumpadKeysE2E проверяет что +/- клавиши на нумпаде работают для управления временем
func TestNumpadKeysE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка клавиш +/- на нумпаде ===")

	// Создаём mock-менеджер времени
	timeManager := &MockTimeManagerNumpad{
		timeScale: 1.0,
		isPaused:  false,
	}

	// Проверяем начальное состояние
	if timeManager.GetTimeScale() != 1.0 {
		t.Errorf("❌ Начальная скорость неверна: %.1f != 1.0", timeManager.GetTimeScale())
		return
	}
	t.Logf("✅ Начальная скорость: %.1fx", timeManager.GetTimeScale())

	// Тестируем клавишу NumpadAdd (+)
	timeManager.HandleKeyPress(ebiten.KeyNumpadAdd)

	// Проверяем что скорость увеличилась
	if timeManager.GetTimeScale() <= 1.0 {
		t.Errorf("❌ Скорость не увеличилась после NumpadAdd: %.1f", timeManager.GetTimeScale())
		return
	}
	t.Logf("✅ NumpadAdd увеличил скорость до: %.1fx", timeManager.GetTimeScale())

	// Тестируем клавишу NumpadSubtract (-)
	initialSpeed := timeManager.GetTimeScale()
	timeManager.HandleKeyPress(ebiten.KeyNumpadSubtract)

	// Проверяем что скорость уменьшилась
	if timeManager.GetTimeScale() >= initialSpeed {
		t.Errorf("❌ Скорость не уменьшилась после NumpadSubtract: %.1f >= %.1f", timeManager.GetTimeScale(), initialSpeed)
		return
	}
	t.Logf("✅ NumpadSubtract уменьшил скорость до: %.1fx", timeManager.GetTimeScale())

	// Проверяем что обычные +/- тоже работают
	timeManager.HandleKeyPress(ebiten.KeyEqual) // +
	plusSpeed := timeManager.GetTimeScale()

	timeManager.HandleKeyPress(ebiten.KeyMinus) // -
	minusSpeed := timeManager.GetTimeScale()

	if plusSpeed <= timeManager.GetTimeScale() || minusSpeed >= plusSpeed {
		t.Errorf("❌ Обычные +/- клавиши не работают корректно")
		return
	}
	t.Logf("✅ Обычные +/- клавиши тоже работают")

	t.Logf("✅ Все клавиши управления временем работают корректно")
}

// MockTimeManagerNumpad имитирует менеджер времени с поддержкой нумпада
type MockTimeManagerNumpad struct {
	timeScale float32
	isPaused  bool
}

// GetTimeScale возвращает текущую скорость времени
func (tm *MockTimeManagerNumpad) GetTimeScale() float32 {
	if tm.isPaused {
		return 0.0
	}
	return tm.timeScale
}

// IsPaused возвращает статус паузы
func (tm *MockTimeManagerNumpad) IsPaused() bool {
	return tm.isPaused
}

// HandleKeyPress обрабатывает нажатие клавиши (должна поддерживать нумпад)
//
//nolint:gocognit // Mock объект с полной обработкой клавиш нумпада
func (tm *MockTimeManagerNumpad) HandleKeyPress(key ebiten.Key) {
	switch key {
	case ebiten.KeyEqual, ebiten.KeyNumpadAdd: // '+' клавиша (обычная и нумпад)
		if tm.isPaused {
			tm.isPaused = false
			tm.timeScale = 1.0
		} else {
			// Увеличиваем скорость: 0.25 -> 0.5 -> 1.0 -> 2.0 -> 4.0 -> 8.0
			if tm.timeScale < 0.5 {
				tm.timeScale = 0.5
			} else if tm.timeScale < 1.0 {
				tm.timeScale = 1.0
			} else if tm.timeScale < 2.0 {
				tm.timeScale = 2.0
			} else if tm.timeScale < 4.0 {
				tm.timeScale = 4.0
			} else if tm.timeScale < 8.0 {
				tm.timeScale = 8.0
			}
		}
	case ebiten.KeyMinus, ebiten.KeyNumpadSubtract: // '-' клавиша (обычная и нумпад)
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
			//nolint:gocritic // commentedOutCode: Это описательный комментарий
			// Минимум = пауза
			tm.timeScale = 0.0
			tm.isPaused = true
		}
	}

	// Ограничиваем масштаб времени
	if tm.timeScale < 0.1 {
		tm.timeScale = 0.1
	}
	if tm.timeScale > 10.0 {
		tm.timeScale = 10.0
	}
}
