package simulation

import (
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
)

// DEPRECATED: константы перенесены в game_balance.go
// Используйте LargeAnimalSizeThreshold и LargeAnimalSaitationRate

// FeedingSystem управляет сытостью и её влиянием на животных
// DEPRECATED: Эта система заменена на специализированные системы:
// - SatiationSystem: только управление сытостью
// - GrassSearchSystem: только поиск травы и создание EatingState
// - GrassEatingSystem: только обработка поедания травы
// - StarvationDamageSystem: только урон от истощения
// - SatiationSpeedModifierSystem: только влияние сытости на скорость
//
// Сохранена для совместимости с тестами. В новом коде используйте специализированные системы.
type FeedingSystem struct {
	healthDamageTimer float32                 // Таймер для нанесения урона здоровью
	vegetation        core.VegetationProvider // Интерфейс для работы с растительностью (соблюдение DIP)
}

// NewFeedingSystem создаёт новую систему питания
func NewFeedingSystem(vegetation core.VegetationProvider) *FeedingSystem {
	return &FeedingSystem{
		healthDamageTimer: 0,
		vegetation:        vegetation,
	}
}

// Update обновляет систему голода для всех животных
// Рефакторинг: использует специализированный интерфейс вместо полного World (ISP)
func (fs *FeedingSystem) Update(world core.SimulationAccess, deltaTime float32) {
	fs.healthDamageTimer += deltaTime

	// Обновляем сытость для всех животных
	world.ForEachWith(core.MaskSatiation, func(entity core.EntityID) {
		fs.updateSatiation(world, entity, deltaTime)
	})

	// Питание зайцев травой
	fs.handleRabbitFeeding(world, deltaTime)

	// ПРИМЕЧАНИЕ: Обновление скорости на основе голода перенесено в HungerSpeedModifierSystem (SRP)

	// Наносим урон здоровью голодающим животным (раз в секунду)
	if fs.healthDamageTimer >= 1.0 { // Раз в секунду
		fs.damageStarvingAnimals(world)
		fs.healthDamageTimer = 0
	}
}

// updateSatiation обновляет сытость животного
func (fs *FeedingSystem) updateSatiation(world core.SimulationAccess, entity core.EntityID, deltaTime float32) {
	hunger, ok := world.GetSatiation(entity)
	if !ok {
		return
	}

	// Определяем скорость снижения сытости в зависимости от размера животного
	hungerRate := float32(BaseSatiationDecreaseRate)
	if size, hasSize := world.GetSize(entity); hasSize {
		// Большие животные (хищники) теряют сытость медленнее
		if size.Radius > LargeAnimalSizeThreshold {
			hungerRate *= LargeAnimalSaitationRate
		}
	}

	// Уменьшаем сытость
	hunger.Value -= hungerRate * deltaTime

	// Ограничиваем снизу
	if hunger.Value < 0 {
		hunger.Value = 0
	}

	world.SetSatiation(entity, hunger)
}

// УДАЛЕНО: updateSpeedBasedOnHunger перенесено в HungerSpeedModifierSystem
//
// ПРИНЦИП SRP: FeedingSystem отвечает только за голод и питание
// Обновление скорости - ответственность HungerSpeedModifierSystem

// damageStarvingAnimals наносит урон здоровью животным с голодом = 0
func (fs *FeedingSystem) damageStarvingAnimals(world core.SimulationAccess) {
	world.ForEachWith(core.MaskSatiation|core.MaskHealth, func(entity core.EntityID) {
		hunger, ok1 := world.GetSatiation(entity)
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
	hunger, ok := world.GetSatiation(entity)
	if !ok {
		return false
	}

	hunger.Value += foodValue

	// Ограничиваем сверху
	if hunger.Value > MaxSatiationLimit {
		hunger.Value = MaxSatiationLimit
	}

	world.SetSatiation(entity, hunger)
	return true
}

// GetSatiationPercentage возвращает процент голода (0-100)
func GetSatiationPercentage(world core.ECSAccess, entity core.EntityID) float32 {
	hunger, ok := world.GetSatiation(entity)
	if !ok {
		return 0
	}
	return hunger.Value
}

// IsStarving проверяет голодает ли животное
func IsStarving(world core.ECSAccess, entity core.EntityID) bool {
	return GetSatiationPercentage(world, entity) <= 0
}

// IsHungry проверяет голодно ли животное (универсальная функция)
// Устраняет нарушение OCP - теперь использует AnimalConfig вместо захардкоженных типов
func IsHungry(world core.ECSAccess, entity core.EntityID) bool {
	hunger := GetSatiationPercentage(world, entity)

	// Получаем конфигурацию животного для порога голода (устраняет захардкоженные типы)
	if config, hasConfig := world.GetAnimalConfig(entity); hasConfig {
		return hunger < config.SatiationThreshold
	}

	// Fallback: используем умеренный порог
	return hunger < FallbackSatiationThreshold
}

// handleRabbitFeeding обрабатывает питание зайцев травой
func (fs *FeedingSystem) handleRabbitFeeding(world core.SimulationAccess, _ float32) {
	if fs.vegetation == nil {
		return
	}

	// Обрабатываем ВСЕХ травоядных животных (устраняет захардкоженность TypeRabbit)
	herbivoreMask := core.MaskBehavior | core.MaskAnimalConfig | core.MaskPosition | core.MaskSatiation
	world.ForEachWith(herbivoreMask, func(entity core.EntityID) {
		fs.processHerbivoreFeeding(world, entity)
	})
}

// processHerbivoreFeeding обрабатывает питание одного травоядного
func (fs *FeedingSystem) processHerbivoreFeeding(world core.SimulationAccess, entity core.EntityID) {
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
	hunger, hasHunger := world.GetSatiation(entity)
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
	world core.SimulationAccess, entity core.EntityID, hunger core.Satiation, config core.AnimalConfig,
) bool {
	// ИСПРАВЛЕНИЕ: Правильная логика гистерезиса для поедания
	isCurrentlyEating := world.HasComponent(entity, core.MaskEatingState)

	if isCurrentlyEating {
		// Если уже ест - прекращаем только при полном насыщении (99.9% с допуском для float32)
		const satietyThreshold = MaxSatiationLimit - constants.SatietyTolerance // Используем константы из game_balance.go
		if hunger.Value >= satietyThreshold {
			world.RemoveEatingState(entity)
			return false
		}
		return true
	}

	// Если не ест - начинаем есть только если голод < SatiationThreshold
	return hunger.Value < config.SatiationThreshold
}

// manageGrassEating управляет состоянием поедания травы
func (fs *FeedingSystem) manageGrassEating(world core.SimulationAccess, entity core.EntityID, pos core.Position) {
	// Проверка скорости не нужна - BehaviorSystem уже устанавливает скорость 0 для еды
	// MovementSystem проверяет EatingState и останавливает движение

	// Проверяем есть ли рядом трава для начала поедания
	grassAmount := fs.vegetation.GetGrassAt(pos.X, pos.Y)
	if grassAmount >= MinGrassAmountToFind {
		// Создаём состояние поедания для анимации (Target = 0 означает поедание травы)
		// НЕ даём сытость здесь - это будет делать GrassEatingSystem дискретно по завершении кадров анимации
		if !world.HasComponent(entity, core.MaskEatingState) {
			eatingState := core.EatingState{
				Target:          GrassEatingTarget,          // 0 = поедание травы (не сущность)
				TargetType:      core.EatingTargetGrass,     // Тип: поедание травы
				EatingProgress:  constants.InitialProgress,  // Начальный прогресс
				NutritionGained: constants.InitialNutrition, // Начальная питательность
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
