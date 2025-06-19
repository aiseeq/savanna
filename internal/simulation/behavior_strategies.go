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
	Behavior core.Behavior
	Position core.Position
	Speed    core.Speed
	Hunger   core.Hunger
}

// UpdateBehavior реализует поведение травоядных (заменяет updateHerbivoreBehavior)
func (h *HerbivoreBehaviorStrategy) UpdateBehavior(
	world core.BehaviorSystemAccess,
	entity core.EntityID,
	components AnimalComponents,
) core.Velocity {
	// ПРИОРИТЕТ 1: Если видит хищника - убегать (всегда)
	nearestPredator, foundPredator := world.FindNearestByType(
		components.Position.X, components.Position.Y,
		components.Behavior.VisionRange, core.TypeWolf,
	)
	if foundPredator {
		// КРИТИЧЕСКИ ВАЖНО: прерываем поедание травы при побеге
		if world.HasComponent(entity, core.MaskEatingState) {
			world.RemoveEatingState(entity)
		}

		predatorPos, _ := world.GetPosition(nearestPredator)
		escapeDir := physics.Vec2{X: components.Position.X - predatorPos.X, Y: components.Position.Y - predatorPos.Y}
		escapeDir = escapeDir.Normalize()

		// Обновляем таймер направления в поведении
		components.Behavior.DirectionTimer = components.Behavior.MinDirectionTime
		world.SetBehavior(entity, components.Behavior)

		return core.Velocity{
			X: escapeDir.X * components.Speed.Current,
			Y: escapeDir.Y * components.Speed.Current,
		}
	}

	// ПРИОРИТЕТ 2: Если голоден ИЛИ уже ест - обрабатываем поедание травы
	isCurrentlyEating := world.HasComponent(entity, core.MaskEatingState)
	if (components.Hunger.Value < components.Behavior.HungerThreshold || isCurrentlyEating) && h.vegetation != nil {
		// Получаем конфигурацию животного для размеров
		// (устраняет нарушение OCP)
		config, hasConfig := world.GetAnimalConfig(entity)
		if !hasConfig {
			// Нет конфигурации - используем случайное движение
			return RandomWalk.GetRandomWalkVelocity(
				world, entity, components.Behavior,
				components.Speed.Current*components.Behavior.WanderingSpeed,
			)
		}

		// Сначала проверим - может быть мы уже на траве и едим?
		localGrassX, localGrassY, hasLocalGrass := h.vegetation.FindNearestGrass(
			components.Position.X, components.Position.Y,
			config.CollisionRadius, MinGrassToFind,
		)
		dx := components.Position.X - localGrassX
		dy := components.Position.Y - localGrassY
		distanceToLocalGrass := math.Sqrt(float64(dx*dx + dy*dy))

		if hasLocalGrass &&
			distanceToLocalGrass <= float64(config.CollisionRadius*GrassProximityMultiplier) {
			// Мы рядом с травой - останавливаемся и едим (возвращаем нулевую скорость)
			return core.Velocity{X: 0, Y: 0}
		}

		// Ищем ближайшую траву в радиусе видимости
		grassX, grassY, foundGrass := h.vegetation.FindNearestGrass(
			components.Position.X, components.Position.Y,
			components.Behavior.VisionRange, MinGrassToFind,
		)
		if foundGrass {
			// Идём к траве
			grassDir := physics.Vec2{X: grassX - components.Position.X, Y: grassY - components.Position.Y}
			grassDir = grassDir.Normalize()

			components.Behavior.DirectionTimer = components.Behavior.MinDirectionTime
			world.SetBehavior(entity, components.Behavior)

			return core.Velocity{
				X: grassDir.X * components.Speed.Current * components.Behavior.SearchSpeed,
				Y: grassDir.Y * components.Speed.Current * components.Behavior.SearchSpeed,
			}
		} else {
			// Трава не найдена - продолжаем случайное движение
			// в поисках
			return RandomWalk.GetRandomWalkVelocity(
				world, entity, components.Behavior,
				components.Speed.Current*components.Behavior.WanderingSpeed,
			)
		}
	}

	// ПРИОРИТЕТ 3: Если сыт - спокойное движение или отдых
	return RandomWalk.GetRandomWalkVelocity(
		world, entity, components.Behavior,
		components.Speed.Current*components.Behavior.ContentSpeed,
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
	if components.Hunger.Value < components.Behavior.HungerThreshold {
		// Ищем ближайшую добычу (травоядных)
		nearestPrey, foundPrey := world.FindNearestByType(
			components.Position.X, components.Position.Y,
			components.Behavior.VisionRange, core.TypeRabbit,
		)
		if foundPrey {
			preyPos, _ := world.GetPosition(nearestPrey)

			// Направление к добыче
			huntDir := physics.Vec2{
				X: preyPos.X - components.Position.X,
				Y: preyPos.Y - components.Position.Y,
			}
			huntDir = huntDir.Normalize()

			// Обновляем таймер направления в поведении
			components.Behavior.DirectionTimer = components.Behavior.MinDirectionTime
			world.SetBehavior(entity, components.Behavior)

			return core.Velocity{
				X: huntDir.X * components.Speed.Current,
				Y: huntDir.Y * components.Speed.Current,
			}
		} else {
			// Добыча не найдена - блуждаем в поисках
			return RandomWalk.GetRandomWalkVelocity(
				world, entity, components.Behavior,
				components.Speed.Current*components.Behavior.WanderingSpeed,
			)
		}
	} else {
		// Сыт - спокойное движение
		return RandomWalk.GetRandomWalkVelocity(
			world, entity, components.Behavior,
			components.Speed.Current*components.Behavior.ContentSpeed,
		)
	}
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// УДАЛЕНО: ScavengerBehaviorStrategy - не используется в игре (нет падальщиков)
