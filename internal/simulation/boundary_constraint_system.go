package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

// BoundaryConstraintSystem отвечает ТОЛЬКО за ограничение сущностей границами мира (SRP)
// Выделена из MovementSystem для соблюдения принципа единственной ответственности
// ТИПОБЕЗОПАСНОСТЬ: размеры мира хранятся в тайлах
type BoundaryConstraintSystem struct {
	worldWidth  physics.Tiles
	worldHeight physics.Tiles
}

// NewBoundaryConstraintSystem создаёт новую систему ограничения границами
func NewBoundaryConstraintSystem(worldWidth, worldHeight float32) *BoundaryConstraintSystem {
	return &BoundaryConstraintSystem{
		worldWidth:  physics.NewTiles(worldWidth),
		worldHeight: physics.NewTiles(worldHeight),
	}
}

// Update ограничивает сущности границами мира (ТИПОБЕЗОПАСНО)
func (bcs *BoundaryConstraintSystem) Update(world core.MovementSystemAccess) {
	world.ForEachWith(core.MaskPosition|core.MaskSize, func(entity core.EntityID) {
		// ИСПРАВЛЕНИЕ: Не двигаем едящих животных
		if world.HasComponent(entity, core.MaskEatingState) {
			return // Животное ест, не ограничиваем границами
		}

		pos, _ := world.GetPosition(entity)
		size, _ := world.GetSize(entity)
		changed := false

		// ТИПОБЕЗОПАСНАЯ ЛОГИКА: Size.Radius уже в тайлах
		radiusInTiles := size.Radius

		// ЭЛЕГАНТНАЯ МАТЕМАТИКА: ограничение границ через комплексные числа
		newPos, hitBounds := bcs.constrainToBounds(pos, physics.NewTiles(radiusInTiles))
		if hitBounds {
			bcs.reflectVelocity(world, entity, pos, newPos)
			pos = newPos
			changed = true
		}

		// Обновляем позицию только если она изменилась
		if changed {
			world.SetPosition(entity, pos)
			world.UpdateSpatialPosition(entity, pos) // Обновляем пространственную систему явно
		}
	})
}

// constrainToBounds ограничивает позицию границами мира (ЭЛЕГАНТНО!)
func (bcs *BoundaryConstraintSystem) constrainToBounds(pos core.Position, radius physics.Tiles) (core.Position, bool) {
	// Конвертируем размеры мира в пиксели
	worldWidth := bcs.worldWidth.ToPixels().Float32()
	worldHeight := bcs.worldHeight.ToPixels().Float32()
	radiusPixels := radius.ToPixels().Float32()
	margin := float32(3.2) // 0.1 тайла в пикселях

	// Получаем текущие координаты
	x, y := pos.X, pos.Y
	newX, newY := x, y
	boundsHit := false

	// Ограничиваем X
	leftBound := margin + radiusPixels
	rightBound := worldWidth - margin - radiusPixels
	if x < leftBound {
		newX = leftBound
		boundsHit = true
	} else if x > rightBound {
		newX = rightBound
		boundsHit = true
	}

	// Ограничиваем Y
	topBound := margin + radiusPixels
	bottomBound := worldHeight - margin - radiusPixels
	if y < topBound {
		newY = topBound
		boundsHit = true
	} else if y > bottomBound {
		newY = bottomBound
		boundsHit = true
	}

	return core.NewPosition(newX, newY), boundsHit
}

// reflectVelocity отражает скорость при столкновении с границами (ЭЛЕГАНТНО!)
func (bcs *BoundaryConstraintSystem) reflectVelocity(world core.MovementSystemAccess, entity core.EntityID, oldPos, newPos core.Position) {
	if !world.HasComponent(entity, core.MaskVelocity) {
		return
	}

	vel, _ := world.GetVelocity(entity)
	velX, velY := vel.X, vel.Y
	oldX, oldY := oldPos.X, oldPos.Y
	newX, newY := newPos.X, newPos.Y

	const (
		ReflectionDamping = 0.8 // Коэффициент затухания при отражении
		MinVelocity       = 1.0 // Минимальная скорость для отражения
	)

	// Отражаем X скорость если X координата изменилась
	if oldX != newX {
		if abs32(velX) > MinVelocity {
			velX = -velX * ReflectionDamping
		} else {
			velX = 0
		}
	}

	// Отражаем Y скорость если Y координата изменилась
	if oldY != newY {
		if abs32(velY) > MinVelocity {
			velY = -velY * ReflectionDamping
		} else {
			velY = 0
		}
	}

	// Обновляем скорость
	newVel := core.NewVelocity(velX, velY)
	world.SetVelocity(entity, newVel)
}
