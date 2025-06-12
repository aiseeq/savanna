package core

import "github.com/aiseeq/savanna/internal/physics"

// Методы для работы с компонентами в World

// HasComponent проверяет наличие компонента у сущности
func (w *World) HasComponent(entity EntityID, component ComponentMask) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	switch component {
	case MaskPosition:
		return testBitMask(w.hasPosition[:], entity)
	case MaskVelocity:
		return testBitMask(w.hasVelocity[:], entity)
	case MaskHealth:
		return testBitMask(w.hasHealth[:], entity)
	case MaskHunger:
		return testBitMask(w.hasHunger[:], entity)
	case MaskAge:
		return testBitMask(w.hasAge[:], entity)
	case MaskAnimalType:
		return testBitMask(w.hasType[:], entity)
	case MaskSize:
		return testBitMask(w.hasSize[:], entity)
	case MaskSpeed:
		return testBitMask(w.hasSpeed[:], entity)
	}
	return false
}

// HasComponents проверяет наличие всех указанных компонентов
func (w *World) HasComponents(entity EntityID, mask ComponentMask) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	// Проверяем каждый бит в маске
	if mask&MaskPosition != 0 && !testBitMask(w.hasPosition[:], entity) {
		return false
	}
	if mask&MaskVelocity != 0 && !testBitMask(w.hasVelocity[:], entity) {
		return false
	}
	if mask&MaskHealth != 0 && !testBitMask(w.hasHealth[:], entity) {
		return false
	}
	if mask&MaskHunger != 0 && !testBitMask(w.hasHunger[:], entity) {
		return false
	}
	if mask&MaskAge != 0 && !testBitMask(w.hasAge[:], entity) {
		return false
	}
	if mask&MaskAnimalType != 0 && !testBitMask(w.hasType[:], entity) {
		return false
	}
	if mask&MaskSize != 0 && !testBitMask(w.hasSize[:], entity) {
		return false
	}
	if mask&MaskSpeed != 0 && !testBitMask(w.hasSpeed[:], entity) {
		return false
	}
	
	return true
}

// Position Component

// AddPosition добавляет компонент Position
func (w *World) AddPosition(entity EntityID, position Position) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.positions[entity] = position
	setBitMask(w.hasPosition[:], entity)
	
	// Обновляем пространственную сетку если есть размер
	if w.HasComponent(entity, MaskSize) {
		size := w.sizes[entity]
		w.spatialGrid.Update(physics.EntityID(entity), 
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}
	
	return true
}

// GetPosition получает компонент Position
func (w *World) GetPosition(entity EntityID) (Position, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasPosition[:], entity) {
		return Position{}, false
	}
	return w.positions[entity], true
}

// SetPosition изменяет позицию сущности
func (w *World) SetPosition(entity EntityID, position Position) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasPosition[:], entity) {
		return false
	}
	
	w.positions[entity] = position
	
	// Обновляем пространственную сетку если есть размер
	if w.HasComponent(entity, MaskSize) {
		size := w.sizes[entity]
		w.spatialGrid.Update(physics.EntityID(entity), 
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}
	
	return true
}

// RemovePosition удаляет компонент Position
func (w *World) RemovePosition(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasPosition[:], entity) {
		return false
	}
	
	clearBitMask(w.hasPosition[:], entity)
	w.spatialGrid.Remove(physics.EntityID(entity))
	return true
}

// Velocity Component

// AddVelocity добавляет компонент Velocity
func (w *World) AddVelocity(entity EntityID, velocity Velocity) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.velocities[entity] = velocity
	setBitMask(w.hasVelocity[:], entity)
	return true
}

// GetVelocity получает компонент Velocity
func (w *World) GetVelocity(entity EntityID) (Velocity, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasVelocity[:], entity) {
		return Velocity{}, false
	}
	return w.velocities[entity], true
}

// SetVelocity изменяет скорость сущности
func (w *World) SetVelocity(entity EntityID, velocity Velocity) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasVelocity[:], entity) {
		return false
	}
	
	w.velocities[entity] = velocity
	return true
}

// RemoveVelocity удаляет компонент Velocity
func (w *World) RemoveVelocity(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasVelocity[:], entity) {
		return false
	}
	
	clearBitMask(w.hasVelocity[:], entity)
	return true
}

// Health Component

// AddHealth добавляет компонент Health
func (w *World) AddHealth(entity EntityID, health Health) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.healths[entity] = health
	setBitMask(w.hasHealth[:], entity)
	return true
}

// GetHealth получает компонент Health
func (w *World) GetHealth(entity EntityID) (Health, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasHealth[:], entity) {
		return Health{}, false
	}
	return w.healths[entity], true
}

// SetHealth изменяет здоровье сущности
func (w *World) SetHealth(entity EntityID, health Health) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasHealth[:], entity) {
		return false
	}
	
	w.healths[entity] = health
	return true
}

