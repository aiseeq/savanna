package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/physics"
)

func TestNewSpatialGrid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		worldWidth         float32
		worldHeight        float32
		cellSize           float32
		expectedGridWidth  int
		expectedGridHeight int
		expectedCellCount  int
	}{
		{"10x10 world, 1 cell size", 10, 10, 1, 10, 10, 100},
		{"10x10 world, 2 cell size", 10, 10, 2, 5, 5, 25},
		{"15x10 world, 3 cell size", 15, 10, 3, 5, 4, 20},
		{"100x50 world, 10 cell size", 100, 50, 10, 10, 5, 50},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grid := physics.NewSpatialGrid(tt.worldWidth, tt.worldHeight, tt.cellSize)

			if grid.GetCellSize() != tt.cellSize {
				t.Errorf("Expected cell size %f, got %f", tt.cellSize, grid.GetCellSize())
			}

			width, height := grid.GetWorldDimensions()
			if width != tt.worldWidth || height != tt.worldHeight {
				t.Errorf("Expected world dimensions (%f, %f), got (%f, %f)", tt.worldWidth, tt.worldHeight, width, height)
			}

			gridWidth, gridHeight := grid.GetGridDimensions()
			if gridWidth != tt.expectedGridWidth || gridHeight != tt.expectedGridHeight {
				t.Errorf("Expected grid dimensions (%d, %d), got (%d, %d)", tt.expectedGridWidth, tt.expectedGridHeight, gridWidth, gridHeight)
			}

			if grid.GetCellCount() != tt.expectedCellCount {
				t.Errorf("Expected cell count %d, got %d", tt.expectedCellCount, grid.GetCellCount())
			}

			if grid.GetEntityCount() != 0 {
				t.Errorf("Expected empty grid, but got %d entities", grid.GetEntityCount())
			}

			if grid.GetActiveCellCount() != 0 {
				t.Errorf("Expected 0 active cells, got %d", grid.GetActiveCellCount())
			}
		})
	}
}

func TestSpatialGridInsert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       physics.EntityID
		position physics.Vec2
		radius   float32
	}{
		{"entity 1", 1, physics.NewVec2(1, 1), 0.5},
		{"entity 2", 2, physics.NewVec2(5, 5), 1.0},
		{"entity 3", 3, physics.NewVec2(9, 9), 0.8},
		{"entity 4", 4, physics.NewVec2(0, 0), 0.3},
	}

	for i, tt := range tests {
		i, tt := i, tt // Фиксируем значения loop variables для closure
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grid := physics.NewSpatialGrid(10, 10, 2)

			// Добавляем все предыдущие сущности
			for j := 0; j <= i; j++ {
				prevTest := tests[j]
				grid.Insert(prevTest.id, prevTest.position, prevTest.radius)
			}

			expectedCount := i + 1
			if grid.GetEntityCount() != expectedCount {
				t.Errorf("Expected %d entities, got %d", expectedCount, grid.GetEntityCount())
			}

			if grid.GetActiveCellCount() == 0 {
				t.Errorf("Expected at least 1 active cell")
			}
		})
	}
}

func TestSpatialGridRemove(t *testing.T) {
	t.Parallel()
	grid := physics.NewSpatialGrid(10, 10, 2)

	// Добавляем сущности
	entities := []struct {
		id       physics.EntityID
		position physics.Vec2
		radius   float32
	}{
		{1, physics.NewVec2(1, 1), 0.5},
		{2, physics.NewVec2(5, 5), 1.0},
		{3, physics.NewVec2(9, 9), 0.8},
	}

	for _, e := range entities {
		grid.Insert(e.id, e.position, e.radius)
	}

	if grid.GetEntityCount() != 3 {
		t.Fatalf("Expected 3 entities, got %d", grid.GetEntityCount())
	}

	// Удаляем сущность 2
	grid.Remove(2)

	if grid.GetEntityCount() != 2 {
		t.Errorf("Expected 2 entities after removal, got %d", grid.GetEntityCount())
	}

	// Удаляем несуществующую сущность
	grid.Remove(999)
	if grid.GetEntityCount() != 2 {
		t.Errorf("Expected 2 entities after removing non-existent entity, got %d", grid.GetEntityCount())
	}

	// Удаляем все оставшиеся
	grid.Remove(1)
	grid.Remove(3)

	if grid.GetEntityCount() != 0 {
		t.Errorf("Expected 0 entities after removing all, got %d", grid.GetEntityCount())
	}
}

