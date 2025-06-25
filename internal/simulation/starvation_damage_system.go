package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// StarvationDamageSystem наносит урон голодающим животным (SRP)
// Единственная ответственность: урон здоровью при голоде
type StarvationDamageSystem struct {
	healthDamageTimer float32 // Таймер для нанесения урона здоровью (раз в секунду)
}

// NewStarvationDamageSystem создаёт новую систему урона от голода
func NewStarvationDamageSystem() *StarvationDamageSystem {
	return &StarvationDamageSystem{
		healthDamageTimer: 0,
	}
}

// Update наносит урон голодающим животным
// ISP Улучшение: использует узкоспециализированный интерфейс
func (sds *StarvationDamageSystem) Update(world core.StarvationDamageSystemAccess, deltaTime float32) {
	sds.healthDamageTimer += deltaTime

	// Наносим урон здоровью голодающим животным (раз в секунду)
	if sds.healthDamageTimer >= 1.0 {
		sds.damageStarvingAnimals(world)
		sds.healthDamageTimer = 0
	}
}

// damageStarvingAnimals наносит урон здоровью голодающим животным
func (sds *StarvationDamageSystem) damageStarvingAnimals(world core.StarvationDamageSystemAccess) {
	world.ForEachWith(core.MaskSatiation|core.MaskHealth, func(entity core.EntityID) {
		hunger, hasHunger := world.GetSatiation(entity)
		if !hasHunger {
			return
		}

		// Проверяем критический голод (0%)
		if hunger.Value > 0 {
			return
		}

		health, hasHealth := world.GetHealth(entity)
		if !hasHealth {
			return
		}

		// Наносим урон от голода
		health.Current -= StarvationDamagePerSecond
		if health.Current < 0 {
			health.Current = 0
		}

		world.SetHealth(entity, health)
	})
}
