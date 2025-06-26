package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

// TestMinimalCore tests core ECS functionality without any GUI dependencies
func TestMinimalCore(t *testing.T) {
	t.Parallel()

	t.Logf("=== MINIMAL CORE TEST: ECS without GUI ===")

	// Test world creation
	world := core.NewWorld(100, 100, 42)
	if world == nil {
		t.Fatalf("Failed to create world")
	}
	t.Logf("✅ World created successfully")

	// Test entity creation
	entity := world.CreateEntity()
	if entity == 0 {
		t.Fatalf("Failed to create entity")
	}
	t.Logf("✅ Entity created: %d", entity)

	// Test position component
	pos := core.NewPosition(10.5, 20.3)
	world.AddPosition(entity, pos)

	retrievedPos, ok := world.GetPosition(entity)
	if !ok {
		t.Errorf("❌ Failed to retrieve position component")
	} else if retrievedPos.X != pos.X || retrievedPos.Y != pos.Y {
		t.Errorf("❌ Position mismatch: expected (%.1f,%.1f), got (%.1f,%.1f)", pos.X, pos.Y, retrievedPos.X, retrievedPos.Y)
	} else {
		t.Logf("✅ Position component works: (%.1f,%.1f)", retrievedPos.X, retrievedPos.Y)
	}

	// Test velocity component
	vel := core.NewVelocity(5.0, -3.0)
	world.AddVelocity(entity, vel)

	retrievedVel, ok := world.GetVelocity(entity)
	if !ok {
		t.Errorf("❌ Failed to retrieve velocity component")
	} else if retrievedVel.X != vel.X || retrievedVel.Y != vel.Y {
		t.Errorf("❌ Velocity mismatch: expected (%.1f,%.1f), got (%.1f,%.1f)", vel.X, vel.Y, retrievedVel.X, retrievedVel.Y)
	} else {
		t.Logf("✅ Velocity component works: (%.1f,%.1f)", retrievedVel.X, retrievedVel.Y)
	}

	// Test health component
	health := core.Health{Current: 80, Max: 100}
	world.AddHealth(entity, health)

	retrievedHealth, ok := world.GetHealth(entity)
	if !ok {
		t.Errorf("❌ Failed to retrieve health component")
	} else if retrievedHealth.Current != health.Current || retrievedHealth.Max != health.Max {
		t.Errorf("❌ Health mismatch: expected %d/%d, got %d/%d", health.Current, health.Max, retrievedHealth.Current, retrievedHealth.Max)
	} else {
		t.Logf("✅ Health component works: %d/%d", retrievedHealth.Current, retrievedHealth.Max)
	}

	// Test component queries
	count := 0
	world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(e core.EntityID) {
		if e == entity {
			count++
		}
	})

	if count != 1 {
		t.Errorf("❌ Query failed: expected 1 entity with position+velocity, got %d", count)
	} else {
		t.Logf("✅ Component queries work")
	}

	// Test physics vector operations
	v1 := physics.NewVec2(3.0, 4.0)
	v2 := physics.NewVec2(1.0, 2.0)
	sum := v1.Add(v2)

	expectedSum := physics.NewVec2(4.0, 6.0)
	if !sum.Equal(expectedSum, 0.001) {
		t.Errorf("❌ Vector addition failed: expected (%.1f,%.1f), got (%.1f,%.1f)", expectedSum.X, expectedSum.Y, sum.X, sum.Y)
	} else {
		t.Logf("✅ Physics vector operations work")
	}

	// Test entity destruction
	world.DestroyEntity(entity)
	if world.IsAlive(entity) {
		t.Errorf("❌ Entity destruction failed: entity %d still alive", entity)
	} else {
		t.Logf("✅ Entity destruction works")
	}

	t.Logf("✅ All core ECS functionality works without GUI dependencies")
}
