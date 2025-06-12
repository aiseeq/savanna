package generator

import (
	"math"
	"math/rand"

	"github.com/aiseeq/savanna/config"
)

// TileType определяет тип тайла на карте
type TileType int8

const (
	TileGrass   TileType = iota // Трава (проходимо)
	TileWater                   // Вода (непроходимо)
	TileBush                    // Куст (непроходимо)
	TileWetland                 // Влажная земля (проходимо, быстрый рост травы)
)

// Terrain представляет сгенерированную карту мира
type Terrain struct {
	Size  int          // Размер мира в тайлах
	Tiles [][]TileType // Типы тайлов [y][x]
	Grass [][]float32  // Количество травы [y][x] (0-100)
}

// TerrainGenerator генерирует детерминированные карты
type TerrainGenerator struct {
	config *config.Config
	rng    *rand.Rand
}

// NewTerrainGenerator создаёт новый генератор ландшафта
func NewTerrainGenerator(cfg *config.Config) *TerrainGenerator {
	// Создаём отдельный источник случайности для генерации ландшафта
	source := rand.NewSource(cfg.World.Seed)
	rng := rand.New(source)

	return &TerrainGenerator{
		config: cfg,
		rng:    rng,
	}
}

// Generate создаёт новую карту согласно конфигурации
func (tg *TerrainGenerator) Generate() *Terrain {
	size := tg.config.World.Size

	terrain := &Terrain{
		Size:  size,
		Tiles: make([][]TileType, size),
		Grass: make([][]float32, size),
	}

	// Инициализируем массивы
	for y := 0; y < size; y++ {
		terrain.Tiles[y] = make([]TileType, size)
		terrain.Grass[y] = make([]float32, size)
	}

	// Генерируем в фиксированном порядке для детерминированности
	tg.generateBaseLayer(terrain)
	tg.generateWaterBodies(terrain)
	tg.generateWetlands(terrain)
	tg.generateBushClusters(terrain)
	tg.generateInitialGrass(terrain)

	return terrain
}

// generateBaseLayer заполняет всю карту травой
func (tg *TerrainGenerator) generateBaseLayer(terrain *Terrain) {
	for y := 0; y < terrain.Size; y++ {
		for x := 0; x < terrain.Size; x++ {
			terrain.Tiles[y][x] = TileGrass
		}
	}
}

// generateWaterBodies создаёт круглые озёра
func (tg *TerrainGenerator) generateWaterBodies(terrain *Terrain) {
	waterBodies := tg.config.Terrain.WaterBodies
	minRadius := float32(tg.config.Terrain.WaterRadiusMin)
	maxRadius := float32(tg.config.Terrain.WaterRadiusMax)

	for i := 0; i < waterBodies; i++ {
		// Случайная позиция с отступом от краёв
		margin := int(maxRadius) + 2
		x := margin + tg.rng.Intn(terrain.Size-2*margin)
		y := margin + tg.rng.Intn(terrain.Size-2*margin)

		// Случайный радиус
		radius := minRadius + tg.rng.Float32()*(maxRadius-minRadius)

		// Создаём круглое озеро
		tg.createCircularWater(terrain, x, y, radius)
	}
}

// createCircularWater создаёт круглое озеро в указанной точке
func (tg *TerrainGenerator) createCircularWater(terrain *Terrain, centerX, centerY int, radius float32) {
	radiusInt := int(radius)

	for dy := -radiusInt; dy <= radiusInt; dy++ {
		for dx := -radiusInt; dx <= radiusInt; dx++ {
			x := centerX + dx
			y := centerY + dy

			// Проверяем границы
			if x < 0 || x >= terrain.Size || y < 0 || y >= terrain.Size {
				continue
			}

			// Проверяем расстояние до центра
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if distance <= radius {
				terrain.Tiles[y][x] = TileWater
				terrain.Grass[y][x] = 0 // На воде трава не растёт
			}
		}
	}
}

// generateWetlands создаёт влажную землю вокруг водоёмов
func (tg *TerrainGenerator) generateWetlands(terrain *Terrain) {
	// Создаём копию для проверки исходного состояния
	for y := 0; y < terrain.Size; y++ {
		for x := 0; x < terrain.Size; x++ {
			// Если это трава рядом с водой - делаем влажной землёй
			if terrain.Tiles[y][x] == TileGrass && tg.isNearWater(terrain, x, y) {
				terrain.Tiles[y][x] = TileWetland
			}
		}
	}
}

// isNearWater проверяет есть ли вода в радиусе 1 тайла
func (tg *TerrainGenerator) isNearWater(terrain *Terrain, x, y int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue // Пропускаем центральную клетку
			}

			checkX := x + dx
			checkY := y + dy

			if checkX >= 0 && checkX < terrain.Size &&
				checkY >= 0 && checkY < terrain.Size {
				if terrain.Tiles[checkY][checkX] == TileWater {
					return true
				}
			}
		}
	}
	return false
}

