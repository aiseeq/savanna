package simulation

import (
	"math"

	"github.com/aiseeq/savanna/internal/core"
)

// RandomWalkUtility утилита для случайного блуждания (устраняет дублирование кода)
// Реализует принцип DRY - одна функция используется всеми стратегиями поведения
type RandomWalkUtility struct{}

// NewRandomWalkUtility создаёт новую утилиту случайного блуждания
func NewRandomWalkUtility() *RandomWalkUtility {
	return &RandomWalkUtility{}
}

// GetRandomWalkVelocity возвращает скорость для случайного блуждания с использованием компонента Behavior
// Устраняет дублирование одинаковой логики в 4 местах кодовой базы
func (rwu *RandomWalkUtility) GetRandomWalkVelocity(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	behavior core.Behavior,
	maxSpeed float32,
) core.Velocity {
	// Проверяем нужно ли сменить направление по таймеру в поведении
	if behavior.DirectionTimer <= 0 {
		// Время сменить направление
		rng := world.GetRNG()

		// Случайный угол от 0 до 2π
		angle := rng.Float64() * 2 * math.Pi

		// Случайная скорость (используем константы из game_balance.go)
		speedMultiplier := RandomSpeedMinMultiplier + rng.Float64()*(RandomSpeedMaxMultiplier-RandomSpeedMinMultiplier)

		velX := float32(math.Cos(angle)) * maxSpeed * float32(speedMultiplier)
		velY := float32(math.Sin(angle)) * maxSpeed * float32(speedMultiplier)
		vel := core.NewVelocity(velX, velY)

		// Устанавливаем новый таймер в поведении
		newTime := behavior.MinDirectionTime + float32(rng.Float64())*(behavior.MaxDirectionTime-behavior.MinDirectionTime)
		behavior.DirectionTimer = newTime
		world.SetBehavior(entity, behavior)

		return vel
	}

	// Сохраняем текущую скорость
	if world.HasComponent(entity, core.MaskVelocity) {
		vel, _ := world.GetVelocity(entity)
		return vel
	}

	return core.NewVelocity(0, 0)
}

// Глобальный экземпляр утилиты для использования во всех стратегиях
var RandomWalk = NewRandomWalkUtility()
