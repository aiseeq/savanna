package simulation

import (
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
	behaviorMask := core.MaskBehavior | core.MaskPosition | core.MaskVelocity | core.MaskSpeed | core.MaskSatiation | core.MaskAnimalConfig
	world.ForEachWith(behaviorMask, func(entity core.EntityID) {
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

// updateAnimalBehavior универсальная система поведения животных
// (устраняет нарушение Open/Closed Principle)
func (abs *AnimalBehaviorSystem) updateAnimalBehavior(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	_ float32,
) {
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
	satiation, _ := world.GetSatiation(entity)
	animalConfig, _ := world.GetAnimalConfig(entity)

	// Используем стратегию поведения (Strategy pattern)
	strategy, hasStrategy := abs.strategies[behavior.Type]
	if hasStrategy {
		components := AnimalComponents{
			Behavior:     behavior,
			AnimalConfig: animalConfig,
			Position:     pos,
			Speed:        speed,
			Satiation:    satiation,
		}
		targetVel := strategy.UpdateBehavior(world, entity, components)
		world.SetVelocity(entity, targetVel)
	} else {
		// Fallback для неизвестных типов поведения - используем AnimalConfig если есть
		var speed_multiplier float32 = behavior.ContentSpeed // Fallback к старому значению
		if animalConfig.ContentSpeed > 0 {
			speed_multiplier = animalConfig.ContentSpeed
		}
		targetVel := RandomWalk.GetRandomWalkVelocity(world, entity, behavior, speed.Current*speed_multiplier)
		world.SetVelocity(entity, targetVel)
	}
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// cleanupTimers очищает таймеры для несуществующих сущностей
func (abs *AnimalBehaviorSystem) cleanupTimers(world core.BehaviorSystemAccess) {
	for entityID := range abs.directionChangeTimers {
		if !world.IsAlive(entityID) {
			delete(abs.directionChangeTimers, entityID)
		}
	}
}
