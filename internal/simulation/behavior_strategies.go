package simulation

import (
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/vec2"
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
	Satiation    core.Satiation
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
		components.AnimalConfig.VisionRange, core.TypeWolf,
	)
	if !foundPredator {
		return nil // Хищника нет
	}

	// КРИТИЧЕСКИ ВАЖНО: прерываем поедание травы при побеге
	if world.HasComponent(entity, core.MaskEatingState) {
		world.RemoveEatingState(entity)
	}

	predatorPos, _ := world.GetPosition(nearestPredator)

	// ОПТИМИЗАЦИЯ: элегантное направление побега используя методы Position
	escapeVector := components.Position.Sub(predatorPos).Normalize() // Вектор от хищника к нам
	escapeDirection := vec2.New(escapeVector.X, escapeVector.Y)

	// ЭЛЕГАНТНАЯ МАТЕМАТИКА: Добавляем отталкивание от границ мира
	worldWidth, worldHeight := world.GetWorldDimensions()
	boundaryRepulsion := h.calculateBoundaryRepulsion(components.Position, worldWidth, worldHeight)

	// Комбинируем направление побега с отталкиванием (комплексная арифметика!)
	finalEscapeDirection := escapeDirection.Add(boundaryRepulsion).Normalize()

	// Обновляем таймер направления в поведении
	components.Behavior.DirectionTimer = components.AnimalConfig.MinDirectionTime
	world.SetBehavior(entity, components.Behavior)

	// Создаем скорость с элегантной арифметикой
	speed := components.Speed.Current
	resultVelocity := core.Velocity{X: finalEscapeDirection.X * speed, Y: finalEscapeDirection.Y * speed}
	return &resultVelocity
}

