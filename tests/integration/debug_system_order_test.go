package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSystemOrder проверяет каждую систему по очереди
func TestSystemOrder(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ ПОРЯДКА СИСТЕМ ===")

	// Создаём мир точно как в больших тестах
	world := core.NewWorld(1600, 1600, 12345)

	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	centerX, centerY := 25, 25
	terrain.SetGrassAmount(centerX, centerY, 100.0)

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(1600, 1600)
	combatSystem := simulation.NewCombatSystem()

	// Создаём зайца
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16)
	rabbit := simulation.CreateRabbit(world, rabbitX, rabbitY)
	world.SetHunger(rabbit, core.Hunger{Value: 70.0})

	deltaTime := float32(1.0 / 60.0)

	checkState := func(stage string) {
		hunger, _ := world.GetHunger(rabbit)
		hasEatingState := world.HasComponent(rabbit, core.MaskEatingState)
		vel, _ := world.GetVelocity(rabbit)
		speed := vel.X*vel.X + vel.Y*vel.Y

		t.Logf("  %s: голод=%.1f%%, EatingState=%v, скорость=%.2f", stage, hunger.Value, hasEatingState, speed)
	}

	t.Logf("ПОШАГОВОЕ ВЫПОЛНЕНИЕ СИСТЕМ:")
	checkState("НАЧАЛО")

	// Система 1: VegetationSystem
	t.Logf("\n--- 1. VegetationSystem ---")
	vegetationSystem.Update(world, deltaTime)
	checkState("После VegetationSystem")

	// Система 2: FeedingSystem (через адаптер)
	t.Logf("\n--- 2. FeedingSystem ---")
	feedingAdapter := &adapters.FeedingSystemAdapter{System: feedingSystem}
	feedingAdapter.Update(world, deltaTime)
	checkState("После FeedingSystem")

	// Система 3: GrassEatingSystem
	t.Logf("\n--- 3. GrassEatingSystem ---")
	grassEatingSystem.Update(world, deltaTime)
	checkState("После GrassEatingSystem")

	// Система 4: BehaviorSystem (через адаптер)
	t.Logf("\n--- 4. BehaviorSystem ---")
	behaviorAdapter := &adapters.BehaviorSystemAdapter{System: animalBehaviorSystem}
	behaviorAdapter.Update(world, deltaTime)
	checkState("После BehaviorSystem")

	// Система 5: MovementSystem (через адаптер)
	t.Logf("\n--- 5. MovementSystem ---")
	movementAdapter := &adapters.MovementSystemAdapter{System: movementSystem}
	movementAdapter.Update(world, deltaTime)
	checkState("После MovementSystem")

	// Система 6: CombatSystem
	t.Logf("\n--- 6. CombatSystem ---")
	combatSystem.Update(world, deltaTime)
	checkState("После CombatSystem")

	// Финальная проверка
	finalHasEatingState := world.HasComponent(rabbit, core.MaskEatingState)

	t.Logf("\nФИНАЛЬНЫЙ РЕЗУЛЬТАТ:")
	if finalHasEatingState {
		t.Logf("✅ EatingState сохранился через все системы")
	} else {
		t.Errorf("❌ EatingState был удалён одной из систем")
		t.Errorf("   Проверьте логи выше чтобы увидеть какая система удалила состояние")
	}
}
