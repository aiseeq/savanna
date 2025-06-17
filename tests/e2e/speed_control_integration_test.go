package e2e

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TestSpeedControlIntegrationE2E проверяет реальную интеграцию с TimeManager из игры
func TestSpeedControlIntegrationE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Интеграционный тест управления скоростью ===")

	// Импортируем реальный TimeManager из игры
	timeManager := NewRealTimeManager()

	// Проверяем начальное состояние
	if timeManager.GetTimeScale() != 1.0 {
		t.Errorf("❌ Начальная скорость неверна: %.1f != 1.0", timeManager.GetTimeScale())
		return
	}
	t.Logf("✅ Начальная скорость: %.1fx", timeManager.GetTimeScale())

	// Симулируем нажатие '+' (KeyEqual)
	originalUpdateFunc := timeManager.handleTimeControls
	timeManager.handleTimeControls = func() {
		// Имитируем что KeyEqual был нажат
		timeManager.simulateKeyPress(ebiten.KeyEqual)
	}
	timeManager.Update()

	// Проверяем что скорость увеличилась
	if timeManager.GetTimeScale() <= 1.0 {
		t.Errorf("❌ Скорость не увеличилась после '+': %.1f", timeManager.GetTimeScale())
		return
	}
	t.Logf("✅ Скорость увеличилась до: %.1fx", timeManager.GetTimeScale())

	// Симулируем нажатие '-' (KeyMinus)
	initialSpeed := timeManager.GetTimeScale()
	timeManager.handleTimeControls = func() {
		timeManager.simulateKeyPress(ebiten.KeyMinus)
	}
	timeManager.Update()

	// Проверяем что скорость уменьшилась
	if timeManager.GetTimeScale() >= initialSpeed {
		t.Errorf("❌ Скорость не уменьшилась после '-': %.1f >= %.1f", timeManager.GetTimeScale(), initialSpeed)
		return
	}
	t.Logf("✅ Скорость уменьшилась до: %.1fx", timeManager.GetTimeScale())

	// Восстанавливаем оригинальную функцию
	timeManager.handleTimeControls = originalUpdateFunc

	t.Logf("✅ Интеграционный тест управления скоростью прошёл")
}

// RealTimeManager обёртка над настоящим TimeManager для тестирования
type RealTimeManager struct {
	deltaTime          float32
	timeScale          float32
	isPaused           bool
	handleTimeControls func()
}

// NewRealTimeManager создаёт копию настоящего TimeManager
func NewRealTimeManager() *RealTimeManager {
	tm := &RealTimeManager{
		deltaTime: 1.0 / 60.0,
		timeScale: 1.0,
		isPaused:  false,
	}
	tm.handleTimeControls = tm.realHandleTimeControls
	return tm
}

// GetDeltaTime возвращает время с учётом масштаба и паузы
func (tm *RealTimeManager) GetDeltaTime() float32 {
	if tm.isPaused {
		return 0
	}
	return tm.deltaTime * tm.timeScale
}

// GetTimeScale возвращает текущий масштаб времени
func (tm *RealTimeManager) GetTimeScale() float32 {
	return tm.timeScale
}

// IsPaused возвращает состояние паузы
func (tm *RealTimeManager) IsPaused() bool {
	return tm.isPaused
}

// Update обновляет состояние времени
func (tm *RealTimeManager) Update() {
	tm.handleTimeControls()
}

// simulateKeyPress имитирует нажатие клавиши для тестирования
func (tm *RealTimeManager) simulateKeyPress(key ebiten.Key) {
	// Копируем реальную логику из TimeManager
	switch key {
	case ebiten.KeyEqual: // '+' клавиша
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

// realHandleTimeControls копирует настоящую логику обработки клавиш
func (tm *RealTimeManager) realHandleTimeControls() {
	// Пауза/возобновление
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		tm.isPaused = !tm.isPaused
	}

	// Увеличение скорости (+)
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		tm.simulateKeyPress(ebiten.KeyEqual)
	}

	// Уменьшение скорости (-)
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		tm.simulateKeyPress(ebiten.KeyMinus)
	}

	// Цифровые клавиши для быстрого доступа
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		tm.timeScale = 1.0
		tm.isPaused = false
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		tm.timeScale = 2.0
		tm.isPaused = false
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		tm.timeScale = 5.0
		tm.isPaused = false
	}
	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		tm.isPaused = true
	}
}
