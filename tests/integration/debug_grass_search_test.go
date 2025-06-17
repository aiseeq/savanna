package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestGrassSearchMethods сравнивает результаты GetGrassAt и FindNearestGrass
func TestGrassSearchMethods(t *testing.T) {
	t.Parallel()

	t.Logf("=== СРАВНЕНИЕ МЕТОДОВ ПОИСКА ТРАВЫ ===")

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Устанавливаем траву в центр
	centerX, centerY := 25, 25
	terrain.SetGrassAmount(centerX, centerY, 100.0)

	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Позиция зайца в центре тайла (25, 25)
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16)

	t.Logf("ПОЗИЦИЯ И ТРАВА:")
	t.Logf("  Позиция зайца: (%.1f, %.1f)", rabbitX, rabbitY)
	t.Logf("  Ожидаемый тайл: (%d, %d)", centerX, centerY)

	// Проверяем тайл зайца
	tileX := int(rabbitX / 32)
	tileY := int(rabbitY / 32)
	grassInTerrain := terrain.GetGrassAmount(tileX, tileY)
	t.Logf("  Реальный тайл зайца: (%d, %d)", tileX, tileY)
	t.Logf("  Трава в terrain тайле: %.1f", grassInTerrain)

	// МЕТОД 1: GetGrassAt (используется в FeedingSystem)
	grassViaGetGrassAt := vegetationSystem.GetGrassAt(rabbitX, rabbitY)
	t.Logf("\nМЕТОД 1 - GetGrassAt (FeedingSystem):")
	t.Logf("  Результат: %.1f", grassViaGetGrassAt)

	// МЕТОД 2: FindNearestGrass (используется в BehaviorStrategy)
	minGrassToFind := float32(simulation.MinGrassToFind)
	searchRadius := float32(16.0) // Радиус зайца (было simulation.RabbitBaseRadius)
	grassX, grassY, foundGrass := vegetationSystem.FindNearestGrass(rabbitX, rabbitY, searchRadius, minGrassToFind)

	t.Logf("\nМЕТОД 2 - FindNearestGrass (BehaviorStrategy):")
	t.Logf("  Параметры поиска: радиус=%.1f, минимум=%.1f", searchRadius, minGrassToFind)
	t.Logf("  Результат: найдено=%v", foundGrass)
	if foundGrass {
		t.Logf("  Координаты найденной травы: (%.1f, %.1f)", grassX, grassY)
		grassTileX := int(grassX / 32)
		grassTileY := int(grassY / 32)
		t.Logf("  Тайл найденной травы: (%d, %d)", grassTileX, grassTileY)
	}

	// СРАВНЕНИЕ
	t.Logf("\nСРАВНЕНИЕ:")
	getGrassAtWorks := grassViaGetGrassAt >= minGrassToFind
	findNearestGrassWorks := foundGrass

	t.Logf("  GetGrassAt работает: %v (%.1f >= %.1f)", getGrassAtWorks, grassViaGetGrassAt, minGrassToFind)
	t.Logf("  FindNearestGrass работает: %v", findNearestGrassWorks)

	if getGrassAtWorks && !findNearestGrassWorks {
		t.Errorf("❌ НЕСООТВЕТСТВИЕ: GetGrassAt находит траву, а FindNearestGrass НЕТ!")
		t.Errorf("   Это объясняет почему FeedingSystem создаёт EatingState,")
		t.Errorf("   а BehaviorStrategy не останавливает зайца для еды")
	} else if getGrassAtWorks && findNearestGrassWorks {
		t.Logf("✅ ОБА МЕТОДА РАБОТАЮТ ПРАВИЛЬНО")
	} else if !getGrassAtWorks && !findNearestGrassWorks {
		t.Logf("⚠️  ОБА МЕТОДА НЕ НАХОДЯТ ТРАВУ")
	} else {
		t.Errorf("❌ ОБРАТНОЕ НЕСООТВЕТСТВИЕ: FindNearestGrass находит, а GetGrassAt НЕТ!")
	}
}
