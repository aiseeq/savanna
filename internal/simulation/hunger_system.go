package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// HungerSystem управляет только голодом животных (SRP - Single Responsibility Principle)
// Единственная ответственность: уменьшение голода со временем
type HungerSystem struct{}

// NewHungerSystem создаёт новую систему голода
func NewHungerSystem() *HungerSystem {
	return &HungerSystem{}
}

// Update обновляет голод для всех животных
// ISP Улучшение: использует узкоспециализированный интерфейс
func (hs *HungerSystem) Update(world core.HungerSystemAccess, deltaTime float32) {
	world.ForEachWith(core.MaskHunger, func(entity core.EntityID) {
		hs.updateHunger(world, entity, deltaTime)
	})
}

// updateHunger обновляет голод животного
func (hs *HungerSystem) updateHunger(world core.HungerSystemAccess, entity core.EntityID, deltaTime float32) {
	hunger, ok := world.GetHunger(entity)
	if !ok {
		return
	}

	// Определяем скорость голода в зависимости от размера животного
	hungerRate := float32(BaseHungerDecreaseRate)
	if size, hasSize := world.GetSize(entity); hasSize {
		// Большие животные (хищники) голодают медленнее
		if size.Radius > LargeAnimalSizeThreshold {
			hungerRate *= LargeAnimalHungerRate
		}
	}

	// Уменьшаем голод
	hunger.Value -= hungerRate * deltaTime

	// Ограничиваем снизу
	if hunger.Value < 0 {
		hunger.Value = 0
	}

	world.SetHunger(entity, hunger)
}
