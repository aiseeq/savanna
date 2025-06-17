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
	UpdateBehavior(world core.BehaviorSystemAccess, entity core.EntityID, behavior core.Behavior, pos core.Position, speed core.Speed, hunger core.Hunger) core.Velocity
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

// UpdateBehavior реализует поведение травоядных (заменяет updateHerbivoreBehavior)
func (h *HerbivoreBehaviorStrategy) UpdateBehavior(world core.BehaviorSystemAccess, entity core.EntityID, behavior core.Behavior, pos core.Position, speed core.Speed, hunger core.Hunger) core.Velocity {
	// ПРИОРИТЕТ 1: Если видит хищника - убегать (всегда)
	nearestPredator, foundPredator := world.FindNearestByType(pos.X, pos.Y, behavior.VisionRange, core.TypeWolf)
	if foundPredator {
		// КРИТИЧЕСКИ ВАЖНО: прерываем поедание травы при побеге
		if world.HasComponent(entity, core.MaskEatingState) {
			world.RemoveEatingState(entity)
		}

		predatorPos, _ := world.GetPosition(nearestPredator)
		escapeDir := physics.Vec2{X: pos.X - predatorPos.X, Y: pos.Y - predatorPos.Y}
		escapeDir = escapeDir.Normalize()

		// Обновляем таймер направления в поведении
		behavior.DirectionTimer = behavior.MinDirectionTime
		world.SetBehavior(entity, behavior)

		return core.Velocity{
			X: escapeDir.X * speed.Current,
			Y: escapeDir.Y * speed.Current,
		}
	}

	// ПРИОРИТЕТ 2: Если голоден ИЛИ уже ест - обрабатываем поедание травы
	isCurrentlyEating := world.HasComponent(entity, core.MaskEatingState)
	if (hunger.Value < behavior.HungerThreshold || isCurrentlyEating) && h.vegetation != nil {
		// Получаем конфигурацию животного для размеров (устраняет нарушение OCP)
		config, hasConfig := world.GetAnimalConfig(entity)
		if !hasConfig {
			// Нет конфигурации - используем случайное движение
			return RandomWalk.GetRandomWalkVelocity(world, entity, behavior, speed.Current*behavior.WanderingSpeed)
		}

		// Сначала проверим - может быть мы уже на траве и едим?
		localGrassX, localGrassY, hasLocalGrass := h.vegetation.FindNearestGrass(pos.X, pos.Y, config.CollisionRadius, MinGrassToFind)
		distanceToLocalGrass := math.Sqrt(float64((pos.X-localGrassX)*(pos.X-localGrassX) + (pos.Y-localGrassY)*(pos.Y-localGrassY)))

		if hasLocalGrass && distanceToLocalGrass <= float64(config.CollisionRadius*GrassProximityMultiplier) {
			// Мы рядом с травой - останавливаемся и едим (возвращаем нулевую скорость)
			return core.Velocity{X: 0, Y: 0}
		}

		// Ищем ближайшую траву в радиусе видимости
		grassX, grassY, foundGrass := h.vegetation.FindNearestGrass(pos.X, pos.Y, behavior.VisionRange, MinGrassToFind)
		if foundGrass {
			// Идём к траве
			grassDir := physics.Vec2{X: grassX - pos.X, Y: grassY - pos.Y}
			grassDir = grassDir.Normalize()

			behavior.DirectionTimer = behavior.MinDirectionTime
			world.SetBehavior(entity, behavior)

			return core.Velocity{
				X: grassDir.X * speed.Current * behavior.SearchSpeed,
				Y: grassDir.Y * speed.Current * behavior.SearchSpeed,
			}
		} else {
			// Трава не найдена - продолжаем случайное движение в поисках
			return RandomWalk.GetRandomWalkVelocity(world, entity, behavior, speed.Current*behavior.WanderingSpeed)
		}
	}

	// ПРИОРИТЕТ 3: Если сыт - спокойное движение или отдых
	return RandomWalk.GetRandomWalkVelocity(world, entity, behavior, speed.Current*behavior.ContentSpeed)
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// PredatorBehaviorStrategy стратегия поведения хищников
type PredatorBehaviorStrategy struct{}

// NewPredatorBehaviorStrategy создаёт новую стратегию хищников
func NewPredatorBehaviorStrategy() *PredatorBehaviorStrategy {
	return &PredatorBehaviorStrategy{}
}

// UpdateBehavior реализует поведение хищников (заменяет updatePredatorBehavior)
func (p *PredatorBehaviorStrategy) UpdateBehavior(world core.BehaviorSystemAccess, entity core.EntityID, behavior core.Behavior, pos core.Position, speed core.Speed, hunger core.Hunger) core.Velocity {
	// ИСПРАВЛЕНИЕ: Если хищник ест - останавливаем движение (решает проблему "волк над зайцем")
	if world.HasComponent(entity, core.MaskEatingState) {
		return core.Velocity{X: 0, Y: 0} // Волк стоит на месте при поедании
	}

	// Хищники охотятся только когда голодны
	if hunger.Value < behavior.HungerThreshold {
		// Ищем ближайшую добычу (травоядных)
		nearestPrey, foundPrey := world.FindNearestByType(pos.X, pos.Y, behavior.VisionRange, core.TypeRabbit)
		if foundPrey {
			preyPos, _ := world.GetPosition(nearestPrey)

			// Направление к добыче
			huntDir := physics.Vec2{X: preyPos.X - pos.X, Y: preyPos.Y - pos.Y}
			huntDir = huntDir.Normalize()

			// Обновляем таймер направления в поведении
			behavior.DirectionTimer = behavior.MinDirectionTime
			world.SetBehavior(entity, behavior)

			return core.Velocity{
				X: huntDir.X * speed.Current,
				Y: huntDir.Y * speed.Current,
			}
		} else {
			// Добыча не найдена - блуждаем в поисках
			return RandomWalk.GetRandomWalkVelocity(world, entity, behavior, speed.Current*behavior.WanderingSpeed)
		}
	} else {
		// Сыт - спокойное движение
		return RandomWalk.GetRandomWalkVelocity(world, entity, behavior, speed.Current*behavior.ContentSpeed)
	}
}

// УДАЛЕНО: getRandomWalkVelocityWithBehavior заменена на RandomWalk.GetRandomWalkVelocity

// УДАЛЕНО: ScavengerBehaviorStrategy - не используется в игре (нет падальщиков)
