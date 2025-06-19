package core

import "github.com/aiseeq/savanna/internal/constants"

// Position component management

// AddPosition добавляет компонент Position к сущности
func (cm *ComponentManager) AddPosition(entity EntityID, position Position) {
	cm.positions[entity] = position

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasPosition[index] |= 1 << bit
}

// GetPosition возвращает компонент Position сущности
func (cm *ComponentManager) GetPosition(entity EntityID) (Position, bool) {
	if !cm.HasComponent(entity, MaskPosition) {
		return Position{}, false
	}
	return cm.positions[entity], true
}

// SetPosition обновляет компонент Position сущности
func (cm *ComponentManager) SetPosition(entity EntityID, position Position) bool {
	if !cm.HasComponent(entity, MaskPosition) {
		return false
	}
	cm.positions[entity] = position
	return true
}

// RemovePosition удаляет компонент Position у сущности
func (cm *ComponentManager) RemovePosition(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskPosition) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasPosition[index] &= ^(1 << bit)
	cm.positions[entity] = Position{} // Очистка данных

	return true
}

// Velocity component management

// AddVelocity добавляет компонент Velocity к сущности
func (cm *ComponentManager) AddVelocity(entity EntityID, velocity Velocity) {
	cm.velocities[entity] = velocity

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasVelocity[index] |= 1 << bit
}

// GetVelocity возвращает компонент Velocity сущности
func (cm *ComponentManager) GetVelocity(entity EntityID) (Velocity, bool) {
	if !cm.HasComponent(entity, MaskVelocity) {
		return Velocity{}, false
	}
	return cm.velocities[entity], true
}

// SetVelocity обновляет компонент Velocity сущности
func (cm *ComponentManager) SetVelocity(entity EntityID, velocity Velocity) bool {
	if !cm.HasComponent(entity, MaskVelocity) {
		return false
	}
	cm.velocities[entity] = velocity
	return true
}

// RemoveVelocity удаляет компонент Velocity у сущности
func (cm *ComponentManager) RemoveVelocity(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskVelocity) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasVelocity[index] &= ^(1 << bit)
	cm.velocities[entity] = Velocity{} // Очистка данных

	return true
}

// Health component management

// AddHealth добавляет компонент Health к сущности
func (cm *ComponentManager) AddHealth(entity EntityID, health Health) {
	cm.healths[entity] = health

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasHealth[index] |= 1 << bit
}

// GetHealth возвращает компонент Health сущности
func (cm *ComponentManager) GetHealth(entity EntityID) (Health, bool) {
	if !cm.HasComponent(entity, MaskHealth) {
		return Health{}, false
	}
	return cm.healths[entity], true
}

// SetHealth обновляет компонент Health сущности
func (cm *ComponentManager) SetHealth(entity EntityID, health Health) bool {
	if !cm.HasComponent(entity, MaskHealth) {
		return false
	}
	cm.healths[entity] = health
	return true
}

// RemoveHealth удаляет компонент Health у сущности
func (cm *ComponentManager) RemoveHealth(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskHealth) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasHealth[index] &= ^(1 << bit)
	cm.healths[entity] = Health{} // Очистка данных

	return true
}

// Hunger component management

// AddHunger добавляет компонент Hunger к сущности
func (cm *ComponentManager) AddHunger(entity EntityID, hunger Hunger) {
	cm.hungers[entity] = hunger

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasHunger[index] |= 1 << bit
}

// GetHunger возвращает компонент Hunger сущности
func (cm *ComponentManager) GetHunger(entity EntityID) (Hunger, bool) {
	if !cm.HasComponent(entity, MaskHunger) {
		return Hunger{}, false
	}
	return cm.hungers[entity], true
}

// SetHunger обновляет компонент Hunger сущности
func (cm *ComponentManager) SetHunger(entity EntityID, hunger Hunger) bool {
	if !cm.HasComponent(entity, MaskHunger) {
		return false
	}
	cm.hungers[entity] = hunger
	return true
}

// RemoveHunger удаляет компонент Hunger у сущности
func (cm *ComponentManager) RemoveHunger(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskHunger) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasHunger[index] &= ^(1 << bit)
	cm.hungers[entity] = Hunger{} // Очистка данных

	return true
}
