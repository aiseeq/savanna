package adapters

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Адаптеры для систем с ISP интерфейсами для совместимости со старым интерфейсом System
// TODO: Удалить когда все системы перейдут на специализированные интерфейсы

// FeedingSystemAdapter адаптирует FeedingSystem к старому интерфейсу System
type FeedingSystemAdapter struct {
	System *simulation.FeedingSystem
}

func (a *FeedingSystemAdapter) Update(world *core.World, deltaTime float32) {
	// Debug: проверяем что система существует
	if a.System == nil {
		return
	}

	// Debug: логируем что адаптер вызывается
	// fmt.Printf("DEBUG: FeedingSystemAdapter.Update вызван\n")

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