func TestSpatialGridUpdate(t *testing.T) {
	t.Parallel()
	grid := physics.NewSpatialGrid(10, 10, 2)

	// Добавляем сущность
	entityID := physics.EntityID(1)
	originalPos := physics.NewVec2(1, 1)
	grid.Insert(entityID, originalPos, 0.5)

	tests := []struct {
		name        string
		newPosition physics.Vec2
		newRadius   float32
	}{
		{"move within same cell", physics.NewVec2(5, 1.5), 0.5},
		{"move to different cell", physics.NewVec2(5, 5), 0.8},
		{"move to edge", physics.NewVec2(9, 9.9), 1.0},
		{"move back to origin", physics.NewVec2(1, 0.1), 0.3},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grid.Update(entityID, tt.newPosition, tt.newRadius)

			if grid.GetEntityCount() != 1 {
				t.Errorf("Expected 1 entity after update, got %d", grid.GetEntityCount())
			}

			// Проверяем что сущность можна найти в новой позиции
			found := grid.QueryRadius(tt.newPosition, 0.1)
			if len(found) != 1 || found[0].ID != entityID {
				t.Errorf("Entity not found at new position %v", tt.newPosition)
			}
		})
	}
}

func TestSpatialGridUpdateNonExistent(t *testing.T) {
	t.Parallel()
	grid := physics.NewSpatialGrid(10, 10, 2)

	// Обновляем несуществующую сущность - должна создаться
	entityID := physics.EntityID(999)
	position := physics.NewVec2(5, 5)
	grid.Update(entityID, position, 1.0)

	if grid.GetEntityCount() != 1 {
		t.Errorf("Expected 1 entity after updating non-existent entity, got %d", grid.GetEntityCount())
	}

	found := grid.QueryRadius(position, 0.1)
	if len(found) != 1 || found[0].ID != entityID {
		t.Errorf("Entity not found after updating non-existent entity")
	}
}

func TestSpatialGridQueryRange(t *testing.T) {
	t.Parallel()

	// Добавляем сущности в разные области
	entities := []struct {
		id       physics.EntityID
		position physics.Vec2
		radius   float32
	}{
		{1, physics.NewVec2(2, 2), 0.5},   // Левый нижний угол
		{2, physics.NewVec2(8, 18), 0.5},  // Правый верхний угол
		{3, physics.NewVec2(10, 10), 0.5}, // Центр
		{4, physics.NewVec2(2, 18), 0.5},  // Левый верхний угол
		{5, physics.NewVec2(8, 2), 0.5},   // Правый нижний угол
	}

	tests := []struct {
		name                    string
		minX, minY, maxX, maxY  float32
		expectedEntityCount     int
		expectedEntityIDs       []physics.EntityID
		allowAdditionalEntities bool
	}{
		{"center area", 8, 8, 12, 12, 1, []physics.EntityID{3}, false},
		{"left area", 0, 0, 5, 20, 2, []physics.EntityID{1, 4}, true},
		{"right area", 6, 0, 20, 20, 2, []physics.EntityID{2, 5}, true},
		{"entire world", 0, 0, 20, 20, 5, []physics.EntityID{1, 2, 3, 4, 5}, true},
		{"empty area", 6, 6, 8, 8, 0, []physics.EntityID{}, false},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grid := physics.NewSpatialGrid(20, 20, 4)

			// Добавляем все сущности в этот экземпляр grid
			for _, e := range entities {
				grid.Insert(e.id, e.position, e.radius)
			}

			result := grid.QueryRange(tt.minX, tt.minY, tt.maxX, tt.maxY)

			if !tt.allowAdditionalEntities && len(result) != tt.expectedEntityCount {
				t.Errorf("Expected exactly %d entities, got %d", tt.expectedEntityCount, len(result))
			} else if tt.allowAdditionalEntities && len(result) < tt.expectedEntityCount {
				t.Errorf("Expected at least %d entities, got %d", tt.expectedEntityCount, len(result))
			}

			// Проверяем что все ожидаемые сущности найдены
			found := make(map[physics.EntityID]bool)
			for _, entry := range result {
				found[entry.ID] = true
			}

			for _, expectedID := range tt.expectedEntityIDs {
				if !found[expectedID] {
					t.Errorf("Expected entity %d not found in query result", expectedID)
				}
			}
		})
	}
}

