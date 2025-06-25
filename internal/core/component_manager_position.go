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

// Satiation component management

// AddSatiation добавляет компонент Satiation к сущности
func (cm *ComponentManager) AddSatiation(entity EntityID, satiation Satiation) {
	cm.satiations[entity] = satiation

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasSatiation[index] |= 1 << bit
}

// GetSatiation возвращает компонент Satiation сущности
func (cm *ComponentManager) GetSatiation(entity EntityID) (Satiation, bool) {
	if !cm.HasComponent(entity, MaskSatiation) {
		return Satiation{}, false
	}
	return cm.satiations[entity], true
}

// SetSatiation обновляет компонент Satiation сущности
func (cm *ComponentManager) SetSatiation(entity EntityID, satiation Satiation) bool {
	if !cm.HasComponent(entity, MaskSatiation) {
		return false
	}
	cm.satiations[entity] = satiation
	return true
}

// RemoveSatiation удаляет компонент Satiation у сущности
func (cm *ComponentManager) RemoveSatiation(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskSatiation) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasSatiation[index] &= ^(1 << bit)
	cm.satiations[entity] = Satiation{} // Очистка данных

	return true
}
