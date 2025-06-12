package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// TestEntityManager тестирует базовую функциональность EntityManager
func TestEntityManager(t *testing.T) {
	em := core.NewEntityManager()

	// Тест создания сущностей
	t.Run("CreateEntity", func(t *testing.T) {
		entity1 := em.CreateEntity()
		entity2 := em.CreateEntity()

		if entity1 == core.INVALID_ENTITY {
			t.Error("Expected valid entity ID, got INVALID_ENTITY")
		}

		if entity2 == core.INVALID_ENTITY {
			t.Error("Expected valid entity ID, got INVALID_ENTITY")
		}

		if entity1 == entity2 {
			t.Error("Expected different entity IDs")
		}

		if em.Count() != 2 {
			t.Errorf("Expected 2 entities, got %d", em.Count())
		}
	})

	// Тест проверки живости
	t.Run("IsAlive", func(t *testing.T) {
		entity := em.CreateEntity()

		if !em.IsAlive(entity) {
			t.Error("Entity should be alive after creation")
		}

		if em.IsAlive(core.INVALID_ENTITY) {
			t.Error("INVALID_ENTITY should not be alive")
		}

		if em.IsAlive(999) {
			t.Error("Non-existent entity should not be alive")
		}
	})

	// Тест уничтожения сущностей
	t.Run("DestroyEntity", func(t *testing.T) {
		entity := em.CreateEntity()
		initialCount := em.Count()

		if !em.DestroyEntity(entity) {
			t.Error("DestroyEntity should return true for valid entity")
		}

		if em.Count() != initialCount-1 {
			t.Errorf("Expected count %d, got %d", initialCount-1, em.Count())
		}

		if em.IsAlive(entity) {
			t.Error("Entity should not be alive after destruction")
		}

		// Попытка повторного уничтожения
		if em.DestroyEntity(entity) {
			t.Error("DestroyEntity should return false for already destroyed entity")
		}
	})

	// Тест переиспользования ID
	t.Run("ReuseIDs", func(t *testing.T) {
		em.Clear()

		// Создаем несколько сущностей
		entities := make([]core.EntityID, 5)
		for i := range entities {
			entities[i] = em.CreateEntity()
		}

		// Уничтожаем некоторые
		em.DestroyEntity(entities[1])
		em.DestroyEntity(entities[3])

		// Создаем новые - должны переиспользовать ID
		newEntity1 := em.CreateEntity()
		newEntity2 := em.CreateEntity()

		// Проверяем что переиспользовались старые ID
		if newEntity1 != entities[3] && newEntity1 != entities[1] {
			t.Errorf("Expected reused ID, got new ID %d", newEntity1)
		}

		if newEntity2 != entities[3] && newEntity2 != entities[1] {
			t.Errorf("Expected reused ID, got new ID %d", newEntity2)
		}

		if newEntity1 == newEntity2 {
			t.Error("Should not reuse the same ID twice")
		}
	})

	// Тест лимита сущностей
	t.Run("EntityLimit", func(t *testing.T) {
		em.Clear()

		// Создаем максимальное количество сущностей
		for i := 0; i < core.MAX_ENTITIES-1; i++ {
			entity := em.CreateEntity()
			if entity == core.INVALID_ENTITY {
				t.Fatalf("Failed to create entity %d", i)
			}
		}

		// Следующая сущность должна вернуть INVALID_ENTITY
		entity := em.CreateEntity()
		if entity != core.INVALID_ENTITY {
			t.Error("Expected INVALID_ENTITY when hitting limit")
		}
	})
}

