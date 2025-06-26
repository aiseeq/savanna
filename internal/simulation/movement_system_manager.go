package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// MovementSystemManager координирует специализированные системы движения (Facade Pattern)
// Соблюдает принципы SRP, OCP и Composition over Inheritance
type MovementSystemManager struct {
	positionUpdateSystem     *PositionUpdateSystem
	collisionSystem          *CollisionSystem
	boundaryConstraintSystem *BoundaryConstraintSystem
}

// NewMovementSystemManager создаёт новый менеджер систем движения
func NewMovementSystemManager(worldWidth, worldHeight float32) *MovementSystemManager {
	return &MovementSystemManager{
		positionUpdateSystem:     NewPositionUpdateSystem(),
		collisionSystem:          NewCollisionSystem(),
		boundaryConstraintSystem: NewBoundaryConstraintSystem(worldWidth, worldHeight),
	}
}

// Update обновляет все системы движения в правильном порядке
func (msm *MovementSystemManager) Update(world core.MovementSystemAccess, deltaTime float32) {
	// 1. Обновляем позиции по скорости
	msm.positionUpdateSystem.Update(world, deltaTime)

	// 2. Обрабатываем коллизии между животными (мягкое расталкивание как в StarCraft)
	msm.collisionSystem.Update(world)

	// 3. ИСПРАВЛЕНИЕ: Ограничиваем границами мира ПОСЛЕ коллизий
	// чтобы расталкивание не выталкивало животных за границы
	msm.boundaryConstraintSystem.Update(world)
}