// generateBushClusters создаёт группы кустов
func (tg *TerrainGenerator) generateBushClusters(terrain *Terrain) {
	clusters := tg.config.Terrain.BushClusters
	bushesPerCluster := tg.config.Terrain.BushPerCluster

	for i := 0; i < clusters; i++ {
		// Находим подходящее место для кластера
		var centerX, centerY int
		attempts := 0
		for attempts < 50 { // Ограничиваем количество попыток
			centerX = 5 + tg.rng.Intn(terrain.Size-10) // Отступ от краёв
			centerY = 5 + tg.rng.Intn(terrain.Size-10)

			// Проверяем что центр на траве или влажной земле
			if terrain.Tiles[centerY][centerX] == TileGrass ||
				terrain.Tiles[centerY][centerX] == TileWetland {
				break
			}
			attempts++
		}

		if attempts >= 50 {
			continue // Не удалось найти подходящее место
		}

		// Размещаем кусты в радиусе 2-3 тайлов от центра
		tg.createBushCluster(terrain, centerX, centerY, bushesPerCluster)
	}
}

// createBushCluster создаёт группу кустов вокруг центральной точки
func (tg *TerrainGenerator) createBushCluster(terrain *Terrain, centerX, centerY, count int) {
	placedBushes := 0
	maxRadius := 3

	for attempts := 0; attempts < count*3 && placedBushes < count; attempts++ {
		// Случайная позиция в радиусе
		angle := tg.rng.Float64() * 2 * math.Pi
		radius := tg.rng.Float64() * float64(maxRadius)

		x := centerX + int(radius*math.Cos(angle))
		y := centerY + int(radius*math.Sin(angle))

		// Проверяем границы и возможность размещения
		if x >= 0 && x < terrain.Size && y >= 0 && y < terrain.Size {
			if terrain.Tiles[y][x] == TileGrass || terrain.Tiles[y][x] == TileWetland {
				terrain.Tiles[y][x] = TileBush
				terrain.Grass[y][x] = 0 // Кусты убивают траву
				placedBushes++
			}
		}
	}
}

// generateInitialGrass устанавливает начальное количество травы
func (tg *TerrainGenerator) generateInitialGrass(terrain *Terrain) {
	for y := 0; y < terrain.Size; y++ {
		for x := 0; x < terrain.Size; x++ {
			switch terrain.Tiles[y][x] {
			case TileGrass:
				// 70% тайлов с травой получают 50-100 единиц
				if tg.rng.Float32() < 0.7 {
					terrain.Grass[y][x] = 50 + tg.rng.Float32()*50
				} else {
					terrain.Grass[y][x] = 0
				}

			case TileWetland:
				// Влажная земля - всегда много травы
				terrain.Grass[y][x] = 80 + tg.rng.Float32()*20

			case TileWater, TileBush:
				// На воде и кустах травы нет
				terrain.Grass[y][x] = 0
			}
		}
	}
}

// IsPassable проверяет можно ли пройти через тайл
func (t *Terrain) IsPassable(x, y int) bool {
	if x < 0 || x >= t.Size || y < 0 || y >= t.Size {
		return false
	}

	tileType := t.Tiles[y][x]
	return tileType == TileGrass || tileType == TileWetland
}

// GetTileType возвращает тип тайла в указанной позиции
func (t *Terrain) GetTileType(x, y int) TileType {
	if x < 0 || x >= t.Size || y < 0 || y >= t.Size {
		return TileWater // За границами мира считаем как вода
	}
	return t.Tiles[y][x]
}

// GetGrassAmount возвращает количество травы в тайле
func (t *Terrain) GetGrassAmount(x, y int) float32 {
	if x < 0 || x >= t.Size || y < 0 || y >= t.Size {
		return 0
	}
	return t.Grass[y][x]
}

// SetGrassAmount устанавливает количество травы в тайле
func (t *Terrain) SetGrassAmount(x, y int, amount float32) {
	if x < 0 || x >= t.Size || y < 0 || y >= t.Size {
		return
	}

	// Ограничиваем количество травы
	if amount < 0 {
		amount = 0
	} else if amount > 100 {
		amount = 100
	}

	t.Grass[y][x] = amount
}

// GetStats возвращает статистику карты
func (t *Terrain) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Подсчитываем тайлы разных типов
	grassTiles := 0
	waterTiles := 0
	bushTiles := 0
	wetlandTiles := 0
	totalGrass := float32(0)

	for y := 0; y < t.Size; y++ {
		for x := 0; x < t.Size; x++ {
			switch t.Tiles[y][x] {
			case TileGrass:
				grassTiles++
			case TileWater:
				waterTiles++
			case TileBush:
				bushTiles++
			case TileWetland:
				wetlandTiles++
			}
			totalGrass += t.Grass[y][x]
		}
	}

	totalTiles := t.Size * t.Size
	stats["total_tiles"] = totalTiles
	stats["grass_tiles"] = grassTiles
	stats["water_tiles"] = waterTiles
	stats["bush_tiles"] = bushTiles
	stats["wetland_tiles"] = wetlandTiles
	stats["total_grass"] = totalGrass
	stats["average_grass"] = totalGrass / float32(totalTiles)

	return stats
}
