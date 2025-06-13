package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// Константы системы голода
const (
	HUNGER_DECREASE_RATE  = 2.0  // Уменьшение голода за один Update (~2% за секунду при 60 TPS)
	HEALTH_DAMAGE_RATE    = 1.0  // Урон здоровью в секунду при голоде = 0
	SLOW_HUNGER_THRESHOLD = 20.0 // При голоде < 20% скорость x0.5
	FAST_HUNGER_THRESHOLD = 80.0 // При голоде > 80% скорость x0.8

	// Константы питания
	GRASS_EATING_RATE       = 10.0 // Трава съедается за секунду
	GRASS_HUNGER_VALUE      = 20.0 // Восстановление голода от травы за секунду
	WOLF_ATTACK_DAMAGE      = 30   // Урон волка за атаку
	WOLF_HUNGER_FROM_RABBIT = 30.0 // Восстановление голода волка от зайца

	// Дистанции
	EATING_DISTANCE      = 15.0 // Дистанция для поедания травы
	WOLF_ATTACK_DISTANCE = 12.0 // Дистанция атаки волков (1.2x от радиуса волка 10.0)
	MIN_GRASS_AMOUNT     = 10.0 // Минимальное количество травы для еды

	// Интервалы атак
	WOLF_ATTACK_COOLDOWN = 1.0 // Волк атакует раз в секунду
)

// FeedingSystem управляет голодом и его влиянием на животных
type FeedingSystem struct {
	healthDamageTimer float32                   // Таймер для нанесения урона здоровью
	wolfAttackTimers  map[core.EntityID]float32 // Таймеры атак для волков
	vegetation        *VegetationSystem         // Ссылка на систему растительности
}

// NewFeedingSystem создаёт новую систему питания
func NewFeedingSystem(vegetation *VegetationSystem) *FeedingSystem {
	return &FeedingSystem{
		healthDamageTimer: 0,
		wolfAttackTimers:  make(map[core.EntityID]float32),
		vegetation:        vegetation,
	}
}

// Update обновляет систему голода для всех животных
func (fs *FeedingSystem) Update(world *core.World, deltaTime float32) {
	fs.healthDamageTimer += deltaTime

	// Обновляем таймеры атак волков
	fs.updateWolfAttackTimers(deltaTime)

	// Обновляем голод для всех животных
	world.ForEachWith(core.MaskHunger, func(entity core.EntityID) {
		fs.updateHunger(world, entity, deltaTime)
	})

	// Питание зайцев травой
	fs.handleRabbitFeeding(world, deltaTime)

	// Охота волков на зайцев
	fs.handleWolfHunting(world, deltaTime)

	// Обновляем скорости на основе голода
	world.ForEachWith(core.MaskHunger|core.MaskSpeed, func(entity core.EntityID) {
		fs.updateSpeedBasedOnHunger(world, entity)
	})

	// Наносим урон здоровью голодающим животным (раз в секунду)
	if fs.healthDamageTimer >= 1.0 {
		fs.damageStarvingAnimals(world)
		fs.healthDamageTimer = 0
	}

	// Удаляем мертвых животных
	fs.removeDeadAnimals(world)

	// Очищаем таймеры для мёртвых волков
	fs.cleanupWolfTimers(world)
}

// updateHunger обновляет голод животного
func (fs *FeedingSystem) updateHunger(world *core.World, entity core.EntityID, deltaTime float32) {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return
	}

	// Определяем скорость голода в зависимости от типа животного
	hungerRate := float32(HUNGER_DECREASE_RATE)
	if animalType, hasType := world.GetAnimalType(entity); hasType {
		if animalType == core.TypeWolf {
			hungerRate *= 0.5 // Волки голодают в 2 раза медленнее зайцев
		}
	}

	// Уменьшаем голод
	hunger.Value -= hungerRate * deltaTime

	// Ограничиваем снизу
	if hunger.Value < 0 {
		hunger.Value = 0
	}

	world.SetHunger(entity, hunger)
}

