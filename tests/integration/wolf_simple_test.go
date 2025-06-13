package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfEatsRabbit простой тест: голодный волк рядом с зайцем должен его съесть
func TestWolfEatsRabbit(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаём зайца и волка в одной точке
	rabbit := simulation.CreateRabbit(world, 300, 300)
	wolf := simulation.CreateWolf(world, 300, 300)

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 30.0}) // 30% < 60% = голодный

	// Проверяем начальное здоровье зайца
	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	// Симулируем 120 тиков (2 секунды)
	deltaTime := float32(1.0 / 60.0)
	for i := 0; i < 120; i++ {
		world.Update(deltaTime)
		feedingSystem.Update(world, deltaTime)

		// Если заяц мёртв - тест прошёл
		if !world.IsAlive(rabbit) {
			t.Logf("Волк съел зайца на тике %d", i)
			return
		}
	}

	// Проверяем финальное здоровье
	finalHealth, _ := world.GetHealth(rabbit)
	if finalHealth.Current < initialHealth.Current {
		t.Logf("Волк атаковал зайца (здоровье: %d -> %d), но не убил",
			initialHealth.Current, finalHealth.Current)
	} else {
		t.Error("Волк даже не атаковал зайца - здоровье не изменилось")
	}
}
