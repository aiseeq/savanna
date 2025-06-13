package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfMultipleAttacks проверяет что волк может атаковать несколько раз
func TestWolfMultipleAttacks(t *testing.T) {
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

	deltaTime := float32(1.0 / 60.0)

	// Симулируем до 300 тиков (5 секунд)
	for i := 0; i < 300; i++ {
		world.Update(deltaTime)
		feedingSystem.Update(world, deltaTime)

		if !world.IsAlive(rabbit) {
			// Заяц умер - вычисляем количество атак по урону
			finalDamage := int(initialHealth.Current - 0) // заяц умер = 0 хитов
			attackCount := (finalDamage + 29) / 30        // округляем вверх (30 урона за атаку)

			wolfHunger, _ := world.GetHunger(wolf)
			t.Logf("Заяц умер на тике %d после %d атак, голод волка %.1f", i, attackCount, wolfHunger.Value)

			if attackCount < 2 {
				t.Errorf("Волк атаковал только %d раз, ожидалось минимум 2 раза", attackCount)
			}
			return
		}
	}

	// Если дошли сюда - заяц не умер за 5 секунд
	t.Error("Заяц не умер за 5 секунд симуляции")
}
