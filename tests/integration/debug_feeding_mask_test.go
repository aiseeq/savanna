package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestFeedingMask проверяет работу маски AnimalType в FeedingSystem
func TestFeedingMask(t *testing.T) {
	t.Parallel()

	t.Logf("=== ТЕСТ МАСКИ ANIMAL TYPE ===")

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

	// Создаём зайца
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)

	// Делаем зайца голодным
	world.SetHunger(rabbit, core.Hunger{Value: 70.0})

	t.Logf("СОЗДАННЫЙ ЗАЯЦ:")
	t.Logf("  EntityID: %d", rabbit)

	// Проверяем компоненты зайца
	hasAnimalType := world.HasComponent(rabbit, core.MaskAnimalType)
	hasHunger := world.HasComponent(rabbit, core.MaskHunger)
	hasPosition := world.HasComponent(rabbit, core.MaskPosition)

	t.Logf("КОМПОНЕНТЫ ЗАЙЦА:")
	t.Logf("  HasAnimalType: %v", hasAnimalType)
	t.Logf("  HasHunger: %v", hasHunger)
	t.Logf("  HasPosition: %v", hasPosition)

	if hasAnimalType {
		animalType, _ := world.GetAnimalType(rabbit)
		t.Logf("  AnimalType: %v", animalType)
	}

	if hasHunger {
		hunger, _ := world.GetHunger(rabbit)
		t.Logf("  Hunger: %.1f%%", hunger.Value)
	}

	// Имитируем поиск зайцев как в FeedingSystem
	t.Logf("\nПОИСК ЗАЙЦЕВ КАК В FEEDING SYSTEM:")
	var foundRabbits []core.EntityID
	entityCount := 0

	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		entityCount++
		t.Logf("  Найдена сущность %d с AnimalType", entity)

		animalType, ok := world.GetAnimalType(entity)
		if ok {
			t.Logf("    AnimalType: %v", animalType)
			if animalType == core.TypeRabbit {
				foundRabbits = append(foundRabbits, entity)
				t.Logf("    ✅ Это заяц!")
			}
		} else {
			t.Logf("    ❌ Не удалось получить AnimalType")
		}
	})

	t.Logf("\nРЕЗУЛЬТАТЫ ПОИСКА:")
	t.Logf("  Всего сущностей с AnimalType: %d", entityCount)
	t.Logf("  Найдено зайцев: %d", len(foundRabbits))

	if len(foundRabbits) == 0 {
		t.Errorf("❌ ПРОБЛЕМА: Зайцы не найдены через ForEachWith(MaskAnimalType)")
		t.Errorf("   Это объясняет почему FeedingSystem не создаёт EatingState")
	} else {
		t.Logf("✅ Зайцы найдены правильно")

		// Проверяем конкретного зайца
		foundRabbit := foundRabbits[0]
		if foundRabbit == rabbit {
			t.Logf("✅ Найден правильный заяц (ID %d)", foundRabbit)
		} else {
			t.Errorf("❌ Найден неправильный заяц: ожидали %d, получили %d", rabbit, foundRabbit)
		}
	}
}
