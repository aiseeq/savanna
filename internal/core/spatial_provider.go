package core

import "github.com/aiseeq/savanna/internal/physics"

// SpatialQueryProvider интерфейс для пространственных запросов (устраняет нарушение DIP)
// Позволяет World работать с любыми системами пространственных запросов, а не только с SpatialGrid
type SpatialQueryProvider interface {
	// UpdateEntity обновляет позицию и радиус сущности в пространственной структуре
	UpdateEntity(id uint32, position physics.Vec2, radius float32)

	// RemoveEntity удаляет сущность из пространственной структуры
	RemoveEntity(id uint32)

	// QueryRadius возвращает все сущности в указанном радиусе
	QueryRadius(center physics.Vec2, radius float32) []physics.SpatialEntry

	// QueryNearest находит ближайшую сущность к указанной позиции
	QueryNearest(center physics.Vec2, maxRadius float32) (physics.SpatialEntry, bool)

	// Clear очищает все данные в пространственной структуре
	Clear()
}

// SpatialGridAdapter адаптер для physics.SpatialGrid (реализует SpatialQueryProvider)
type SpatialGridAdapter struct {
	grid *physics.SpatialGrid
}

// NewSpatialGridAdapter создаёт новый адаптер для SpatialGrid
func NewSpatialGridAdapter(worldWidth, worldHeight float32) SpatialQueryProvider {
	return &SpatialGridAdapter{
		grid: physics.NewSpatialGrid(worldWidth, worldHeight, float32(physics.TileSize)),
	}
}

// UpdateEntity обновляет позицию и радиус сущности в пространственной сетке
func (sga *SpatialGridAdapter) UpdateEntity(id uint32, position physics.Vec2, radius float32) {
	sga.grid.Update(physics.EntityID(id), position, radius)
}

// RemoveEntity удаляет сущность из пространственной сетки
func (sga *SpatialGridAdapter) RemoveEntity(id uint32) {
	sga.grid.Remove(physics.EntityID(id))
}

// QueryRadius возвращает все сущности в указанном радиусе
func (sga *SpatialGridAdapter) QueryRadius(center physics.Vec2, radius float32) []physics.SpatialEntry {
	return sga.grid.QueryRadius(center, radius)
}

// QueryNearest находит ближайшую сущность к указанной позиции
func (sga *SpatialGridAdapter) QueryNearest(center physics.Vec2, maxRadius float32) (physics.SpatialEntry, bool) {
	return sga.grid.QueryNearest(center, maxRadius)
}

// Clear очищает все данные в пространственной сетке
func (sga *SpatialGridAdapter) Clear() {
	sga.grid.Clear()
}