// TestComponentMasks тестирует работу с битовыми масками компонентов
func TestComponentMasks(t *testing.T) {
	t.Run("HasComponent", func(t *testing.T) {
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
func TestWorld(t *testing.T) {
	world := core.NewWorld(100, 100, 42)
	defer world.Clear()

	t.Run("CreateDestroyEntity", func(t *testing.T) {
		entity := world.CreateEntity()

		if entity == core.INVALID_ENTITY {
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
		width, height := world.GetWorldDimensions()

		if width != 100 {
			t.Errorf("Expected width 100, got %f", width)
		}

		if height != 100 {
			t.Errorf("Expected height 100, got %f", height)
		}
	})
}

// TestComponents тестирует работу с компонентами
func TestComponents(t *testing.T) {
	world := core.NewWorld(100, 100, 42)
	defer world.Clear()

	entity := world.CreateEntity()

	t.Run("Position", func(t *testing.T) {
		// Проверяем что компонента изначально нет
		if world.HasComponent(entity, core.MaskPosition) {
			t.Error("Entity should not have Position initially")
		}

		_, ok := world.GetPosition(entity)
		if ok {
			t.Error("GetPosition should return false for non-existent component")
		}

		// Добавляем компонент
		pos := core.Position{X: 10, Y: 20}
		if !world.AddPosition(entity, pos) {
			t.Error("AddPosition should succeed")
		}

		if !world.HasComponent(entity, core.MaskPosition) {
			t.Error("Entity should have Position after adding")
		}

		// Получаем компонент
		retrievedPos, ok := world.GetPosition(entity)
		if !ok {
			t.Error("GetPosition should return true for existing component")
		}

		if retrievedPos.X != pos.X || retrievedPos.Y != pos.Y {
			t.Errorf("Expected position %+v, got %+v", pos, retrievedPos)
		}

		// Изменяем позицию
		newPos := core.Position{X: 30, Y: 40}
		if !world.SetPosition(entity, newPos) {
			t.Error("SetPosition should succeed")
		}

		retrievedPos, _ = world.GetPosition(entity)
		if retrievedPos.X != newPos.X || retrievedPos.Y != newPos.Y {
			t.Errorf("Expected position %+v, got %+v", newPos, retrievedPos)
		}

		// Удаляем компонент
		if !world.RemovePosition(entity) {
			t.Error("RemovePosition should succeed")
		}

		if world.HasComponent(entity, core.MaskPosition) {
			t.Error("Entity should not have Position after removal")
		}
	})

	t.Run("MultipleComponents", func(t *testing.T) {
		// Добавляем несколько компонентов
		world.AddPosition(entity, core.Position{X: 5, Y: 5})
		world.AddVelocity(entity, core.Velocity{X: 1, Y: 1})
		world.AddHealth(entity, core.Health{Current: 100, Max: 100})

		// Проверяем наличие компонентов
		if !world.HasComponents(entity, core.MaskPosition|core.MaskVelocity|core.MaskHealth) {
			t.Error("Entity should have all three components")
		}

		if world.HasComponents(entity, core.MaskPosition|core.MaskHunger) {
			t.Error("Entity should not have Hunger component")
		}

		// Проверяем что можем получить все компоненты
		pos, posOk := world.GetPosition(entity)
		vel, velOk := world.GetVelocity(entity)
		health, healthOk := world.GetHealth(entity)

		if !posOk || !velOk || !healthOk {
			t.Error("Should be able to get all added components")
		}

		if pos.X != 5 || vel.X != 1 || health.Current != 100 {
			t.Error("Component values should match what was set")
		}
	})

	t.Run("AnimalType", func(t *testing.T) {
		world.AddAnimalType(entity, core.TypeRabbit)

		animalType, ok := world.GetAnimalType(entity)
		if !ok {
			t.Error("Should be able to get AnimalType")
		}

		if animalType != core.TypeRabbit {
			t.Errorf("Expected TypeRabbit, got %v", animalType)
		}

		// Проверяем строковое представление
		if animalType.String() != "Rabbit" {
			t.Errorf("Expected 'Rabbit', got '%s'", animalType.String())
		}
	})
}

// TestQueries тестирует системы запросов
func TestQueries(t *testing.T) {
	world := core.NewWorld(100, 100, 42)
	defer world.Clear()

	// Создаем несколько сущностей с разными компонентами
	rabbit1 := world.CreateEntity()
	world.AddPosition(rabbit1, core.Position{X: 10, Y: 10})
	world.AddVelocity(rabbit1, core.Velocity{X: 1, Y: 0})
	world.AddAnimalType(rabbit1, core.TypeRabbit)
	world.AddHealth(rabbit1, core.Health{Current: 50, Max: 100})

	rabbit2 := world.CreateEntity()
	world.AddPosition(rabbit2, core.Position{X: 20, Y: 20})
	world.AddAnimalType(rabbit2, core.TypeRabbit)

	wolf := world.CreateEntity()
	world.AddPosition(wolf, core.Position{X: 30, Y: 30})
	world.AddVelocity(wolf, core.Velocity{X: -1, Y: -1})
	world.AddAnimalType(wolf, core.TypeWolf)

	t.Run("ForEachWith", func(t *testing.T) {
		// Считаем сущности с позицией
		count := 0
		world.ForEachWith(core.MaskPosition, func(entity core.EntityID) {
			count++
		})

		if count != 3 {
			t.Errorf("Expected 3 entities with Position, got %d", count)
		}

		// Считаем сущности с позицией И скоростью
		count = 0
		world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
			count++
		})

		if count != 2 {
			t.Errorf("Expected 2 entities with Position and Velocity, got %d", count)
		}
	})

	t.Run("QueryEntitiesWith", func(t *testing.T) {
		entities := world.QueryEntitiesWith(core.MaskPosition)

		if len(entities) != 3 {
			t.Errorf("Expected 3 entities, got %d", len(entities))
		}

		// Проверяем что все возвращенные сущности действительно имеют позицию
		for _, entity := range entities {
			if !world.HasComponent(entity, core.MaskPosition) {
				t.Errorf("Entity %d should have Position component", entity)
			}
		}
	})

	t.Run("CountEntitiesWith", func(t *testing.T) {
		count := world.CountEntitiesWith(core.MaskAnimalType)
		if count != 3 {
			t.Errorf("Expected 3 animals, got %d", count)
		}

		count = world.CountEntitiesWith(core.MaskHealth)
		if count != 1 {
			t.Errorf("Expected 1 entity with health, got %d", count)
		}
	})

	t.Run("QueryByType", func(t *testing.T) {
		rabbits := world.QueryByType(core.TypeRabbit)
		if len(rabbits) != 2 {
			t.Errorf("Expected 2 rabbits, got %d", len(rabbits))
		}

		wolves := world.QueryByType(core.TypeWolf)
		if len(wolves) != 1 {
			t.Errorf("Expected 1 wolf, got %d", len(wolves))
		}

		grass := world.QueryByType(core.TypeGrass)
		if len(grass) != 0 {
			t.Errorf("Expected 0 grass entities, got %d", len(grass))
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := world.GetStats()

		if stats[core.TypeRabbit] != 2 {
			t.Errorf("Expected 2 rabbits in stats, got %d", stats[core.TypeRabbit])
		}

		if stats[core.TypeWolf] != 1 {
			t.Errorf("Expected 1 wolf in stats, got %d", stats[core.TypeWolf])
		}

		if stats[core.TypeGrass] != 0 {
			t.Errorf("Expected 0 grass in stats, got %d", stats[core.TypeGrass])
		}
	})
}

// TestSpatialIntegration тестирует интеграцию с пространственной системой
func TestSpatialIntegration(t *testing.T) {
	world := core.NewWorld(100, 100, 42)
	defer world.Clear()

	// Создаем сущности с позицией и размером
	entity1 := world.CreateEntity()
	world.AddPosition(entity1, core.Position{X: 10, Y: 10})
	world.AddSize(entity1, core.Size{Radius: 5})

	entity2 := world.CreateEntity()
	world.AddPosition(entity2, core.Position{X: 20, Y: 20})
	world.AddSize(entity2, core.Size{Radius: 3})

	entity3 := world.CreateEntity()
	world.AddPosition(entity3, core.Position{X: 80, Y: 80})
	world.AddSize(entity3, core.Size{Radius: 2})

	t.Run("QueryInRadius", func(t *testing.T) {
		// Поиск в радиусе 15 от точки (10, 10)
		nearby := world.QueryInRadius(10, 10, 15)

		// Должны найти entity1 и entity2, но не entity3
		if len(nearby) < 1 || len(nearby) > 2 {
			t.Errorf("Expected 1-2 entities in radius, got %d", len(nearby))
		}

		// Проверяем что entity1 точно есть
		found := false
		for _, entity := range nearby {
			if entity == entity1 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Should find entity1 in radius")
		}
	})

	t.Run("FindNearestAnimal", func(t *testing.T) {
		// Добавляем типы животных
		world.AddAnimalType(entity1, core.TypeRabbit)
		world.AddAnimalType(entity2, core.TypeRabbit)
		world.AddAnimalType(entity3, core.TypeWolf)

		// Ищем ближайшее животное к точке (12, 12)
		nearest, found := world.FindNearestAnimal(12, 12, 50)

		if !found {
			t.Error("Should find nearest animal")
		}

		// Ближайшим должен быть entity1 (расстояние ~2.8)
		if nearest != entity1 {
			t.Errorf("Expected entity1 as nearest, got %d", nearest)
		}
	})

	t.Run("FindNearestByType", func(t *testing.T) {
		// Ищем ближайшего волка к точке (10, 10)
		nearest, found := world.FindNearestByType(10, 10, 100, core.TypeWolf)

		if !found {
			t.Error("Should find nearest wolf")
		}

		if nearest != entity3 {
			t.Errorf("Expected entity3 as nearest wolf, got %d", nearest)
		}

		// Ищем ближайшего зайца
		nearest, found = world.FindNearestByType(10, 10, 100, core.TypeRabbit)

		if !found {
			t.Error("Should find nearest rabbit")
		}

		if nearest != entity1 {
			t.Errorf("Expected entity1 as nearest rabbit, got %d", nearest)
		}
	})
}

// TestDestroyEntityCleanup тестирует что уничтожение сущности очищает все данные
func TestDestroyEntityCleanup(t *testing.T) {
	world := core.NewWorld(100, 100, 42)
	defer world.Clear()

	entity := world.CreateEntity()

	// Добавляем все возможные компоненты
	world.AddPosition(entity, core.Position{X: 10, Y: 10})
	world.AddVelocity(entity, core.Velocity{X: 1, Y: 1})
	world.AddHealth(entity, core.Health{Current: 100, Max: 100})
	world.AddHunger(entity, core.Hunger{Value: 50})
	world.AddAge(entity, core.Age{Seconds: 10})
	world.AddAnimalType(entity, core.TypeRabbit)
	world.AddSize(entity, core.Size{Radius: 5})
	world.AddSpeed(entity, core.Speed{Base: 20, Current: 15})

	// Проверяем что все компоненты есть
	if !world.HasComponents(entity, core.MaskPosition|core.MaskVelocity|core.MaskHealth|core.MaskHunger|core.MaskAge|core.MaskAnimalType|core.MaskSize|core.MaskSpeed) {
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
