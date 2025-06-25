package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// УДАЛЕНО: TestEntityManager - перенесён в entity_manager_test.go

// TestComponentMasks тестирует работу с битовыми масками компонентов
//
//nolint:gocognit // Комплексный unit тест масок компонентов ECS
func TestComponentMasks(t *testing.T) {
	t.Parallel()
	t.Run("HasComponent", func(t *testing.T) {
		t.Parallel()
		mask := core.MaskPosition | core.MaskVelocity

		if !mask.HasComponent(core.MaskPosition) {
			t.Error("Mask should have Position component")
		}

		if !mask.HasComponent(core.MaskVelocity) {
			t.Error("Mask should have Velocity component")
		}

		if mask.HasComponent(core.MaskHealth) {
			t.Error("Mask should not have Health component")
		}
	})

	t.Run("AddRemoveComponent", func(t *testing.T) {
		t.Parallel()
		mask := core.MaskPosition

		// Добавляем компонент
		mask = mask.AddComponent(core.MaskVelocity)
		if !mask.HasComponent(core.MaskVelocity) {
			t.Error("Should have Velocity after adding")
		}

		// Удаляем компонент
		mask = mask.RemoveComponent(core.MaskPosition)
		if mask.HasComponent(core.MaskPosition) {
			t.Error("Should not have Position after removing")
		}

		if !mask.HasComponent(core.MaskVelocity) {
			t.Error("Should still have Velocity")
		}
	})

	t.Run("ComponentSet", func(t *testing.T) {
		t.Parallel()
		cs := core.NewComponentSet(core.MaskPosition, core.MaskVelocity)

		if !cs.Has(core.MaskPosition) {
			t.Error("ComponentSet should have Position")
		}

		if !cs.HasAll(core.MaskPosition | core.MaskVelocity) {
			t.Error("ComponentSet should have both Position and Velocity")
		}

		cs.Add(core.MaskHealth)
		if !cs.Has(core.MaskHealth) {
			t.Error("ComponentSet should have Health after adding")
		}

		cs.Remove(core.MaskPosition)
		if cs.Has(core.MaskPosition) {
			t.Error("ComponentSet should not have Position after removing")
		}
	})
}

// TestWorld тестирует основную функциональность World
//
//nolint:gocognit // Комплексный unit тест мира ECS
func TestWorld(t *testing.T) {
	t.Parallel()

	t.Run("CreateDestroyEntity", func(t *testing.T) {
		t.Parallel()
		world := core.NewWorld(100, 100, 42)
		entity := world.CreateEntity()

		if entity == core.InvalidEntity {
			t.Error("Expected valid entity")
		}

		if !world.IsAlive(entity) {
			t.Error("Entity should be alive")
		}

		if world.GetEntityCount() != 1 {
			t.Errorf("Expected 1 entity, got %d", world.GetEntityCount())
		}

		if !world.DestroyEntity(entity) {
			t.Error("Should successfully destroy entity")
		}

		if world.IsAlive(entity) {
			t.Error("Entity should not be alive after destruction")
		}

		if world.GetEntityCount() != 0 {
			t.Errorf("Expected 0 entities, got %d", world.GetEntityCount())
		}
	})

	t.Run("TimeManagement", func(t *testing.T) {
		t.Parallel()
		world := core.NewWorld(100, 100, 42)
		initialTime := world.GetTime()

		world.Update(1.0 / 60.0) // 60 FPS

		if world.GetDeltaTime() != 1.0/60.0 {
			t.Errorf("Expected delta time %f, got %f", 1.0/60.0, world.GetDeltaTime())
		}

		if world.GetTime() <= initialTime {
			t.Error("Time should advance after Update")
		}

		// Тест масштаба времени
		world.SetTimeScale(2.0)
		if world.GetTimeScale() != 2.0 {
			t.Errorf("Expected time scale 2.0, got %f", world.GetTimeScale())
		}

		prevTime := world.GetTime()
		world.Update(1.0)
		expectedTime := prevTime + 2.0 // 1.0 * 2.0 time scale

		if world.GetTime() != expectedTime {
			t.Errorf("Expected time %f, got %f", expectedTime, world.GetTime())
		}
	})

	t.Run("WorldDimensions", func(t *testing.T) {
		t.Parallel()
		world := core.NewWorld(100, 100, 42)
		width, height := world.GetWorldDimensions()

		if width != 100 {
			t.Errorf("Expected width 100, got %f", width)
		}

		if height != 100 {
			t.Errorf("Expected height 100, got %f", height)
		}
	})
}

// УДАЛЕНО: TestComponents - перенесён в components_test.go

// УДАЛЕНО: TestQueries - перенесён в queries_test.go

// УДАЛЕНО: TestSpatialIntegration - перенесён в spatial_integration_test.go

// TestDestroyEntityCleanup тестирует что уничтожение сущности очищает все данные
func TestDestroyEntityCleanup(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(100, 100, 42)
	defer world.Clear()

	entity := world.CreateEntity()

	// Добавляем все возможные компоненты
	world.AddPosition(entity, core.Position{X: 10, Y: 10})
	world.AddVelocity(entity, core.Velocity{X: 1, Y: 1})
	world.AddHealth(entity, core.Health{Current: 100, Max: 100})
	world.AddSatiation(entity, core.Satiation{Value: 50})
	world.AddAnimalType(entity, core.TypeRabbit)
	world.AddSize(entity, core.Size{Radius: 5, AttackRange: 0})
	world.AddSpeed(entity, core.Speed{Base: 20, Current: 15})

	// Проверяем что все компоненты есть
	if !world.HasComponents(entity,
		core.MaskPosition|core.MaskVelocity|core.MaskHealth|core.MaskSatiation|
			core.MaskAnimalType|core.MaskSize|core.MaskSpeed) {
		t.Error("Entity should have all components before destruction")
	}

	// Уничтожаем сущность
	world.DestroyEntity(entity)

	// Проверяем что все компоненты удалены
	if world.HasComponent(entity, core.MaskPosition) {
		t.Error("Entity should not have Position after destruction")
	}

	if world.HasComponent(entity, core.MaskVelocity) {
		t.Error("Entity should not have Velocity after destruction")
	}

	// Проверяем что сущность не появляется в запросах
	count := world.CountEntitiesWith(core.MaskAnimalType)
	if count != 0 {
		t.Errorf("Expected 0 animals after destruction, got %d", count)
	}

	entities := world.QueryEntitiesWith(core.MaskPosition)
	if len(entities) != 0 {
		t.Errorf("Expected 0 entities with position after destruction, got %d", len(entities))
	}
}
