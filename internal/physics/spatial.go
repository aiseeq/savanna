package physics

import (
	"math"

	"github.com/aiseeq/savanna/internal/constants"
)

// Константы размеров
const (
	TileSize   = constants.TileSizePixels // РЕФАКТОРИНГ: используем константу из constants.go
	EdgeOffset = 0.1                      // Небольшой отступ от границы мира для предотвращения выхода за пределы
)

// EntityID представляет уникальный идентификатор сущности
type EntityID uint16

// SpatialEntry представляет запись в пространственной сетке
type SpatialEntry struct {
	ID       EntityID
	Position Vec2
	Radius   float32
}

// SpatialGrid представляет пространственную сетку для быстрого поиска соседей
type SpatialGrid struct {
	cellSize    float32
	worldWidth  float32
	worldHeight float32
	gridWidth   int
	gridHeight  int
	cells       [][]SpatialEntry // 2D массив ячеек
	entities    map[EntityID]SpatialEntry
}

// NewSpatialGrid создает новую пространственную сетку
func NewSpatialGrid(worldWidth, worldHeight, cellSize float32) *SpatialGrid {
	gridWidth := int(math.Ceil(float64(worldWidth / cellSize)))
	gridHeight := int(math.Ceil(float64(worldHeight / cellSize)))

	grid := &SpatialGrid{
		cellSize:    cellSize,
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
		gridWidth:   gridWidth,
		gridHeight:  gridHeight,
		cells:       make([][]SpatialEntry, gridWidth*gridHeight),
		entities:    make(map[EntityID]SpatialEntry),
	}

	// Инициализируем ячейки
	for i := range grid.cells {
		grid.cells[i] = make([]SpatialEntry, 0, 8) // Предварительно выделяем место для 8 сущностей
	}

	return grid
}

// getCellIndex возвращает индекс ячейки для данной позиции
func (sg *SpatialGrid) getCellIndex(x, y float32) int {
	// Ограничиваем координаты границами мира
	x = float32(math.Max(0, math.Min(float64(x), float64(sg.worldWidth-EdgeOffset))))
	y = float32(math.Max(0, math.Min(float64(y), float64(sg.worldHeight-EdgeOffset))))

	cellX := int(x / sg.cellSize)
	cellY := int(y / sg.cellSize)

	// Убеждаемся что индексы не выходят за границы
	cellX = int(math.Max(0, math.Min(float64(cellX), float64(sg.gridWidth-1))))
	cellY = int(math.Max(0, math.Min(float64(cellY), float64(sg.gridHeight-1))))

	return cellY*sg.gridWidth + cellX
}

// getCellCoords возвращает координаты ячейки для данной позиции
func (sg *SpatialGrid) getCellCoords(x, y float32) (cellX, cellY int) {
	x = float32(math.Max(0, math.Min(float64(x), float64(sg.worldWidth-EdgeOffset))))
	y = float32(math.Max(0, math.Min(float64(y), float64(sg.worldHeight-EdgeOffset))))

	cellX = int(x / sg.cellSize)
	cellY = int(y / sg.cellSize)

	cellX = int(math.Max(0, math.Min(float64(cellX), float64(sg.gridWidth-1))))
	cellY = int(math.Max(0, math.Min(float64(cellY), float64(sg.gridHeight-1))))

	return
}

// Insert добавляет сущность в пространственную сетку
func (sg *SpatialGrid) Insert(id EntityID, position Vec2, radius float32) {
	entry := SpatialEntry{
		ID:       id,
		Position: position,
		Radius:   radius,
	}

	// Удаляем старую запись если она существует
	sg.Remove(id)

	// Добавляем в новую позицию
	cellIndex := sg.getCellIndex(position.X, position.Y)
	sg.cells[cellIndex] = append(sg.cells[cellIndex], entry)
	sg.entities[id] = entry
}

// Remove удаляет сущность из пространственной сетки
func (sg *SpatialGrid) Remove(id EntityID) {
	entry, exists := sg.entities[id]
	if !exists {
		return
	}

	cellIndex := sg.getCellIndex(entry.Position.X, entry.Position.Y)
	cell := sg.cells[cellIndex]

	// Находим и удаляем запись из ячейки
	for i, cellEntry := range cell {
		if cellEntry.ID == id {
			// Удаляем элемент, сохраняя порядок
			sg.cells[cellIndex] = append(cell[:i], cell[i+1:]...)
			break
		}
	}

	delete(sg.entities, id)
}

// Update обновляет позицию сущности в сетке
func (sg *SpatialGrid) Update(id EntityID, newPosition Vec2, newRadius float32) {
	entry, exists := sg.entities[id]
	if !exists {
		sg.Insert(id, newPosition, newRadius)
		return
	}

	// Проверяем изменилась ли ячейка
	oldCellIndex := sg.getCellIndex(entry.Position.X, entry.Position.Y)
	newCellIndex := sg.getCellIndex(newPosition.X, newPosition.Y)

	entry.Position = newPosition
	entry.Radius = newRadius

	if oldCellIndex == newCellIndex {
		// Ячейка не изменилась, просто обновляем данные
		sg.entities[id] = entry
		for i := range sg.cells[oldCellIndex] {
			if sg.cells[oldCellIndex][i].ID == id {
				sg.cells[oldCellIndex][i] = entry
				break
			}
		}
	} else {
		// Ячейка изменилась, переносим сущность
		sg.Remove(id)
		sg.Insert(id, newPosition, newRadius)
	}
}