// handleFeeding обрабатывает поиск и поедание травы (KISS: выделено в отдельный метод)
func (h *HerbivoreBehaviorStrategy) handleFeeding(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) *core.Velocity {
	isCurrentlyEating := world.HasComponent(entity, core.MaskEatingState)
	if !(components.Satiation.Value < components.AnimalConfig.SatiationThreshold || isCurrentlyEating) || h.vegetation == nil {
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
	// ТИПОБЕЗОПАСНОСТЬ: Конвертируем радиус коллизий из тайлов в пиксели для FindNearestGrass
	collisionRadiusPixels := constants.TilesToPixels(config.CollisionRadius) // Конвертируем physics.Tiles
	localGrassX, localGrassY, hasLocalGrass := h.vegetation.FindNearestGrass(
		components.Position.X, components.Position.Y,
		collisionRadiusPixels, MinGrassAmountToFind,
	)
	if !hasLocalGrass {
		return nil
	}

	// ОПТИМИЗАЦИЯ: элегантное расстояние до травы через методы Position + сравнение квадратов
	grassPos := core.NewPosition(localGrassX, localGrassY)
	distanceSquared := components.Position.DistanceSquaredTo(grassPos)

	// ИСПРАВЛЕНИЕ: Конвертируем радиус коллизий из тайлов в пиксели для проверки расстояния
	collisionRadiusPixelsForCheck := constants.TilesToPixels(config.CollisionRadius)
	threshold := collisionRadiusPixelsForCheck * GrassProximityMultiplier
	if distanceSquared <= threshold*threshold {
		// Мы рядом с травой - останавливаемся и едим
		zeroVel := core.NewVelocity(0, 0)
		return &zeroVel
	}

	return nil
}

// searchForGrass ищет траву в радиусе видимости (KISS: выделено в отдельный метод)
func (h *HerbivoreBehaviorStrategy) searchForGrass(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) *core.Velocity {
	// ИСПРАВЛЕНИЕ: Конвертируем дальность зрения из тайлов в пиксели для FindNearestGrass
	visionRangePixels := constants.TilesToPixels(components.AnimalConfig.VisionRange)
	grassX, grassY, foundGrass := h.vegetation.FindNearestGrass(
		components.Position.X, components.Position.Y,
		visionRangePixels, MinGrassAmountToFind,
	)
	if foundGrass {
		// ОПТИМИЗАЦИЯ: элегантное направление к траве через методы Position
		grassPos := core.NewPosition(grassX, grassY)
		grassVector := grassPos.Sub(components.Position).Normalize()
		grassDir := vec2.New(grassVector.X, grassVector.Y)

		// КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Добавляем избегание близких животных
		avoidanceDir := h.calculateAvoidanceDirection(world, entity, components)

		// Комбинируем направление к траве с избеганием (ЭЛЕГАНТНО!)
		finalDir := grassDir.Add(avoidanceDir)

		// Нормализуем итоговое направление
		if finalDir.Length() > 0 {
			finalDir = finalDir.Normalize()
		} else {
			finalDir = grassDir // Fallback к исходному направлению
		}

		components.Behavior.DirectionTimer = components.AnimalConfig.MinDirectionTime
		world.SetBehavior(entity, components.Behavior)

		// Конвертируем в скорость
		speed := components.Speed.Current * components.AnimalConfig.SearchSpeed
		resultVel := core.Velocity{X: finalDir.X * speed, Y: finalDir.Y * speed}
		return &resultVel
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
		return core.NewVelocity(0, 0) // Волк стоит на месте при поедании
	}

	// Хищники охотятся только когда голодны
	if components.Satiation.Value < components.AnimalConfig.SatiationThreshold {
		// ЭЛЕГАНТНАЯ МАТЕМАТИКА: прямое использование комплексной позиции

		// Ищем ближайшую добычу (травоядных)
		nearestPrey, foundPrey := world.FindNearestByTypeInTiles(
			components.Position.X, components.Position.Y,
			components.AnimalConfig.VisionRange, core.TypeRabbit, // РЕФАКТОРИНГ: используем AnimalConfig вместо Behavior
		)
		if foundPrey {
			preyPos, _ := world.GetPosition(nearestPrey)

			// ОПТИМИЗАЦИЯ: элегантное направление к добыче через методы Position
			huntVector := preyPos.Sub(components.Position).Normalize()
			huntDir := vec2.New(huntVector.X, huntVector.Y)

			// Обновляем таймер направления в поведении используя значения из AnimalConfig
			components.Behavior.DirectionTimer = components.AnimalConfig.MinDirectionTime
			world.SetBehavior(entity, components.Behavior)

			// ИСПРАВЛЕНИЕ: Используем SearchSpeed множитель для скорости охоты (как у травоядных при поиске травы)
			speed := components.Speed.Current * components.AnimalConfig.SearchSpeed
			return core.Velocity{X: huntDir.X * speed, Y: huntDir.Y * speed}
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
// Предотвращает кластеризацию животных в углах карты - ЭЛЕГАНТНАЯ МАТЕМАТИКА
func (h *HerbivoreBehaviorStrategy) calculateBoundaryRepulsion(position core.Position, worldWidth, worldHeight float32) vec2.Vec2 {
	// Процент от размера мира для начала отталкивания (5% от каждого края)
	const boundaryZonePercent = 0.05

	boundaryThresholdX := worldWidth * boundaryZonePercent  // 5% от ширины
	boundaryThresholdY := worldHeight * boundaryZonePercent // 5% от высоты

	var repulsionX, repulsionY float32

	// Отталкивание от левой границы (ЭЛЕГАНТНАЯ МАТЕМАТИКА)
	if position.X < boundaryThresholdX {
		force := (boundaryThresholdX - position.X) / boundaryThresholdX // 0-1, сильнее ближе к границе
		repulsionX += force                                             // Толкает вправо
	}

	// Отталкивание от правой границы
	if position.X > worldWidth-boundaryThresholdX {
		distanceFromRightEdge := worldWidth - position.X
		force := (boundaryThresholdX - distanceFromRightEdge) / boundaryThresholdX
		repulsionX -= force // Толкает влево
	}

	// Отталкивание от верхней границы
	if position.Y < boundaryThresholdY {
		force := (boundaryThresholdY - position.Y) / boundaryThresholdY
		repulsionY += force // Толкает вниз
	}

	// Отталкивание от нижней границы
	if position.Y > worldHeight-boundaryThresholdY {
		distanceFromBottomEdge := worldHeight - position.Y
		force := (boundaryThresholdY - distanceFromBottomEdge) / boundaryThresholdY
		repulsionY -= force // Толкает вверх
	}

	// Создаем вектор отталкивания из компонентов
	repulsion := vec2.New(repulsionX, repulsionY)

	// Нормализуем силу отталкивания чтобы она не была слишком сильной
	// Максимальная сила отталкивания составляет 50% от направления побега
	const maxRepulsionStrength = 0.5
	repulsionMagnitude := repulsion.Length()
	if repulsionMagnitude > maxRepulsionStrength {
		repulsion = repulsion.Normalize().Scale(maxRepulsionStrength)
	}

	return repulsion
}

// calculateAvoidanceDirection вычисляет направление для избегания близких животных
// Предотвращает петли столкновений когда животные постоянно бегут навстречу друг другу
func (h *HerbivoreBehaviorStrategy) calculateAvoidanceDirection(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) vec2.Vec2 {
	avoidanceForce := vec2.New(0, 0)

	// Радиус поиска соседей - использует константу из game_balance.go (ТИПОБЕЗОПАСНО)
	searchRadius := constants.TilesToPixels(components.AnimalConfig.CollisionRadius * BehaviorAvoidanceRadiusMultiplier)

	// Ищем близких животных в радиусе (ТИПОБЕЗОПАСНО)
	nearbyAnimals := world.QueryInRadius(components.Position.X, components.Position.Y, searchRadius)

	for _, neighborID := range nearbyAnimals {
		if neighborID == entity {
			continue // Пропускаем себя
		}

		neighborPos, hasPos := world.GetPosition(neighborID)
		if !hasPos {
			continue
		}

		// ОПТИМИЗАЦИЯ: элегантное расстояние и направление через методы Position (без дублирования!)
		directionVector := components.Position.Sub(neighborPos) // Направление от соседа к нам
		distance := directionVector.Length()

		if distance > 0.1 && distance < searchRadius { // Избегаем деления на ноль
			// Сила обратно пропорциональна расстоянию (чем ближе, тем сильнее отталкивание)
			force := 1.0 / distance

			// Нормализованное направление от соседа (используем уже вычисленный вектор!)
			normalizedDirection := directionVector.Scale(1.0 / distance) // Вместо повторного вычисления
			repulsionDirection := vec2.New(normalizedDirection.X*float32(force), normalizedDirection.Y*float32(force))
			avoidanceForce = avoidanceForce.Add(repulsionDirection)
		}
	}

	// Ограничиваем силу избегания чтобы она не перебивала движение к траве
	if avoidanceForce.Length() > BehaviorAvoidanceMaxStrength {
		avoidanceForce = avoidanceForce.Normalize().Scale(BehaviorAvoidanceMaxStrength)
	}

	return avoidanceForce
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// УДАЛЕНО: ScavengerBehaviorStrategy - не используется в игре (нет падальщиков)
