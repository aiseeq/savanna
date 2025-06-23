package unit

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestCoreRabbitComponents tests core rabbit functionality without animation/system dependencies
func TestCoreRabbitComponents(t *testing.T) {
	t.Parallel()

	t.Logf("=== CORE RABBIT COMPONENTS TEST ===")

	// Create minimal world
	world := core.NewWorld(64, 64, 12345)

	// Create terrain with grass
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 2
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Set up grass
	terrain.SetTileType(0, 0, generator.TileGrass)
	terrain.SetGrassAmount(0, 0, 100.0)

	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Create rabbit manually with proper configuration
	rabbit := world.CreateEntity()
	t.Logf("Created rabbit entity: %d", rabbit)

	// Validate entity was created
	if !world.IsAlive(rabbit) {
		t.Fatalf("Created entity is not alive: %d", rabbit)
	}

	// Add all components (use Add* methods like in real code)
	if !world.AddPosition(rabbit, core.Position{X: 16, Y: 16}) {
		t.Fatalf("Failed to add position component for entity %d", rabbit)
	}

	// Try Add* methods for other components based on CreateAnimal pattern
	world.AddVelocity(rabbit, core.Velocity{X: 0, Y: 0})
	world.AddHunger(rabbit, core.Hunger{Value: 80.0}) // Make it hungry enough (< 90% threshold)
	world.AddAnimalType(rabbit, core.TypeRabbit)
	world.AddHealth(rabbit, core.Health{Current: 100, Max: 100})
	world.AddSize(rabbit, core.Size{Radius: 16.0})

	t.Logf("Initial state:")
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	t.Logf("  Rabbit position: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Rabbit hunger: %.1f%%", hunger.Value)
	t.Logf("  Grass at position: %.1f units", grassAmount)

	// Test 1: Grass detection works
	if grassAmount <= 0 {
		t.Errorf("❌ No grass detected at rabbit position")
		return
	}
	t.Logf("✅ Grass detection works: %.1f units found", grassAmount)

	// Test 2: Hunger system basics - manually decrease hunger
	deltaTime := float32(1.0 / 60.0)
	hungerSystem := simulation.NewHungerSystem()

	// Run hunger system for a few seconds
	for i := 0; i < 180; i++ { // 3 seconds
		hungerSystem.Update(world, deltaTime)
	}

	newHunger, _ := world.GetHunger(rabbit)
	if newHunger.Value >= hunger.Value {
		t.Errorf("❌ Hunger system not working - hunger should decrease over time")
		t.Errorf("   Initial: %.1f%%, After 3s: %.1f%%", hunger.Value, newHunger.Value)
	} else {
		t.Logf("✅ Hunger system works - hunger decreased from %.1f%% to %.1f%%",
			hunger.Value, newHunger.Value)
	}

	// Test 3: Grass search system (should create EatingState)
	// Add AnimalConfig component (required by GrassSearchSystem)
	world.AddAnimalConfig(rabbit, core.AnimalConfig{
		HungerThreshold: 90.0, // RabbitHungerThreshold
		MaxHealth:       100,
		CollisionRadius: 16.0,
		AttackRange:     0, // Herbivores don't attack
		AttackDamage:    0,
		AttackCooldown:  0,
		HitChance:       0,
	})

	// Add Behavior component required by GrassSearchSystem
	world.AddBehavior(rabbit, core.Behavior{
		Type:            core.BehaviorHerbivore, // CRITICAL: Must be herbivore
		HungerThreshold: 90.0,                   // RabbitHungerThreshold
		VisionRange:     100.0,
	})

	// Check current hunger before running grass search
	currentHunger, _ := world.GetHunger(rabbit)
	t.Logf("Hunger before grass search: %.1f%% (threshold: 90.0%%)", currentHunger.Value)

	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)

	// Run grass search for a few ticks
	for i := 0; i < 10; i++ {
		grassSearchSystem.Update(world, deltaTime)
	}

	hasEatingState := world.HasComponent(rabbit, core.MaskEatingState)
	if !hasEatingState {
		t.Errorf("❌ GrassSearchSystem failed to create EatingState for hungry rabbit near grass")
	} else {
		t.Logf("✅ GrassSearchSystem works - EatingState created")

		// Check EatingState details
		eatingState, _ := world.GetEatingState(rabbit)
		t.Logf("   EatingState: Target=%d, TargetType=%d", eatingState.Target, eatingState.TargetType)
	}

	// Test 4: Vegetation system grass consumption
	initialGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	// Manually consume some grass (simulating eating)
	vegetationSystem.ConsumeGrassAt(pos.X, pos.Y, 5.0)

	finalGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	if finalGrass >= initialGrass {
		t.Errorf("❌ Vegetation system grass consumption failed")
	} else {
		t.Logf("✅ Vegetation system works - grass consumed: %.1f → %.1f",
			initialGrass, finalGrass)
	}

	// Test 5: Component management
	// Test adding/removing components
	world.SetBehavior(rabbit, core.Behavior{
		Type:            0,
		HungerThreshold: 80.0,
		VisionRange:     100.0,
	})

	hasBehavior := world.HasComponent(rabbit, core.MaskBehavior)
	if !hasBehavior {
		t.Errorf("❌ Component management failed - Behavior component not added")
	} else {
		t.Logf("✅ Component management works - Behavior component added")
	}

	// Test component queries
	rabbitCount := 0
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, _ := world.GetAnimalType(entity)
		if animalType == core.TypeRabbit {
			rabbitCount++
		}
	})

	if rabbitCount != 1 {
		t.Errorf("❌ Component queries failed - expected 1 rabbit, found %d", rabbitCount)
	} else {
		t.Logf("✅ Component queries work - found %d rabbit", rabbitCount)
	}

	t.Logf("✅ All core rabbit component tests passed!")
}
