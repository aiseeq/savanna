package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// DEPRECATED: константы перенесены в game_balance.go
// Используйте LargeAnimalSizeThreshold и LargeAnimalHungerRate

// FeedingSystem управляет голодом и его влиянием на животных
type FeedingSystem struct {
	healthDamageTimer float32           // Таймер для нанесения урона здоровью
	vegetation        *VegetationSystem // Ссылка на систему растительности
}

// NewFeedingSystem создаёт новую систему питания
func NewFeedingSystem(vegetation *VegetationSystem) *FeedingSystem {
	return &FeedingSystem{
		healthDamageTimer: 0,
		vegetation:        vegetation,
	}
}

// Update обновляет систему голода для всех животных
// Рефакторинг: использует специализированный интерфейс вместо полного World (ISP)
func (fs *FeedingSystem) Update(world core.FeedingSystemAccess, deltaTime float32) {
	fs.healthDamageTimer += deltaTime

	// Обновляем голод для всех животных
	world.ForEachWith(core.MaskHunger, func(entity core.EntityID) {
		fs.updateHunger(world, entity, deltaTime)
	})

	// Питание зайцев травой
	fs.handleRabbitFeeding(world, deltaTime)

	// Обновляем скорости на основе голода
	world.ForEachWith(core.MaskHunger|core.MaskSpeed, func(entity core.EntityID) {
		fs.updateSpeedBasedOnHunger(world, entity)
	})

	// Наносим урон здоровью голодающим животным (раз в секунду)
	if fs.healthDamageTimer >= 1.0 { // Раз в секунду
		fs.damageStarvingAnimals(world)
		fs.healthDamageTimer = 0
	}
}

// updateHunger обновляет голод животного
func (fs *FeedingSystem) updateHunger(world core.FeedingSystemAccess, entity core.EntityID, deltaTime float32) {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return
	}

	// Определяем скорость голода в зависимости от размера животного
	hungerRate := float32(BaseHungerDecreaseRate)
	if size, hasSize := world.GetSize(entity); hasSize {
		// Большие животные (хищники) голодают медленнее
		if size.Radius > LargeAnimalSizeThreshold {
			hungerRate *= LargeAnimalHungerRate
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

// updateSpeedBasedOnHunger обновляет скорость животного на основе сытости и здоровья
// НОВАЯ ЛОГИКА (по требованию пользователя):
// 1. Голодные (< 80%) бегают с полной скоростью (1.0)
// 2. Сытые (> 80%) замедляются: скорость *= (1 + 0.8 - сытость/100)
// 3. Раненые: скорость *= (процент_здоровья / 100)
func (fs *FeedingSystem) updateSpeedBasedOnHunger(world core.FeedingSystemAccess, entity core.EntityID) {
	hunger, hasHunger := world.GetHunger(entity)
	speed, hasSpeed := world.GetSpeed(entity)

	if !hasHunger || !hasSpeed {
		return
	}

	var speedMultiplier float32 = 1.0

	// НОВАЯ ЛОГИКА 1: Сытость влияет на скорость только при > 80%
	if hunger.Value > OverfedSpeedThreshold {
		// Сытые животные замедляются: скорость *= (1 + 0.8 - сытость)
		// где сытость в долях от 1.0 (90% = 0.9, 95% = 0.95)
		satietyRatio := hunger.Value / 100.0       //nolint:gomnd // Переводим проценты в доли (90% → 0.9)
		speedMultiplier = 1.0 + 0.8 - satietyRatio //nolint:gomnd // Формула замедления сытых животных

		// Минимальная скорость не меньше 0.1 (для безопасности)
		if speedMultiplier < 0.1 { //nolint:gomnd // Минимальная скорость для безопасности
			speedMultiplier = 0.1 //nolint:gomnd // Минимальная скорость для безопасности
		}
	}
	// Голодные (< 80%) бегают с полной скоростью (speedMultiplier = 1.0)

	// НОВАЯ ЛОГИКА 2: Здоровье влияет на скорость линейно (только если теряет хиты)
	if health, hasHealth := world.GetHealth(entity); hasHealth {
		if health.Current < health.Max {
			// Раненое животное: скорость *= (процент_здоровья / 100)
			healthRatio := float32(health.Current) / float32(health.Max)
			speedMultiplier *= healthRatio
		}
		// Здоровые животные (100% хитов) не получают штрафа
	}

	// Обновляем текущую скорость
	speed.Current = speed.Base * speedMultiplier
	world.SetSpeed(entity, speed)
}

// damageStarvingAnimals наносит урон здоровью животным с голодом = 0
func (fs *FeedingSystem) damageStarvingAnimals(world core.FeedingSystemAccess) {
	world.ForEachWith(core.MaskHunger|core.MaskHealth, func(entity core.EntityID) {
		hunger, ok1 := world.GetHunger(entity)
		health, ok2 := world.GetHealth(entity)

		if !ok1 || !ok2 {
			return
		}

		// Если голод = 0, наносим урон
		if hunger.Value <= 0 {
			health.Current -= int16(BaseHealthDamageRate)

			// Не даем здоровью стать отрицательным
			if health.Current < 0 {
				health.Current = 0
			}

			world.SetHealth(entity, health)
		}
	})
}

// FeedAnimal восстанавливает голод животного (для будущего - поедание травы/добычи)
func FeedAnimal(world core.ECSAccess, entity core.EntityID, foodValue float32) bool {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return false
	}

	hunger.Value += foodValue

	// Ограничиваем сверху
	if hunger.Value > MaxHungerLimit {
		hunger.Value = MaxHungerLimit
	}

	world.SetHunger(entity, hunger)
	return true
}

// GetHungerPercentage возвращает процент голода (0-100)
func GetHungerPercentage(world core.ECSAccess, entity core.EntityID) float32 {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return 0
	}
	return hunger.Value
}

