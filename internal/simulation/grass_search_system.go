package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// GrassSearchSystem ищет траву и создаёт EatingState для травоядных (SRP)
// Единственная ответственность: поиск травы и создание состояния поедания
type GrassSearchSystem struct {
	vegetation core.VegetationProvider // Интерфейс для работы с растительностью (соблюдение DIP)
}

// NewGrassSearchSystem создаёт новую систему поиска травы
func NewGrassSearchSystem(vegetation core.VegetationProvider) *GrassSearchSystem {
	return &GrassSearchSystem{
		vegetation: vegetation,
	}
}

// Update ищет траву для голодных травоядных
// ISP Улучшение: использует узкоспециализированный интерфейс
func (gss *GrassSearchSystem) Update(world core.GrassSearchSystemAccess, _ float32) {
	gss.handleRabbitFeeding(world)
}

// handleRabbitFeeding обрабатывает питание травоядных (в основном зайцев)
func (gss *GrassSearchSystem) handleRabbitFeeding(world core.GrassSearchSystemAccess) {
	// Обрабатываем всех животных с голодом и позицией
	world.ForEachWith(core.MaskHunger|core.MaskPosition, func(entity core.EntityID) {
		// Пропускаем животных которые уже едят
		if world.HasComponent(entity, core.MaskEatingState) {
			return
		}

		// Обрабатываем только травоядных
		gss.processHerbivoreFeeding(world, entity)
	})
}

// processHerbivoreFeeding обрабатывает питание конкретного травоядного
func (gss *GrassSearchSystem) processHerbivoreFeeding(world core.GrassSearchSystemAccess, entity core.EntityID) {
	hunger, hasHunger := world.GetHunger(entity)
	if !hasHunger {
		return
	}

	config, hasConfig := world.GetAnimalConfig(entity)
	if !hasConfig {
		return
	}

	// Проверяем является ли животное травоядным через тип поведения
	behavior, hasBehavior := world.GetBehavior(entity)
	if !hasBehavior {
		return
	}

	isHerbivore := behavior.Type == core.BehaviorHerbivore
	isHungry := hunger.Value < config.HungerThreshold

	if !isHerbivore || !isHungry {
		return
	}

	pos, hasPos := world.GetPosition(entity)
	if !hasPos {
		return
	}

	// Ищем и управляем поеданием травы
	gss.manageGrassEating(world, entity, pos)
}

// manageGrassEating управляет поеданием травы для конкретного животного
func (gss *GrassSearchSystem) manageGrassEating(
	world core.GrassSearchSystemAccess,
	entity core.EntityID,
	pos core.Position,
) {
	// Ищем ближайшую траву в радиусе 1 тайла (32 пикселя в старой системе)
	const grassSearchRadius = 1.0
	_, _, found := gss.vegetation.FindNearestGrass(pos.X, pos.Y, grassSearchRadius, MinGrassAmountToFind)

	if found {
		// Создаём состояние поедания травы
		eatingState := core.EatingState{
			Target:          GrassEatingTarget,      // 0 = поедание травы
			TargetType:      core.EatingTargetGrass, // Тип: поедание травы
			EatingProgress:  0.0,                    // Прогресс поедания
			NutritionGained: 0.0,                    // Полученная питательность
		}

		world.AddEatingState(entity, eatingState)
	}
}
