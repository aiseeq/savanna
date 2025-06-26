package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// PositionUpdateSystem отвечает ТОЛЬКО за обновление позиций по скорости (SRP)
// Выделена из MovementSystem для соблюдения принципа единственной ответственности
type PositionUpdateSystem struct{}

// NewPositionUpdateSystem создаёт новую систему обновления позиций
func NewPositionUpdateSystem() *PositionUpdateSystem {
	return &PositionUpdateSystem{}
}

// Update обновляет позиции всех сущностей по их скорости
func (pus *PositionUpdateSystem) Update(world core.MovementSystemAccess, deltaTime float32) {
	world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
		if pus.shouldSkipPositionUpdate(world, entity) {
			return
		}

		vel, _ := world.GetVelocity(entity)
		pos, _ := world.GetPosition(entity)

		pus.updateEntityPosition(world, entity, pos, vel, deltaTime)
	})
}

// shouldSkipPositionUpdate проверяет нужно ли пропустить обновление позиции
func (pus *PositionUpdateSystem) shouldSkipPositionUpdate(world core.MovementSystemAccess, entity core.EntityID) bool {
	// КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Проверяем EatingState в САМОМ НАЧАЛЕ
	if world.HasComponent(entity, core.MaskEatingState) {
		// Сбрасываем скорость в ноль для едящих животных
		world.SetVelocity(entity, core.Velocity{X: 0, Y: 0})
		return true // Животное ест, не двигается
	}

	// Читаем скорость
	vel, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return true // Нет скорости
	}

	// Пропускаем неподвижных животных (оптимизация)
	return vel.X == 0 && vel.Y == 0
}

// updateEntityPosition обновляет позицию конкретной сущности (ТИПОБЕЗОПАСНО)
func (pus *PositionUpdateSystem) updateEntityPosition(
	world core.MovementSystemAccess,
	entity core.EntityID,
	pos core.Position,
	vel core.Velocity,
	deltaTime float32,
) {
	// Обновление позиции

	// Вычисляем смещение (скорость * время) и конвертируем из тайлов в пиксели
	displacementX := vel.X * deltaTime * 32.0 // 32 пикселя = 1 тайл
	displacementY := vel.Y * deltaTime * 32.0

	// Обновляем позицию
	newPosition := core.Position{
		X: pos.X + displacementX,
		Y: pos.Y + displacementY,
	}

	world.SetPosition(entity, newPosition)
	world.UpdateSpatialPosition(entity, newPosition) // Обновляем пространственную систему явно
}
