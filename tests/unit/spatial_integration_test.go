package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// setupSpatialWorld создаёт тестовый мир с сущностями для пространственных тестов
//
//nolint:revive // function-result-limit: тестовая функция требует возврата нескольких сущностей
func setupSpatialWorld() (world *core.World, entity1, entity2, entity3 core.EntityID) {
	world = core.NewWorld(100, 100, 42) //nolint:gomnd // Тестовые параметры мира

	// Создаем сущности с позицией и размером
	entity1 = world.CreateEntity()
	world.AddPosition(entity1, core.Position{X: 10, Y: 10})      //nolint:gomnd // Тестовые координаты
	world.AddSize(entity1, core.Size{Radius: 5, AttackRange: 0}) //nolint:gomnd // Тестовые значения

	entity2 = world.CreateEntity()
	world.AddPosition(entity2, core.Position{X: 20, Y: 20})      //nolint:gomnd // Тестовые координаты
	world.AddSize(entity2, core.Size{Radius: 3, AttackRange: 0}) //nolint:gomnd // Тестовые значения

	entity3 = world.CreateEntity()
	world.AddPosition(entity3, core.Position{X: 80, Y: 80})      //nolint:gomnd // Тестовые координаты
	world.AddSize(entity3, core.Size{Radius: 2, AttackRange: 0}) //nolint:gomnd // Тестовые значения

	return world, entity1, entity2, entity3
}

// TestSpatialQueryInRadius тестирует QueryInRadius
func TestSpatialQueryInRadius(t *testing.T) {
	t.Parallel()

	world, entity1, _, _ := setupSpatialWorld()

	// Поиск в радиусе 15 от точки (10, 10)
	searchRadius := float32(15)
	searchX := float32(10)
	searchY := float32(10)
	nearby := world.QueryInRadius(searchX, searchY, searchRadius)

	// Должны найти entity1 и entity2, но не entity3
	minExpected := 1
	maxExpected := 2
	if len(nearby) < minExpected || len(nearby) > maxExpected {
		t.Errorf("Expected %d-%d entities in radius, got %d", minExpected, maxExpected, len(nearby))
	}

	// Проверяем что entity1 точно есть
	found := false
	for _, entity := range nearby {
		if entity == entity1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should find entity1 in radius")
	}
}

// TestSpatialFindNearestAnimal тестирует FindNearestAnimal
func TestSpatialFindNearestAnimal(t *testing.T) {
	t.Parallel()

	world, entity1, entity2, entity3 := setupSpatialWorld()

	// Добавляем типы животных
	world.AddAnimalType(entity1, core.TypeRabbit)
	world.AddAnimalType(entity2, core.TypeRabbit)
	world.AddAnimalType(entity3, core.TypeWolf)

	// Ищем ближайшее животное к точке (12, 12)
	searchX := float32(12)
	searchY := float32(12)
	searchRadius := float32(50)
	nearest, found := world.FindNearestAnimal(searchX, searchY, searchRadius)

	if !found {
		t.Error("Should find nearest animal")
	}

	// Ближайшим должен быть entity1 (расстояние ~2.8)
	if nearest != entity1 {
		t.Errorf("Expected entity1 as nearest, got %d", nearest)
	}
}

// TestSpatialFindNearestByType тестирует FindNearestByType
func TestSpatialFindNearestByType(t *testing.T) {
	t.Parallel()

	world, entity1, _, entity3 := setupSpatialWorld()

	// Добавляем типы животных
	world.AddAnimalType(entity1, core.TypeRabbit)
	world.AddAnimalType(entity3, core.TypeWolf)

	// Ищем ближайшего волка к точке (10, 10)
	searchX := float32(10)
	searchY := float32(10)
	searchRadius := float32(100)
	nearest, found := world.FindNearestByType(searchX, searchY, searchRadius, core.TypeWolf)

	if !found {
		t.Error("Should find nearest wolf")
	}

	if nearest != entity3 {
		t.Errorf("Expected entity3 as nearest wolf, got %d", nearest)
	}

	// Ищем ближайшего зайца
	nearest, found = world.FindNearestByType(searchX, searchY, searchRadius, core.TypeRabbit)

	if !found {
		t.Error("Should find nearest rabbit")
	}

	if nearest != entity1 {
		t.Errorf("Expected entity1 as nearest rabbit, got %d", nearest)
	}
}
