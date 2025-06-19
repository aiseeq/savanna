package core

import (
	"math/rand"

	"github.com/aiseeq/savanna/internal/physics"
)

// World главная структура мира симуляции
// Использует Composition Pattern и SRP - разделён на специализированные менеджеры
type World struct {
	// Специализированные менеджеры (соблюдают SRP)
	entityManager    *EntityManager    // Создание/удаление сущностей
	componentManager *ComponentManager // Управление компонентами
	queryManager     *QueryManager     // Запросы и итерации
	worldState       *WorldState       // Состояние мира (время, размеры, RNG)
}

// NewWorld создаёт новый мир симуляции
func NewWorld(worldWidth, worldHeight float32, seed int64) *World {
	// Создаём специализированные менеджеры (применяем Composition Pattern)
	entityManager := NewEntityManager()
	componentManager := NewComponentManager()
	worldState := NewWorldState(worldWidth, worldHeight, seed)
	queryManager := NewQueryManager(componentManager, entityManager)

	return &World{
		entityManager:    entityManager,
		componentManager: componentManager,
		queryManager:     queryManager,
		worldState:       worldState,
	}
}

// GetTime возвращает текущее время симуляции (делегирование к WorldState)
func (w *World) GetTime() float32 {
	return w.worldState.GetTime()
}

// GetDeltaTime возвращает время с последнего обновления (делегирование к WorldState)
func (w *World) GetDeltaTime() float32 {
	return w.worldState.GetDeltaTime()
}

// SetTimeScale устанавливает масштаб времени (делегирование к WorldState)
func (w *World) SetTimeScale(scale float32) {
	w.worldState.SetTimeScale(scale)
}

// GetTimeScale возвращает текущий масштаб времени (делегирование к WorldState)
func (w *World) GetTimeScale() float32 {
	return w.worldState.GetTimeScale()
}

// Update обновляет мир на один кадр (делегирование к WorldState)
func (w *World) Update(realDeltaTime float32) {
	w.worldState.Update(realDeltaTime)
}

// CreateEntity создаёт новую сущность (делегирование к EntityManager)
func (w *World) CreateEntity() EntityID {
	return w.entityManager.CreateEntity()
}

// DestroyEntity уничтожает сущность и все её компоненты
func (w *World) DestroyEntity(entity EntityID) bool {
	if !w.entityManager.IsAlive(entity) {
		return false
	}

	// Удаляем из пространственной сетки если есть позиция
	if w.componentManager.HasComponent(entity, MaskPosition) {
		w.worldState.GetSpatialProvider().RemoveEntity(uint32(entity))
	}

	// Очищаем все компоненты (делегирование к ComponentManager)
	w.componentManager.ClearAllComponents(entity)

	// Уничтожаем сущность (делегирование к EntityManager)
	return w.entityManager.DestroyEntity(entity)
}

// IsAlive проверяет, существует ли сущность (делегирование к EntityManager)
func (w *World) IsAlive(entity EntityID) bool {
	return w.entityManager.IsAlive(entity)
}

// GetEntityCount возвращает количество живых сущностей (делегирование к QueryManager)
func (w *World) GetEntityCount() int {
	return w.queryManager.GetEntityCount()
}

// GetWorldDimensions возвращает размеры мира (делегирование к WorldState)
func (w *World) GetWorldDimensions() (width, height float32) {
	return w.worldState.GetWorldWidth(), w.worldState.GetWorldHeight()
}

// GetSpatialGrid возвращает пространственную сетку для прямого доступа
// Deprecated: нарушает DIP, будет удалена в будущем
func (w *World) GetSpatialGrid() *physics.SpatialGrid {
	// Приведение к конкретному типу через адаптер
	if adapter, ok := w.worldState.GetSpatialProvider().(*SpatialGridAdapter); ok {
		return adapter.grid
	}
	return nil
}

// GetRNG возвращает генератор случайных чисел (делегирование к WorldState)
func (w *World) GetRNG() *rand.Rand {
	return w.worldState.GetRNG()
}

// Clear очищает весь мир (для тестов и перезапуска)
func (w *World) Clear() {
	// Делегируем очистку к соответствующим менеджерам
	w.entityManager.Clear()
	w.worldState.GetSpatialProvider().Clear()

	// Сбрасываем время через WorldState
	w.worldState = NewWorldState(w.worldState.GetWorldWidth(), w.worldState.GetWorldHeight(), 0)

	// Очищаем компоненты через ComponentManager (он имеет метод для этого)
	w.componentManager = NewComponentManager()
}
