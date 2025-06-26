package simulation

import (
	"fmt"
	"math"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
	"github.com/aiseeq/savanna/internal/vec2"
)

// CollisionSystem отвечает ТОЛЬКО за обработку коллизий между животными (SRP)
// Выделена из MovementSystem для соблюдения принципа единственной ответственности
type CollisionSystem struct{}

// NewCollisionSystem создаёт новую систему коллизий
func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{}
}

// Update обрабатывает коллизии между животными
func (cs *CollisionSystem) Update(world core.MovementSystemAccess) {
	cs.broadPhaseCollisionDetection(world)
}

// broadPhaseCollisionDetection ищет потенциальные коллизии
func (cs *CollisionSystem) broadPhaseCollisionDetection(world core.MovementSystemAccess) {
	world.ForEachWith(core.MaskPosition|core.MaskSize, func(entity core.EntityID) {
		// ИСПРАВЛЕНИЕ: Не двигаем едящих животных
		if world.HasComponent(entity, core.MaskEatingState) {
			return // Животное ест, не участвует в коллизиях
		}

		candidates := cs.findCollisionCandidates(world, entity)
		cs.processCollisionCandidates(world, entity, candidates)
	})
}

// findCollisionCandidates находит кандидатов для проверки коллизий
func (cs *CollisionSystem) findCollisionCandidates(
	world core.MovementSystemAccess,
	entity core.EntityID,
) []core.EntityID {
	pos, _ := world.GetPosition(entity)
	size, _ := world.GetSize(entity)

	// Size.Radius в тайлах, конвертируем в пиксели для QueryInRadius
	searchRadiusPixels := size.Radius * CollisionConstants.SearchRadiusMultiplier * float32(constants.TileSizePixels)

	// QueryInRadius ожидает радиус в пикселях (так как позиция в пикселях)
	return world.QueryInRadius(pos.X, pos.Y, searchRadiusPixels)
}

// processCollisionCandidates проверяет кандидатов и обрабатывает коллизии
func (cs *CollisionSystem) processCollisionCandidates(
	world core.MovementSystemAccess,
	entity core.EntityID,
	candidates []core.EntityID,
) {
	for _, candidate := range candidates {
		if cs.shouldCheckCollision(entity, candidate) {
			cs.checkAndHandleCollision(world, entity, candidate)
		}
	}
}

// shouldCheckCollision определяет нужно ли проверять коллизию
func (cs *CollisionSystem) shouldCheckCollision(entity1, entity2 core.EntityID) bool {
	return entity2 != entity1 && entity2 > entity1 // Предотвращает дублирование проверок
}

// checkAndHandleCollision проверяет и обрабатывает коллизию
func (cs *CollisionSystem) checkAndHandleCollision(world core.MovementSystemAccess, entity1, entity2 core.EntityID) {
	circles := cs.createCollisionCircles(world, entity1, entity2)
	collision := cs.detectCollision(circles.circle1, circles.circle2)

	if cs.shouldApplySeparation(circles, collision) {
		finalCollision := cs.calculateFinalCollision(circles, collision)
		cs.separateEntities(world, entity1, entity2, finalCollision)
	}
}

// shouldApplySeparation определяет нужно ли применять расталкивание
func (cs *CollisionSystem) shouldApplySeparation(circles collisionCircles, collision physics.CollisionDetails) bool {
	// ОПТИМИЗАЦИЯ: Сначала проверяем манхеттенское расстояние как быстрый предфильтр
	dx := circles.circle1.Center.X - circles.circle2.Center.X
	dy := circles.circle1.Center.Y - circles.circle2.Center.Y
	manhattanDistance := float32(math.Abs(float64(dx))) + float32(math.Abs(float64(dy)))
	maxPossibleRadius := circles.circle1.Radius + circles.circle2.Radius + 0.1 // +0.1 тайла буфер

	// Быстрое отсечение: если манхеттенское расстояние больше чем возможный радиус * 1.5, то точно нет коллизии
	if manhattanDistance > maxPossibleRadius*1.5 {
		return false // Нет смысла проверять евклидово расстояние
	}

	// ОПТИМИЗАЦИЯ: сравниваем квадраты расстояний вместо sqrt
	distanceSquared := dx*dx + dy*dy
	safeDistanceSquared := maxPossibleRadius * maxPossibleRadius

	// Применяем расталкивание если пересекаются ИЛИ слишком близко
	return collision.Colliding || distanceSquared < safeDistanceSquared
}

// calculateFinalCollision вычисляет финальные параметры коллизии
func (cs *CollisionSystem) calculateFinalCollision(circles collisionCircles, collision physics.CollisionDetails) physics.CollisionDetails {
	// Если не пересекаются, но близко - создаём искусственную коллизию
	if !collision.Colliding {
		// ОПТИМИЗАЦИЯ: используем методы physics.Vec2 для элегантности
		direction := circles.circle1.Center.Sub(circles.circle2.Center)
		distance := direction.Length()
		maxPossibleRadius := circles.circle1.Radius + circles.circle2.Radius + 0.1
		safeDistance := maxPossibleRadius

		penetration := float32(safeDistance - distance)
		normal := direction.Normalize() // Элегантная нормализация
		return physics.CollisionDetails{
			Colliding:   true,
			Penetration: penetration,
			Normal:      normal,
		}
	}
	return collision
}

