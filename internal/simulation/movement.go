package simulation

import (
	"math"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

// Константы для оптимизированной системы коллизий (устраняет магические числа)
const (
	// Множители для радиусов и поиска
	SearchRadiusMultiplier    = 3.0 // Ищем в радиусе в 3 раза больше размера объекта для broad-phase
	SeparationForceMultiplier = 0.6 // Коэффициент силы разделения позиций при коллизии

	// Параметры взаимодействий хищник-добыча
	PredatorPreyDamping = 0.7 // Замедление при коллизии хищник-добыча (70% скорости)

	// Параметры мягкого расталкивания
	SoftPushThreshold = 1.0 // Порог для активации мягкого расталкивания (пикс)
	SoftPushForce     = 1.0 // Базовая сила мягкого расталкивания

	// Параметры жёсткого расталкивания
	PenetrationThreshold = 0.1  // Порог проникновения для жёсткого расталкивания
	PushForceMultiplier  = 10.0 // Множитель силы жёсткого расталкивания

	// Пороги движения
	SlowMovementThreshold = 1.0 // Порог медленного движения (пикс/сек)
)

// CollisionConstants структура констант для обратной совместимости
var CollisionConstants = struct {
	SearchRadiusMultiplier    float32
	SeparationForceMultiplier float32
	PredatorPreyDamping       float32
	SoftPushThreshold         float32
	SoftPushForce             float32
	PenetrationThreshold      float32
	PushForceMultiplier       float32
	SlowMovementThreshold     float32
}{
	SearchRadiusMultiplier:    SearchRadiusMultiplier,
	SeparationForceMultiplier: SeparationForceMultiplier,
	PredatorPreyDamping:       PredatorPreyDamping,
	SoftPushThreshold:         SoftPushThreshold,
	SoftPushForce:             SoftPushForce,
	PenetrationThreshold:      PenetrationThreshold,
	PushForceMultiplier:       PushForceMultiplier,
	SlowMovementThreshold:     SlowMovementThreshold,
}

// MovementSystem отвечает за обновление позиций по скорости и обработку коллизий
type MovementSystem struct {
	worldWidth  float32
	worldHeight float32
}

// NewMovementSystem создаёт новую систему движения
func NewMovementSystem(worldWidth, worldHeight float32) *MovementSystem {
	return &MovementSystem{
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
	}
}

// Update обновляет движение всех сущностей
// Рефакторинг: использует специализированный интерфейс вместо полного World (ISP)
func (ms *MovementSystem) Update(world core.MovementSystemAccess, deltaTime float32) {
	// Рефакторинг: полностью используем специализированный интерфейс

	// Обновляем позиции по скорости
	ms.updatePositions(world, deltaTime)

	// Ограничиваем границами мира
	ms.constrainToBounds(world)

	// Обрабатываем коллизии между животными
	ms.handleCollisions(world)
}

// updatePositions обновляет позиции по скорости
func (ms *MovementSystem) updatePositions(world core.MovementSystemAccess, deltaTime float32) {
	world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
		// ВАЖНО: Животные не двигаются во время поедания (реализм!)
		if world.HasComponent(entity, core.MaskEatingState) {
			// Сбрасываем скорость в ноль для едящих животных
			world.SetVelocity(entity, core.Velocity{X: 0, Y: 0})
			return // Животное ест, не двигается
		}

		pos, _ := world.GetPosition(entity)
		vel, _ := world.GetVelocity(entity)

		// Обновляем позицию
		pos.X += vel.X * deltaTime
		pos.Y += vel.Y * deltaTime

		world.SetPosition(entity, pos)
	})
}

