package simulation

import (
	"fmt"
	"math"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

// Константы для оптимизированной системы коллизий (устраняет магические числа)
const (
	// Множители для радиусов и поиска
	// ИСПРАВЛЕНИЕ: Расталкивание только при физическом контакте
	SearchRadiusMultiplier    = 1.1 // Ищем в радиусе чуть больше размера объекта (было 2.0)
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

	// Обрабатываем коллизии между животными (мягкое расталкивание как в StarCraft)
	ms.handleCollisions(world)

	// ИСПРАВЛЕНИЕ: Ограничиваем границами мира ПОСЛЕ коллизий
	// чтобы расталкивание не выталкивало животных за границы
	ms.constrainToBounds(world)
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

		// РЕФАКТОРИНГ: Конвертируем скорость из тайлов/сек в пиксели/сек
		// vel в тайлах/сек, pos в пикселях, используем унифицированную функцию конвертации
		pixelVelX := constants.TilesToPixels(vel.X)
		pixelVelY := constants.TilesToPixels(vel.Y)

		// Обновляем позицию
		pos.X += pixelVelX * deltaTime
		pos.Y += pixelVelY * deltaTime

		world.SetPosition(entity, pos)
		world.UpdateSpatialPosition(entity, pos) // Обновляем пространственную систему явно
	})
}

// constrainToBounds ограничивает сущности границами мира (рефакторинг: устранено дублирование кода)
func (ms *MovementSystem) constrainToBounds(world core.MovementSystemAccess) {
	world.ForEachWith(core.MaskPosition|core.MaskSize, func(entity core.EntityID) {
		pos, _ := world.GetPosition(entity)
		size, _ := world.GetSize(entity)
		// ИСПРАВЛЕНИЕ: Радиус уже в тайлах после перехода к тайловой системе
		radiusInTiles := constants.SizeRadiusToTiles(size.Radius) // Конвертируем пиксели в тайлы
		changed := false

		// Проверяем все границы через helper функции
		if ms.constrainXBounds(&pos, radiusInTiles, &changed) {
			ms.reflectVelocityX(world, entity, pos.X, radiusInTiles)
		}
		if ms.constrainYBounds(&pos, radiusInTiles, &changed) {
			ms.reflectVelocityY(world, entity, pos.Y, radiusInTiles)
		}

		// Обновляем позицию только если она изменилась
		if changed {
			world.SetPosition(entity, pos)
			world.UpdateSpatialPosition(entity, pos) // Обновляем пространственную систему явно
		}
	})
}

// constrainXBounds проверяет и исправляет X координату
func (ms *MovementSystem) constrainXBounds(pos *core.Position, radius float32, changed *bool) bool {
	boundsHit := false

	// ИСПРАВЛЕНИЕ: Симметричные границы с минимальным отступом
	const margin = 0.1 // Минимальный отступ для избежания проблем с округлением

	// Левая граница
	if pos.X-radius < margin {
		pos.X = margin + radius
		*changed = true
		boundsHit = true
	}

	// Правая граница
	if pos.X+radius > ms.worldWidth-margin {
		pos.X = ms.worldWidth - margin - radius
		*changed = true
		boundsHit = true
	}

	return boundsHit
}

// constrainYBounds проверяет и исправляет Y координату
func (ms *MovementSystem) constrainYBounds(pos *core.Position, radius float32, changed *bool) bool {
	boundsHit := false

	// ИСПРАВЛЕНИЕ: Симметричные границы с минимальным отступом
	const margin = 0.1 // Минимальный отступ для избежания проблем с округлением

	// Верхняя граница
	if pos.Y-radius < margin {
		pos.Y = margin + radius
		*changed = true
		boundsHit = true
	}

	// Нижняя граница
	if pos.Y+radius > ms.worldHeight-margin {
		pos.Y = ms.worldHeight - margin - radius
		*changed = true
		boundsHit = true
	}

	return boundsHit
}

// reflectVelocityX отражает X скорость при столкновении с границей
func (ms *MovementSystem) reflectVelocityX(
	world core.MovementSystemAccess,
	entity core.EntityID,
	posX, radius float32,
) {
	if !world.HasComponent(entity, core.MaskVelocity) {
		return
	}

	vel, _ := world.GetVelocity(entity)

	// ИСПРАВЛЕНИЕ: Симметричные границы с минимальным отступом
	const margin = 0.1 // Синхронизировано с constrainXBounds
	shouldReflect := (posX <= margin+radius && vel.X < 0) || (posX >= ms.worldWidth-margin-radius && vel.X > 0)

	if shouldReflect {
		ms.reflectVelocityComponent(&vel.X)
		world.SetVelocity(entity, vel)
	}
}

