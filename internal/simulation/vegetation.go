package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

// Константы системы растительности
const (
	GRASS_GROWTH_RATE = 0.5   // Рост травы за один Update (при 60 TPS = 30 за секунду)
	GRASS_MAX_AMOUNT  = 100.0 // Максимальное количество травы на тайле
	TILE_SIZE         = 32    // Размер тайла в пикселях
)

// VegetationSystem управляет ростом и распределением травы
type VegetationSystem struct {
	terrain   generator.TerrainInterface
	worldSize int // Размер мира в тайлах
}

// NewVegetationSystem создаёт новую систему растительности
func NewVegetationSystem(terrain generator.TerrainInterface) *VegetationSystem {
	return &VegetationSystem{
		terrain:   terrain,
		worldSize: terrain.GetSize(),
	}
}

// Update обновляет рост травы на всех тайлах
func (vs *VegetationSystem) Update(world *core.World, deltaTime float32) {
	if vs.terrain == nil {
		return
	}

	// Проходим по всем тайлам и обновляем рост травы
	for y := 0; y < vs.worldSize; y++ {
		for x := 0; x < vs.worldSize; x++ {
			vs.updateGrassTile(x, y, deltaTime)
		}
	}
}

// updateGrassTile обновляет рост травы на одном тайле
func (vs *VegetationSystem) updateGrassTile(x, y int, deltaTime float32) {
	tileType := vs.terrain.GetTileType(x, y)

	// Трава растёт только на подходящих тайлах
	canGrow := false
	switch tileType {
	case generator.TileGrass:
		canGrow = true
	case generator.TileWetland:
		canGrow = true // Влажная земля - трава растёт быстрее
	case generator.TileWater, generator.TileBush:
		canGrow = false // На воде и кустах трава не растёт
	}

	if !canGrow {
		return
	}

	currentGrass := vs.terrain.GetGrassAmount(x, y)
	if currentGrass >= GRASS_MAX_AMOUNT {
		return // Уже максимум
	}

	// Вычисляем скорость роста
	growthRate := GRASS_GROWTH_RATE * deltaTime

	// На влажной земле трава растёт в 1.5 раза быстрее
	if tileType == generator.TileWetland {
		growthRate *= 1.5
	}

	// Не растёт рядом с водой (кроме влажной земли)
	if tileType == generator.TileGrass && vs.isNearWater(x, y) {
		growthRate *= 0.3 // Замедленный рост рядом с водой
	}

	// Увеличиваем количество травы
	newAmount := currentGrass + growthRate
	if newAmount > GRASS_MAX_AMOUNT {
		newAmount = GRASS_MAX_AMOUNT
	}

	vs.terrain.SetGrassAmount(x, y, newAmount)
}

// isNearWater проверяет есть ли вода в радиусе 1 тайла (исключая влажную землю)
func (vs *VegetationSystem) isNearWater(x, y int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue // Пропускаем центральную клетку
			}

			checkX := x + dx
			checkY := y + dy

			if checkX >= 0 && checkX < vs.worldSize &&
				checkY >= 0 && checkY < vs.worldSize {
				if vs.terrain.GetTileType(checkX, checkY) == generator.TileWater {
					return true
				}
			}
		}
	}
	return false
}

// GetGrassAt возвращает количество травы в указанной позиции в пикселях
func (vs *VegetationSystem) GetGrassAt(worldX, worldY float32) float32 {
	tileX := int(worldX / TILE_SIZE)
	tileY := int(worldY / TILE_SIZE)

	if tileX < 0 || tileX >= vs.worldSize || tileY < 0 || tileY >= vs.worldSize {
		return 0
	}

	return vs.terrain.GetGrassAmount(tileX, tileY)
}

// ConsumeGrassAt поедает траву в указанной позиции и возвращает сколько было съедено
func (vs *VegetationSystem) ConsumeGrassAt(worldX, worldY, amount float32) float32 {
	tileX := int(worldX / TILE_SIZE)
	tileY := int(worldY / TILE_SIZE)

	if tileX < 0 || tileX >= vs.worldSize || tileY < 0 || tileY >= vs.worldSize {
		return 0
	}

	currentGrass := vs.terrain.GetGrassAmount(tileX, tileY)
	if currentGrass <= 0 {
		return 0
	}

	// Определяем сколько можем съесть
	consumed := amount
	if consumed > currentGrass {
		consumed = currentGrass
	}

	// Уменьшаем количество травы
	newAmount := currentGrass - consumed
	vs.terrain.SetGrassAmount(tileX, tileY, newAmount)

	return consumed
}

