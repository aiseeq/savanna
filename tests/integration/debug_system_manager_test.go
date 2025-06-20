package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSystemManagerExecution проверяет что SystemManager вызывает системы
//
//nolint:revive // function-length: Детальный тест системного менеджера
func TestSystemManagerExecution(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ ВЫПОЛНЕНИЯ SYSTEM MANAGER ===")

	// Создаём простейший мир
	world := core.NewWorld(1600, 1600, 12345)

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	terrain.SetTileType(25, 25, generator.TileGrass)
	terrain.SetGrassAmount(25, 25, 100.0)
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём зайца
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 816.0, 816.0)
	world.SetHunger(rabbit, core.Hunger{Value: 70.0}) // Голодный

	// Создаём систему и менеджер
	systemManager := core.NewSystemManager()

	// Проверяем начальное состояние
	hungerBefore, _ := world.GetHunger(rabbit)
	eatingStateBefore := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("ДО SystemManager:")
	t.Logf("  Голод: %.1f%%", hungerBefore.Value)
	t.Logf("  EatingState: %v", eatingStateBefore)

	// ТЕСТ 1: Вызов DeprecatedFeedingSystem напрямую (должен работать)
	t.Logf("\n--- ТЕСТ 1: Прямой вызов DeprecatedFeedingSystem ---")
	deltaTime := float32(1.0 / 60.0)
	feedingSystemAdapter := adapters.NewDeprecatedFeedingSystemAdapter(vegetationSystem)
	feedingSystemAdapter.Update(world, deltaTime)

	hungerAfterDirect, _ := world.GetHunger(rabbit)
	eatingStateAfterDirect := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("ПОСЛЕ прямого вызова:")
	t.Logf("  Голод: %.1f%%", hungerAfterDirect.Value)
	t.Logf("  EatingState: %v", eatingStateAfterDirect)

	if eatingStateAfterDirect {
		t.Logf("✅ Прямой вызов работает")
	} else {
		t.Errorf("❌ Прямой вызов НЕ работает")
	}

	// Сбрасываем состояние
	world.RemoveEatingState(rabbit)
	world.SetHunger(rabbit, core.Hunger{Value: 70.0})

	// ТЕСТ 2: Вызов через FeedingSystemAdapter (должен работать)
	t.Logf("\n--- ТЕСТ 2: Вызов через Adapter ---")
	adapter := adapters.NewFeedingSystemAdapter(vegetationSystem)
	adapter.Update(world, deltaTime)

	hungerAfterAdapter, _ := world.GetHunger(rabbit)
	eatingStateAfterAdapter := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("ПОСЛЕ вызова через Adapter:")
	t.Logf("  Голод: %.1f%%", hungerAfterAdapter.Value)
	t.Logf("  EatingState: %v", eatingStateAfterAdapter)

	if eatingStateAfterAdapter {
		t.Logf("✅ Adapter работает")
	} else {
		t.Errorf("❌ Adapter НЕ работает")
	}

	// Сбрасываем состояние
	world.RemoveEatingState(rabbit)
	world.SetHunger(rabbit, core.Hunger{Value: 70.0})

	// ТЕСТ 3: Вызов через SystemManager (должен работать)
	t.Logf("\n--- ТЕСТ 3: Вызов через SystemManager ---")
	systemManager.AddSystem(adapter)

	world.Update(deltaTime) // Это может быть важно
	systemManager.Update(world, deltaTime)

	hungerAfterManager, _ := world.GetHunger(rabbit)
	eatingStateAfterManager := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("ПОСЛЕ вызова через SystemManager:")
	t.Logf("  Голод: %.1f%%", hungerAfterManager.Value)
	t.Logf("  EatingState: %v", eatingStateAfterManager)

	if eatingStateAfterManager {
		t.Logf("✅ SystemManager работает")
		t.Logf("🎯 ПРОБЛЕМА НЕ В SYSTEM MANAGER")
	} else {
		t.Errorf("❌ SystemManager НЕ работает")
		t.Errorf("🔍 ПРОБЛЕМА В SYSTEM MANAGER или в интеграции")
	}
}
