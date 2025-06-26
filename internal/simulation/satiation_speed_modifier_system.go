package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// SatiationSpeedModifierSystem изменяет скорость в зависимости от сытости (SRP)
// Единственная ответственность: влияние сытости на скорость движения
type SatiationSpeedModifierSystem struct{}

// NewSatiationSpeedModifierSystem создаёт новую систему изменения скорости
func NewSatiationSpeedModifierSystem() *SatiationSpeedModifierSystem {
	return &SatiationSpeedModifierSystem{}
}

// Update обновляет скорости на основе сытости
// ISP Улучшение: использует узкоспециализированный интерфейс
func (ssms *SatiationSpeedModifierSystem) Update(world core.SatiationSpeedModifierSystemAccess, _ float32) {
	world.ForEachWith(core.MaskSatiation|core.MaskSpeed, func(entity core.EntityID) {
		ssms.updateSpeedBasedOnSatiation(world, entity)
	})
}

// updateSpeedBasedOnSatiation обновляет скорость животного на основе сытости и здоровья
// НОВАЯ ЛОГИКА (по требованию пользователя):
// 1. Малосытные (< 80%) бегают с полной скоростью (1.0)
// 2. Сытые (> 80%) замедляются: скорость *= (1 + 0.8 - сытость/100)
func (ssms *SatiationSpeedModifierSystem) updateSpeedBasedOnSatiation(
	world core.SatiationSpeedModifierSystemAccess,
	entity core.EntityID,
) {
	satiation, hasSatiation := world.GetSatiation(entity)
	if !hasSatiation {
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
	if satiation.Value > SatiatedThreshold {
		// Сытые животные замедляются: скорость *= (1 + 0.8 - сытость)
		// где сытость в долях от 1.0 (90% = 0.9, 95% = 0.95)
		satietyRatio := satiation.Value / PercentToRatioConversion
		speedMultiplier = NormalSpeedMultiplier + SatietySlowdownOffset - satietyRatio

		// Минимальная скорость не меньше 0.1 (для безопасности)
		if speedMultiplier < MinimumSpeedMultiplier {
			speedMultiplier = MinimumSpeedMultiplier
		}
	}
	// Малосытные (< 80%) бегают с полной скоростью (speedMultiplier = 1.0)

	// НОВАЯ ЛОГИКА 2: Здоровье влияет на скорость линейно (только если теряет хиты)
	if health.Current < health.Max {
		// Раненое животное: скорость *= (процент_здоровья / 100)
		healthRatio := float32(health.Current) / float32(health.Max)
		speedMultiplier *= healthRatio
	}
	// Здоровые животные (100% хитов) не получают штрафа

	// Обновляем текущую скорость (ТИПОБЕЗОПАСНО)
	speed.Current = speed.Base * speedMultiplier
	world.SetSpeed(entity, speed)
}
