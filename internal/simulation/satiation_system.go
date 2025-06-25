package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// SatiationSystem управляет только сытостью животных (SRP - Single Responsibility Principle)
// Единственная ответственность: уменьшение сытости со временем
type SatiationSystem struct{}

// NewSatiationSystem создаёт новую систему сытости
func NewSatiationSystem() *SatiationSystem {
	return &SatiationSystem{}
}

// Update обновляет сытость для всех животных
// ISP Улучшение: использует узкоспециализированный интерфейс
func (ss *SatiationSystem) Update(world core.SatiationSystemAccess, deltaTime float32) {
	world.ForEachWith(core.MaskSatiation, func(entity core.EntityID) {
		ss.updateSatiation(world, entity, deltaTime)
	})
}

// updateSatiation обновляет сытость животного
func (ss *SatiationSystem) updateSatiation(world core.SatiationSystemAccess, entity core.EntityID, deltaTime float32) {
	satiation, ok := world.GetSatiation(entity)
	if !ok {
		return
	}

	// ИСПРАВЛЕНИЕ: Животные не теряют сытость когда едят!
	// Проверяем есть ли EatingState - если есть, пропускаем снижение сытости
	if world.HasComponent(entity, core.MaskEatingState) {
		return // Животное ест - сытость не снижается
	}

	// Определяем скорость снижения сытости в зависимости от размера животного
	satiationRate := float32(BaseSatiationDecreaseRate)
	if size, hasSize := world.GetSize(entity); hasSize {
		// Большие животные (хищники) теряют сытость медленнее
		if size.Radius > LargeAnimalSizeThreshold {
			satiationRate *= LargeAnimalSaitationRate
		}
	}

	// Уменьшаем сытость
	satiation.Value -= satiationRate * deltaTime

	// Ограничиваем снизу
	if satiation.Value < 0 {
		satiation.Value = 0
	}

	world.SetSatiation(entity, satiation)
}
