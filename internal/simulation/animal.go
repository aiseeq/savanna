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
	VISION_RANGE_RABBIT = 100.0 // Увеличено в 2 раза
	VISION_RANGE_WOLF   = 200.0 // Увеличено в 2 раза

	// Параметры поведения
	WOLF_HUNGER_THRESHOLD = 60.0 // Волк начинает охотиться если голод < 60%
	RANDOM_WALK_MIN_TIME  = 2.0  // Минимальное время случайного движения
	RANDOM_WALK_MAX_TIME  = 5.0  // Максимальное время случайного движения
)

// AnimalBehaviorSystem управляет поведением животных
type AnimalBehaviorSystem struct {
	// Время до смены направления для каждого животного
	directionChangeTimers map[core.EntityID]float32
	// Ссылка на систему растительности для поиска травы
	vegetation *VegetationSystem
}

// NewAnimalBehaviorSystem создаёт новую систему поведения животных
func NewAnimalBehaviorSystem(vegetation *VegetationSystem) *AnimalBehaviorSystem {
	return &AnimalBehaviorSystem{
		directionChangeTimers: make(map[core.EntityID]float32),
		vegetation:            vegetation,
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

// updateRabbitBehavior обновляет поведение зайцев с новыми приоритетами
func (abs *AnimalBehaviorSystem) updateRabbitBehavior(world *core.World, deltaTime float32) {
	rabbits := world.QueryByType(core.TypeRabbit)

	for _, rabbit := range rabbits {
		if !world.HasComponents(rabbit, core.MaskPosition|core.MaskVelocity|core.MaskSpeed|core.MaskHunger) {
			continue
		}

		pos, _ := world.GetPosition(rabbit)
		speed, _ := world.GetSpeed(rabbit)
		hunger, _ := world.GetHunger(rabbit)

		var targetVel core.Velocity

		// ПРИОРИТЕТ 1: Если видит волка - убегать (всегда)
		nearestWolf, foundWolf := world.FindNearestByType(pos.X, pos.Y, VISION_RANGE_RABBIT, core.TypeWolf)
		if foundWolf {
			wolfPos, _ := world.GetPosition(nearestWolf)
			escapeDir := physics.Vec2{X: pos.X - wolfPos.X, Y: pos.Y - wolfPos.Y}
			escapeDir = escapeDir.Normalize()

			targetVel = core.Velocity{
				X: escapeDir.X * speed.Current,
				Y: escapeDir.Y * speed.Current,
			}
			abs.directionChangeTimers[rabbit] = RANDOM_WALK_MIN_TIME

		} else if hunger.Value < 70.0 && abs.vegetation != nil {
			// ПРИОРИТЕТ 2: Если голоден - идти к ближайшей траве
			grassX, grassY, foundGrass := abs.vegetation.FindNearestGrass(pos.X, pos.Y, VISION_RANGE_RABBIT, 10.0)
			if foundGrass {
				// Идём к траве
				grassDir := physics.Vec2{X: grassX - pos.X, Y: grassY - pos.Y}
				grassDir = grassDir.Normalize()

				targetVel = core.Velocity{
					X: grassDir.X * speed.Current * 0.8, // Немного медленнее при поиске еды
					Y: grassDir.Y * speed.Current * 0.8,
				}
				abs.directionChangeTimers[rabbit] = RANDOM_WALK_MIN_TIME
			} else {
				// Трава не найдена - продолжаем случайное движение в поисках
				targetVel = abs.getRandomWalkVelocity(world, rabbit, speed.Current*0.7)
			}

		} else {
			// ПРИОРИТЕТ 3: Если сыт - спокойное движение или отдых
			targetVel = abs.getRandomWalkVelocity(world, rabbit, speed.Current*0.3) // Медленно когда сыт
		}

		world.SetVelocity(rabbit, targetVel)
	}
}

// updateWolfBehavior обновляет поведение волков с приоритетами охоты
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

				// Вычисляем расстояние до зайца
				distance := physics.Vec2{X: rabbitPos.X - pos.X, Y: rabbitPos.Y - pos.Y}.Length()

				// Направление к зайцу
				huntDir := physics.Vec2{X: rabbitPos.X - pos.X, Y: rabbitPos.Y - pos.Y}
				huntDir = huntDir.Normalize()

				// Простое и надежное правило против перепрыгивания
				huntSpeed := speed.Current

				if distance <= 12.0 { // WOLF_ATTACK_DISTANCE
					rabbitSpeed, hasRabbitSpeed := world.GetSpeed(nearestRabbit)
					if hasRabbitSpeed && rabbitSpeed.Current <= 1.0 {
						// Цель почти неподвижна - полная остановка
						huntSpeed = 0
					} else if hasRabbitSpeed && speed.Current > rabbitSpeed.Current {
						// Обычное правило - снижаем до скорости цели
						huntSpeed = rabbitSpeed.Current
					}
				}

				targetVel = core.Velocity{
					X: huntDir.X * huntSpeed,
					Y: huntDir.Y * huntSpeed,
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
	world.AddHunger(entity, core.Hunger{Value: 50.0}) // Начинаем голодным для охоты
	world.AddAge(entity, core.Age{Seconds: 0})
	world.AddAnimalType(entity, core.TypeWolf)
	world.AddSize(entity, core.Size{Radius: WOLF_RADIUS})
	world.AddSpeed(entity, core.Speed{Base: WOLF_SPEED, Current: WOLF_SPEED})

	return entity
}