// RemoveHealth удаляет компонент Health
func (w *World) RemoveHealth(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasHealth[:], entity) {
		return false
	}
	
	clearBitMask(w.hasHealth[:], entity)
	return true
}

// Hunger Component

// AddHunger добавляет компонент Hunger
func (w *World) AddHunger(entity EntityID, hunger Hunger) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.hungers[entity] = hunger
	setBitMask(w.hasHunger[:], entity)
	return true
}

// GetHunger получает компонент Hunger
func (w *World) GetHunger(entity EntityID) (Hunger, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasHunger[:], entity) {
		return Hunger{}, false
	}
	return w.hungers[entity], true
}

// SetHunger изменяет голод сущности
func (w *World) SetHunger(entity EntityID, hunger Hunger) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasHunger[:], entity) {
		return false
	}
	
	w.hungers[entity] = hunger
	return true
}

// RemoveHunger удаляет компонент Hunger
func (w *World) RemoveHunger(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasHunger[:], entity) {
		return false
	}
	
	clearBitMask(w.hasHunger[:], entity)
	return true
}

// Age Component

// AddAge добавляет компонент Age
func (w *World) AddAge(entity EntityID, age Age) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.ages[entity] = age
	setBitMask(w.hasAge[:], entity)
	return true
}

// GetAge получает компонент Age
func (w *World) GetAge(entity EntityID) (Age, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasAge[:], entity) {
		return Age{}, false
	}
	return w.ages[entity], true
}

// SetAge изменяет возраст сущности
func (w *World) SetAge(entity EntityID, age Age) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasAge[:], entity) {
		return false
	}
	
	w.ages[entity] = age
	return true
}

// RemoveAge удаляет компонент Age
func (w *World) RemoveAge(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasAge[:], entity) {
		return false
	}
	
	clearBitMask(w.hasAge[:], entity)
	return true
}

// AnimalType Component

// AddAnimalType добавляет компонент AnimalType
func (w *World) AddAnimalType(entity EntityID, animalType AnimalType) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.types[entity] = animalType
	setBitMask(w.hasType[:], entity)
	return true
}

// GetAnimalType получает компонент AnimalType
func (w *World) GetAnimalType(entity EntityID) (AnimalType, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasType[:], entity) {
		return TypeNone, false
	}
	return w.types[entity], true
}

// SetAnimalType изменяет тип животного сущности
func (w *World) SetAnimalType(entity EntityID, animalType AnimalType) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasType[:], entity) {
		return false
	}
	
	w.types[entity] = animalType
	return true
}

// RemoveAnimalType удаляет компонент AnimalType
func (w *World) RemoveAnimalType(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasType[:], entity) {
		return false
	}
	
	clearBitMask(w.hasType[:], entity)
	return true
}

// Size Component

// AddSize добавляет компонент Size
func (w *World) AddSize(entity EntityID, size Size) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.sizes[entity] = size
	setBitMask(w.hasSize[:], entity)
	
	// Обновляем пространственную сетку если есть позиция
	if w.HasComponent(entity, MaskPosition) {
		position := w.positions[entity]
		w.spatialGrid.Update(physics.EntityID(entity), 
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}
	
	return true
}

// GetSize получает компонент Size
func (w *World) GetSize(entity EntityID) (Size, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasSize[:], entity) {
		return Size{}, false
	}
	return w.sizes[entity], true
}

// SetSize изменяет размер сущности
func (w *World) SetSize(entity EntityID, size Size) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasSize[:], entity) {
		return false
	}
	
	w.sizes[entity] = size
	
	// Обновляем пространственную сетку если есть позиция
	if w.HasComponent(entity, MaskPosition) {
		position := w.positions[entity]
		w.spatialGrid.Update(physics.EntityID(entity), 
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}
	
	return true
}

// RemoveSize удаляет компонент Size
func (w *World) RemoveSize(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasSize[:], entity) {
		return false
	}
	
	clearBitMask(w.hasSize[:], entity)
	return true
}

// Speed Component

// AddSpeed добавляет компонент Speed
func (w *World) AddSpeed(entity EntityID, speed Speed) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	w.speeds[entity] = speed
	setBitMask(w.hasSpeed[:], entity)
	return true
}

// GetSpeed получает компонент Speed
func (w *World) GetSpeed(entity EntityID) (Speed, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasSpeed[:], entity) {
		return Speed{}, false
	}
	return w.speeds[entity], true
}

// SetSpeed изменяет скорость сущности
func (w *World) SetSpeed(entity EntityID, speed Speed) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasSpeed[:], entity) {
		return false
	}
	
	w.speeds[entity] = speed
	return true
}

// RemoveSpeed удаляет компонент Speed
func (w *World) RemoveSpeed(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasSpeed[:], entity) {
		return false
	}
	
	clearBitMask(w.hasSpeed[:], entity)
	return true
}