// collisionCircles простая структура для передачи кругов
type collisionCircles struct {
	circle1, circle2 physics.Circle
}

// createCollisionCircles создаёт круги для проверки коллизий
func (cs *CollisionSystem) createCollisionCircles(
	world core.MovementSystemAccess,
	entity1, entity2 core.EntityID,
) collisionCircles {
	pos1, _ := world.GetPosition(entity1)
	pos2, _ := world.GetPosition(entity2)
	size1, _ := world.GetSize(entity1)
	size2, _ := world.GetSize(entity2)

	// Конвертация позиций через векторы
	pos1TilesX, pos1TilesY := constants.PositionToTiles(pos1.X, pos1.Y) // Позиции в пикселях -> тайлы
	pos2TilesX, pos2TilesY := constants.PositionToTiles(pos2.X, pos2.Y) // Позиции в пикселях -> тайлы

	return collisionCircles{
		circle1: physics.Circle{
			Center: physics.Vec2{X: pos1TilesX, Y: pos1TilesY},
			Radius: size1.Radius, // Size.Radius уже в тайлах как float32
		},
		circle2: physics.Circle{
			Center: physics.Vec2{X: pos2TilesX, Y: pos2TilesY},
			Radius: size2.Radius, // Size.Radius уже в тайлах как float32
		},
	}
}

// detectCollision проверяет коллизию двух кругов
func (cs *CollisionSystem) detectCollision(circle1, circle2 physics.Circle) physics.CollisionDetails {
	return physics.CircleCircleCollisionWithDetails(circle1, circle2)
}

// separateEntities мягко расталкивает две сущности при коллизии
func (cs *CollisionSystem) separateEntities(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID, collision physics.CollisionDetails,
) {
	isPredatorPreyCollision := cs.isPredatorPreyCollision(world, entity1, entity2)

	// Разделяем позиции только если это не коллизия хищник-добыча
	if !isPredatorPreyCollision {
		cs.applyPositionSeparation(world, entity1, entity2, collision)
	}

	// Мягкое расталкивание через скорость
	cs.applyVelocitySeparation(world, entity1, entity2, collision, isPredatorPreyCollision)
}

// isPredatorPreyCollision проверяет является ли коллизия между хищником и добычей
func (cs *CollisionSystem) isPredatorPreyCollision(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID,
) bool {
	if !world.HasComponent(entity1, core.MaskSize) || !world.HasComponent(entity2, core.MaskSize) {
		return false
	}

	size1, _ := world.GetSize(entity1)
	size2, _ := world.GetSize(entity2)

	// Если один из них хищник (AttackRange > 0), а другой потенциальная добыча
	return (size1.AttackRange > 0 && size2.AttackRange == 0) ||
		(size1.AttackRange == 0 && size2.AttackRange > 0)
}

// applyPositionSeparation применяет разделение позиций
func (cs *CollisionSystem) applyPositionSeparation(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID, collision physics.CollisionDetails,
) {
	separationForce := cs.calculateSeparationForce(collision.Penetration)
	pos1, pos2 := cs.getEntityPositions(world, entity1, entity2)

	cs.logCriticalCollision(entity1, entity2, collision.Penetration)
	newPos1, newPos2 := cs.applySeparationToPositions(pos1, pos2, collision, separationForce)
	cs.updatePositions(world, entity1, entity2, newPos1, newPos2)
}

// calculateSeparationForce вычисляет силу расталкивания с ограничениями
func (cs *CollisionSystem) calculateSeparationForce(penetration float32) float32 {
	separationForce := penetration * CollisionConstants.SeparationForceMultiplier

	// КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Ограничиваем силу расталкивания!
	const MaxSeparationForce = 2.0 // Максимум 2 тайла за раз
	if separationForce > MaxSeparationForce {
		separationForce = MaxSeparationForce
	}
	return separationForce
}

// getEntityPositions получает позиции двух сущностей
func (cs *CollisionSystem) getEntityPositions(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID,
) (core.Position, core.Position) {
	pos1, _ := world.GetPosition(entity1)
	pos2, _ := world.GetPosition(entity2)
	return pos1, pos2
}

// logCriticalCollision логирует критические коллизии
func (cs *CollisionSystem) logCriticalCollision(entity1, entity2 core.EntityID, penetration float32) {
	if penetration > constants.CriticalCollisionThreshold {
		fmt.Printf("CRITICAL Collision: entities %d-%d, penetration=%.2f\n",
			entity1, entity2, penetration)
	}
}

// applySeparationToPositions применяет расталкивание к позициям
func (cs *CollisionSystem) applySeparationToPositions(
	pos1, pos2 core.Position, collision physics.CollisionDetails, separationForce float32,
) (core.Position, core.Position) {
	// Простая математика: разделение через векторы
	separationForcePixels := constants.TilesToPixels(separationForce)

	// Создаем вектор разделения
	separationX := collision.Normal.X * separationForcePixels
	separationY := collision.Normal.Y * separationForcePixels

	// Применяем разделение
	newPos1 := core.Position{X: pos1.X - separationX, Y: pos1.Y - separationY}
	newPos2 := core.Position{X: pos2.X + separationX, Y: pos2.Y + separationY}

	return newPos1, newPos2
}

