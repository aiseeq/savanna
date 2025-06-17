package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestFeedingSystemDirect проверяет работу FeedingSystem напрямую
func TestFeedingSystemDirect(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ ПРЯМОЙ РАБОТЫ FEEDING SYSTEM ===")

	// Создаём мир
	world := core.NewWorld(1600, 1600, 12345)

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Устанавливаем траву в центр
	centerX, centerY := 25, 25
	terrain.SetGrassAmount(centerX, centerY, 100.0)
	grassInTerrain := terrain.GetGrassAmount(centerX, centerY)
	t.Logf("Трава в terrain тайле (%d, %d): %.1f", centerX, centerY, grassInTerrain)

	// Создаём системы
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаём зайца точно в центре тайла с травой
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16)
	rabbit := simulation.CreateRabbit(world, rabbitX, rabbitY)

	// Делаем зайца голодным
	world.SetHunger(rabbit, core.Hunger{Value: 70.0}) // 70% < 90% = голодный

	// Проверяем начальное состояние
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	tileX := int(pos.X / 32)
	tileY := int(pos.Y / 32)
	grassViaVegetation := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("НАЧАЛЬНОЕ СОСТОЯНИЕ:")
	t.Logf("  Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Тайл зайца: (%d, %d), ожидаем (%d, %d)", tileX, tileY, centerX, centerY)
	t.Logf("  Голод зайца: %.1f%% (порог %.1f%%)", hunger.Value, simulation.RabbitHungryThreshold)
	t.Logf("  Трава через VegetationSystem: %.1f", grassViaVegetation)
	t.Logf("  Минимум травы для поедания: %.1f", simulation.MinGrassToFind)

	// Проверяем условия
	isHungry := hunger.Value < simulation.RabbitHungryThreshold
	hasEnoughGrass := grassViaVegetation >= simulation.MinGrassToFind
	hasEatingStateBefore := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("ПРОВЕРКА УСЛОВИЙ:")
	t.Logf("  Заяц голоден: %v (%.1f < %.1f)", isHungry, hunger.Value, simulation.RabbitHungryThreshold)
	t.Logf("  Достаточно травы: %v (%.1f >= %.1f)", hasEnoughGrass, grassViaVegetation, simulation.MinGrassToFind)
	t.Logf("  EatingState до FeedingSystem: %v", hasEatingStateBefore)

	// ВЫЗЫВАЕМ ТОЛЬКО FEEDING SYSTEM
	deltaTime := float32(1.0 / 60.0)
	t.Logf("\n--- ВЫЗОВ FEEDING SYSTEM ---")
	feedingSystem.Update(world, deltaTime)

	// Проверяем результат
	hasEatingStateAfter := world.HasComponent(rabbit, core.MaskEatingState)
	hungerAfter, _ := world.GetHunger(rabbit)

	t.Logf("РЕЗУЛЬТАТ:")
	t.Logf("  EatingState после FeedingSystem: %v", hasEatingStateAfter)
	t.Logf("  Голод после FeedingSystem: %.1f%%", hungerAfter.Value)

	if hasEatingStateAfter {
		eatingState, _ := world.GetEatingState(rabbit)
		t.Logf("  EatingState детали: Target=%d, Progress=%.2f, Nutrition=%.2f",
			eatingState.Target, eatingState.EatingProgress, eatingState.NutritionGained)
		t.Logf("✅ УСПЕХ: FeedingSystem создал EatingState!")
	} else {
		t.Errorf("❌ ОШИБКА: FeedingSystem НЕ создал EatingState!")
		t.Errorf("   Все условия выполнены, но состояние не создано")
		t.Errorf("   Возможные причины:")
		t.Errorf("   1. Заяц не найден в ForEachWith")
		t.Errorf("   2. Ошибка в проверке условий")
		t.Errorf("   3. Ошибка в AddEatingState")
	}
}
