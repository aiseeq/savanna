package simulation

import (
	"math"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

// BehaviorStrategy интерфейс для различных стратегий поведения животных
// Реализация Strategy pattern для устранения нарушения Open/Closed Principle
type BehaviorStrategy interface {
	// UpdateBehavior обновляет поведение животного и возвращает целевую скорость
	UpdateBehavior(world core.BehaviorSystemAccess, entity core.EntityID, components AnimalComponents) core.Velocity
}

// HerbivoreBehaviorStrategy стратегия поведения травоядных
type HerbivoreBehaviorStrategy struct {
	vegetation VegetationProvider
}

// NewHerbivoreBehaviorStrategy создаёт новую стратегию травоядных
func NewHerbivoreBehaviorStrategy(vegetation VegetationProvider) *HerbivoreBehaviorStrategy {
	return &HerbivoreBehaviorStrategy{
		vegetation: vegetation,
	}
}

// AnimalComponents группирует компоненты животного для поведения
type AnimalComponents struct {
	Behavior     core.Behavior
	AnimalConfig core.AnimalConfig
	Position     core.Position
	Speed        core.Speed
	Hunger       core.Hunger
}

// UpdateBehavior реализует поведение травоядных (KISS: упрощено разбиением на методы)
func (h *HerbivoreBehaviorStrategy) UpdateBehavior(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) core.Velocity {
	// ПРИОРИТЕТ 1: Если видит хищника - убегать (всегда)
	if velocity := h.handlePredatorEscape(world, entity, components); velocity != nil {
		return *velocity
	}

	// ПРИОРИТЕТ 2: Если голоден ИЛИ уже ест - обрабатываем поедание травы
	if velocity := h.handleFeeding(world, entity, components); velocity != nil {
		return *velocity
	}

	// ПРИОРИТЕТ 3: Если сыт - спокойное движение или отдых
	return h.handleIdleBehavior(world, entity, components)
}

// handlePredatorEscape обрабатывает побег от хищника (KISS: выделено в отдельный метод)
func (h *HerbivoreBehaviorStrategy) handlePredatorEscape(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) *core.Velocity {
	nearestPredator, foundPredator := world.FindNearestByTypeInTiles(
		components.Position.X, components.Position.Y,
		components.AnimalConfig.VisionRange, core.TypeWolf, // РЕФАКТОРИНГ: используем AnimalConfig вместо Behavior
	)
	if !foundPredator {
		return nil // Хищника нет
	}

	// КРИТИЧЕСКИ ВАЖНО: прерываем поедание травы при побеге
	if world.HasComponent(entity, core.MaskEatingState) {
		world.RemoveEatingState(entity)
	}

	predatorPos, _ := world.GetPosition(nearestPredator)

	// Базовое направление побега (от хищника)
	escapeDir := physics.Vec2{X: components.Position.X - predatorPos.X, Y: components.Position.Y - predatorPos.Y}
	escapeDir = escapeDir.Normalize()

	// ИСПРАВЛЕНИЕ: Добавляем отталкивание от границ мира для предотвращения кластеризации в углах
	worldWidth, worldHeight := world.GetWorldDimensions()
	boundaryRepulsion := h.calculateBoundaryRepulsion(components.Position, worldWidth, worldHeight)

	// Комбинируем направление побега с отталкиванием от границ
	finalEscapeDir := physics.Vec2{
		X: escapeDir.X + boundaryRepulsion.X,
		Y: escapeDir.Y + boundaryRepulsion.Y,
	}
	finalEscapeDir = finalEscapeDir.Normalize()

	// Обновляем таймер направления в поведении используя значения из AnimalConfig
	components.Behavior.DirectionTimer = components.AnimalConfig.MinDirectionTime
	world.SetBehavior(entity, components.Behavior)

	return &core.Velocity{
		X: finalEscapeDir.X * components.Speed.Current,
		Y: finalEscapeDir.Y * components.Speed.Current,
	}
}

// handleFeeding обрабатывает поиск и поедание травы (KISS: выделено в отдельный метод)
func (h *HerbivoreBehaviorStrategy) handleFeeding(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) *core.Velocity {
	isCurrentlyEating := world.HasComponent(entity, core.MaskEatingState)
	if !(components.Hunger.Value < components.AnimalConfig.HungerThreshold || isCurrentlyEating) || h.vegetation == nil {
		return nil // Не голоден и не ест
	}

	// Проверяем находимся ли мы рядом с травой используя AnimalConfig из компонентов
	if velocity := h.checkLocalGrass(components, components.AnimalConfig); velocity != nil {
		return velocity
	}

	// Ищем траву в радиусе видимости
	return h.searchForGrass(world, entity, components)
}

// checkLocalGrass проверяет траву рядом с животным (KISS: выделено в отдельный метод)
func (h *HerbivoreBehaviorStrategy) checkLocalGrass(
	components AnimalComponents,
	config core.AnimalConfig,
) *core.Velocity {
	localGrassX, localGrassY, hasLocalGrass := h.vegetation.FindNearestGrass(
		components.Position.X, components.Position.Y,
		config.CollisionRadius, MinGrassAmountToFind,
	)
	if !hasLocalGrass {
		return nil
	}

	dx := components.Position.X - localGrassX
	dy := components.Position.Y - localGrassY
	distanceToLocalGrass := math.Sqrt(float64(dx*dx + dy*dy))

	if distanceToLocalGrass <= float64(config.CollisionRadius*GrassProximityMultiplier) {
		// Мы рядом с травой - останавливаемся и едим
		return &core.Velocity{X: 0, Y: 0}
	}

	return nil
}

// searchForGrass ищет траву в радиусе видимости (KISS: выделено в отдельный метод)
func (h *HerbivoreBehaviorStrategy) searchForGrass(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) *core.Velocity {
	grassX, grassY, foundGrass := h.vegetation.FindNearestGrass(
		components.Position.X, components.Position.Y,
		components.AnimalConfig.VisionRange, MinGrassAmountToFind,
	)
	if foundGrass {
		// Идём к траве
		grassDir := physics.Vec2{X: grassX - components.Position.X, Y: grassY - components.Position.Y}
		grassDir = grassDir.Normalize()

		components.Behavior.DirectionTimer = components.AnimalConfig.MinDirectionTime
		world.SetBehavior(entity, components.Behavior)

		return &core.Velocity{
			X: grassDir.X * components.Speed.Current * components.AnimalConfig.SearchSpeed,
			Y: grassDir.Y * components.Speed.Current * components.AnimalConfig.SearchSpeed,
		}
	}

	// Трава не найдена - продолжаем случайное движение в поисках
	vel := RandomWalk.GetRandomWalkVelocity(
		world, entity, components.Behavior,
		components.Speed.Current*components.AnimalConfig.WanderingSpeed,
	)
	return &vel
}

// handleIdleBehavior обрабатывает спокойное поведение (KISS: выделено в отдельный метод)
func (h *HerbivoreBehaviorStrategy) handleIdleBehavior(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) core.Velocity {
	return RandomWalk.GetRandomWalkVelocity(
		world, entity, components.Behavior,
		components.Speed.Current*components.AnimalConfig.ContentSpeed,
	)
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// PredatorBehaviorStrategy стратегия поведения хищников
type PredatorBehaviorStrategy struct{}

// NewPredatorBehaviorStrategy создаёт новую стратегию хищников
func NewPredatorBehaviorStrategy() *PredatorBehaviorStrategy {
	return &PredatorBehaviorStrategy{}
}

// UpdateBehavior реализует поведение хищников (заменяет updatePredatorBehavior)
func (p *PredatorBehaviorStrategy) UpdateBehavior(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) core.Velocity {
	// ИСПРАВЛЕНИЕ: Если хищник ест - останавливаем движение (решает проблему "волк над зайцем")
	if world.HasComponent(entity, core.MaskEatingState) {
		return core.Velocity{X: 0, Y: 0} // Волк стоит на месте при поедании
	}

	// Хищники охотятся только когда голодны
	if components.Hunger.Value < components.AnimalConfig.HungerThreshold {
		// Ищем ближайшую добычу (травоядных)
		nearestPrey, foundPrey := world.FindNearestByTypeInTiles(
			components.Position.X, components.Position.Y,
			components.AnimalConfig.VisionRange, core.TypeRabbit, // РЕФАКТОРИНГ: используем AnimalConfig вместо Behavior
		)
		if foundPrey {
			preyPos, _ := world.GetPosition(nearestPrey)

			// Направление к добыче
			huntDir := physics.Vec2{
				X: preyPos.X - components.Position.X,
				Y: preyPos.Y - components.Position.Y,
			}
			huntDir = huntDir.Normalize()

			// Обновляем таймер направления в поведении используя значения из AnimalConfig
			components.Behavior.DirectionTimer = components.AnimalConfig.MinDirectionTime
			world.SetBehavior(entity, components.Behavior)

			return core.Velocity{
				X: huntDir.X * components.Speed.Current,
				Y: huntDir.Y * components.Speed.Current,
			}
		} else {
			// Добыча не найдена - блуждаем в поисках
			return RandomWalk.GetRandomWalkVelocity(
				world, entity, components.Behavior,
				components.Speed.Current*components.AnimalConfig.WanderingSpeed,
			)
		}
	} else {
		// Сыт - спокойное движение
		return RandomWalk.GetRandomWalkVelocity(
			world, entity, components.Behavior,
			components.Speed.Current*components.AnimalConfig.ContentSpeed,
		)
	}
}

// calculateBoundaryRepulsion вычисляет вектор отталкивания от границ мира
// Предотвращает кластеризацию животных в углах карты
func (h *HerbivoreBehaviorStrategy) calculateBoundaryRepulsion(position core.Position, worldWidth, worldHeight float32) physics.Vec2 {
	// Процент от размера мира для начала отталкивания (5% от каждого края)
	const boundaryZonePercent = 0.05

	boundaryThresholdX := worldWidth * boundaryZonePercent  // 5% от ширины
	boundaryThresholdY := worldHeight * boundaryZonePercent // 5% от высоты

	repulsion := physics.Vec2{X: 0, Y: 0}

	// Отталкивание от левой границы
	if position.X < boundaryThresholdX {
		force := (boundaryThresholdX - position.X) / boundaryThresholdX // 0-1, сильнее ближе к границе
		repulsion.X += force                                            // Толкает вправо
	}

	// Отталкивание от правой границы
	if position.X > worldWidth-boundaryThresholdX {
		distanceFromRightEdge := worldWidth - position.X
		force := (boundaryThresholdX - distanceFromRightEdge) / boundaryThresholdX
		repulsion.X -= force // Толкает влево
	}

	// Отталкивание от верхней границы
	if position.Y < boundaryThresholdY {
		force := (boundaryThresholdY - position.Y) / boundaryThresholdY
		repulsion.Y += force // Толкает вниз
	}

	// Отталкивание от нижней границы
	if position.Y > worldHeight-boundaryThresholdY {
		distanceFromBottomEdge := worldHeight - position.Y
		force := (boundaryThresholdY - distanceFromBottomEdge) / boundaryThresholdY
		repulsion.Y -= force // Толкает вверх
	}

	// Нормализуем силу отталкивания чтобы она не была слишком сильной
	// Максимальная сила отталкивания составляет 50% от направления побега
	const maxRepulsionStrength = 0.5
	repulsionMagnitude := repulsion.Length()
	if repulsionMagnitude > maxRepulsionStrength {
		repulsion = repulsion.Normalize()
		repulsion.X *= maxRepulsionStrength
		repulsion.Y *= maxRepulsionStrength
	}

	return repulsion
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// УДАЛЕНО: ScavengerBehaviorStrategy - не используется в игре (нет падальщиков)