// updateSpeedBasedOnHunger обновляет скорость животного на основе уровня голода
func (fs *FeedingSystem) updateSpeedBasedOnHunger(world *core.World, entity core.EntityID) {
	hunger, ok1 := world.GetHunger(entity)
	speed, ok2 := world.GetSpeed(entity)

	if !ok1 || !ok2 {
		return
	}

	var speedMultiplier float32 = 1.0

	if hunger.Value < SLOW_HUNGER_THRESHOLD {
		// Очень голоден - медленно
		speedMultiplier = 0.5
	} else if hunger.Value > FAST_HUNGER_THRESHOLD {
		// Сыт - чуть медленнее
		speedMultiplier = 0.8
	}
	// При 20-80% голода - нормальная скорость (1.0)

	// Обновляем текущую скорость
	speed.Current = speed.Base * speedMultiplier
	world.SetSpeed(entity, speed)
}

// damageStarvingAnimals наносит урон здоровью животным с голодом = 0
func (fs *FeedingSystem) damageStarvingAnimals(world *core.World) {
	world.ForEachWith(core.MaskHunger|core.MaskHealth, func(entity core.EntityID) {
		hunger, ok1 := world.GetHunger(entity)
		health, ok2 := world.GetHealth(entity)

		if !ok1 || !ok2 {
			return
		}

		// Если голод = 0, наносим урон
		if hunger.Value <= 0 {
			health.Current -= int16(HEALTH_DAMAGE_RATE)

			// Не даем здоровью стать отрицательным
			if health.Current < 0 {
				health.Current = 0
			}

			world.SetHealth(entity, health)
		}
	})
}

// removeDeadAnimals удаляет животных с здоровьем <= 0
func (fs *FeedingSystem) removeDeadAnimals(world *core.World) {
	// Собираем ID мертвых животных
	var deadAnimals []core.EntityID

	world.ForEachWith(core.MaskHealth, func(entity core.EntityID) {
		health, ok := world.GetHealth(entity)
		if ok && health.Current <= 0 {
			deadAnimals = append(deadAnimals, entity)
		}
	})

	// Удаляем мертвых животных
	for _, entity := range deadAnimals {
		world.DestroyEntity(entity)
	}
}

// FeedAnimal восстанавливает голод животного (для будущего - поедание травы/добычи)
func FeedAnimal(world *core.World, entity core.EntityID, foodValue float32) bool {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return false
	}

	hunger.Value += foodValue

	// Ограничиваем сверху
	if hunger.Value > 100 {
		hunger.Value = 100
	}

	world.SetHunger(entity, hunger)
	return true
}

// GetHungerPercentage возвращает процент голода (0-100)
func GetHungerPercentage(world *core.World, entity core.EntityID) float32 {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return 0
	}
	return hunger.Value
}

// IsStarving проверяет голодает ли животное
func IsStarving(world *core.World, entity core.EntityID) bool {
	return GetHungerPercentage(world, entity) <= 0
}

// IsHungry проверяет голодно ли животное
func IsHungry(world *core.World, entity core.EntityID) bool {
	return GetHungerPercentage(world, entity) < 50
}

// handleRabbitFeeding обрабатывает питание зайцев травой
func (fs *FeedingSystem) handleRabbitFeeding(world *core.World, deltaTime float32) {
	if fs.vegetation == nil {
		return
	}

	// Обходим баг QueryByType - ищем зайцев вручную
	var rabbits []core.EntityID
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, ok := world.GetAnimalType(entity)
		if ok && animalType == core.TypeRabbit {
			rabbits = append(rabbits, entity)
		}
	})

	for _, rabbit := range rabbits {
		pos, hasPos := world.GetPosition(rabbit)
		if !hasPos {
			continue
		}

		// Проверяем голод зайца - едят только если голодны
		hunger, hasHunger := world.GetHunger(rabbit)
		if !hasHunger || hunger.Value >= 100.0 {
			continue // Заяц сыт - не ест
		}

		// Проверяем есть ли рядом трава
		grassAmount := fs.vegetation.GetGrassAt(pos.X, pos.Y)
		if grassAmount < MIN_GRASS_AMOUNT {
			continue
		}

		// Заяц ест траву
		grassToEat := GRASS_EATING_RATE * deltaTime
		consumedGrass := fs.vegetation.ConsumeGrassAt(pos.X, pos.Y, grassToEat)

		if consumedGrass > 0 {
			// Восстанавливаем голод пропорционально съеденной траве
			hungerToRestore := (consumedGrass / GRASS_EATING_RATE) * GRASS_HUNGER_VALUE
			FeedAnimal(world, rabbit, hungerToRestore)
		}
	}
}