// QueryRange возвращает все сущности в указанной области (рефакторинг: снижена когнитивная сложность)
func (sg *SpatialGrid) QueryRange(minX, minY, maxX, maxY float32) []SpatialEntry {
	result := make([]SpatialEntry, 0, 32)

	// Определяем диапазон ячеек для проверки
	startCellX, startCellY := sg.getCellCoords(minX, minY)
	endCellX, endCellY := sg.getCellCoords(maxX, maxY)

	query := RangeQuery{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
	for cellY := startCellY; cellY <= endCellY; cellY++ {
		for cellX := startCellX; cellX <= endCellX; cellX++ {
			sg.processGridCell(cellX, cellY, query, &result)
		}
	}

	return result
}

// RangeQuery представляет параметры запроса по диапазону (устранение нарушения argument-limit)
type RangeQuery struct {
	MinX, MinY, MaxX, MaxY float32
}

// processGridCell обрабатывает одну ячейку сетки для QueryRange (helper-функция для снижения сложности)
func (sg *SpatialGrid) processGridCell(cellX, cellY int, query RangeQuery, result *[]SpatialEntry) {
	cellIndex := cellY*sg.gridWidth + cellX
	if !sg.isValidCellIndex(cellIndex) {
		return
	}

	for _, entry := range sg.cells[cellIndex] {
		if sg.isEntryInRange(entry, query.MinX, query.MinY, query.MaxX, query.MaxY) {
			*result = append(*result, entry)
		}
	}
}

// isValidCellIndex проверяет валидность индекса ячейки
func (sg *SpatialGrid) isValidCellIndex(cellIndex int) bool {
	return cellIndex >= 0 && cellIndex < len(sg.cells)
}

// isEntryInRange проверяет находится ли сущность в указанном диапазоне
func (sg *SpatialGrid) isEntryInRange(entry SpatialEntry, minX, minY, maxX, maxY float32) bool {
	return entry.Position.X >= minX && entry.Position.X <= maxX &&
		entry.Position.Y >= minY && entry.Position.Y <= maxY
}

// QueryRadius возвращает все сущности в радиусе от указанной точки
func (sg *SpatialGrid) QueryRadius(center Vec2, radius float32) []SpatialEntry {
	// Определяем квадратную область поиска
	minX := center.X - radius
	minY := center.Y - radius
	maxX := center.X + radius
	maxY := center.Y + radius

	candidates := sg.QueryRange(minX, minY, maxX, maxY)
	result := make([]SpatialEntry, 0, len(candidates))

	for _, entry := range candidates {
		// Проверяем действительное расстояние с учетом радиуса сущности
		distanceSquared := center.DistanceSquared(entry.Position)
		combinedRadius := radius + entry.Radius
		if distanceSquared <= combinedRadius*combinedRadius {
			result = append(result, entry)
		}
	}

	return result
}

// QueryNearest возвращает ближайшую сущность к указанной точке
func (sg *SpatialGrid) QueryNearest(center Vec2, maxRadius float32) (SpatialEntry, bool) {
	candidates := sg.QueryRadius(center, maxRadius)
	if len(candidates) == 0 {
		return SpatialEntry{}, false
	}

	nearest := candidates[0]
	minDistanceSquared := center.DistanceSquared(nearest.Position)

	for i := 1; i < len(candidates); i++ {
		distanceSquared := center.DistanceSquared(candidates[i].Position)
		if distanceSquared < minDistanceSquared {
			minDistanceSquared = distanceSquared
			nearest = candidates[i]
		}
	}

	return nearest, true
}

// Clear очищает всю сетку
func (sg *SpatialGrid) Clear() {
	for i := range sg.cells {
		sg.cells[i] = sg.cells[i][:0] // Очищаем slice но сохраняем capacity
	}
	// Очищаем map
	for k := range sg.entities {
		delete(sg.entities, k)
	}
}

// GetEntityCount возвращает общее количество сущностей в сетке
func (sg *SpatialGrid) GetEntityCount() int {
	return len(sg.entities)
}

// GetCellCount возвращает общее количество ячеек в сетке
func (sg *SpatialGrid) GetCellCount() int {
	return len(sg.cells)
}

// GetActiveCellCount возвращает количество непустых ячеек
func (sg *SpatialGrid) GetActiveCellCount() int {
	count := 0
	for _, cell := range sg.cells {
		if len(cell) > 0 {
			count++
		}
	}
	return count
}

// GetCellSize возвращает размер ячейки
func (sg *SpatialGrid) GetCellSize() float32 {
	return sg.cellSize
}

// GetWorldDimensions возвращает размеры мира
func (sg *SpatialGrid) GetWorldDimensions() (width, height float32) {
	return sg.worldWidth, sg.worldHeight
}

// GetGridDimensions возвращает размеры сетки в ячейках
func (sg *SpatialGrid) GetGridDimensions() (width, height int) {
	return sg.gridWidth, sg.gridHeight
}
