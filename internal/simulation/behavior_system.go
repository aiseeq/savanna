package simulation

import (
	"math"

	"github.com/aiseeq/savanna/internal/core"
)

// VegetationProvider интерфейс для поиска растительности (устраняет нарушение DIP)
// Позволяет AnimalBehaviorSystem работать с любыми источниками пищи
type VegetationProvider interface {
	// FindNearestGrass ищет ближайшую траву в радиусе
	FindNearestGrass(worldX, worldY, searchRadius, minAmount float32) (grassX, grassY float32, found bool)
}

// AnimalBehaviorSystem управляет поведением животных через стратегии
// Универсальная система поведения через Strategy pattern (устраняет нарушение Open/Closed Principle)
type AnimalBehaviorSystem struct {
	// Время до смены направления для каждого животного (старый механизм)
	directionChangeTimers map[core.EntityID]float32
	// Абстракция для поиска растительности (соблюдение DIP)
	vegetation VegetationProvider
	// Стратегии поведения для разных типов животных (Strategy pattern)
	strategies map[core.BehaviorType]BehaviorStrategy
}

// NewAnimalBehaviorSystem создаёт новую систему поведения животных
// Принимает абстракцию VegetationProvider вместо конкретной реализации
func NewAnimalBehaviorSystem(vegetation VegetationProvider) *AnimalBehaviorSystem {
	abs := &AnimalBehaviorSystem{
		directionChangeTimers: make(map[core.EntityID]float32),
		vegetation:            vegetation,
		strategies:            make(map[core.BehaviorType]BehaviorStrategy),
	}

	// Инициализируем стратегии поведения (Strategy pattern)
	abs.strategies[core.BehaviorHerbivore] = NewHerbivoreBehaviorStrategy(vegetation)
	abs.strategies[core.BehaviorPredator] = NewPredatorBehaviorStrategy()
	// УДАЛЕНО: BehaviorScavenger - не используется в игре (нет падальщиков)

	return abs
}

// Update обновляет поведение всех животных через универсальную систему поведения
// Update обновляет поведение всех животных
// Рефакторинг: использует специализированный интерфейс вместо полного World (ISP)
func (abs *AnimalBehaviorSystem) Update(world core.BehaviorSystemAccess, deltaTime float32) {
	// Рефакторинг: используем только специализированный интерфейс

	// Обновляем таймеры поведения для всех животных с компонентом Behavior
	abs.updateBehaviorTimers(world, deltaTime)

	// Обрабатываем поведение всех животных через универсальную логику
	world.ForEachWith(core.MaskBehavior|core.MaskPosition|core.MaskVelocity|core.MaskSpeed|core.MaskHunger, func(entity core.EntityID) {
		abs.updateAnimalBehavior(world, entity, deltaTime)
	})

	// Очищаем таймеры для несуществующих сущностей
	abs.cleanupTimers(world)
}

// updateBehaviorTimers обновляет таймеры поведения для всех животных
func (abs *AnimalBehaviorSystem) updateBehaviorTimers(world core.BehaviorSystemAccess, deltaTime float32) {
	// Обновляем таймеры смены направления в старом стиле (для совместимости)
	for entityID, timeLeft := range abs.directionChangeTimers {
		abs.directionChangeTimers[entityID] = timeLeft - deltaTime
	}

	// Обновляем таймеры в компонентах Behavior
	world.ForEachWith(core.MaskBehavior, func(entity core.EntityID) {
		behavior, ok := world.GetBehavior(entity)
		if !ok {
			return
		}

		behavior.DirectionTimer -= deltaTime
		world.SetBehavior(entity, behavior)
	})
}

// updateAnimalBehavior универсальная система поведения животных (устраняет нарушение Open/Closed Principle)
func (abs *AnimalBehaviorSystem) updateAnimalBehavior(world core.BehaviorSystemAccess, entity core.EntityID, _ float32) {
	// ВАЖНО: Только атакующие животные не меняют поведение
	// Травоядные должны прекратить есть и убегать при виде хищника!
	if world.HasComponent(entity, core.MaskAttackState) {
		return // Животное атакует
	}

	behavior, ok := world.GetBehavior(entity)
	if !ok {
		return
	}

	pos, _ := world.GetPosition(entity)
	speed, _ := world.GetSpeed(entity)
	hunger, _ := world.GetHunger(entity)

	// Используем стратегию поведения (Strategy pattern)
	strategy, hasStrategy := abs.strategies[behavior.Type]
	if hasStrategy {
		targetVel := strategy.UpdateBehavior(world, entity, behavior, pos, speed, hunger)
		world.SetVelocity(entity, targetVel)
	} else {
		// Fallback для неизвестных типов поведения
		targetVel := RandomWalk.GetRandomWalkVelocity(world, entity, behavior, speed.Current*behavior.ContentSpeed)
		world.SetVelocity(entity, targetVel)
	}
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// getRandomWalkVelocity возвращает скорость для случайного блуждания
func (abs *AnimalBehaviorSystem) getRandomWalkVelocity(
	world core.BehaviorSystemAccess, entity core.EntityID, maxSpeed float32,
) core.Velocity {
	// Проверяем нужно ли сменить направление
	timeLeft, exists := abs.directionChangeTimers[entity]
	if !exists || timeLeft <= 0 {
		// Время сменить направление
		rng := world.GetRNG()

		// Случайный угол от 0 до 2π
		angle := rng.Float64() * 2 * math.Pi

		// Случайная скорость в диапазоне RandomSpeedMin до RandomSpeedMax
		speedMultiplier := RandomSpeedMin + rng.Float64()*(RandomSpeedMax-RandomSpeedMin)

		vel := core.Velocity{
			X: float32(math.Cos(angle)) * maxSpeed * float32(speedMultiplier),
			Y: float32(math.Sin(angle)) * maxSpeed * float32(speedMultiplier),
		}

		// Устанавливаем новый таймер
		newTime := RandomWalkMinTime + rng.Float64()*(RandomWalkMaxTime-RandomWalkMinTime)
		abs.directionChangeTimers[entity] = float32(newTime)

		return vel
	}

	// Сохраняем текущую скорость
	if world.HasComponent(entity, core.MaskVelocity) {
		vel, _ := world.GetVelocity(entity)
		return vel
	}

	return core.Velocity{X: 0, Y: 0}
}

// cleanupTimers очищает таймеры для несуществующих сущностей
func (abs *AnimalBehaviorSystem) cleanupTimers(world core.BehaviorSystemAccess) {
	for entityID := range abs.directionChangeTimers {
		if !world.IsAlive(entityID) {
			delete(abs.directionChangeTimers, entityID)
		}
	}
}