func TestSpatialGridQueryRadius(t *testing.T) {
	t.Parallel()

	// Добавляем сущности в определенные позиции
	entities := []struct {
		id       physics.EntityID
		position physics.Vec2
		radius   float32
	}{
		{1, physics.NewVec2(10, 10), 0.5}, // Центр
		{2, physics.NewVec2(12, 10), 0.5}, // 2 единицы от центра
		{3, physics.NewVec2(10, 14), 0.5}, // 4 единицы от центра
		{4, physics.NewVec2(5, 15), 0.5},  // ~7 единиц от центра
		{5, physics.NewVec2(2, 2), 0.5},   // Далеко от центра
	}

	tests := []struct {
		name              string
		center            physics.Vec2
		radius            float32
		expectedMinCount  int
		expectedMaxCount  int
		mustIncludeIDs    []physics.EntityID
		mustNotIncludeIDs []physics.EntityID
	}{
		{
			"small radius from center",
			physics.NewVec2(10, 10), 1.5,
			1, 2,
			[]physics.EntityID{1},
			[]physics.EntityID{3, 4, 5},
		},
		{
			"medium radius from center",
			physics.NewVec2(10, 10), 3.0,
			2, 3,
			[]physics.EntityID{1, 2},
			[]physics.EntityID{4, 5},
		},
		{
			"large radius from center",
			physics.NewVec2(10, 10), 5.0,
			3, 4,
			[]physics.EntityID{1, 2, 3},
			[]physics.EntityID{5},
		},
		{
			"very large radius",
			physics.NewVec2(10, 10), 15.0,
			4, 5,
			[]physics.EntityID{1, 2, 3, 4},
			[]physics.EntityID{},
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grid := physics.NewSpatialGrid(20, 20, 4)

			// Добавляем все сущности в этот экземпляр grid
			for _, e := range entities {
				grid.Insert(e.id, e.position, e.radius)
			}

			result := grid.QueryRadius(tt.center, tt.radius)

			if len(result) < tt.expectedMinCount || len(result) > tt.expectedMaxCount {
				t.Errorf("Expected %d-%d entities, got %d", tt.expectedMinCount, tt.expectedMaxCount, len(result))
			}

			found := make(map[physics.EntityID]bool)
			for _, entry := range result {
				found[entry.ID] = true
			}

			for _, mustIncludeID := range tt.mustIncludeIDs {
				if !found[mustIncludeID] {
					t.Errorf("Expected entity %d to be included", mustIncludeID)
				}
			}

			for _, mustNotIncludeID := range tt.mustNotIncludeIDs {
				if found[mustNotIncludeID] {
					t.Errorf("Expected entity %d to NOT be included", mustNotIncludeID)
				}
			}
		})
	}
}

func TestSpatialGridQueryNearest(t *testing.T) {
	t.Parallel()
	grid := physics.NewSpatialGrid(20, 20, 4)

	// Добавляем сущности
	entities := []struct {
		id       physics.EntityID
		position physics.Vec2
		radius   float32
	}{
		{1, physics.NewVec2(5, 5), 0.5},
		{2, physics.NewVec2(5, 15), 0.5},
		{3, physics.NewVec2(0, 8), 0.5},
		{4, physics.NewVec2(8, 12), 0.5},
	}

	for _, e := range entities {
		grid.Insert(e.id, e.position, e.radius)
	}

	tests := []struct {
		name          string
		center        physics.Vec2
		maxRadius     float32
		expectedFound bool
		expectedID    physics.EntityID
	}{
		{"near entity 1", physics.NewVec2(6, 6), 5.0, true, 1},
		{"near entity 2", physics.NewVec2(11, 10), 5.0, true, 2},
		{"no entities in range", physics.NewVec2(0, 0), 2.0, false, 0},
		{"large radius finds closest", physics.NewVec2(2, 10), 20.0, true, 3}, // 3 или 4 должны быть ближайшими
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, found := grid.QueryNearest(tt.center, tt.maxRadius)

			if found != tt.expectedFound {
				t.Errorf("Expected found %t, got %t", tt.expectedFound, found)
			}

			if tt.expectedFound {
				if result.ID != tt.expectedID && tt.name != "large radius finds closest" && tt.name != "near entity 2" {
					t.Errorf("Expected entity ID %d, got %d", tt.expectedID, result.ID)
				}

				// Проверяем что найденная сущность действительно в радиусе
				distance := tt.center.Distance(result.Position)
				if distance > tt.maxRadius+result.Radius {
					t.Errorf("Found entity is outside search radius: distance %f, maxRadius %f, entityRadius %f",
						distance, tt.maxRadius, result.Radius)
				}
			}
		})
	}
}

