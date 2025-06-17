package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// DamageSystem отвечает ТОЛЬКО за эффекты урона (устраняет нарушение SRP)
type DamageSystem struct{}

// NewDamageSystem создаёт новую систему эффектов урона
func NewDamageSystem() *DamageSystem {
	return &DamageSystem{}
}

// Update обновляет систему эффектов урона
func (ds *DamageSystem) Update(world *core.World, deltaTime float32) {
	// Обновляем эффекты мигания при уроне
	ds.updateDamageFlashes(world, deltaTime)
}

// updateDamageFlashes обновляет эффекты мигания при уроне
func (ds *DamageSystem) updateDamageFlashes(world *core.World, deltaTime float32) {
	// Список сущностей для удаления эффекта
	var toRemove []core.EntityID

	world.ForEachWith(core.MaskDamageFlash, func(entity core.EntityID) {
		flash, hasFlash := world.GetDamageFlash(entity)
		if !hasFlash {
			return
		}

		// Уменьшаем таймер
		flash.Timer -= deltaTime
		if flash.Timer <= 0 {
			// Эффект закончился
			toRemove = append(toRemove, entity)
		} else {
			// Обновляем интенсивность (уменьшается со временем для плавного затухания)
			flash.Intensity = flash.Timer / flash.Duration
			world.SetDamageFlash(entity, flash)
		}
	})

	// Удаляем завершенные эффекты
	for _, entity := range toRemove {
		world.RemoveDamageFlash(entity)
	}
}
