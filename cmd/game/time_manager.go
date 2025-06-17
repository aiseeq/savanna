package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TimeManager управляет временем симуляции
// Соблюдает SRP - единственная ответственность: управление временем
type TimeManager struct {
	deltaTime float32 // Время с последнего кадра
	timeScale float32 // Масштаб времени (1.0 = нормально, 2.0 = в 2 раза быстрее)

	// Состояние паузы
	isPaused bool
}

// NewTimeManager создаёт новый менеджер времени
func NewTimeManager() *TimeManager {
	return &TimeManager{
		deltaTime: 1.0 / 60.0, // 60 FPS
		timeScale: 1.0,        // Нормальная скорость
		isPaused:  false,
	}
}

// GetDeltaTime возвращает время с учётом масштаба и паузы
func (tm *TimeManager) GetDeltaTime() float32 {
	if tm.isPaused {
		return 0
	}
	return tm.deltaTime * tm.timeScale
}

// GetTimeScale возвращает текущий масштаб времени
func (tm *TimeManager) GetTimeScale() float32 {
	return tm.timeScale
}

// IsPaused возвращает состояние паузы
func (tm *TimeManager) IsPaused() bool {
	return tm.isPaused
}

// Update обновляет состояние времени на основе ввода
func (tm *TimeManager) Update() {
	tm.handleTimeControls()
}

// handleTimeControls обрабатывает управление временем
func (tm *TimeManager) handleTimeControls() {
	// Пауза/возобновление
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		tm.isPaused = !tm.isPaused
	}

	// Увеличение скорости (+ и NumpadAdd)
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
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
			// Максимум 8x
		}
	}

	// Уменьшение скорости (- и NumpadSubtract)
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
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
			tm.isPaused = true
		}
	}

	// Масштаб времени (цифровые клавиши для быстрого доступа)
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		tm.timeScale = 1.0 // Нормальная скорость
		tm.isPaused = false
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		tm.timeScale = 2.0 // 2x скорость
		tm.isPaused = false
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		tm.timeScale = 5.0 // 5x скорость
		tm.isPaused = false
	}
	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		tm.isPaused = true // Пауза
	}

	// Ограничиваем масштаб времени
	if tm.timeScale < 0.1 {
		tm.timeScale = 0.1
	}
	if tm.timeScale > 10.0 {
		tm.timeScale = 10.0
	}
}

// SetTimeScale устанавливает масштаб времени
func (tm *TimeManager) SetTimeScale(scale float32) {
	if scale <= 0 {
		tm.isPaused = true
	} else {
		tm.timeScale = scale
		tm.isPaused = false
	}
}

// TogglePause переключает состояние паузы
func (tm *TimeManager) TogglePause() {
	tm.isPaused = !tm.isPaused
}