// constrainToBounds ограничивает сущности границами мира (рефакторинг: устранено дублирование кода)
func (ms *MovementSystem) constrainToBounds(world core.MovementSystemAccess) {
	world.ForEachWith(core.MaskPosition|core.MaskSize, func(entity core.EntityID) {
		pos, _ := world.GetPosition(entity)
		size, _ := world.GetSize(entity)
		radius := size.Radius
		changed := false

		// Проверяем все границы через helper функции
		if ms.constrainXBounds(&pos, radius, &changed) {
			ms.reflectVelocityX(world, entity, pos.X, radius)
		}
		if ms.constrainYBounds(&pos, radius, &changed) {
			ms.reflectVelocityY(world, entity, pos.Y, radius)
		}

		if changed {
			world.SetPosition(entity, pos)
		}
	})
}

// constrainXBounds проверяет и исправляет X координату
func (ms *MovementSystem) constrainXBounds(pos *core.Position, radius float32, changed *bool) bool {
	boundsHit := false

	// Левая граница
	if pos.X-radius < 0 {
		pos.X = radius
		*changed = true
		boundsHit = true
	}

	// Правая граница
	if pos.X+radius > ms.worldWidth {
		pos.X = ms.worldWidth - radius
		*changed = true
		boundsHit = true
	}

	return boundsHit
}

// constrainYBounds проверяет и исправляет Y координату
func (ms *MovementSystem) constrainYBounds(pos *core.Position, radius float32, changed *bool) bool {
	boundsHit := false

	// Верхняя граница
	if pos.Y-radius < 0 {
		pos.Y = radius
		*changed = true
		boundsHit = true
	}

	// Нижняя граница
	if pos.Y+radius > ms.worldHeight {
		pos.Y = ms.worldHeight - radius
		*changed = true
		boundsHit = true
	}

	return boundsHit
}

// reflectVelocityX отражает X скорость при столкновении с границей
func (ms *MovementSystem) reflectVelocityX(world core.MovementSystemAccess, entity core.EntityID, posX, radius float32) {
	if !world.HasComponent(entity, core.MaskVelocity) {
		return
	}

	vel, _ := world.GetVelocity(entity)

	// Определяем направление отражения
	shouldReflect := (posX <= radius && vel.X < 0) || (posX >= ms.worldWidth-radius && vel.X > 0)

	if shouldReflect {
		ms.reflectVelocityComponent(&vel.X)
		world.SetVelocity(entity, vel)
	}
}

// reflectVelocityY отражает Y скорость при столкновении с границей
func (ms *MovementSystem) reflectVelocityY(world core.MovementSystemAccess, entity core.EntityID, posY, radius float32) {
	if !world.HasComponent(entity, core.MaskVelocity) {
		return
	}

	vel, _ := world.GetVelocity(entity)

	// Определяем направление отражения
	shouldReflect := (posY <= radius && vel.Y < 0) || (posY >= ms.worldHeight-radius && vel.Y > 0)

	if shouldReflect {
		ms.reflectVelocityComponent(&vel.Y)
		world.SetVelocity(entity, vel)
	}
}

// reflectVelocityComponent отражает компонент скорости с затуханием
func (ms *MovementSystem) reflectVelocityComponent(velocity *float32) {
	const (
		ReflectionDamping = 0.8 // Коэффициент затухания при отражении
		MinVelocity       = 1.0 // Минимальная скорость для отражения
	)

	if math.Abs(float64(*velocity)) > MinVelocity {
		*velocity = -*velocity * ReflectionDamping // Отражаем с затуханием
	} else {
		*velocity = 0 // Останавливаем медленные объекты
	}
}

// handleCollisions обрабатывает мягкие коллизии между животными
// ОПТИМИЗИРОВАНО: O(n²) → O(n·k) используя SpatialGrid для broad-phase detection
func (ms *MovementSystem) handleCollisions(world core.MovementSystemAccess) {
	// Используем пространственные запросы вместо проверки "каждый с каждым"
	world.ForEachWith(core.MaskPosition|core.MaskSize, func(entity core.EntityID) {
		pos, _ := world.GetPosition(entity)
		size, _ := world.GetSize(entity)

		// Ищем потенциальных соседей в радиусе возможных коллизий
		// Увеличиваем радиус поиска для учёта движения и размеров других объектов
		searchRadius := size.Radius * CollisionConstants.SearchRadiusMultiplier
		nearby := world.QueryInRadius(pos.X, pos.Y, searchRadius)

		// Проверяем коллизии только с близкими объектами
		for _, nearbyEntity := range nearby {
			if nearbyEntity != entity && nearbyEntity > entity {
				// Условие > entity предотвращает дублирование проверок
				ms.checkAndHandleCollision(world, entity, nearbyEntity)
			}
		}
	})
}