// handleWolfHunting обрабатывает охоту волков на зайцев
func (fs *FeedingSystem) handleWolfHunting(world *core.World, deltaTime float32) {
	// Обходим баг QueryByType - ищем волков и зайцев вручную
	var wolves []core.EntityID
	var rabbits []core.EntityID

	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, ok := world.GetAnimalType(entity)
		if !ok {
			return
		}

		if animalType == core.TypeWolf {
			wolves = append(wolves, entity)
		} else if animalType == core.TypeRabbit {
			rabbits = append(rabbits, entity)
		}
	})

	if len(wolves) == 0 || len(rabbits) == 0 {
		return // Нет волков или зайцев
	}

	for _, wolf := range wolves {
		wolfPos, hasWolfPos := world.GetPosition(wolf)
		if !hasWolfPos {
			continue
		}

		// Проверяем голод волка - атакует только если голоден
		wolfHunger, hasHunger := world.GetHunger(wolf)
		if !hasHunger {
			continue // Нет компонента голода
		}
		if wolfHunger.Value > WOLF_HUNGER_THRESHOLD {
			continue // Волк сыт - не атакует
		}

		// Ищем ближайшего зайца в радиусе атаки
		for _, rabbit := range rabbits {
			rabbitPos, hasRabbitPos := world.GetPosition(rabbit)
			if !hasRabbitPos {
				continue
			}

			// Проверяем дистанцию
			dx := wolfPos.X - rabbitPos.X
			dy := wolfPos.Y - rabbitPos.Y
			distanceSquared := dx*dx + dy*dy

			if distanceSquared <= WOLF_ATTACK_DISTANCE*WOLF_ATTACK_DISTANCE {
				// Проверяем кулдаун атаки
				if fs.canWolfAttack(wolf) {
					// Волк атакует зайца
					fs.attackRabbit(world, wolf, rabbit)
					fs.setWolfAttackCooldown(wolf)
					break // Атакуем только одного зайца за раз
				}
			}
		}
	}
}

// updateWolfAttackTimers обновляет таймеры атак для всех волков
func (fs *FeedingSystem) updateWolfAttackTimers(deltaTime float32) {
	for wolfID, timeLeft := range fs.wolfAttackTimers {
		fs.wolfAttackTimers[wolfID] = timeLeft - deltaTime
		if fs.wolfAttackTimers[wolfID] < 0 {
			fs.wolfAttackTimers[wolfID] = 0
		}
	}
}

// cleanupWolfTimers очищает таймеры для несуществующих волков
func (fs *FeedingSystem) cleanupWolfTimers(world *core.World) {
	for wolfID := range fs.wolfAttackTimers {
		if !world.IsAlive(wolfID) {
			delete(fs.wolfAttackTimers, wolfID)
		}
	}
}

// canWolfAttack проверяет может ли волк атаковать (кулдаун прошёл)
func (fs *FeedingSystem) canWolfAttack(wolf core.EntityID) bool {
	timeLeft, exists := fs.wolfAttackTimers[wolf]
	return !exists || timeLeft <= 0
}

// setWolfAttackCooldown устанавливает кулдаун атаки для волка
func (fs *FeedingSystem) setWolfAttackCooldown(wolf core.EntityID) {
	fs.wolfAttackTimers[wolf] = WOLF_ATTACK_COOLDOWN
}

// attackRabbit волк атакует зайца
func (fs *FeedingSystem) attackRabbit(world *core.World, wolf, rabbit core.EntityID) {
	// Наносим урон зайцу
	rabbitHealth, hasHealth := world.GetHealth(rabbit)
	if !hasHealth {
		return
	}

	rabbitHealth.Current -= int16(WOLF_ATTACK_DAMAGE)
	if rabbitHealth.Current < 0 {
		rabbitHealth.Current = 0
	}

	world.SetHealth(rabbit, rabbitHealth)

	// Если заяц умер, волк восстанавливает голод
	if rabbitHealth.Current <= 0 {
		FeedAnimal(world, wolf, WOLF_HUNGER_FROM_RABBIT)
	}
}
