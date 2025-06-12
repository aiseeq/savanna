package core

// System интерфейс для всех систем симуляции
type System interface {
	// Update выполняет один кадр обновления системы
	Update(world *World, deltaTime float32)
}

// SystemManager управляет набором систем и их выполнением
type SystemManager struct {
	systems []System
}

// NewSystemManager создаёт новый менеджер систем
func NewSystemManager() *SystemManager {
	return &SystemManager{
		systems: make([]System, 0, 8), // Предварительно выделяем место для 8 систем
	}
}

// AddSystem добавляет систему в менеджер
func (sm *SystemManager) AddSystem(system System) {
	sm.systems = append(sm.systems, system)
}

// Update обновляет все системы в порядке их добавления
func (sm *SystemManager) Update(world *World, deltaTime float32) {
	for _, system := range sm.systems {
		system.Update(world, deltaTime)
	}
}

// GetSystemCount возвращает количество зарегистрированных систем
func (sm *SystemManager) GetSystemCount() int {
	return len(sm.systems)
}

// Clear очищает все системы (для тестов)
func (sm *SystemManager) Clear() {
	sm.systems = sm.systems[:0]
}
