package simulation

import (
	"math"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

// Константы животных
const (
	// Размеры и скорости
	RABBIT_RADIUS = 5.0
	RABBIT_SPEED  = 20.0

	WOLF_RADIUS = 10.0
	WOLF_SPEED  = 30.0

	// Здоровье
	RABBIT_MAX_HEALTH = 50
	WOLF_MAX_HEALTH   = 100

	// Дальность видения
	VISION_RANGE_RABBIT = 50.0
	VISION_RANGE_WOLF   = 100.0

	// Параметры поведения
	WOLF_HUNGER_THRESHOLD = 60.0 // Волк начинает охотиться если голод < 60%
	RANDOM_WALK_MIN_TIME  = 2.0  // Минимальное время случайного движения
	RANDOM_WALK_MAX_TIME  = 5.0  // Максимальное время случайного движения
)

// AnimalBehaviorSystem управляет поведением животных
type AnimalBehaviorSystem struct {
	// Время до смены направления для каждого животного
	directionChangeTimers map[core.EntityID]float32
}

// NewAnimalBehaviorSystem создаёт новую систему поведения животных
func NewAnimalBehaviorSystem() *AnimalBehaviorSystem {
	return &AnimalBehaviorSystem{
		directionChangeTimers: make(map[core.EntityID]float32),
	}
}

// Update обновляет поведение всех животных
func (abs *AnimalBehaviorSystem) Update(world *core.World, deltaTime float32) {
	// Обновляем таймеры смены направления
	abs.updateDirectionTimers(deltaTime)

	// Обрабатываем поведение зайцев
	abs.updateRabbitBehavior(world, deltaTime)

	// Обрабатываем поведение волков
	abs.updateWolfBehavior(world, deltaTime)

	// Очищаем таймеры для несуществующих сущностей
	abs.cleanupTimers(world)
}

// updateDirectionTimers обновляет таймеры смены направления
func (abs *AnimalBehaviorSystem) updateDirectionTimers(deltaTime float32) {
	for entityID, timeLeft := range abs.directionChangeTimers {
		abs.directionChangeTimers[entityID] = timeLeft - deltaTime
	}
}

// updateRabbitBehavior обновляет поведение зайцев
func (abs *AnimalBehaviorSystem) updateRabbitBehavior(world *core.World, deltaTime float32) {
	rabbits := world.QueryByType(core.TypeRabbit)

	for _, rabbit := range rabbits {
		if !world.HasComponents(rabbit, core.MaskPosition|core.MaskVelocity|core.MaskSpeed) {
			continue
		}

		pos, _ := world.GetPosition(rabbit)
		speed, _ := world.GetSpeed(rabbit)

		// Ищем ближайшего волка
		nearestWolf, foundWolf := world.FindNearestByType(pos.X, pos.Y, VISION_RANGE_RABBIT, core.TypeWolf)

		var targetVel core.Velocity

		if foundWolf {
			// ПАНИКА! Убегаем от волка
			wolfPos, _ := world.GetPosition(nearestWolf)

			// Направление от волка к зайцу
			escapeDir := physics.Vec2{X: pos.X - wolfPos.X, Y: pos.Y - wolfPos.Y}
			escapeDir = escapeDir.Normalize()

			// Максимальная скорость убегания
			targetVel = core.Velocity{
				X: escapeDir.X * speed.Current,
				Y: escapeDir.Y * speed.Current,
			}

			// Сбрасываем таймер - в панике не меняем направление
			abs.directionChangeTimers[rabbit] = RANDOM_WALK_MIN_TIME

		} else {
			// Спокойное состояние - случайное блуждание
			targetVel = abs.getRandomWalkVelocity(world, rabbit, speed.Current*0.5) // Медленнее когда спокоен
		}

		world.SetVelocity(rabbit, targetVel)
	}
}

// updateWolfBehavior обновляет поведение волков
func (abs *AnimalBehaviorSystem) updateWolfBehavior(world *core.World, deltaTime float32) {
	wolves := world.QueryByType(core.TypeWolf)

	for _, wolf := range wolves {
		if !world.HasComponents(wolf, core.MaskPosition|core.MaskVelocity|core.MaskSpeed|core.MaskHunger) {
			continue
		}

		pos, _ := world.GetPosition(wolf)
		speed, _ := world.GetSpeed(wolf)
		hunger, _ := world.GetHunger(wolf)

		var targetVel core.Velocity

		if hunger.Value < WOLF_HUNGER_THRESHOLD {
			// Голоден - ищем добычу
			nearestRabbit, foundRabbit := world.FindNearestByType(pos.X, pos.Y, VISION_RANGE_WOLF, core.TypeRabbit)

			if foundRabbit {
				// Преследуем зайца
				rabbitPos, _ := world.GetPosition(nearestRabbit)

				// Направление к зайцу
				huntDir := physics.Vec2{X: rabbitPos.X - pos.X, Y: rabbitPos.Y - pos.Y}
				huntDir = huntDir.Normalize()

				// Максимальная скорость охоты
				targetVel = core.Velocity{
					X: huntDir.X * speed.Current,
					Y: huntDir.Y * speed.Current,
				}

				// Сбрасываем таймер - во время охоты не блуждаем
				abs.directionChangeTimers[wolf] = RANDOM_WALK_MIN_TIME

			} else {
				// Добычи не видно - медленное блуждание в поисках
				targetVel = abs.getRandomWalkVelocity(world, wolf, speed.Current*0.7)
			}

		} else {
			// Сыт - медленное блуждание
			targetVel = abs.getRandomWalkVelocity(world, wolf, speed.Current*0.3)
		}

		world.SetVelocity(wolf, targetVel)
	}
}

// getRandomWalkVelocity возвращает скорость для случайного блуждания
func (abs *AnimalBehaviorSystem) getRandomWalkVelocity(world *core.World, entity core.EntityID, maxSpeed float32) core.Velocity {
	// Проверяем нужно ли сменить направление
	timeLeft, exists := abs.directionChangeTimers[entity]
	if !exists || timeLeft <= 0 {
		// Время сменить направление
		rng := world.GetRNG()

		// Случайный угол от 0 до 2π
		angle := rng.Float64() * 2 * math.Pi

		// Случайная скорость от 0.5 до 1.0 от максимальной
		speedMultiplier := 0.5 + rng.Float64()*0.5

		vel := core.Velocity{
			X: float32(math.Cos(angle)) * maxSpeed * float32(speedMultiplier),
			Y: float32(math.Sin(angle)) * maxSpeed * float32(speedMultiplier),
		}

		// Устанавливаем новый таймер
		newTime := RANDOM_WALK_MIN_TIME + rng.Float64()*(RANDOM_WALK_MAX_TIME-RANDOM_WALK_MIN_TIME)
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
func (abs *AnimalBehaviorSystem) cleanupTimers(world *core.World) {
	for entityID := range abs.directionChangeTimers {
		if !world.IsAlive(entityID) {
			delete(abs.directionChangeTimers, entityID)
		}
	}
}

// CreateRabbit создаёт зайца в указанной позиции
func CreateRabbit(world *core.World, x, y float32) core.EntityID {
	entity := world.CreateEntity()

	world.AddPosition(entity, core.Position{X: x, Y: y})
	world.AddVelocity(entity, core.Velocity{X: 0, Y: 0})
	world.AddHealth(entity, core.Health{Current: RABBIT_MAX_HEALTH, Max: RABBIT_MAX_HEALTH})
	world.AddHunger(entity, core.Hunger{Value: 80.0}) // Начинаем с 80% сытости
	world.AddAge(entity, core.Age{Seconds: 0})
	world.AddAnimalType(entity, core.TypeRabbit)
	world.AddSize(entity, core.Size{Radius: RABBIT_RADIUS})
	world.AddSpeed(entity, core.Speed{Base: RABBIT_SPEED, Current: RABBIT_SPEED})

	return entity
}

// CreateWolf создаёт волка в указанной позиции
func CreateWolf(world *core.World, x, y float32) core.EntityID {
	entity := world.CreateEntity()

	world.AddPosition(entity, core.Position{X: x, Y: y})
	world.AddVelocity(entity, core.Velocity{X: 0, Y: 0})
	world.AddHealth(entity, core.Health{Current: WOLF_MAX_HEALTH, Max: WOLF_MAX_HEALTH})
	world.AddHunger(entity, core.Hunger{Value: 60.0}) // Начинаем немного голодным
	world.AddAge(entity, core.Age{Seconds: 0})
	world.AddAnimalType(entity, core.TypeWolf)
	world.AddSize(entity, core.Size{Radius: WOLF_RADIUS})
	world.AddSpeed(entity, core.Speed{Base: WOLF_SPEED, Current: WOLF_SPEED})

	return entity
}
