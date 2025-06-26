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
		w.removeSpatialEntity(entity)
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

// GetRNG возвращает генератор случайных чисел (делегирование к WorldState)
func (w *World) GetRNG() *rand.Rand {
	return w.worldState.GetRNG()
}

// Clear очищает весь мир (для тестов и перезапуска)
// Улучшено для соблюдения LoD - избегаем цепочек вызовов
func (w *World) Clear() {
	// Делегируем очистку к соответствующим менеджерам
	w.entityManager.Clear()

	// Очищаем пространственную систему через worldState
	w.worldState.ClearSpatialProvider()

	// Сохраняем размеры мира для пересоздания состояния
	width, height := w.worldState.GetWorldWidth(), w.worldState.GetWorldHeight()
	w.worldState = NewWorldState(width, height, 0)

	// Очищаем компоненты через ComponentManager
	w.componentManager = NewComponentManager()
}

// ===== МЕТОДЫ-ФАСАДЫ ДЛЯ СОБЛЮДЕНИЯ LAW OF DEMETER =====

// updateSpatialEntity обновляет позицию сущности в пространственной системе
// Скрывает сложность доступа к SpatialProvider через WorldState (LoD)
func (w *World) updateSpatialEntity(entity EntityID, x, y float32) {
	// КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Получаем реальный радиус животного для spatial grid
	radius := float32(0)
	if size, hasSize := w.componentManager.GetSize(entity); hasSize {
		radius = size.Radius
	}

	// ИСПРАВЛЕНИЕ ЕДИНИЦ ИЗМЕРЕНИЯ: координаты уже в тайлах
	posInTiles := physics.Vec2{
		X: x,
		Y: y,
	}
	radiusInTiles := radius

	w.worldState.GetSpatialProvider().UpdateEntity(uint32(entity), posInTiles, radiusInTiles)
}

// removeSpatialEntity удаляет сущность из пространственной системы
// Скрывает сложность доступа к SpatialProvider через WorldState (LoD)
func (w *World) removeSpatialEntity(entity EntityID) {
	w.worldState.GetSpatialProvider().RemoveEntity(uint32(entity))
}

// querySpatialRadius возвращает сущности в радиусе
// Скрывает сложность доступа к SpatialProvider через WorldState (LoD)
// ИСПРАВЛЕНИЕ ЕДИНИЦ ИЗМЕРЕНИЯ: координаты уже в тайлах
func (w *World) querySpatialRadius(x, y, radius float32) []EntityID {
	// Spatial система работает в тайлах, конвертации не нужны
	xInTiles := x
	yInTiles := y
	radiusInTiles := radius

	entries := w.worldState.GetSpatialProvider().QueryRadius(physics.Vec2{X: xInTiles, Y: yInTiles}, radiusInTiles)

	// Конвертируем SpatialEntry в EntityID
	result := make([]EntityID, len(entries))
	for i, entry := range entries {
		result[i] = EntityID(entry.ID)
	}
	return result
}