// CanGrassGrowAt проверяет может ли трава расти в указанной позиции
func (vs *VegetationSystem) CanGrassGrowAt(worldX, worldY float32) bool {
	tileX := int(worldX / TILE_SIZE)
	tileY := int(worldY / TILE_SIZE)

	if tileX < 0 || tileX >= vs.worldSize || tileY < 0 || tileY >= vs.worldSize {
		return false
	}

	tileType := vs.terrain.GetTileType(tileX, tileY)
	return tileType == generator.TileGrass || tileType == generator.TileWetland
}

// FindNearestGrass ищет ближайший тайл с травой (количество > minAmount)
func (vs *VegetationSystem) FindNearestGrass(worldX, worldY, searchRadius, minAmount float32) (float32, float32, bool) {
	centerTileX := int(worldX / TILE_SIZE)
	centerTileY := int(worldY / TILE_SIZE)
	searchRadiusTiles := int(searchRadius / TILE_SIZE)

	bestDistance := float32(1e9)
	var bestX, bestY float32
	found := false

	// Ищем по спирали от центра
	for radius := 1; radius <= searchRadiusTiles; radius++ {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				// Проверяем только границу текущего радиуса для оптимизации
				if abs(dx) != radius && abs(dy) != radius {
					continue
				}

				tileX := centerTileX + dx
				tileY := centerTileY + dy

				if tileX < 0 || tileX >= vs.worldSize || tileY < 0 || tileY >= vs.worldSize {
					continue
				}

				grassAmount := vs.terrain.GetGrassAmount(tileX, tileY)
				if grassAmount < minAmount {
					continue
				}

				// Проверяем что тип тайла подходит для роста травы
				if !vs.CanGrassGrowAt(float32(tileX*TILE_SIZE), float32(tileY*TILE_SIZE)) {
					continue
				}

				// Вычисляем расстояние до центра тайла
				grassWorldX := float32(tileX*TILE_SIZE + TILE_SIZE/2)
				grassWorldY := float32(tileY*TILE_SIZE + TILE_SIZE/2)

				dx_world := grassWorldX - worldX
				dy_world := grassWorldY - worldY
				distance := dx_world*dx_world + dy_world*dy_world // Квадрат расстояния для скорости

				if distance < bestDistance {
					bestDistance = distance
					bestX = grassWorldX
					bestY = grassWorldY
					found = true
				}
			}
		}

		// Если нашли траву на этом радиусе, возвращаем (ближайшую)
		if found {
			return bestX, bestY, true
		}
	}

	return 0, 0, false
}

// GetStats возвращает статистику растительности
func (vs *VegetationSystem) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalGrass := float32(0)
	tilesWithGrass := 0
	maxGrass := float32(0)
	minGrass := float32(1000)

	for y := 0; y < vs.worldSize; y++ {
		for x := 0; x < vs.worldSize; x++ {
			if vs.CanGrassGrowAt(float32(x*TILE_SIZE), float32(y*TILE_SIZE)) {
				grass := vs.terrain.GetGrassAmount(x, y)
				totalGrass += grass

				if grass > 0 {
					tilesWithGrass++
					if grass > maxGrass {
						maxGrass = grass
					}
					if grass < minGrass {
						minGrass = grass
					}
				}
			}
		}
	}

	totalTiles := vs.worldSize * vs.worldSize

	stats["total_grass"] = totalGrass
	stats["average_grass"] = totalGrass / float32(totalTiles)
	stats["tiles_with_grass"] = tilesWithGrass
	stats["max_grass"] = maxGrass
	stats["min_grass"] = minGrass
	stats["grass_coverage"] = float32(tilesWithGrass) / float32(totalTiles) * 100

	return stats
}

// abs возвращает абсолютное значение integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
