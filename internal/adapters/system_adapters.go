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
	manager.AddSystem(&SatiationSystemAdapter{System: simulation.NewSatiationSystem()})
	manager.AddSystem(&GrassSearchSystemAdapter{System: simulation.NewGrassSearchSystem(vegetation)})
	manager.AddSystem(&GrassEatingSystemAdapter{System: simulation.NewGrassEatingSystem(vegetation)})
	manager.AddSystem(&SatiationSpeedModifierSystemAdapter{System: simulation.NewSatiationSpeedModifierSystem()})
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

// SatiationSystemAdapter адаптирует SatiationSystem к старому интерфейсу System
type SatiationSystemAdapter struct {
	System *simulation.SatiationSystem
}

func (a *SatiationSystemAdapter) Update(world *core.World, deltaTime float32) {
	if a.System == nil {
		return
	}
	a.System.Update(world, deltaTime)
}

// HungerSystemAdapter DEPRECATED: алиас для совместимости
type HungerSystemAdapter = SatiationSystemAdapter

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

// SatiationSpeedModifierSystemAdapter адаптирует SatiationSpeedModifierSystem к старому интерфейсу System
type SatiationSpeedModifierSystemAdapter struct {
	System *simulation.SatiationSpeedModifierSystem
}

func (a *SatiationSpeedModifierSystemAdapter) Update(world *core.World, deltaTime float32) {
	if a.System == nil {
		return
	}
	a.System.Update(world, deltaTime)
}

// HungerSpeedModifierSystemAdapter DEPRECATED: алиас для совместимости
type HungerSpeedModifierSystemAdapter = SatiationSpeedModifierSystemAdapter

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
