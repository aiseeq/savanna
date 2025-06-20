package adapters

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Адаптеры для систем с ISP интерфейсами для совместимости со старым интерфейсом System
// Рефакторинг SRP: разделены специализированные системы вместо монолитного FeedingSystem

// DeprecatedFeedingSystemAdapter DEPRECATED: используйте новые специализированные системы
// Оставлен для обратной совместимости с тестами
type DeprecatedFeedingSystemAdapter struct {
	hungerSystem        *simulation.HungerSystem
	grassSearchSystem   *simulation.GrassSearchSystem
	grassEatingSystem   *simulation.GrassEatingSystem
	hungerSpeedModifier *simulation.HungerSpeedModifierSystem
	starvationDamage    *simulation.StarvationDamageSystem
}

// NewDeprecatedFeedingSystemAdapter создаёт адаптер для обратной совместимости
func NewDeprecatedFeedingSystemAdapter(vegetation *simulation.VegetationSystem) *DeprecatedFeedingSystemAdapter {
	return &DeprecatedFeedingSystemAdapter{
		hungerSystem:        simulation.NewHungerSystem(),
		grassSearchSystem:   simulation.NewGrassSearchSystem(vegetation),
		grassEatingSystem:   simulation.NewGrassEatingSystem(vegetation),
		hungerSpeedModifier: simulation.NewHungerSpeedModifierSystem(),
		starvationDamage:    simulation.NewStarvationDamageSystem(),
	}
}

func (a *DeprecatedFeedingSystemAdapter) Update(world *core.World, deltaTime float32) {
	// Выполняем все 5 систем в правильном порядке (согласно CLAUDE.md)
	a.hungerSystem.Update(world, deltaTime)        // 1. Управление голодом
	a.grassSearchSystem.Update(world, deltaTime)   // 2. Поиск травы и создание EatingState
	a.grassEatingSystem.Update(world, deltaTime)   // 3. Дискретное поедание травы
	a.hungerSpeedModifier.Update(world, deltaTime) // 4. Влияние голода на скорость
	a.starvationDamage.Update(world, deltaTime)    // 5. Урон от голода
}

// FeedingSystemAdapter DEPRECATED: структура для обратной совместимости
type FeedingSystemAdapter struct {
	*DeprecatedFeedingSystemAdapter
}

// NewFeedingSystemAdapter DEPRECATED: создаёт адаптер для обратной совместимости
func NewFeedingSystemAdapter(vegetation *simulation.VegetationSystem) *FeedingSystemAdapter {
	return &FeedingSystemAdapter{
		DeprecatedFeedingSystemAdapter: NewDeprecatedFeedingSystemAdapter(vegetation),
	}
}

// HungerSystemAdapter адаптирует HungerSystem к старому интерфейсу System
type HungerSystemAdapter struct {
	System *simulation.HungerSystem
}

func (a *HungerSystemAdapter) Update(world *core.World, deltaTime float32) {
	if a.System == nil {
		return
	}
	a.System.Update(world, deltaTime)
}

// GrassSearchSystemAdapter адаптирует GrassSearchSystem к старому интерфейсу System
type GrassSearchSystemAdapter struct {
	System *simulation.GrassSearchSystem
}

func (a *GrassSearchSystemAdapter) Update(world *core.World, deltaTime float32) {
	if a.System == nil {
		return
	}
	a.System.Update(world, deltaTime)
}

// HungerSpeedModifierSystemAdapter адаптирует HungerSpeedModifierSystem к старому интерфейсу System
type HungerSpeedModifierSystemAdapter struct {
	System *simulation.HungerSpeedModifierSystem
}

func (a *HungerSpeedModifierSystemAdapter) Update(world *core.World, deltaTime float32) {
	if a.System == nil {
		return
	}
	a.System.Update(world, deltaTime)
}

// StarvationDamageSystemAdapter адаптирует StarvationDamageSystem к старому интерфейсу System
type StarvationDamageSystemAdapter struct {
	System *simulation.StarvationDamageSystem
}

func (a *StarvationDamageSystemAdapter) Update(world *core.World, deltaTime float32) {
	if a.System == nil {
		return
	}
	a.System.Update(world, deltaTime)
}

// BehaviorSystemAdapter адаптирует AnimalBehaviorSystem к старому интерфейсу System
type BehaviorSystemAdapter struct {
	System *simulation.AnimalBehaviorSystem
}

func (a *BehaviorSystemAdapter) Update(world *core.World, deltaTime float32) {
	a.System.Update(world, deltaTime)
}

// MovementSystemAdapter адаптирует MovementSystem к старому интерфейсу System
type MovementSystemAdapter struct {
	System *simulation.MovementSystem
}

func (a *MovementSystemAdapter) Update(world *core.World, deltaTime float32) {
	a.System.Update(world, deltaTime)
}

// GrassEatingSystemAdapter адаптирует GrassEatingSystem к старому интерфейсу System
type GrassEatingSystemAdapter struct {
	System *simulation.GrassEatingSystem
}

func (a *GrassEatingSystemAdapter) Update(world *core.World, deltaTime float32) {
	a.System.Update(world, deltaTime)
}
