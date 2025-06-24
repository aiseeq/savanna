package adapters

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Адаптеры для систем с ISP интерфейсами для совместимости со старым интерфейсом System
// Рефакторинг SRP: разделены специализированные системы вместо монолитного FeedingSystem

// ВРЕМЕННОЕ ВОССТАНОВЛЕНИЕ для совместимости тестов
// TODO: Заменить все тесты на использование common.CreateTestSystemManager()

// DeprecatedFeedingSystemAdapter DEPRECATED: используйте common.CreateTestSystemManager()
type DeprecatedFeedingSystemAdapter struct {
	systemManager *core.SystemManager
}

// NewDeprecatedFeedingSystemAdapter создаёт временный адаптер для совместимости
func NewDeprecatedFeedingSystemAdapter(vegetation *simulation.VegetationSystem) *DeprecatedFeedingSystemAdapter {
	// Создаем мини-менеджер только с системами питания
	manager := core.NewSystemManager()
	manager.AddSystem(&HungerSystemAdapter{System: simulation.NewHungerSystem()})
	manager.AddSystem(&GrassSearchSystemAdapter{System: simulation.NewGrassSearchSystem(vegetation)})
	manager.AddSystem(&GrassEatingSystemAdapter{System: simulation.NewGrassEatingSystem(vegetation)})
	manager.AddSystem(&HungerSpeedModifierSystemAdapter{System: simulation.NewHungerSpeedModifierSystem()})
	manager.AddSystem(&StarvationDamageSystemAdapter{System: simulation.NewStarvationDamageSystem()})

	return &DeprecatedFeedingSystemAdapter{systemManager: manager}
}

func (a *DeprecatedFeedingSystemAdapter) Update(world *core.World, deltaTime float32) {
	a.systemManager.Update(world, deltaTime)
}

// FeedingSystemAdapter DEPRECATED: алиас для совместимости
type FeedingSystemAdapter = DeprecatedFeedingSystemAdapter

// NewFeedingSystemAdapter DEPRECATED: алиас для совместимости
func NewFeedingSystemAdapter(vegetation *simulation.VegetationSystem) *FeedingSystemAdapter {
	return NewDeprecatedFeedingSystemAdapter(vegetation)
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
	if a.System == nil {
		return
	}
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
