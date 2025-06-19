package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// TestEntityManagerCreateEntity тестирует создание сущностей
func TestEntityManagerCreateEntity(t *testing.T) {
	t.Parallel()

	em := core.NewEntityManager()
	entity1 := em.CreateEntity()
	entity2 := em.CreateEntity()

	if entity1 == core.InvalidEntity {
		t.Error("Expected valid entity ID, got InvalidEntity")
	}

	if entity2 == core.InvalidEntity {
		t.Error("Expected valid entity ID, got InvalidEntity")
	}

	if entity1 == entity2 {
		t.Error("Expected different entity IDs")
	}

	if em.Count() != 2 {
		t.Errorf("Expected 2 entities, got %d", em.Count())
	}
}

// TestEntityManagerIsAlive тестирует проверку живости сущностей
func TestEntityManagerIsAlive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entityID core.EntityID
		want     bool
	}{
		{"InvalidEntity should not be alive", core.InvalidEntity, false},
		{"Non-existent entity should not be alive", 999, false},
	}

	em := core.NewEntityManager()
	entity := em.CreateEntity()

	// Тест живой сущности
	if !em.IsAlive(entity) {
		t.Error("Entity should be alive after creation")
	}

	// Table-driven тесты для разных случаев
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := em.IsAlive(tt.entityID); got != tt.want {
				t.Errorf("IsAlive(%v) = %v, want %v", tt.entityID, got, tt.want)
			}
		})
	}
}

// TestEntityManagerDestroyEntity тестирует уничтожение сущностей
func TestEntityManagerDestroyEntity(t *testing.T) {
	t.Parallel()

	em := core.NewEntityManager()
	entity := em.CreateEntity()

	// Успешное уничтожение
	if !em.DestroyEntity(entity) {
		t.Error("DestroyEntity should return true for valid entity")
	}

	if em.IsAlive(entity) {
		t.Error("Entity should not be alive after destruction")
	}

	if em.Count() != 0 {
		t.Errorf("Expected 0 entities after destruction, got %d", em.Count())
	}

	// Попытки уничтожить несуществующие сущности
	failureCases := []struct {
		name     string
		entityID core.EntityID
	}{
		{"already destroyed entity", entity},
		{"InvalidEntity", core.InvalidEntity},
	}

	for _, tc := range failureCases {
		t.Run(tc.name, func(t *testing.T) {
			if em.DestroyEntity(tc.entityID) {
				t.Errorf("DestroyEntity should return false for %s", tc.name)
			}
		})
	}
}

// TestEntityManagerReuseIDs тестирует переиспользование ID
func TestEntityManagerReuseIDs(t *testing.T) {
	t.Parallel()

	em := core.NewEntityManager()

	// Создаём и уничтожаем сущность
	entity1 := em.CreateEntity()
	em.DestroyEntity(entity1)

	// Создаём новую сущность - должна переиспользовать ID
	entity2 := em.CreateEntity()
	if entity2 != entity1 {
		t.Error("Expected ID reuse")
	}

	if em.Count() != 1 {
		t.Errorf("Expected 1 entity, got %d", em.Count())
	}
}

// TestEntityManagerEntityLimit тестирует лимит сущностей
func TestEntityManagerEntityLimit(t *testing.T) {
	t.Parallel()

	em := core.NewEntityManager()

	// Создаём максимальное количество сущностей
	for i := 0; i < core.MaxEntities-1; i++ {
		entity := em.CreateEntity()
		if entity == core.InvalidEntity {
			t.Fatalf("Failed to create entity %d", i)
		}
	}

	// Следующая сущность должна вернуть InvalidEntity
	entity := em.CreateEntity()
	if entity != core.InvalidEntity {
		t.Error("Expected InvalidEntity when hitting limit")
	}
}
