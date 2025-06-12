package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// Константы системы голода
const (
	HUNGER_DECREASE_RATE  = 0.2  // Уменьшение голода за один Update (1% за 5 сек при 60 TPS)
	HEALTH_DAMAGE_RATE    = 1.0  // Урон здоровью в секунду при голоде = 0
	SLOW_HUNGER_THRESHOLD = 20.0 // При голоде < 20% скорость x0.5
	FAST_HUNGER_THRESHOLD = 80.0 // При голоде > 80% скорость x0.8
)

// FeedingSystem управляет голодом и его влиянием на животных
type FeedingSystem struct {
	healthDamageTimer float32 // Таймер для нанесения урона здоровью
}

// NewFeedingSystem создаёт новую систему питания
func NewFeedingSystem() *FeedingSystem {
	return &FeedingSystem{
		healthDamageTimer: 0,
	}
}

// Update обновляет систему голода для всех животных
func (fs *FeedingSystem) Update(world *core.World, deltaTime float32) {
	fs.healthDamageTimer += deltaTime

	// Обновляем голод для всех животных
	world.ForEachWith(core.MaskHunger, func(entity core.EntityID) {
		fs.updateHunger(world, entity, deltaTime)
	})

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
}

// updateHunger обновляет голод животного
func (fs *FeedingSystem) updateHunger(world *core.World, entity core.EntityID, deltaTime float32) {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return
	}

	// Уменьшаем голод
	hunger.Value -= HUNGER_DECREASE_RATE * deltaTime

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
