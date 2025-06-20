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

	// Настройка тестовой среды
	world, systems, rabbit := setupSystemOrderTest(t)

	// Пошаговое тестирование систем
	testSystemsStepByStep(t, world, systems, rabbit)
}

// systemTestData содержит все системы для тестирования
type systemTestData struct {
	vegetationSystem     *simulation.VegetationSystem
	feedingSystem        *simulation.FeedingSystem
	grassEatingSystem    *simulation.GrassEatingSystem
	animalBehaviorSystem *simulation.AnimalBehaviorSystem
	movementSystem       *simulation.MovementSystem
	combatSystem         *simulation.CombatSystem
}

// setupSystemOrderTest настраивает тестовую среду для тестирования порядка систем
func setupSystemOrderTest(_ *testing.T) (*core.World, *systemTestData, core.EntityID) {
	// Создаём мир точно как в больших тестах
	world := core.NewWorld(1600, 1600, 12345) //nolint:gomnd // Тестовые параметры

	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50 //nolint:gomnd // Тестовый размер мира
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	centerX, centerY := 25, 25 //nolint:gomnd // Центр тестового мира
	terrain.SetTileType(centerX, centerY, generator.TileGrass)
	terrain.SetGrassAmount(centerX, centerY, 100.0) //nolint:gomnd // Максимум травы

	systems := &systemTestData{
		vegetationSystem:     simulation.NewVegetationSystem(terrain),
		feedingSystem:        simulation.NewFeedingSystem(simulation.NewVegetationSystem(terrain)),
		grassEatingSystem:    simulation.NewGrassEatingSystem(simulation.NewVegetationSystem(terrain)),
		animalBehaviorSystem: simulation.NewAnimalBehaviorSystem(simulation.NewVegetationSystem(terrain)),
		movementSystem:       simulation.NewMovementSystem(1600, 1600), //nolint:gomnd // Размер мира
		combatSystem:         simulation.NewCombatSystem(),
	}

	// Создаём зайца
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16) //nolint:gomnd // Центр тайла
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)
	world.SetHunger(rabbit, core.Hunger{Value: 70.0}) //nolint:gomnd // 70% = голодный

	return world, systems, rabbit
}

// testSystemsStepByStep выполняет пошаговое тестирование систем
func testSystemsStepByStep(t *testing.T, world *core.World, systems *systemTestData, rabbit core.EntityID) {
	deltaTime := float32(1.0 / 60.0) //nolint:gomnd // 60 FPS

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
	systems.vegetationSystem.Update(world, deltaTime)
	checkState("После VegetationSystem")

	// Система 2: FeedingSystem (через адаптер)
	t.Logf("\n--- 2. FeedingSystem ---")
	feedingAdapter := adapters.NewFeedingSystemAdapter(systems.vegetationSystem)
	feedingAdapter.Update(world, deltaTime)
	checkState("После FeedingSystem")

	// Система 3: GrassEatingSystem
	t.Logf("\n--- 3. GrassEatingSystem ---")
	systems.grassEatingSystem.Update(world, deltaTime)
	checkState("После GrassEatingSystem")

	// Система 4: BehaviorSystem (через адаптер)
	t.Logf("\n--- 4. BehaviorSystem ---")
	behaviorAdapter := &adapters.BehaviorSystemAdapter{System: systems.animalBehaviorSystem}
	behaviorAdapter.Update(world, deltaTime)
	checkState("После BehaviorSystem")

	// Система 5: MovementSystem (через адаптер)
	t.Logf("\n--- 5. MovementSystem ---")
	movementAdapter := &adapters.MovementSystemAdapter{System: systems.movementSystem}
	movementAdapter.Update(world, deltaTime)
	checkState("После MovementSystem")

	// Система 6: CombatSystem
	t.Logf("\n--- 6. CombatSystem ---")
	systems.combatSystem.Update(world, deltaTime)
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
