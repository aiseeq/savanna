package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfContinuousHunting проверяет что волк продолжает охотиться после поедания зайца
func TestWolfContinuousHunting(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаём несколько зайцев и одного волка в одной точке
	rabbit1 := simulation.CreateRabbit(world, 300, 300)
	rabbit2 := simulation.CreateRabbit(world, 300, 300)
	rabbit3 := simulation.CreateRabbit(world, 300, 300)
	wolf := simulation.CreateWolf(world, 300, 300)

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 20.0}) // 20% = очень голодный

	killedRabbits := 0
	deltaTime := float32(1.0 / 60.0)

	// Симулируем до 1800 тиков (30 секунд)
	for i := 0; i < 1800; i++ {
		world.Update(deltaTime)
		feedingSystem.Update(world, deltaTime)

		// Подсчитываем мёртвых зайцев
		if !world.IsAlive(rabbit1) && killedRabbits == 0 {
			killedRabbits = 1
			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("Заяц 1 умер на тике %d, голод волка %.1f", i, wolfHunger.Value)
		}
		if !world.IsAlive(rabbit2) && killedRabbits == 1 {
			killedRabbits = 2
			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("Заяц 2 умер на тике %d, голод волка %.1f", i, wolfHunger.Value)
		}
		if !world.IsAlive(rabbit3) && killedRabbits == 2 {
			killedRabbits = 3
			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("Заяц 3 умер на тике %d, голод волка %.1f", i, wolfHunger.Value)
			break
		}
	}

	t.Logf("Волк убил %d зайцев за 30 секунд", killedRabbits)

	if killedRabbits < 2 {
		t.Errorf("Ожидалось что волк убьёт минимум 2 зайцев, но убил только %d", killedRabbits)
	}
}
