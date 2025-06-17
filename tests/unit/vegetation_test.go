package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// MockTerrain реализует TerrainInterface для тестов
type MockTerrain struct {
	size        int
	grassAmount map[int]float32
	tileTypes   map[int]generator.TileType
}

func NewMockTerrain(size int) *MockTerrain {
	return &MockTerrain{
		size:        size,
		grassAmount: make(map[int]float32),
		tileTypes:   make(map[int]generator.TileType),
	}
}

func (m *MockTerrain) GetSize() int {
	return m.size
}

func (m *MockTerrain) GetGrassAmount(x, y int) float32 {
	key := y*m.size + x
	return m.grassAmount[key]
}

func (m *MockTerrain) SetGrassAmount(x, y int, amount float32) {
	key := y*m.size + x
	m.grassAmount[key] = amount
}

func (m *MockTerrain) GetTileType(x, y int) generator.TileType {
	key := y*m.size + x
	if tileType, exists := m.tileTypes[key]; exists {
		return tileType
	}
	return generator.TileGrass // По умолчанию трава
}

func (m *MockTerrain) SetTileType(x, y int, tileType generator.TileType) {
	key := y*m.size + x
	m.tileTypes[key] = tileType
}

func (m *MockTerrain) ConsumeGrass(x, y int, amount float32) float32 {
	current := m.GetGrassAmount(x, y)
	consumed := amount
	if consumed > current {
		consumed = current
	}
	m.SetGrassAmount(x, y, current-consumed)
	return consumed
}

func TestVegetationSystem_GrassGrowth(t *testing.T) {
	t.Parallel()
	// Создаем тестовое окружение
	terrain := NewMockTerrain(10)
	vegSystem := simulation.NewVegetationSystem(terrain)
	world := core.NewWorld(320.0, 320.0, 12345)

	// Устанавливаем начальное количество травы
	terrain.SetGrassAmount(5, 5, 50.0)
	terrain.SetTileType(5, 5, generator.TileGrass)

	// Обновляем систему - трава должна расти
	deltaTime := float32(2.0) // 2 секунды
	vegSystem.Update(world, deltaTime)

	// Проверяем что трава выросла (+ 1 единица за 2 секунды)
	newAmount := terrain.GetGrassAmount(5, 5)
	expected := float32(51.0)  // 50 + 1
	tolerance := float32(0.01) // Допуск для чисел с плавающей точкой
	if newAmount < expected-tolerance || newAmount > expected+tolerance {
		t.Errorf("Ожидали %f травы (±%.3f), получили %f", expected, tolerance, newAmount)
	}
}

func TestVegetationSystem_GrassGrowthCap(t *testing.T) {
	t.Parallel()
	// Создаем тестовое окружение
	terrain := NewMockTerrain(10)
	vegSystem := simulation.NewVegetationSystem(terrain)
	world := core.NewWorld(320.0, 320.0, 12345)

	// Устанавливаем максимальное количество травы
	terrain.SetGrassAmount(5, 5, 100.0)
	terrain.SetTileType(5, 5, generator.TileGrass)

	// Обновляем систему
	deltaTime := float32(2.0)
	vegSystem.Update(world, deltaTime)

	// Проверяем что трава не превысила максимум
	newAmount := terrain.GetGrassAmount(5, 5)
	if newAmount > 100.0 {
		t.Errorf("Трава превысила максимум: %f", newAmount)
	}
}

func TestVegetationSystem_NonGrassTiles(t *testing.T) {
	t.Parallel()
	// Создаем тестовое окружение
	terrain := NewMockTerrain(10)
	vegSystem := simulation.NewVegetationSystem(terrain)
	world := core.NewWorld(320.0, 320.0, 12345)

	// Устанавливаем водную клетку
	terrain.SetGrassAmount(5, 5, 0.0)
	terrain.SetTileType(5, 5, generator.TileWater)

	// Обновляем систему
	deltaTime := float32(2.0)
	vegSystem.Update(world, deltaTime)

	// Проверяем что на воде трава не растет
	newAmount := terrain.GetGrassAmount(5, 5)
	if newAmount != 0.0 {
		t.Errorf("На воде не должна расти трава, но есть %f", newAmount)
	}
}

func TestVegetationSystem_FindNearestGrass(t *testing.T) {
	t.Parallel()
	// Создаем тестовое окружение
	terrain := NewMockTerrain(10)
	vegSystem := simulation.NewVegetationSystem(terrain)

	// Размещаем траву в разных местах
	terrain.SetGrassAmount(3, 3, 80.0) // Много травы
	terrain.SetGrassAmount(7, 7, 20.0) // Мало травы
	terrain.SetTileType(3, 3, generator.TileGrass)
	terrain.SetTileType(7, 7, generator.TileGrass)

	// Ищем ближайшую траву от позиции (160, 160) (центр тайла 5,5) с минимумом 50
	grassX, grassY, found := vegSystem.FindNearestGrass(160.0, 160.0, 320.0, 50.0)

	if !found {
		t.Error("Должна была найтись трава")
		return
	}

	// Должна найтись клетка (3, 3) так как там больше минимума
	// Ожидаем центр тайла: (3*32 + 16, 3*32 + 16) = (112, 112)
	expectedX := float32(3*32 + 16)
	expectedY := float32(3*32 + 16)
	if grassX != expectedX || grassY != expectedY {
		t.Errorf("Ожидали траву в (%f, %f), получили (%f, %f)", expectedX, expectedY, grassX, grassY)
	}
}

func TestVegetationSystem_FindNearestGrass_NotFound(t *testing.T) {
	t.Parallel()
	// Создаем тестовое окружение
	terrain := NewMockTerrain(10)
	vegSystem := simulation.NewVegetationSystem(terrain)

	// Размещаем мало травы
	terrain.SetGrassAmount(3, 3, 10.0) // Мало травы
	terrain.SetTileType(3, 3, generator.TileGrass)

	// Ищем траву с высоким минимумом от позиции (160, 160)
	_, _, found := vegSystem.FindNearestGrass(160.0, 160.0, 320.0, 50.0)

	if found {
		t.Error("Не должна была найтись трава с таким минимумом")
	}
}

func TestVegetationSystem_ConsumeGrass(t *testing.T) {
	t.Parallel()
	// Создаем тестовое окружение
	terrain := NewMockTerrain(10)

	// Устанавливаем траву
	terrain.SetGrassAmount(5, 5, 60.0)

	// Потребляем траву
	consumed := terrain.ConsumeGrass(5, 5, 20.0)

	// Проверяем результаты
	if consumed != 20.0 {
		t.Errorf("Ожидали потребить 20.0, потребили %f", consumed)
	}

	remaining := terrain.GetGrassAmount(5, 5)
	if remaining != 40.0 {
		t.Errorf("Ожидали остаток 40.0, получили %f", remaining)
	}
}

func TestVegetationSystem_ConsumeGrass_PartialConsumption(t *testing.T) {
	t.Parallel()
	// Создаем тестовое окружение
	terrain := NewMockTerrain(10)

	// Устанавливаем мало травы
	terrain.SetGrassAmount(5, 5, 10.0)

	// Пытаемся потребить больше чем есть
	consumed := terrain.ConsumeGrass(5, 5, 30.0)

	// Проверяем что потребили только доступное количество
	if consumed != 10.0 {
		t.Errorf("Ожидали потребить 10.0, потребили %f", consumed)
	}

	remaining := terrain.GetGrassAmount(5, 5)
	if remaining != 0.0 {
		t.Errorf("Ожидали остаток 0.0, получили %f", remaining)
	}
}
