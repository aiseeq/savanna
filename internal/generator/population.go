package generator

import (
	"math"
	"math/rand"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
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
	source := rand.NewSource(cfg.World.Seed + 1000) // +1000 чтобы отличаться от terrain
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
			// Случайное смещение в радиусе 2 тайлов от центра группы
			angle := pg.rng.Float64() * 2 * math.Pi
			radius := pg.rng.Float64() * 2.0 * 32.0 // 2 тайла в пикселях

			x := groupCenterX + float32(radius*math.Cos(angle))
			y := groupCenterY + float32(radius*math.Sin(angle))

			// Проверяем что позиция валидна
			tileX := int(x / 32.0)
			tileY := int(y / 32.0)

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
	minDistance := float32(pg.config.Population.MinWolfDistance) * 32.0 // Конвертируем в пиксели

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
func (pg *PopulationGenerator) findSuitableLocation(existingPositions []struct{ x, y float32 }, minDistance float32) (float32, float32, bool) {
	worldSizePixels := float32(pg.terrain.Size * 32) // Размер мира в пикселях
	margin := float32(64)                            // Отступ от краёв (2 тайла)

	for attempts := 0; attempts < 1000; attempts++ {
		// Случайная позиция с отступом от краёв
		x := margin + pg.rng.Float32()*(worldSizePixels-2*margin)
		y := margin + pg.rng.Float32()*(worldSizePixels-2*margin)

		// Проверяем что тайл проходим
		tileX := int(x / 32.0)
		tileY := int(y / 32.0)

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
		tileX := int(placement.X / 32.0)
		tileY := int(placement.Y / 32.0)

		if !pg.terrain.IsPassable(tileX, tileY) {
			errors = append(errors, "Animal placed on impassable tile")
		}
	}

	// Проверяем минимальные расстояния между волками
	minWolfDistance := float32(pg.config.Population.MinWolfDistance) * 32.0
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

// GetStats возвращает статистику размещения животных
func (pg *PopulationGenerator) GetStats(placements []AnimalPlacement) map[string]interface{} {
	stats := make(map[string]interface{})

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

	stats["total_animals"] = len(placements)
	stats["rabbits"] = rabbits
	stats["wolves"] = wolves
	stats["rabbit_groups"] = (rabbits + pg.config.Population.RabbitGroupSize - 1) / pg.config.Population.RabbitGroupSize // Округление вверх

	// Вычисляем среднее расстояние между волками
	if wolves > 1 {
		totalDistance := float32(0)
		pairs := 0

		wolfPlacements := make([]AnimalPlacement, 0)
		for _, placement := range placements {
			if placement.Type == core.TypeWolf {
				wolfPlacements = append(wolfPlacements, placement)
			}
		}

		for i := 0; i < len(wolfPlacements); i++ {
			for j := i + 1; j < len(wolfPlacements); j++ {
				dx := wolfPlacements[i].X - wolfPlacements[j].X
				dy := wolfPlacements[i].Y - wolfPlacements[j].Y
				distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
				totalDistance += distance
				pairs++
			}
		}

		if pairs > 0 {
			stats["average_wolf_distance"] = totalDistance / float32(pairs)
		}
	}

	return stats
}