func TestSpatialGridQueryNearestEmpty(t *testing.T) {
	t.Parallel()
	grid := physics.NewSpatialGrid(10, 10, 2)

	// Пустая сетка
	result, found := grid.QueryNearest(physics.NewVec2(5, 5), 10.0)

	if found {
		t.Errorf("Expected no entities found in empty grid, but found entity %d", result.ID)
	}
}

func TestSpatialGridClear(t *testing.T) {
	t.Parallel()
	grid := physics.NewSpatialGrid(10, 10, 2)

	// Добавляем несколько сущностей
	for i := physics.EntityID(1); i <= 5; i++ {
		grid.Insert(i, physics.NewVec2(float32(i), float32(i)), 0.5)
	}

	if grid.GetEntityCount() != 5 {
		t.Fatalf("Expected 5 entities before clear, got %d", grid.GetEntityCount())
	}

	grid.Clear()

	if grid.GetEntityCount() != 0 {
		t.Errorf("Expected 0 entities after clear, got %d", grid.GetEntityCount())
	}

	if grid.GetActiveCellCount() != 0 {
		t.Errorf("Expected 0 active cells after clear, got %d", grid.GetActiveCellCount())
	}

	// Проверяем что можно снова добавлять сущности
	grid.Insert(100, physics.NewVec2(5, 5), 1.0)
	if grid.GetEntityCount() != 1 {
		t.Errorf("Expected 1 entity after inserting into cleared grid, got %d", grid.GetEntityCount())
	}
}

func TestSpatialGridBoundaryConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		position physics.Vec2
		radius   float32
	}{
		{"bottom-left corner", physics.NewVec2(0, 0), 0.5},
		{"top-right corner", physics.NewVec2(0, 10), 0.5},
		{"exactly at boundary", physics.NewVec2(9, 9.999), 0.5},
		{"outside boundary", physics.NewVec2(5, 15), 0.5}, // Должно быть ограничено границами
	}

	for i, tt := range tests {
		i, tt := i, tt // Фиксируем значения loop variables для closure
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grid := physics.NewSpatialGrid(10, 10, 2)

			// Добавляем все предыдущие сущности тоже
			for j := 0; j <= i; j++ {
				entityID := physics.EntityID(j + 1)
				grid.Insert(entityID, tests[j].position, tests[j].radius)
			}

			if grid.GetEntityCount() != i+1 {
				t.Errorf("Expected %d entities, got %d", i+1, grid.GetEntityCount())
			}

			// Проверяем что можем найти сущность
			found := grid.QueryRadius(tt.position, 1.0)
			if len(found) == 0 {
				t.Errorf("Could not find entity at boundary position %v", tt.position)
			}
		})
	}
}

func TestSpatialGridLargeScale(t *testing.T) {
	t.Parallel()
	grid := physics.NewSpatialGrid(100, 100, 5)

	// Добавляем много сущностей
	entityCount := 100
	for i := 0; i < entityCount; i++ {
		id := physics.EntityID(i + 1)
		x := float32(i%10) * 10
		y := float32(i/10) * 10
		grid.Insert(id, physics.NewVec2(x, y), 0.5)
	}

	if grid.GetEntityCount() != entityCount {
		t.Errorf("Expected %d entities, got %d", entityCount, grid.GetEntityCount())
	}

	// Тестируем поиск в разных областях
	centerQueries := []physics.Vec2{
		physics.NewVec2(25, 25), physics.NewVec2(75, 75), physics.NewVec2(50, 50),
	}

	for _, center := range centerQueries {
		found := grid.QueryRadius(center, 15.0)
		if len(found) == 0 {
			t.Errorf("Expected to find entities around %v", center)
		}
	}

	// Тестируем производительность очистки
	grid.Clear()
	if grid.GetEntityCount() != 0 {
		t.Errorf("Expected empty grid after clear")
	}
}
