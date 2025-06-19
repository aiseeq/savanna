package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

// Константы системы растительности
const (
	// Базовые параметры роста (выводятся от логики игры)
	TileSizeVegetation = 32    // Размер тайла в пикселях
	GrassMaxAmount     = 100.0 // Максимальное количество травы на тайле
	GrassGrowthRate    = 0.5   // Рост травы за секунду (0.5 единиц/сек)

	// Множители скорости роста для разных типов почвы
	WetlandGrowthMultiplier = 1.5 // Влажная земля - трава растёт в 1.5 раза быстрее
	NearWaterGrowthPenalty  = 0.3 // Замедленный рост рядом с водой (30% от нормального)
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
	if currentGrass >= GrassMaxAmount {
		return // Уже максимум
	}

	// Вычисляем скорость роста
	growthRate := GrassGrowthRate * deltaTime

	// На влажной земле трава растёт быстрее
	if tileType == generator.TileWetland {
		growthRate *= WetlandGrowthMultiplier
	}

	// Не растёт рядом с водой (кроме влажной земли)
	if tileType == generator.TileGrass && vs.isNearWater(x, y) {
		growthRate *= NearWaterGrowthPenalty
	}

	// Увеличиваем количество травы
	newAmount := currentGrass + growthRate
	if newAmount > GrassMaxAmount {
		newAmount = GrassMaxAmount
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
	tileX := int(worldX / TileSizeVegetation)
	tileY := int(worldY / TileSizeVegetation)

	if tileX < 0 || tileX >= vs.worldSize || tileY < 0 || tileY >= vs.worldSize {
		return 0
	}

	return vs.terrain.GetGrassAmount(tileX, tileY)
}

// ConsumeGrassAt поедает траву в указанной позиции и возвращает сколько было съедено
func (vs *VegetationSystem) ConsumeGrassAt(worldX, worldY, amount float32) float32 {
	tileX := int(worldX / TileSizeVegetation)
	tileY := int(worldY / TileSizeVegetation)

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
	tileX := int(worldX / TileSizeVegetation)
	tileY := int(worldY / TileSizeVegetation)

	if tileX < 0 || tileX >= vs.worldSize || tileY < 0 || tileY >= vs.worldSize {
		return false
	}

	tileType := vs.terrain.GetTileType(tileX, tileY)
	return tileType == generator.TileGrass || tileType == generator.TileWetland
}

// FindNearestGrass ищет ближайший тайл с травой (количество > minAmount)
// Рефакторинг: разбито на вспомогательные функции для снижения когнитивной сложности
func (vs *VegetationSystem) FindNearestGrass(
	worldX, worldY, searchRadius, minAmount float32,
) (grassX, grassY float32, found bool) {
	centerTileX := int(worldX / TileSizeVegetation)
	centerTileY := int(worldY / TileSizeVegetation)
	searchRadiusTiles := int(searchRadius / TileSizeVegetation)

	bestDistance := float32(1e9) //nolint:gomnd // Большое число для поиска минимума
	var bestX, bestY float32
	found = false

	// Ищем по спирали от центра, включая сам центральный тайл (radius = 0)
	for radius := 0; radius <= searchRadiusTiles; radius++ {
		tiles := vs.getSpiralRingTiles(centerTileX, centerTileY, radius)

		for _, tile := range tiles {
			if !vs.isValidTile(tile.x, tile.y) {
				continue
			}

			if !vs.checkGrassTile(tile.x, tile.y, minAmount) {
				continue
			}

			grassWorldX, grassWorldY := vs.tileToWorldCenter(tile.x, tile.y)
			distanceSquared := vs.calculateDistanceSquared(grassWorldX, grassWorldY, worldX, worldY)

			if distanceSquared < bestDistance {
				bestDistance = distanceSquared
				bestX = grassWorldX
				bestY = grassWorldY
				found = true
			}
		}

		// Если нашли траву на этом радиусе, возвращаем (ближайшую)
		if found {
			return bestX, bestY, true
		}
	}

	return 0, 0, false
}

// GetStats возвращает статистику растительности (рефакторинг: снижена когнитивная сложность)
func (vs *VegetationSystem) GetStats() map[string]interface{} {
	grassData := vs.collectGrassData()
	totalTiles := vs.worldSize * vs.worldSize

	return vs.buildStatsMap(grassData, totalTiles)
}

// grassStatistics содержит агрегированные данные о траве (helper-структура)
type grassStatistics struct {
	totalGrass     float32
	tilesWithGrass int
	maxGrass       float32
	minGrass       float32
}

// collectGrassData собирает данные о траве по всем тайлам (helper-функция)
func (vs *VegetationSystem) collectGrassData() grassStatistics {
	data := grassStatistics{
		totalGrass:     0,
		tilesWithGrass: 0,
		maxGrass:       0,
		minGrass:       GrassMaxAmount,
	}

	for y := 0; y < vs.worldSize; y++ {
		for x := 0; x < vs.worldSize; x++ {
			vs.processTileGrassData(x, y, &data)
		}
	}

	return data
}

// processTileGrassData обрабатывает данные одного тайла (helper-функция)
func (vs *VegetationSystem) processTileGrassData(x, y int, data *grassStatistics) {
	if !vs.CanGrassGrowAt(float32(x*TileSizeVegetation), float32(y*TileSizeVegetation)) {
		return
	}

	grass := vs.terrain.GetGrassAmount(x, y)
	data.totalGrass += grass

	if grass > 0 {
		data.tilesWithGrass++
		if grass > data.maxGrass {
			data.maxGrass = grass
		}
		if grass < data.minGrass {
			data.minGrass = grass
		}
	}
}

// buildStatsMap создаёт карту статистики из собранных данных (helper-функция)
func (vs *VegetationSystem) buildStatsMap(data grassStatistics, totalTiles int) map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_grass"] = data.totalGrass
	stats["average_grass"] = data.totalGrass / float32(totalTiles)
	stats["tiles_with_grass"] = data.tilesWithGrass
	stats["max_grass"] = data.maxGrass
	stats["min_grass"] = data.minGrass
	stats["grass_coverage"] = float32(data.tilesWithGrass) / float32(totalTiles) * 100

	return stats
}

