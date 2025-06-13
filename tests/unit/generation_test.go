package unit

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

// TestTerrainGeneration проверяет генерацию ландшафта
func TestTerrainGeneration(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20 // Маленький мир для тестов
	cfg.Terrain.WaterBodies = 2
	cfg.Terrain.BushClusters = 3

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Проверяем базовые свойства
	if terrain.Size != cfg.World.Size {
		t.Errorf("Expected terrain size %d, got %d", cfg.World.Size, terrain.Size)
	}

	// Проверяем что массивы правильного размера
	if len(terrain.Tiles) != cfg.World.Size {
		t.Errorf("Expected %d rows in tiles array, got %d", cfg.World.Size, len(terrain.Tiles))
	}

	if len(terrain.Grass) != cfg.World.Size {
		t.Errorf("Expected %d rows in grass array, got %d", cfg.World.Size, len(terrain.Grass))
	}

	// Проверяем что есть вода и кусты
	stats := terrain.GetStats()
	waterTiles := stats["water_tiles"].(int)
	bushTiles := stats["bush_tiles"].(int)

	if waterTiles == 0 {
		t.Error("Expected some water tiles, got 0")
	}

	if bushTiles == 0 {
		t.Error("Expected some bush tiles, got 0")
	}

	// Проверяем что на воде нет травы
	for y := 0; y < terrain.Size; y++ {
		for x := 0; x < terrain.Size; x++ {
			if terrain.GetTileType(x, y) == generator.TileWater {
				grass := terrain.GetGrassAmount(x, y)
				if grass != 0 {
					t.Errorf("Water tile at (%d,%d) has grass: %f", x, y, grass)
				}
			}
		}
	}
}

// TestDeterministicTerrainGeneration проверяет детерминированность генерации ландшафта
func TestDeterministicTerrainGeneration(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 15
	cfg.World.Seed = 12345

	// Генерируем карту 10 раз с одним seed
	var terrains []*generator.Terrain
	for i := 0; i < 10; i++ {
		terrainGen := generator.NewTerrainGenerator(cfg)
		terrain := terrainGen.Generate()
		terrains = append(terrains, terrain)
	}

	// Проверяем что все карты идентичны
	firstTerrain := terrains[0]
	for i := 1; i < len(terrains); i++ {
		if !terrainsEqual(firstTerrain, terrains[i]) {
			t.Errorf("Terrain %d differs from first terrain", i)
		}
	}

	// Проверяем что с другим seed результат отличается
	cfg.World.Seed = 54321
	terrainGen2 := generator.NewTerrainGenerator(cfg)
	differentTerrain := terrainGen2.Generate()

	if terrainsEqual(firstTerrain, differentTerrain) {
		t.Error("Different seeds produced identical terrains")
	}
}

// TestPopulationGeneration проверяет размещение животных
func TestPopulationGeneration(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20
	cfg.Population.Rabbits = 10
	cfg.Population.Wolves = 2
	cfg.Population.RabbitGroupSize = 3
	cfg.Population.MinWolfDistance = 5

	// Генерируем ландшафт
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Генерируем популяцию
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	// Проверяем количество животных
	rabbits := 0
	wolves := 0
	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			rabbits++
		case core.TypeWolf:
			wolves++
		}
	}

	if rabbits != cfg.Population.Rabbits {
		t.Errorf("Expected %d rabbits, got %d", cfg.Population.Rabbits, rabbits)
	}

	if wolves != cfg.Population.Wolves {
		t.Errorf("Expected %d wolves, got %d", cfg.Population.Wolves, wolves)
	}

	// Проверяем валидность размещения
	errors := popGen.ValidatePlacement(placements)
	if len(errors) > 0 {
		t.Errorf("Placement validation failed: %v", errors)
	}
}

// TestAnimalPlacementValidation проверяет что животные не размещаются на воде/кустах
func TestAnimalPlacementValidation(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 15
	cfg.Terrain.WaterBodies = 5  // Много воды
	cfg.Terrain.BushClusters = 5 // Много кустов
	cfg.Population.Rabbits = 20
	cfg.Population.Wolves = 3

	// Генерируем ландшафт с большим количеством препятствий
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Генерируем популяцию
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	// Проверяем что все животные на проходимых тайлах
	for _, placement := range placements {
		tileX := int(placement.X / 32.0)
		tileY := int(placement.Y / 32.0)

		if !terrain.IsPassable(tileX, tileY) {
			t.Errorf("Animal placed on impassable tile at (%d,%d), tile type: %d",
				tileX, tileY, terrain.GetTileType(tileX, tileY))
		}
	}
}

// TestWolfMinimumDistance проверяет минимальную дистанцию между волками
func TestWolfMinimumDistance(t *testing.T) {
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 30        // Большой мир
	cfg.Population.Rabbits = 5 // Мало зайцев
	cfg.Population.Wolves = 4  // Несколько волков
	cfg.Population.MinWolfDistance = 10

	// Генерируем ландшафт
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Генерируем популяцию
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	// Находим всех волков
	var wolfPlacements []generator.AnimalPlacement
	for _, placement := range placements {
		if placement.Type == core.TypeWolf {
			wolfPlacements = append(wolfPlacements, placement)
		}
	}

	// Проверяем дистанции между волками
	minDistancePixels := float32(cfg.Population.MinWolfDistance) * 32.0

	for i := 0; i < len(wolfPlacements); i++ {
		for j := i + 1; j < len(wolfPlacements); j++ {
			dx := wolfPlacements[i].X - wolfPlacements[j].X
			dy := wolfPlacements[i].Y - wolfPlacements[j].Y
			distance := float32(dx*dx + dy*dy) // Квадрат расстояния для скорости

			if distance < minDistancePixels*minDistancePixels {
				t.Errorf("Wolves too close: distance %f, minimum required %f",
					distance, minDistancePixels*minDistancePixels)
			}
		}
	}
}

// TestConfigValidation проверяет валидацию конфигурации
func TestConfigValidation(t *testing.T) {
	// Тест корректной конфигурации
	validConfig := config.LoadDefaultConfig()
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}

	// Тест некорректного размера мира
	invalidConfig := config.LoadDefaultConfig()
	invalidConfig.World.Size = 5 // Слишком мало
	if err := invalidConfig.Validate(); err == nil {
		t.Error("Invalid world size should fail validation")
	}

	// Тест некорректных радиусов воды
	invalidConfig2 := config.LoadDefaultConfig()
	invalidConfig2.Terrain.WaterRadiusMin = 10
	invalidConfig2.Terrain.WaterRadiusMax = 5 // Мин > макс
	if err := invalidConfig2.Validate(); err == nil {
		t.Error("Invalid water radius should fail validation")
	}

	// Тест отрицательных популяций
	invalidConfig3 := config.LoadDefaultConfig()
	invalidConfig3.Population.Rabbits = -5
	if err := invalidConfig3.Validate(); err == nil {
		t.Error("Negative population should fail validation")
	}
}

// terrainsEqual проверяет идентичность двух карт
func terrainsEqual(t1, t2 *generator.Terrain) bool {
	if t1.Size != t2.Size {
		return false
	}

	for y := 0; y < t1.Size; y++ {
		for x := 0; x < t1.Size; x++ {
			if t1.Tiles[y][x] != t2.Tiles[y][x] {
				return false
			}
			if t1.Grass[y][x] != t2.Grass[y][x] {
				return false
			}
		}
	}

	return true
}
