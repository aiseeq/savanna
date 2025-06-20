package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestGrassEatingSystemDirect проверяет работу GrassEatingSystem напрямую
func TestGrassEatingSystemDirect(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ ПРЯМОЙ РАБОТЫ GRASS EATING SYSTEM ===")

	// Создаём мир
	world := core.NewWorld(1600, 1600, 12345)

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Устанавливаем траву в центр
	centerX, centerY := 25, 25
	terrain.SetTileType(centerX, centerY, generator.TileGrass)
	terrain.SetGrassAmount(centerX, centerY, 100.0)

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)

	// Создаём зайца
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)

	// Делаем зайца голодным и ВРУЧНУЮ создаём EatingState (как FeedingSystem)
	world.SetHunger(rabbit, core.Hunger{Value: 70.0})
	eatingState := core.EatingState{
		Target:          0,                      // 0 = поедание травы
		TargetType:      core.EatingTargetGrass, // Тип: поедание травы
		EatingProgress:  0.0,
		NutritionGained: 0.0,
	}
	world.AddEatingState(rabbit, eatingState)

	t.Logf("НАЧАЛЬНОЕ СОСТОЯНИЕ:")
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	hasEatingState := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("  Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Голод зайца: %.1f%%", hunger.Value)
	t.Logf("  Трава в позиции: %.1f единиц", grassAmount)
	t.Logf("  EatingState создан: %v", hasEatingState)
	t.Logf("  MinGrassToFind: %.1f", simulation.MinGrassAmountToFind)

	// Проверяем условия
	hasEnoughGrass := grassAmount >= simulation.MinGrassAmountToFind
	t.Logf("  Достаточно травы: %v (%.1f >= %.1f)", hasEnoughGrass, grassAmount, simulation.MinGrassAmountToFind)

	// ВЫЗЫВАЕМ ТОЛЬКО GRASS EATING SYSTEM
	deltaTime := float32(1.0 / 60.0)
	t.Logf("\n--- ВЫЗОВ GRASS EATING SYSTEM ---")
	grassEatingSystem.Update(world, deltaTime)

	// Проверяем результат
	hasEatingStateAfter := world.HasComponent(rabbit, core.MaskEatingState)
	hungerAfter, _ := world.GetHunger(rabbit)
	grassAfter := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("РЕЗУЛЬТАТ:")
	t.Logf("  EatingState после GrassEatingSystem: %v", hasEatingStateAfter)
	t.Logf("  Голод после: %.1f%%", hungerAfter.Value)
	t.Logf("  Трава после: %.1f", grassAfter)

	if !hasEatingStateAfter {
		t.Errorf("❌ ОШИБКА: GrassEatingSystem УДАЛИЛ EatingState!")
		t.Errorf("   Возможные причины:")
		t.Errorf("   1. grassAmount < MinGrassToFind (%.1f < %.1f)", grassAmount, simulation.MinGrassAmountToFind)
		t.Errorf("   2. Ошибка в логике isEatingAnimationFrameComplete")
		t.Errorf("   3. Другая ошибка в Update логике")
	} else {
		t.Logf("✅ УСПЕХ: GrassEatingSystem сохранил EatingState")

		// Проверяем изменился ли голод
		if hungerAfter.Value > hunger.Value {
			t.Logf("✅ Голод увеличился с %.1f%% до %.1f%%", hunger.Value, hungerAfter.Value)
		} else {
			t.Logf("⚠️  Голод не изменился - анимация ещё не завершена")
		}
	}
}
