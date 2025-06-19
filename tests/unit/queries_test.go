package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// setupQueryWorld создаёт тестовый мир с несколькими сущностями для тестирования запросов
func setupQueryWorld() *core.World {
	world := core.NewWorld(100, 100, 42) //nolint:gomnd // Тестовые параметры мира

	// Создаем несколько сущностей с разными компонентами
	rabbit1 := world.CreateEntity()
	world.AddPosition(rabbit1, core.Position{X: 10, Y: 10}) //nolint:gomnd // Тестовые координаты
	world.AddVelocity(rabbit1, core.Velocity{X: 1, Y: 0})   //nolint:gomnd // Тестовые значения
	world.AddAnimalType(rabbit1, core.TypeRabbit)
	world.AddHealth(rabbit1, core.Health{Current: 50, Max: 100}) //nolint:gomnd // Тестовые значения

	rabbit2 := world.CreateEntity()
	world.AddPosition(rabbit2, core.Position{X: 20, Y: 20}) //nolint:gomnd // Тестовые координаты
	world.AddAnimalType(rabbit2, core.TypeRabbit)

	wolf := world.CreateEntity()
	world.AddPosition(wolf, core.Position{X: 30, Y: 30}) //nolint:gomnd // Тестовые координаты
	world.AddVelocity(wolf, core.Velocity{X: -1, Y: -1}) //nolint:gomnd // Тестовые значения
	world.AddAnimalType(wolf, core.TypeWolf)

	return world
}

// TestQueryForEachWith тестирует запросы через ForEachWith
func TestQueryForEachWith(t *testing.T) {
	t.Parallel()

	world := setupQueryWorld()

	// Считаем сущности с позицией
	count := 0
	world.ForEachWith(core.MaskPosition, func(entity core.EntityID) {
		count++
	})

	expectedPositionCount := 3
	if count != expectedPositionCount {
		t.Errorf("Expected %d entities with Position, got %d", expectedPositionCount, count)
	}

	// Считаем сущности с позицией И скоростью
	count = 0
	world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
		count++
	})

	expectedVelocityCount := 2
	if count != expectedVelocityCount {
		t.Errorf("Expected %d entities with Position and Velocity, got %d", expectedVelocityCount, count)
	}
}

// TestQueryEntitiesWith тестирует QueryEntitiesWith
func TestQueryEntitiesWith(t *testing.T) {
	t.Parallel()

	world := setupQueryWorld()
	entities := world.QueryEntitiesWith(core.MaskPosition)

	expectedEntityCount := 3
	if len(entities) != expectedEntityCount {
		t.Errorf("Expected %d entities, got %d", expectedEntityCount, len(entities))
	}

	// Проверяем что все возвращенные сущности действительно имеют позицию
	for _, entity := range entities {
		if !world.HasComponent(entity, core.MaskPosition) {
			t.Errorf("Entity %d should have Position component", entity)
		}
	}
}

// TestQueryCountEntitiesWith тестирует CountEntitiesWith
func TestQueryCountEntitiesWith(t *testing.T) {
	t.Parallel()

	world := setupQueryWorld()

	count := world.CountEntitiesWith(core.MaskAnimalType)
	expectedAnimalCount := 3
	if count != expectedAnimalCount {
		t.Errorf("Expected %d animals, got %d", expectedAnimalCount, count)
	}

	count = world.CountEntitiesWith(core.MaskHealth)
	expectedHealthCount := 1
	if count != expectedHealthCount {
		t.Errorf("Expected %d entity with health, got %d", expectedHealthCount, count)
	}
}

// TestQueryByType тестирует QueryByType
func TestQueryByType(t *testing.T) {
	t.Parallel()

	world := setupQueryWorld()

	rabbits := world.QueryByType(core.TypeRabbit)
	expectedRabbitCount := 2
	if len(rabbits) != expectedRabbitCount {
		t.Errorf("Expected %d rabbits, got %d", expectedRabbitCount, len(rabbits))
	}

	wolves := world.QueryByType(core.TypeWolf)
	expectedWolfCount := 1
	if len(wolves) != expectedWolfCount {
		t.Errorf("Expected %d wolf, got %d", expectedWolfCount, len(wolves))
	}

	grass := world.QueryByType(core.TypeGrass)
	expectedGrassCount := 0
	if len(grass) != expectedGrassCount {
		t.Errorf("Expected %d grass entities, got %d", expectedGrassCount, len(grass))
	}
}

// TestQueryGetStats тестирует GetStats
func TestQueryGetStats(t *testing.T) {
	t.Parallel()

	world := setupQueryWorld()
	stats := world.GetStats()

	expectedRabbitCount := 2
	if stats[core.TypeRabbit] != expectedRabbitCount {
		t.Errorf("Expected %d rabbits in stats, got %d", expectedRabbitCount, stats[core.TypeRabbit])
	}

	expectedWolfCount := 1
	if stats[core.TypeWolf] != expectedWolfCount {
		t.Errorf("Expected %d wolf in stats, got %d", expectedWolfCount, stats[core.TypeWolf])
	}

	expectedGrassCount := 0
	if stats[core.TypeGrass] != expectedGrassCount {
		t.Errorf("Expected %d grass in stats, got %d", expectedGrassCount, stats[core.TypeGrass])
	}
}