// checkAndHandleCollision проверяет и обрабатывает коллизию между двумя сущностями (устраняет дублирование)
func (ms *MovementSystem) checkAndHandleCollision(world core.MovementSystemAccess, entity1, entity2 core.EntityID) {
	pos1, _ := world.GetPosition(entity1)
	pos2, _ := world.GetPosition(entity2)
	size1, _ := world.GetSize(entity1)
	size2, _ := world.GetSize(entity2)

	// Проверяем коллизию кругов
	circle1 := physics.Circle{
		Center: physics.Vec2{X: pos1.X, Y: pos1.Y},
		Radius: size1.Radius,
	}
	circle2 := physics.Circle{
		Center: physics.Vec2{X: pos2.X, Y: pos2.Y},
		Radius: size2.Radius,
	}

	collision := physics.CircleCircleCollisionWithDetails(circle1, circle2)
	if collision.Colliding {
		// Мягкое расталкивание
		ms.separateEntities(world, entity1, entity2, collision)
	}
}

// separateEntities мягко расталкивает две сущности при коллизии (рефакторинг: разбито на helper-функции)
func (ms *MovementSystem) separateEntities(world core.MovementSystemAccess, entity1, entity2 core.EntityID, collision physics.CollisionDetails) {
	isPredatorPreyCollision := ms.isPredatorPreyCollision(world, entity1, entity2)

	// Разделяем позиции только если это не коллизия хищник-добыча
	if !isPredatorPreyCollision {
		ms.applyPositionSeparation(world, entity1, entity2, collision)
	}

	// Мягкое расталкивание через скорость
	ms.applyVelocitySeparation(world, entity1, entity2, collision, isPredatorPreyCollision)

}

// limitVelocity ограничивает скорость животного его максимальной скоростью
func (ms *MovementSystem) limitVelocity(world core.MovementSystemAccess, entity core.EntityID, velocity core.Velocity) core.Velocity {
	// Получаем максимальную скорость животного
	speed, hasSpeed := world.GetSpeed(entity)
	if !hasSpeed {
		return velocity // Если нет компонента скорости, не ограничиваем
	}

	// Вычисляем текущую скорость
	currentSpeed := math.Sqrt(float64(velocity.X*velocity.X + velocity.Y*velocity.Y))
	maxSpeed := float64(speed.Base)

	// Если скорость превышает максимальную, масштабируем вектор
	if currentSpeed > maxSpeed {
		scale := maxSpeed / currentSpeed
		velocity.X = float32(float64(velocity.X) * scale)
		velocity.Y = float32(float64(velocity.Y) * scale)
	}

	return velocity
}

// isPredatorPreyCollision проверяет является ли коллизия между хищником и добычей (helper-функция)
func (ms *MovementSystem) isPredatorPreyCollision(world core.MovementSystemAccess, entity1, entity2 core.EntityID) bool {
	if !world.HasComponent(entity1, core.MaskSize) || !world.HasComponent(entity2, core.MaskSize) {
		return false
	}

	size1, _ := world.GetSize(entity1)
	size2, _ := world.GetSize(entity2)

	// Если один из них хищник (AttackRange > 0), а другой потенциальная добыча
	return (size1.AttackRange > 0 && size2.AttackRange == 0) ||
		(size1.AttackRange == 0 && size2.AttackRange > 0)
}

