package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

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

// Update обновляет позиции всех движущихся сущностей
func (ms *MovementSystem) Update(world *core.World, deltaTime float32) {
	// Обновляем позиции по скорости
	ms.updatePositions(world, deltaTime)

	// Ограничиваем границами мира
	ms.constrainToBounds(world)

	// Обрабатываем коллизии между животными
	ms.handleCollisions(world)
}

// updatePositions обновляет позиции по скорости
func (ms *MovementSystem) updatePositions(world *core.World, deltaTime float32) {
	world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
		pos, _ := world.GetPosition(entity)
		vel, _ := world.GetVelocity(entity)

		// Обновляем позицию
		pos.X += vel.X * deltaTime
		pos.Y += vel.Y * deltaTime

		world.SetPosition(entity, pos)
	})
}

// constrainToBounds ограничивает сущности границами мира
func (ms *MovementSystem) constrainToBounds(world *core.World) {
	world.ForEachWith(core.MaskPosition|core.MaskSize, func(entity core.EntityID) {
		pos, _ := world.GetPosition(entity)
		size, _ := world.GetSize(entity)

		radius := size.Radius
		changed := false

		// Левая граница
		if pos.X-radius < 0 {
			pos.X = radius
			changed = true

			// Если есть скорость, отражаем её
			if world.HasComponent(entity, core.MaskVelocity) {
				vel, _ := world.GetVelocity(entity)
				if vel.X < 0 {
					vel.X = -vel.X * 0.8 // Немного гасим скорость
					world.SetVelocity(entity, vel)
				}
			}
		}

		// Правая граница
		if pos.X+radius > ms.worldWidth {
			pos.X = ms.worldWidth - radius
			changed = true

			if world.HasComponent(entity, core.MaskVelocity) {
				vel, _ := world.GetVelocity(entity)
				if vel.X > 0 {
					vel.X = -vel.X * 0.8
					world.SetVelocity(entity, vel)
				}
			}
		}

		// Верхняя граница
		if pos.Y-radius < 0 {
			pos.Y = radius
			changed = true

			if world.HasComponent(entity, core.MaskVelocity) {
				vel, _ := world.GetVelocity(entity)
				if vel.Y < 0 {
					vel.Y = -vel.Y * 0.8
					world.SetVelocity(entity, vel)
				}
			}
		}

		// Нижняя граница
		if pos.Y+radius > ms.worldHeight {
			pos.Y = ms.worldHeight - radius
			changed = true

			if world.HasComponent(entity, core.MaskVelocity) {
				vel, _ := world.GetVelocity(entity)
				if vel.Y > 0 {
					vel.Y = -vel.Y * 0.8
					world.SetVelocity(entity, vel)
				}
			}
		}

		if changed {
			world.SetPosition(entity, pos)
		}
	})
}

// handleCollisions обрабатывает мягкие коллизии между животными
func (ms *MovementSystem) handleCollisions(world *core.World) {
	// Получаем всех животных с позицией и размером
	animals := world.QueryEntitiesWith(core.MaskPosition | core.MaskSize)

	// Проверяем каждую пару на коллизию
	for i := 0; i < len(animals); i++ {
		for j := i + 1; j < len(animals); j++ {
			entity1 := animals[i]
			entity2 := animals[j]

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
	}
}

// separateEntities мягко расталкивает две сущности при коллизии
func (ms *MovementSystem) separateEntities(world *core.World, entity1, entity2 core.EntityID, collision physics.CollisionDetails) {
	// Разделяем сущности пополам
	separationForce := collision.Penetration * 0.5

	pos1, _ := world.GetPosition(entity1)
	pos2, _ := world.GetPosition(entity2)

	// Применяем разделение
	pos1.X += collision.Normal.X * separationForce
	pos1.Y += collision.Normal.Y * separationForce

	pos2.X -= collision.Normal.X * separationForce
	pos2.Y -= collision.Normal.Y * separationForce

	world.SetPosition(entity1, pos1)
	world.SetPosition(entity2, pos2)

	// Если у сущностей есть скорости, корректируем их
	if world.HasComponent(entity1, core.MaskVelocity) && world.HasComponent(entity2, core.MaskVelocity) {
		vel1, _ := world.GetVelocity(entity1)
		vel2, _ := world.GetVelocity(entity2)

		// Добавляем небольшое отталкивание
		pushForce := float32(20.0) // единиц в секунду

		vel1.X += collision.Normal.X * pushForce
		vel1.Y += collision.Normal.Y * pushForce

		vel2.X -= collision.Normal.X * pushForce
		vel2.Y -= collision.Normal.Y * pushForce

		world.SetVelocity(entity1, vel1)
		world.SetVelocity(entity2, vel2)
	}
}
