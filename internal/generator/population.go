package generator

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

const (
	// Константы для детерминированной генерации
	PopulationSeedOffset = 1000 // Смещение seed для популяций относительно terrain

	// Константы размещения (в пикселях, привязаны к TILE_SIZE)
	TileSizePixels       = float32(physics.TileSize) // 32.0 пикселя на тайл
	MaxAttemptsPlacement = 1000                      // Максимум попыток размещения
	MarginPixels         = 2 * TileSizePixels        // Отступ от краёв карты (2 тайла)

	// Группировка животных
	MaxGroupRadiusPixels = 2 * TileSizePixels // Радиус группы зайцев (2 тайла)
)

// PopulationGenerator генерирует размещение животных на карте
type PopulationGenerator struct {
	config  *config.Config
	terrain *Terrain
	rng     *rand.Rand
}

// AnimalPlacement содержит информацию о размещённом животном
type AnimalPlacement struct {
	Type core.AnimalType
	X, Y float32
}

// NewPopulationGenerator создаёт новый генератор популяций
func NewPopulationGenerator(cfg *config.Config, terrain *Terrain) *PopulationGenerator {
	// Создаём отдельный источник случайности для размещения животных
	// Используем отдельный seed для детерминированности
	source := rand.NewSource(cfg.World.Seed + PopulationSeedOffset) // Смещение для детерминированности
	rng := rand.New(source)

	return &PopulationGenerator{
		config:  cfg,
		terrain: terrain,
		rng:     rng,
	}
}

// Generate генерирует позиции для размещения животных согласно конфигурации
func (pg *PopulationGenerator) Generate() []AnimalPlacement {
	var placements []AnimalPlacement

	// Размещаем зайцев группами
	rabbitPlacements := pg.placeRabbits()
	placements = append(placements, rabbitPlacements...)

	// Размещаем волков поодиночке
	wolfPlacements := pg.placeWolves()
	placements = append(placements, wolfPlacements...)

	return placements
}

// placeRabbits размещает зайцев группами по 2-4 особи
func (pg *PopulationGenerator) placeRabbits() []AnimalPlacement {
	var placements []AnimalPlacement

	totalRabbits := pg.config.Population.Rabbits
	groupSize := pg.config.Population.RabbitGroupSize
	placedRabbits := 0

	for placedRabbits < totalRabbits {
		// Определяем размер текущей группы
		remainingRabbits := totalRabbits - placedRabbits
		currentGroupSize := groupSize
		if remainingRabbits < groupSize {
			currentGroupSize = remainingRabbits
		}

		// Находим место для группы
		groupCenterX, groupCenterY, found := pg.findSuitableLocation(nil, 0)
		if !found {
			break // Не удалось найти место
		}

		// Размещаем зайцев в группе
		for i := 0; i < currentGroupSize; i++ {
			// Случайное смещение в радиусе группы от центра
			angle := pg.rng.Float64() * 2 * math.Pi
			radius := pg.rng.Float64() * float64(MaxGroupRadiusPixels)

			x := groupCenterX + float32(radius*math.Cos(angle))
			y := groupCenterY + float32(radius*math.Sin(angle))

			// Проверяем что позиция валидна
			tileX := int(x / TileSizePixels)
			tileY := int(y / TileSizePixels)

			if !pg.terrain.IsPassable(tileX, tileY) {
				// Если место неподходящее, размещаем в центре группы
				x = groupCenterX
				y = groupCenterY
			}

			// Добавляем позицию зайца
			placements = append(placements, AnimalPlacement{
				Type: core.TypeRabbit,
				X:    x,
				Y:    y,
			})
			placedRabbits++
		}
	}

	return placements
}

// placeWolves размещает волков поодиночке с минимальной дистанцией
func (pg *PopulationGenerator) placeWolves() []AnimalPlacement {
	var placements []AnimalPlacement
	var wolfPositions []struct{ x, y float32 }

	totalWolves := pg.config.Population.Wolves
	minDistance := float32(pg.config.Population.MinWolfDistance) * TileSizePixels // Конвертируем в пиксели

	for placedWolves := 0; placedWolves < totalWolves; placedWolves++ {
		// Находим место для волка с учётом минимальной дистанции
		x, y, found := pg.findSuitableLocation(wolfPositions, minDistance)
		if !found {
			break // Не удалось найти подходящее место
		}

		// Добавляем позицию волка
		placements = append(placements, AnimalPlacement{
			Type: core.TypeWolf,
			X:    x,
			Y:    y,
		})

		// Запоминаем позицию волка
		wolfPositions = append(wolfPositions, struct{ x, y float32 }{x, y})
	}

	return placements
}

