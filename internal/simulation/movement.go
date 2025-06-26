package simulation

import (
	"math"

	"github.com/aiseeq/savanna/internal/core"
)

// Локальные константы для системы коллизий (основные перенесены в game_balance.go)
const (
	// Параметры взаимодействий хищник-добыча (оставляем локальные)
	PredatorPreyDamping = 0.7 // Замедление при коллизии хищник-добыча (70% скорости)

	// Пороги движения (оставляем локальные)
	SoftPushThreshold     = 1.0 // Порог для активации мягкого расталкивания (пикс)
	SlowMovementThreshold = 1.0 // Порог медленного движения (пикс/сек)
)

// Вспомогательные функции для работы с float32
func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

// РЕФАКТОРИНГ: Основные константы коллизий перенесены в game_balance.go:
// - CollisionSearchRadiusMultiplier (было SearchRadiusMultiplier = 2.2)
// - CollisionSeparationForceMultiplier (было SeparationForceMultiplier = 1.5)
// - SoftCollisionPushForce (было SoftPushForce = 3.0)
// - HardCollisionPenetrationThreshold (было PenetrationThreshold = 0.05)
// - HardCollisionPushForceMultiplier (было PushForceMultiplier = 25.0)

// CollisionConstants структура констант для обратной совместимости
// РЕФАКТОРИНГ: Теперь использует константы из game_balance.go
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
	SearchRadiusMultiplier:    CollisionSearchRadiusMultiplier,
	SeparationForceMultiplier: CollisionSeparationForceMultiplier,
	PredatorPreyDamping:       PredatorPreyDamping,
	SoftPushThreshold:         SoftPushThreshold,
	SoftPushForce:             SoftCollisionPushForce,
	PenetrationThreshold:      HardCollisionPenetrationThreshold,
	PushForceMultiplier:       HardCollisionPushForceMultiplier,
	SlowMovementThreshold:     SlowMovementThreshold,
}

// MovementSystem backward-compatible обертка над MovementSystemManager
// РЕФАКТОРИНГ SRP: Теперь делегирует работу специализированным системам
type MovementSystem struct {
	manager *MovementSystemManager // Композиция вместо наследования
}

// NewMovementSystem создаёт новую систему движения
func NewMovementSystem(worldWidth, worldHeight float32) *MovementSystem {
	return &MovementSystem{
		manager: NewMovementSystemManager(worldWidth, worldHeight),
	}
}

// Update обновляет движение всех сущностей
// РЕФАКТОРИНГ SRP: Делегирует работу специализированному менеджеру
func (ms *MovementSystem) Update(world core.MovementSystemAccess, deltaTime float32) {
	ms.manager.Update(world, deltaTime)
}