// updatePositions обновляет позиции сущностей
func (cs *CollisionSystem) updatePositions(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID, pos1, pos2 core.Position,
) {
	world.SetPosition(entity1, pos1)
	world.UpdateSpatialPosition(entity1, pos1)
	world.SetPosition(entity2, pos2)
	world.UpdateSpatialPosition(entity2, pos2)
}

// applyVelocitySeparation применяет мягкое расталкивание через скорость
func (cs *CollisionSystem) applyVelocitySeparation(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID,
	collision physics.CollisionDetails, isPredatorPrey bool,
) {
	if !world.HasComponent(entity1, core.MaskVelocity) || !world.HasComponent(entity2, core.MaskVelocity) {
		return
	}

	vel1, _ := world.GetVelocity(entity1)
	vel2, _ := world.GetVelocity(entity2)

	if isPredatorPrey {
		cs.applyPredatorPreyVelocity(&vel1, &vel2, collision)
	} else {
		cs.applyNormalVelocitySeparation(&vel1, &vel2, collision)
	}

	// Ограничиваем скорость максимальной скоростью животного
	vel1 = cs.limitVelocity(world, entity1, vel1)
	vel2 = cs.limitVelocity(world, entity2, vel2)

	world.SetVelocity(entity1, vel1)
	world.SetVelocity(entity2, vel2)
}

// limitVelocity ограничивает скорость животного его максимальной скоростью
func (cs *CollisionSystem) limitVelocity(
	world core.MovementSystemAccess, entity core.EntityID, velocity core.Velocity,
) core.Velocity {
	// Получаем максимальную скорость животного
	speed, hasSpeed := world.GetSpeed(entity)
	if !hasSpeed {
		return velocity // Если нет компонента скорости, не ограничиваем
	}

	// Простая математика: ограничение скорости через vec2
	velocityVec := vec2.Vec2{X: velocity.X, Y: velocity.Y}
	currentSpeed := velocityVec.Length()
	maxSpeed := speed.Base

	// Если скорость превышает максимальную, масштабируем вектор
	if currentSpeed > maxSpeed {
		normalized := velocityVec.Normalize()
		velocity = core.Velocity{X: normalized.X * maxSpeed, Y: normalized.Y * maxSpeed}
	}

	return velocity
}

// applyPredatorPreyVelocity применяет скорость для коллизий хищник-добыча
func (cs *CollisionSystem) applyPredatorPreyVelocity(vel1, vel2 *core.Velocity, collision physics.CollisionDetails) {
	// Простая математика: замедление через масштабирование
	vel1.X *= CollisionConstants.PredatorPreyDamping
	vel1.Y *= CollisionConstants.PredatorPreyDamping
	vel2.X *= CollisionConstants.PredatorPreyDamping
	vel2.Y *= CollisionConstants.PredatorPreyDamping

	// Добавляем очень мягкое расталкивание при сильном проникновении
	if collision.Penetration > CollisionConstants.SoftPushThreshold {
		pushX := collision.Normal.X * CollisionConstants.SoftPushForce
		pushY := collision.Normal.Y * CollisionConstants.SoftPushForce
		vel1.X += pushX
		vel1.Y += pushY
		vel2.X -= pushX
		vel2.Y -= pushY
	}
}

// applyNormalVelocitySeparation применяет обычное расталкивание как в StarCraft 2
func (cs *CollisionSystem) applyNormalVelocitySeparation(
	vel1, vel2 *core.Velocity,
	collision physics.CollisionDetails,
) {
	// Простая математика: векторные операции
	normalX := collision.Normal.X
	normalY := collision.Normal.Y

	// Останавливаем движение в сторону коллизии
	vel1Vec := vec2.Vec2{X: vel1.X, Y: vel1.Y}
	vel2Vec := vec2.Vec2{X: vel2.X, Y: vel2.Y}
	normal := vec2.Vec2{X: normalX, Y: normalY}

	dotProduct1 := vel1Vec.Dot(normal)
	dotProduct2 := vel2Vec.Dot(vec2.Vec2{X: -normalX, Y: -normalY})

	if dotProduct1 > 0 { // entity1 движется в сторону коллизии
		vel1.X -= normalX * dotProduct1
		vel1.Y -= normalY * dotProduct1
	}

	if dotProduct2 > 0 { // entity2 движется в сторону коллизии
		vel2.X += normalX * dotProduct2
		vel2.Y += normalY * dotProduct2
	}

	// Добавляем расталкивание пропорционально проникновению
	if collision.Penetration > CollisionConstants.PenetrationThreshold {
		softPushForce := collision.Penetration * CollisionConstants.PushForceMultiplier
		pushX := normalX * softPushForce
		pushY := normalY * softPushForce

		vel1.X += pushX
		vel1.Y += pushY
		vel2.X -= pushX
		vel2.Y -= pushY
	}
}