// findSuitableLocation ищет подходящее место для размещения животного
func (pg *PopulationGenerator) findSuitableLocation(
	existingPositions []struct{ x, y float32 }, minDistance float32,
) (x, y float32, found bool) {
	// ИСПРАВЛЕНИЕ: Используем реальные размеры карты (Width x Height) вместо квадратного Size
	worldWidthPixels := float32(pg.terrain.Width) * TileSizePixels   // Ширина мира в пикселях
	worldHeightPixels := float32(pg.terrain.Height) * TileSizePixels // Высота мира в пикселях
	margin := MarginPixels                                           // Отступ от краёв карты

	for attempts := 0; attempts < MaxAttemptsPlacement; attempts++ {
		// Случайная позиция с отступом от краёв - используем правильные границы для прямоугольной карты
		x := margin + pg.rng.Float32()*(worldWidthPixels-2*margin)
		y := margin + pg.rng.Float32()*(worldHeightPixels-2*margin)

		// DEBUG: Проверяем границы
		if attempts == 0 {
			fmt.Printf("DEBUG PopulationGenerator: terrain %dx%d, worldPixels %.1fx%.1f, margin %.1f\n",
				pg.terrain.Width, pg.terrain.Height, worldWidthPixels, worldHeightPixels, margin)
			fmt.Printf("DEBUG PopulationGenerator: x range [%.1f, %.1f], y range [%.1f, %.1f]\n",
				margin, worldWidthPixels-margin, margin, worldHeightPixels-margin)
		}

		// Проверяем что тайл проходим
		tileX := int(x / TileSizePixels)
		tileY := int(y / TileSizePixels)

		if !pg.terrain.IsPassable(tileX, tileY) {
			continue
		}

		// Проверяем минимальную дистанцию до существующих позиций
		if minDistance > 0 {
			tooClose := false
			for _, pos := range existingPositions {
				dx := x - pos.x
				dy := y - pos.y
				distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

				if distance < minDistance {
					tooClose = true
					break
				}
			}

			if tooClose {
				continue
			}
		}

		return x, y, true
	}

	return 0, 0, false // Не удалось найти подходящее место
}

// ValidatePlacement проверяет корректность размещения животных
func (pg *PopulationGenerator) ValidatePlacement(placements []AnimalPlacement) []string {
	var errors []string

	// Проверяем что все животные на проходимых тайлах
	for _, placement := range placements {
		tileX := int(placement.X / TileSizePixels)
		tileY := int(placement.Y / TileSizePixels)

		if !pg.terrain.IsPassable(tileX, tileY) {
			errors = append(errors, "Animal placed on impassable tile")
		}
	}

	// Проверяем минимальные расстояния между волками
	minWolfDistance := float32(pg.config.Population.MinWolfDistance) * TileSizePixels
	wolves := make([]AnimalPlacement, 0)

	for _, placement := range placements {
		if placement.Type == core.TypeWolf {
			wolves = append(wolves, placement)
		}
	}

	for i := 0; i < len(wolves); i++ {
		for j := i + 1; j < len(wolves); j++ {
			dx := wolves[i].X - wolves[j].X
			dy := wolves[i].Y - wolves[j].Y
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

			if distance < minWolfDistance {
				errors = append(errors, "Wolves placed too close to each other")
			}
		}
	}

	return errors
}

// GetStats возвращает статистику размещения животных (рефакторинг: снижена когнитивная сложность)
func (pg *PopulationGenerator) GetStats(placements []AnimalPlacement) map[string]interface{} {
	stats := make(map[string]interface{})

	rabbits, wolves := pg.countAnimalsByType(placements)

	stats["total_animals"] = len(placements)
	stats["rabbits"] = rabbits
	stats["wolves"] = wolves
	stats["rabbit_groups"] = pg.calculateRabbitGroups(rabbits)

	// Добавляем статистику расстояний между волками
	if avgDistance, hasDistance := pg.calculateAverageWolfDistance(placements, wolves); hasDistance {
		stats["average_wolf_distance"] = avgDistance
	}

	return stats
}

// countAnimalsByType подсчитывает количество животных по типам (helper-функция)
func (pg *PopulationGenerator) countAnimalsByType(placements []AnimalPlacement) (rabbits, wolves int) {
	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			rabbits++
		case core.TypeWolf:
			wolves++
		}
	}
	return rabbits, wolves
}

// calculateRabbitGroups вычисляет количество групп зайцев (helper-функция)
func (pg *PopulationGenerator) calculateRabbitGroups(rabbits int) int {
	// Округление вверх
	return (rabbits + pg.config.Population.RabbitGroupSize - 1) / pg.config.Population.RabbitGroupSize
}

// calculateAverageWolfDistance вычисляет среднее расстояние между волками (helper-функция)
func (pg *PopulationGenerator) calculateAverageWolfDistance(placements []AnimalPlacement, wolves int) (float32, bool) {
	if wolves <= 1 {
		return 0, false
	}

	wolfPlacements := pg.filterWolfPlacements(placements)
	totalDistance, pairs := pg.calculatePairwiseDistances(wolfPlacements)

	if pairs > 0 {
		return totalDistance / float32(pairs), true
	}
	return 0, false
}

// filterWolfPlacements фильтрует размещения волков (helper-функция)
func (pg *PopulationGenerator) filterWolfPlacements(placements []AnimalPlacement) []AnimalPlacement {
	wolfPlacements := make([]AnimalPlacement, 0)
	for _, placement := range placements {
		if placement.Type == core.TypeWolf {
			wolfPlacements = append(wolfPlacements, placement)
		}
	}
	return wolfPlacements
}

// calculatePairwiseDistances вычисляет суммарное расстояние между всеми парами волков (helper-функция)
func (pg *PopulationGenerator) calculatePairwiseDistances(
	wolfPlacements []AnimalPlacement,
) (totalDistance float32, pairs int) {
	for i := 0; i < len(wolfPlacements); i++ {
		for j := i + 1; j < len(wolfPlacements); j++ {
			dx := wolfPlacements[i].X - wolfPlacements[j].X
			dy := wolfPlacements[i].Y - wolfPlacements[j].Y
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			totalDistance += distance
			pairs++
		}
	}
	return totalDistance, pairs
}
