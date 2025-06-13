package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDirectAttack тестирует прямую атаку волка на зайца
func TestDirectAttack(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаём зайца
	rabbit := simulation.CreateRabbit(world, 300, 300)
	wolf := simulation.CreateWolf(world, 300, 300)

	// Проверяем начальное здоровье
	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Здоровье зайца до атаки: %d", initialHealth.Current)

	// Проверяем типы созданных животных
	wolfType, _ := world.GetAnimalType(wolf)
	rabbitType, _ := world.GetAnimalType(rabbit)

	t.Logf("Тип волка: %d (ожидается %d)", wolfType, core.TypeWolf)
	t.Logf("Тип зайца: %d (ожидается %d)", rabbitType, core.TypeRabbit)

	if wolfType != core.TypeWolf {
		t.Errorf("Волк имеет неправильный тип: %d, ожидается %d", wolfType, core.TypeWolf)
	}
	if rabbitType != core.TypeRabbit {
		t.Errorf("Заяц имеет неправильный тип: %d, ожидается %d", rabbitType, core.TypeRabbit)
	}

	// Тестируем механику охоты
	world.SetHunger(wolf, core.Hunger{Value: 30.0}) // Делаем волка голодным
	feedingSystem.Update(world, 1.0/60.0)

	finalHealth, _ := world.GetHealth(rabbit)
	if finalHealth.Current < initialHealth.Current {
		t.Logf("Волк успешно атаковал зайца (HP: %d -> %d)", initialHealth.Current, finalHealth.Current)
	} else {
		t.Error("Волк не атаковал зайца")
	}
}
