package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestHungerRate проверяет скорость голода без еды
func TestHungerRate(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаём зайца без травы чтобы проверить скорость голода
	rabbit := simulation.CreateRabbit(world, 300, 300)

	tileX, tileY := int(300/32), int(300/32)
	terrain.SetGrassAmount(tileX, tileY, 0.0) // Нет травы - заяц будет голодать

	// Устанавливаем начальный голод 100% (сытый)
	world.SetHunger(rabbit, core.Hunger{Value: 100.0})

	deltaTime := float32(1.0 / 60.0)

	t.Log("=== Тест скорости голода без еды ===")

	// Симулируем 5 секунд (300 тиков)
	for i := 0; i < 300; i++ {
		feedingSystem.Update(world, deltaTime)

		// Отладка каждую секунду
		if i%60 == 59 {
			currentHunger, _ := world.GetHunger(rabbit)
			t.Logf("  Секунда %d: голод %.1f", (i+1)/60, currentHunger.Value)
		}
	}

	finalHunger, _ := world.GetHunger(rabbit)
	t.Logf("Итого за 5 сек: голод %.1f", finalHunger.Value)

	// За 5 секунд голод должен уменьшиться примерно на 5% (0.033 * 60 * 5 ≈ 10%)
	expectedDecrease := 5.0 * 0.033 * 60.0 // 5 сек * rate * 60 тиков/сек = 9.9%
	actualDecrease := 100.0 - finalHunger.Value

	t.Logf("Ожидалось уменьшение: %.1f%%, получили: %.1f%%", expectedDecrease, actualDecrease)

	if actualDecrease < 8.0 || actualDecrease > 12.0 {
		t.Errorf("Голод уменьшился неправильно: ожидалось ~10%%, получили %.1f%%", actualDecrease)
	}
}