// reflectVelocityY отражает Y скорость при столкновении с границей
func (ms *MovementSystem) reflectVelocityY(
	world core.MovementSystemAccess,
	entity core.EntityID,
	posY, radius float32,
) {
	if !world.HasComponent(entity, core.MaskVelocity) {
		return
	}

	vel, _ := world.GetVelocity(entity)

	// ИСПРАВЛЕНИЕ: Симметричные границы с минимальным отступом
	const margin = 0.1 // Синхронизировано с constrainYBounds
	shouldReflect := (posY <= margin+radius && vel.Y < 0) || (posY >= ms.worldHeight-margin-radius && vel.Y > 0)

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

// handleCollisions обрабатывает мягкие коллизии между животными (KISS: простая структура)
func (ms *MovementSystem) handleCollisions(world core.MovementSystemAccess) {
	ms.broadPhaseCollisionDetection(world)
}

// broadPhaseCollisionDetection ищет потенциальные коллизии (KISS: одна ответственность)
func (ms *MovementSystem) broadPhaseCollisionDetection(world core.MovementSystemAccess) {
	world.ForEachWith(core.MaskPosition|core.MaskSize, func(entity core.EntityID) {
		candidates := ms.findCollisionCandidates(world, entity)
		ms.processCollisionCandidates(world, entity, candidates)
	})
}

// findCollisionCandidates находит кандидатов для проверки коллизий (KISS: простая логика)
func (ms *MovementSystem) findCollisionCandidates(
	world core.MovementSystemAccess,
	entity core.EntityID,
) []core.EntityID {
	pos, _ := world.GetPosition(entity)
	size, _ := world.GetSize(entity)

	searchRadiusPixels := size.Radius * CollisionConstants.SearchRadiusMultiplier
	searchRadius := constants.SizeRadiusToTiles(searchRadiusPixels) // Конвертируем в тайлы для поиска
	return world.QueryInRadius(pos.X, pos.Y, searchRadius)
}

// processCollisionCandidates проверяет кандидатов и обрабатывает коллизии (KISS: простой цикл)
func (ms *MovementSystem) processCollisionCandidates(
	world core.MovementSystemAccess,
	entity core.EntityID,
	candidates []core.EntityID,
) {
	for _, candidate := range candidates {
		if ms.shouldCheckCollision(entity, candidate) {
			ms.checkAndHandleCollision(world, entity, candidate)
		}
	}
}

// shouldCheckCollision определяет нужно ли проверять коллизию (KISS: простая логика)
func (ms *MovementSystem) shouldCheckCollision(entity1, entity2 core.EntityID) bool {
	return entity2 != entity1 && entity2 > entity1 // Предотвращает дублирование проверок
}

// checkAndHandleCollision проверяет и обрабатывает коллизию (KISS: разделено на простые шаги)
func (ms *MovementSystem) checkAndHandleCollision(world core.MovementSystemAccess, entity1, entity2 core.EntityID) {
	circles := ms.createCollisionCircles(world, entity1, entity2)
	collision := ms.detectCollision(circles.circle1, circles.circle2)

	if collision.Colliding {
		ms.separateEntities(world, entity1, entity2, collision)
	}
}

// collisionCircles простая структура для передачи кругов (KISS)
type collisionCircles struct {
	circle1, circle2 physics.Circle
}

// createCollisionCircles создаёт круги для проверки коллизий (KISS: простое создание)
func (ms *MovementSystem) createCollisionCircles(
	world core.MovementSystemAccess,
	entity1, entity2 core.EntityID,
) collisionCircles {
	pos1, _ := world.GetPosition(entity1)
	pos2, _ := world.GetPosition(entity2)
	size1, _ := world.GetSize(entity1)
	size2, _ := world.GetSize(entity2)

	return collisionCircles{
		circle1: physics.Circle{
			Center: physics.Vec2{X: pos1.X, Y: pos1.Y},
			Radius: constants.SizeRadiusToTiles(size1.Radius), // Конвертируем в тайлы
		},
		circle2: physics.Circle{
			Center: physics.Vec2{X: pos2.X, Y: pos2.Y},
			Radius: constants.SizeRadiusToTiles(size2.Radius), // Конвертируем в тайлы
		},
	}
}

// detectCollision проверяет коллизию двух кругов (KISS: простая проверка)
func (ms *MovementSystem) detectCollision(circle1, circle2 physics.Circle) physics.CollisionDetails {
	return physics.CircleCircleCollisionWithDetails(circle1, circle2)
}

// separateEntities мягко расталкивает две сущности при коллизии
// (рефакторинг: разбито на helper-функции)
func (ms *MovementSystem) separateEntities(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID, collision physics.CollisionDetails,
) {
	isPredatorPreyCollision := ms.isPredatorPreyCollision(world, entity1, entity2)

	// Разделяем позиции только если это не коллизия хищник-добыча
	if !isPredatorPreyCollision {
		ms.applyPositionSeparation(world, entity1, entity2, collision)
	}

	// Мягкое расталкивание через скорость
	ms.applyVelocitySeparation(world, entity1, entity2, collision, isPredatorPreyCollision)

}

// limitVelocity ограничивает скорость животного его максимальной скоростью
func (ms *MovementSystem) limitVelocity(
	world core.MovementSystemAccess, entity core.EntityID, velocity core.Velocity,
) core.Velocity {
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

// isPredatorPreyCollision проверяет является ли коллизия между хищником и добычей
// (helper-функция)
func (ms *MovementSystem) isPredatorPreyCollision(
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

// applyPositionSeparation применяет разделение позиций (helper-функция)
func (ms *MovementSystem) applyPositionSeparation(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID, collision physics.CollisionDetails,
) {
	separationForce := collision.Penetration * CollisionConstants.SeparationForceMultiplier

	// КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Ограничиваем силу расталкивания!
	const MaxSeparationForce = 2.0 // Максимум 2 тайла за раз
	if separationForce > MaxSeparationForce {
		separationForce = MaxSeparationForce
	}

	pos1, _ := world.GetPosition(entity1)
	pos2, _ := world.GetPosition(entity2)

	// Логируем только критические коллизии
	if collision.Penetration > constants.CriticalCollisionThreshold {
		fmt.Printf("CRITICAL Collision: entities %d-%d, penetration=%.2f\n",
			entity1, entity2, collision.Penetration)
	}

	// Сохраняем исходные позиции для проверки (убрано - не используются)

	// Применяем разделение
	pos1.X -= collision.Normal.X * separationForce
	pos1.Y -= collision.Normal.Y * separationForce
	pos2.X += collision.Normal.X * separationForce
	pos2.Y += collision.Normal.Y * separationForce

	// КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Проверяем что новые позиции в границах мира
	pos1Bounded := ms.boundPosition(pos1)
	pos2Bounded := ms.boundPosition(pos2)

	// Если расталкивание выходит за границы, применяем ограниченные позиции
	if pos1Bounded != pos1 || pos2Bounded != pos2 {
		pos1, pos2 = pos1Bounded, pos2Bounded
	}

	world.SetPosition(entity1, pos1)
	world.UpdateSpatialPosition(entity1, pos1) // Обновляем пространственную систему явно
	world.SetPosition(entity2, pos2)
	world.UpdateSpatialPosition(entity2, pos2) // Обновляем пространственную систему явно
}

// boundPosition ограничивает позицию границами мира
func (ms *MovementSystem) boundPosition(pos core.Position) core.Position {
	result := pos

	// ИСПРАВЛЕНИЕ: Симметричные границы с минимальным отступом
	const margin = 0.1 // Синхронизировано с constrainToBounds

	if result.X < margin {
		result.X = margin
	} else if result.X > ms.worldWidth-margin {
		result.X = ms.worldWidth - margin
	}

	if result.Y < margin {
		result.Y = margin
	} else if result.Y > ms.worldHeight-margin {
		result.Y = ms.worldHeight - margin
	}

	return result
}

// applyVelocitySeparation применяет мягкое расталкивание через скорость (helper-функция)
func (ms *MovementSystem) applyVelocitySeparation(
	world core.MovementSystemAccess, entity1, entity2 core.EntityID,
	collision physics.CollisionDetails, isPredatorPrey bool,
) {
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
func (ms *MovementSystem) applyNormalVelocitySeparation(
	vel1, vel2 *core.Velocity,
	collision physics.CollisionDetails,
) {

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