// IsStarving проверяет голодает ли животное
func IsStarving(world core.ECSAccess, entity core.EntityID) bool {
	return GetHungerPercentage(world, entity) <= 0
}

// IsHungry проверяет голодно ли животное (универсальная функция)
// Устраняет нарушение OCP - теперь использует AnimalConfig вместо захардкоженных типов
func IsHungry(world core.ECSAccess, entity core.EntityID) bool {
	hunger := GetHungerPercentage(world, entity)

	// Получаем конфигурацию животного для порога голода (устраняет захардкоженные типы)
	if config, hasConfig := world.GetAnimalConfig(entity); hasConfig {
		return hunger < config.HungerThreshold
	}

	// Fallback: используем умеренный порог
	return hunger < 75.0 //nolint:gomnd // Fallback порог голода для неизвестных животных
}

// handleRabbitFeeding обрабатывает питание зайцев травой
func (fs *FeedingSystem) handleRabbitFeeding(world core.FeedingSystemAccess, _ float32) {
	if fs.vegetation == nil {
		return
	}

	// Обрабатываем ВСЕХ травоядных животных (устраняет захардкоженность TypeRabbit)
	herbivoreMask := core.MaskBehavior | core.MaskAnimalConfig | core.MaskPosition | core.MaskHunger
	world.ForEachWith(herbivoreMask, func(entity core.EntityID) {
		fs.processHerbivoreFeeding(world, entity)
	})
}

// processHerbivoreFeeding обрабатывает питание одного травоядного
func (fs *FeedingSystem) processHerbivoreFeeding(world core.FeedingSystemAccess, entity core.EntityID) {
	// Проверяем что это травоядное
	behavior, hasBehavior := world.GetBehavior(entity)
	if !hasBehavior || behavior.Type != core.BehaviorHerbivore {
		return
	}

	pos, hasPos := world.GetPosition(entity)
	if !hasPos {
		return
	}

	// Проверяем голод животного
	hunger, hasHunger := world.GetHunger(entity)
	if !hasHunger {
		return
	}

	// Получаем конфигурацию для порога голода (устраняет RabbitHungryThreshold)
	config, hasConfig := world.GetAnimalConfig(entity)
	if !hasConfig {
		return
	}

	// Проверяем логику гистерезиса для поедания
	if !fs.shouldContinueOrStartEating(world, entity, hunger, config) {
		return
	}

	// Проверяем наличие травы и управляем состоянием поедания
	fs.manageGrassEating(world, entity, pos)
}

// shouldContinueOrStartEating проверяет должно ли животное продолжать или начать есть
func (fs *FeedingSystem) shouldContinueOrStartEating(
	world core.FeedingSystemAccess, entity core.EntityID, hunger core.Hunger, config core.AnimalConfig,
) bool {
	// ИСПРАВЛЕНИЕ: Правильная логика гистерезиса для поедания
	isCurrentlyEating := world.HasComponent(entity, core.MaskEatingState)

	if isCurrentlyEating {
		// Если уже ест - прекращаем только при полном насыщении (99.9% с допуском для float32)
		const satietyThreshold = MaxHungerLimit - SatietyTolerance // Используем константы из game_balance.go
		if hunger.Value >= satietyThreshold {
			world.RemoveEatingState(entity)
			return false
		}
		return true
	}

	// Если не ест - начинаем есть только если голод < HungerThreshold
	return hunger.Value < config.HungerThreshold
}

// manageGrassEating управляет состоянием поедания травы
func (fs *FeedingSystem) manageGrassEating(world core.FeedingSystemAccess, entity core.EntityID, pos core.Position) {
	// Проверка скорости не нужна - BehaviorSystem уже устанавливает скорость 0 для еды
	// MovementSystem проверяет EatingState и останавливает движение

	// Проверяем есть ли рядом трава для начала поедания
	grassAmount := fs.vegetation.GetGrassAt(pos.X, pos.Y)
	if grassAmount >= MinGrassAmountToFind {
		// Создаём состояние поедания для анимации (Target = 0 означает поедание травы)
		// НЕ даём сытость здесь - это будет делать GrassEatingSystem дискретно по завершении кадров анимации
		if !world.HasComponent(entity, core.MaskEatingState) {
			eatingState := core.EatingState{
				Target:          GrassEatingTarget, // 0 = поедание травы (не сущность)
				EatingProgress:  0.0,               //nolint:gomnd // Начальный прогресс
				NutritionGained: 0.0,               //nolint:gomnd // Начальная питательность
			}
			world.AddEatingState(entity, eatingState)
		}
	} else {
		// Нет травы - убираем состояние поедания
		if world.HasComponent(entity, core.MaskEatingState) {
			world.RemoveEatingState(entity)
		}
	}
}
