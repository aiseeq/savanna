package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// HungerSpeedModifierSystem изменяет скорость в зависимости от голода (SRP)
// Единственная ответственность: влияние голода на скорость движения
type HungerSpeedModifierSystem struct{}

// NewHungerSpeedModifierSystem создаёт новую систему изменения скорости
func NewHungerSpeedModifierSystem() *HungerSpeedModifierSystem {
	return &HungerSpeedModifierSystem{}
}

// Update обновляет скорости на основе голода
// ISP Улучшение: использует узкоспециализированный интерфейс
func (hsms *HungerSpeedModifierSystem) Update(world core.HungerSpeedModifierSystemAccess, _ float32) {
	world.ForEachWith(core.MaskHunger|core.MaskSpeed, func(entity core.EntityID) {
		hsms.updateSpeedBasedOnHunger(world, entity)
	})
}

// updateSpeedBasedOnHunger обновляет скорость животного на основе сытости и здоровья
// НОВАЯ ЛОГИКА (по требованию пользователя):
// 1. Голодные (< 80%) бегают с полной скоростью (1.0)
// 2. Сытые (> 80%) замедляются: скорость *= (1 + 0.8 - сытость/100)
func (hsms *HungerSpeedModifierSystem) updateSpeedBasedOnHunger(
	world core.HungerSpeedModifierSystemAccess,
	entity core.EntityID,
) {
	hunger, hasHunger := world.GetHunger(entity)
	if !hasHunger {
		return
	}

	speed, hasSpeed := world.GetSpeed(entity)
	if !hasSpeed {
		return
	}

	health, hasHealth := world.GetHealth(entity)
	if !hasHealth {
		return
	}

	var speedMultiplier float32 = NormalSpeedMultiplier

	// НОВАЯ ЛОГИКА 1: Сытость влияет на скорость только при > 80%
	if hunger.Value > SatiatedThreshold {
		// Сытые животные замедляются: скорость *= (1 + 0.8 - сытость)
		// где сытость в долях от 1.0 (90% = 0.9, 95% = 0.95)
		satietyRatio := hunger.Value / PercentToRatioConversion
		speedMultiplier = NormalSpeedMultiplier + SatietySlowdownOffset - satietyRatio

		// Минимальная скорость не меньше 0.1 (для безопасности)
		if speedMultiplier < MinimumSpeedMultiplier {
			speedMultiplier = MinimumSpeedMultiplier
		}
	}
	// Голодные (< 80%) бегают с полной скоростью (speedMultiplier = 1.0)

	// НОВАЯ ЛОГИКА 2: Здоровье влияет на скорость линейно (только если теряет хиты)
	if health.Current < health.Max {
		// Раненое животное: скорость *= (процент_здоровья / 100)
		healthRatio := float32(health.Current) / float32(health.Max)
		speedMultiplier *= healthRatio
	}
	// Здоровые животные (100% хитов) не получают штрафа

	// Обновляем текущую скорость
	speed.Current = speed.Base * speedMultiplier
	world.SetSpeed(entity, speed)
}
