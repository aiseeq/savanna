package main

import (
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Константы управления временем (устраняет магические числа)
const (
	NormalTimeScale    = 1.0 // Нормальная скорость времени
	FastTimeScale      = 2.0 // Быстрая скорость (2x)
	SuperFastTimeScale = 5.0 // Сверхбыстрая скорость (5x)
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
		deltaTime: constants.StandardDeltaTime, // Стандартный deltaTime (1/60)
		timeScale: 1.0,                         // Нормальная скорость
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
	tm.handlePauseToggle()
	tm.handleSpeedIncrease()
	tm.handleSpeedDecrease()
	tm.handleDirectSpeedKeys()
	tm.clampTimeScale()
}

// handlePauseToggle обрабатывает переключение паузы
func (tm *TimeManager) handlePauseToggle() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		tm.isPaused = !tm.isPaused
	}
}

// handleSpeedIncrease обрабатывает увеличение скорости
func (tm *TimeManager) handleSpeedIncrease() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEqual) && !inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		return
	}

	if tm.isPaused {
		tm.isPaused = false
		tm.timeScale = NormalTimeScale
		return
	}

	// Увеличиваем скорость: 0.25 -> 0.5 -> 1.0 -> 2.0 -> 4.0 -> 8.0
	switch {
	case tm.timeScale < constants.TimeScaleHalf:
		tm.timeScale = constants.TimeScaleHalf
	case tm.timeScale < NormalTimeScale:
		tm.timeScale = 1.0
	case tm.timeScale < constants.TimeScaleDouble:
		tm.timeScale = constants.TimeScaleDouble
	case tm.timeScale < constants.TimeScaleQuad:
		tm.timeScale = constants.TimeScaleQuad
	case tm.timeScale < constants.TimeScaleOcta:
		tm.timeScale = constants.TimeScaleOcta
	}
	// Максимум 8x
}

// handleSpeedDecrease обрабатывает уменьшение скорости
func (tm *TimeManager) handleSpeedDecrease() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyMinus) && !inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		return
	}

	switch {
	case tm.timeScale > constants.TimeScaleQuad:
		tm.timeScale = constants.TimeScaleQuad
	case tm.timeScale > constants.TimeScaleDouble:
		tm.timeScale = constants.TimeScaleDouble
	case tm.timeScale > NormalTimeScale:
		tm.timeScale = 1.0
	case tm.timeScale > constants.TimeScaleHalf:
		tm.timeScale = constants.TimeScaleHalf
	case tm.timeScale > constants.TimeScaleQuarter:
		tm.timeScale = constants.TimeScaleQuarter
	default:
		//nolint:gocritic // commentedOutCode: Это описательный комментарий, не код
		// Минимум = пауза
		tm.isPaused = true
	}
}

// handleDirectSpeedKeys обрабатывает прямые клавиши скорости
func (tm *TimeManager) handleDirectSpeedKeys() {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.Key1):
		tm.timeScale = NormalTimeScale
		tm.isPaused = false
	case inpututil.IsKeyJustPressed(ebiten.Key2):
		tm.timeScale = FastTimeScale
		tm.isPaused = false
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		tm.timeScale = SuperFastTimeScale
		tm.isPaused = false
	case inpututil.IsKeyJustPressed(ebiten.Key0):
		tm.isPaused = true // Пауза
	}
}

// clampTimeScale ограничивает масштаб времени допустимыми значениями
func (tm *TimeManager) clampTimeScale() {
	if tm.timeScale < constants.TimeScaleMinimum {
		tm.timeScale = constants.TimeScaleMinimum
	}
	if tm.timeScale > constants.TimeScaleMaximum {
		tm.timeScale = constants.TimeScaleMaximum
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
