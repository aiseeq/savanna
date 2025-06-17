package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

func TestVegetationIntegration_BasicFunctionality(t *testing.T) {
	t.Parallel()
	// Создаем простую симуляцию
	cfg := &config.Config{
		World: config.WorldConfig{
			Size: 10,
			Seed: 12345,
		},
	}
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаем мир и системы
	world := core.NewWorld(320.0, 320.0, 12345)
	systemManager := core.NewSystemManager()

	// Создаем систему растительности
	vegSystem := simulation.NewVegetationSystem(terrain)
	systemManager.AddSystem(vegSystem)

	// Устанавливаем начальное количество травы
	terrain.SetGrassAmount(5, 5, 50.0)
	initialGrass := terrain.GetGrassAmount(5, 5)

	// Запускаем симуляцию на время роста травы
	deltaTime := float32(2.0) // 2 секунды за шаг
	for i := 0; i < 10; i++ { // 20 секунд симуляции
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
	}

	// Проверяем что трава выросла
	finalGrass := terrain.GetGrassAmount(5, 5)
	if finalGrass <= initialGrass {
		t.Errorf("Трава должна была вырасти. Было: %f, стало: %f", initialGrass, finalGrass)
	}

	// Проверяем что трава не превысила максимум
	if finalGrass > 100.0 {
		t.Errorf("Трава превысила максимум: %f", finalGrass)
	}
}