// applyPositionSeparation применяет разделение позиций (helper-функция)
func (ms *MovementSystem) applyPositionSeparation(world core.MovementSystemAccess, entity1, entity2 core.EntityID, collision physics.CollisionDetails) {
	separationForce := collision.Penetration * CollisionConstants.SeparationForceMultiplier

	pos1, _ := world.GetPosition(entity1)
	pos2, _ := world.GetPosition(entity2)

	// Применяем разделение
	pos1.X -= collision.Normal.X * separationForce
	pos1.Y -= collision.Normal.Y * separationForce
	pos2.X += collision.Normal.X * separationForce
	pos2.Y += collision.Normal.Y * separationForce

	world.SetPosition(entity1, pos1)
	world.SetPosition(entity2, pos2)
}

// applyVelocitySeparation применяет мягкое расталкивание через скорость (helper-функция)
func (ms *MovementSystem) applyVelocitySeparation(world core.MovementSystemAccess, entity1, entity2 core.EntityID, collision physics.CollisionDetails, isPredatorPrey bool) {
	if !world.HasComponent(entity1, core.MaskVelocity) || !world.HasComponent(entity2, core.MaskVelocity) {
		return
	}

	vel1, _ := world.GetVelocity(entity1)
	vel2, _ := world.GetVelocity(entity2)

	if isPredatorPrey {
		ms.applyPredatorPreyVelocity(&vel1, &vel2, collision)
	} else {
		ms.applyNormalVelocitySeparation(&vel1, &vel2, collision)
	}

	// Ограничиваем скорость максимальной скоростью животного
	vel1 = ms.limitVelocity(world, entity1, vel1)
	vel2 = ms.limitVelocity(world, entity2, vel2)

	world.SetVelocity(entity1, vel1)
	world.SetVelocity(entity2, vel2)
}

// applyPredatorPreyVelocity применяет скорость для коллизий хищник-добыча (helper-функция)
func (ms *MovementSystem) applyPredatorPreyVelocity(vel1, vel2 *core.Velocity, collision physics.CollisionDetails) {
	// Замедляем оба объекта чтобы волк мог атаковать
	vel1.X *= CollisionConstants.PredatorPreyDamping
	vel1.Y *= CollisionConstants.PredatorPreyDamping
	vel2.X *= CollisionConstants.PredatorPreyDamping
	vel2.Y *= CollisionConstants.PredatorPreyDamping

	// Добавляем очень мягкое расталкивание при сильном проникновении
	if collision.Penetration > CollisionConstants.SoftPushThreshold {
		vel1.X += collision.Normal.X * CollisionConstants.SoftPushForce
		vel1.Y += collision.Normal.Y * CollisionConstants.SoftPushForce
		vel2.X -= collision.Normal.X * CollisionConstants.SoftPushForce
		vel2.Y -= collision.Normal.Y * CollisionConstants.SoftPushForce
	}
}

// applyNormalVelocitySeparation применяет обычное расталкивание как в StarCraft 2 (helper-функция)
func (ms *MovementSystem) applyNormalVelocitySeparation(vel1, vel2 *core.Velocity, collision physics.CollisionDetails) {

	// Сначала останавливаем движение в сторону коллизии
	dotProduct1 := vel1.X*collision.Normal.X + vel1.Y*collision.Normal.Y
	dotProduct2 := vel2.X*(-collision.Normal.X) + vel2.Y*(-collision.Normal.Y)

	if dotProduct1 > 0 { // entity1 движется в сторону коллизии
		vel1.X -= collision.Normal.X * dotProduct1
		vel1.Y -= collision.Normal.Y * dotProduct1
	}

	if dotProduct2 > 0 { // entity2 движется в сторону коллизии
		vel2.X += collision.Normal.X * dotProduct2
		vel2.Y += collision.Normal.Y * dotProduct2
	}

	// Добавляем расталкивание пропорционально проникновению
	if collision.Penetration > CollisionConstants.PenetrationThreshold {
		softPushForce := collision.Penetration * CollisionConstants.PushForceMultiplier

		vel1.X += collision.Normal.X * softPushForce
		vel1.Y += collision.Normal.Y * softPushForce
		vel2.X -= collision.Normal.X * softPushForce
		vel2.Y -= collision.Normal.Y * softPushForce
	}
}
