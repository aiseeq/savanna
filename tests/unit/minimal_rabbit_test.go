package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// TestMinimalRabbitFunctionality tests only core ECS functionality without any simulation systems
func TestMinimalRabbitFunctionality(t *testing.T) {
	t.Parallel()

	t.Logf("=== MINIMAL RABBIT FUNCTIONALITY TEST ===")

	// Create minimal world
	world := core.NewWorld(64, 64, 12345)

	// Create rabbit entity manually
	rabbit := world.CreateEntity()

	// Test basic component adding/setting
	world.AddPosition(rabbit, core.Position{X: 16, Y: 16})
	world.AddVelocity(rabbit, core.Velocity{X: 0, Y: 0})
	world.AddSatiation(rabbit, core.Satiation{Value: 50.0})
	world.AddAnimalType(rabbit, core.TypeRabbit)
	world.AddHealth(rabbit, core.Health{Current: 100, Max: 100})

	// Test 1: Position component
	pos, hasPos := world.GetPosition(rabbit)
	if !hasPos {
		t.Errorf("❌ Position component not found")
		return
	}
	if pos.X != 16 || pos.Y != 16 {
		t.Errorf("❌ Position incorrect: expected (16,16), got (%.1f,%.1f)", pos.X, pos.Y)
	} else {
		t.Logf("✅ Position component works: (%.1f, %.1f)", pos.X, pos.Y)
	}

	// Test 2: Hunger component
	hunger, hasHunger := world.GetSatiation(rabbit)
	if !hasHunger {
		t.Errorf("❌ Hunger component not found")
		return
	}
	if hunger.Value != 50.0 {
		t.Errorf("❌ Hunger incorrect: expected 50.0, got %.1f", hunger.Value)
	} else {
		t.Logf("✅ Hunger component works: %.1f%%", hunger.Value)
	}

	// Test 3: Animal type component
	animalType, hasType := world.GetAnimalType(rabbit)
	if !hasType {
		t.Errorf("❌ AnimalType component not found")
		return
	}
	if animalType != core.TypeRabbit {
		t.Errorf("❌ AnimalType incorrect: expected %d, got %d", core.TypeRabbit, animalType)
	} else {
		t.Logf("✅ AnimalType component works: type %d (rabbit)", animalType)
	}

	// Test 4: Health component
	health, hasHealth := world.GetHealth(rabbit)
	if !hasHealth {
		t.Errorf("❌ Health component not found")
		return
	}
	if health.Current != 100 || health.Max != 100 {
		t.Errorf("❌ Health incorrect: expected (100,100), got (%d,%d)", health.Current, health.Max)
	} else {
		t.Logf("✅ Health component works: %d/%d", health.Current, health.Max)
	}

	// Test 5: Component masks and queries
	positionExists := world.HasComponent(rabbit, core.MaskPosition)
	hungerExists := world.HasComponent(rabbit, core.MaskSatiation)
	animalTypeExists := world.HasComponent(rabbit, core.MaskAnimalType)

	if !positionExists || !hungerExists || !animalTypeExists {
		t.Errorf("❌ Component mask query failed: pos=%t, hunger=%t, type=%t",
			positionExists, hungerExists, animalTypeExists)
	} else {
		t.Logf("✅ Component mask queries work")
	}

	// Test 6: Modify hunger (simulate hunger system effect)
	world.SetSatiation(rabbit, core.Satiation{Value: 40.0})
	newHunger, _ := world.GetSatiation(rabbit)
	if newHunger.Value != 40.0 {
		t.Errorf("❌ Hunger modification failed: expected 40.0, got %.1f", newHunger.Value)
	} else {
		t.Logf("✅ Hunger modification works: %.1f%% → %.1f%%", hunger.Value, newHunger.Value)
	}

	// Test 7: Entity iteration
	rabbitCount := 0
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		entityType, _ := world.GetAnimalType(entity)
		if entityType == core.TypeRabbit {
			rabbitCount++
		}
	})

	if rabbitCount != 1 {
		t.Errorf("❌ Entity iteration failed: expected 1 rabbit, found %d", rabbitCount)
	} else {
		t.Logf("✅ Entity iteration works: found %d rabbit", rabbitCount)
	}

	// Test 8: Test eating state creation (without systems)
	world.AddEatingState(rabbit, core.EatingState{
		Target:     0, // Eating grass tile
		TargetType: 1, // Grass type
	})

	hasEatingState := world.HasComponent(rabbit, core.MaskEatingState)
	if !hasEatingState {
		t.Errorf("❌ EatingState component creation failed")
	} else {
		eatingState, _ := world.GetEatingState(rabbit)
		t.Logf("✅ EatingState component works: Target=%d, TargetType=%d",
			eatingState.Target, eatingState.TargetType)
	}

	// Test 9: World time management
	initialTime := world.GetTime()
	deltaTime := float32(1.0 / 60.0)
	world.Update(deltaTime)
	newTime := world.GetTime()

	if newTime <= initialTime {
		t.Errorf("❌ World time update failed: time did not advance")
	} else {
		t.Logf("✅ World time management works: %.3f → %.3f", initialTime, newTime)
	}

	// Test 10: Entity lifecycle
	wolf := world.CreateEntity()
	world.AddAnimalType(wolf, core.TypeWolf)

	// Count entities
	totalEntities := 0
	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		totalEntities++
	})

	if totalEntities != 2 {
		t.Errorf("❌ Entity creation failed: expected 2 entities, found %d", totalEntities)
	} else {
		t.Logf("✅ Entity lifecycle works: created %d entities", totalEntities)
	}

	t.Logf("✅ All minimal rabbit functionality tests passed!")
	t.Logf("Core ECS system is fully functional for rabbit simulation")
}