// tileCoord представляет координаты тайла
type tileCoord struct {
	x, y int
}

// getSpiralRingTiles возвращает тайлы на границе кольца указанного радиуса
func (vs *VegetationSystem) getSpiralRingTiles(centerX, centerY, radius int) []tileCoord {
	tiles := make([]tileCoord, 0, radius*8)

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			// Проверяем только границу текущего радиуса для оптимизации
			if abs(dx) != radius && abs(dy) != radius {
				continue
			}

			tiles = append(tiles, tileCoord{x: centerX + dx, y: centerY + dy})
		}
	}

	return tiles
}

// isValidTile проверяет находится ли тайл в границах мира
func (vs *VegetationSystem) isValidTile(tileX, tileY int) bool {
	return tileX >= 0 && tileX < vs.worldSize && tileY >= 0 && tileY < vs.worldSize
}

// checkGrassTile проверяет подходит ли тайл (достаточно травы и может расти)
func (vs *VegetationSystem) checkGrassTile(tileX, tileY int, minAmount float32) bool {
	grassAmount := vs.terrain.GetGrassAmount(tileX, tileY)
	if grassAmount < minAmount {
		return false
	}

	// Проверяем что тип тайла подходит для роста травы
	return vs.CanGrassGrowAt(float32(tileX*TileSizeVegetation), float32(tileY*TileSizeVegetation))
}

// tileToWorldCenter конвертирует координаты тайла в мировые координаты центра тайла
func (vs *VegetationSystem) tileToWorldCenter(tileX, tileY int) (worldX, worldY float32) {
	worldX = float32(tileX*TileSizeVegetation + TileSizeVegetation/2)
	worldY = float32(tileY*TileSizeVegetation + TileSizeVegetation/2)
	return worldX, worldY
}

// calculateDistanceSquared вычисляет квадрат расстояния между двумя точками
func (vs *VegetationSystem) calculateDistanceSquared(x1, y1, x2, y2 float32) float32 {
	dx := x1 - x2
	dy := y1 - y2
	return dx*dx + dy*dy
}

// abs возвращает абсолютное значение integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
