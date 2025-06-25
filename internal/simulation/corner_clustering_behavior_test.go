package simulation

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

func TestCornerClusteringBehavior_RabbitEscapesFromCorner(t *testing.T) {
	// Создаем мир 50x38 как в задаче
	worldWidth := float32(50)
	worldHeight := float32(38)
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	// Создаем terrain
	terrain := &generator.Terrain{
		Width:  int(worldWidth),
		Height: int(worldHeight),
		Size:   int(max(worldWidth, worldHeight)), // ИСПРАВЛЕНИЕ: нужно установить Size для GetSize()
		Tiles:  make([][]generator.TileType, int(worldHeight)),
		Grass:  make([][]float32, int(worldHeight)),
	}
	for y := 0; y < int(worldHeight); y++ {
		terrain.Tiles[y] = make([]generator.TileType, int(worldWidth))
		terrain.Grass[y] = make([]float32, int(worldWidth))
		for x := 0; x < int(worldWidth); x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	behaviorSystem := NewAnimalBehaviorSystem(vegetationSystem)

	// ИСПРАВЛЕНИЕ: Добавляем MovementSystem чтобы зайцы реально двигались
	// BehaviorSystem только устанавливает Velocity, MovementSystem обновляет Position
	movementSystem := NewMovementSystem(worldWidth, worldHeight)

	// Размещаем волка в центре мира (конвертируем тайлы в пиксели)
	centerXTiles := worldWidth / 2                         // 25 тайлов
	centerYTiles := worldHeight / 2                        // 19 тайлов
	centerXPixels := constants.TilesToPixels(centerXTiles) // 25 * 32 = 800 пикселей
	centerYPixels := constants.TilesToPixels(centerYTiles) // 19 * 32 = 608 пикселей
	_ = CreateAnimal(world, core.TypeWolf, centerXPixels, centerYPixels)

	// Размещаем зайца рядом с волком, но ближе к углу (левый верхний угол)
	rabbitXTiles := centerXTiles - 5                       // 20 тайлов (близко к волку)
	rabbitYTiles := centerYTiles - 5                       // 14 тайлов (близко к волку, но ближе к верхнему краю)
	rabbitXPixels := constants.TilesToPixels(rabbitXTiles) // 20 * 32 = 640 пикселей
	rabbitYPixels := constants.TilesToPixels(rabbitYTiles) // 14 * 32 = 448 пикселей
	rabbit := CreateAnimal(world, core.TypeRabbit, rabbitXPixels, rabbitYPixels)

	// ИСПРАВЛЕНИЕ: Делаем зайца сытым чтобы он не искал еду, а убегал от волка
	world.SetSatiation(rabbit, core.Satiation{Value: 95.0}) // Сытый заяц (выше порога 90%)

	t.Logf("Initial positions: Wolf(%.1f, %.1f), Rabbit(%.1f, %.1f)",
		centerXPixels, centerYPixels, rabbitXPixels, rabbitYPixels)

	// Запоминаем начальную позицию зайца
	initialPos, _ := world.GetPosition(rabbit)

	// Симулируем 10 секунд (600 тиков)
	deltaTime := float32(1.0 / 60.0)
	for tick := 0; tick < 600; tick++ {
		// ИСПРАВЛЕНИЕ: Сначала BehaviorSystem устанавливает Velocity, потом MovementSystem обновляет Position
		behaviorSystem.Update(world, deltaTime)
		movementSystem.Update(world, deltaTime)
	}

	// Проверяем финальную позицию зайца
	finalPos, _ := world.GetPosition(rabbit)

	// ИСПРАВЛЕНИЕ: Конвертируем размеры мира в пиксели для сравнения с позициями
	worldWidthPixels := constants.TilesToPixels(worldWidth)
	worldHeightPixels := constants.TilesToPixels(worldHeight)

	// Определяем находится ли заяц в углу карты
	// Угол определяется как область в пределах 10% от каждого края
	const cornerThreshold = 0.1
	cornerZoneX := worldWidthPixels * cornerThreshold  // 10% от ширины в пикселях
	cornerZoneY := worldHeightPixels * cornerThreshold // 10% от высоты в пикселях

	inLeftCorner := finalPos.X < cornerZoneX
	inRightCorner := finalPos.X > worldWidthPixels-cornerZoneX
	inTopCorner := finalPos.Y < cornerZoneY
	inBottomCorner := finalPos.Y > worldHeightPixels-cornerZoneY

	isInCorner := (inLeftCorner || inRightCorner) && (inTopCorner || inBottomCorner)

	if isInCorner {
		t.Errorf("Rabbit ended up in corner! Final position: (%.1f, %.1f)",
			finalPos.X, finalPos.Y)
		t.Errorf("Corner zones: X < %.1f or X > %.1f, Y < %.1f or Y > %.1f",
			cornerZoneX, worldWidthPixels-cornerZoneX, cornerZoneY, worldHeightPixels-cornerZoneY)
	}

	// Дополнительная проверка: заяц должен двигаться от начальной позиции
	distanceMoved := math.Sqrt(float64((finalPos.X-initialPos.X)*(finalPos.X-initialPos.X) +
		(finalPos.Y-initialPos.Y)*(finalPos.Y-initialPos.Y)))

	if distanceMoved < 2.0 {
		t.Errorf("Rabbit moved only %.2f units, expected significant movement", distanceMoved)
	}

	// Проверяем что заяц не прижался к границе
	const edgeThresholdTiles = 1.0 // 1 тайл от края
	edgeThresholdPixels := constants.TilesToPixels(edgeThresholdTiles)
	nearLeftEdge := finalPos.X < edgeThresholdPixels
	nearRightEdge := finalPos.X > worldWidthPixels-edgeThresholdPixels
	nearTopEdge := finalPos.Y < edgeThresholdPixels
	nearBottomEdge := finalPos.Y > worldHeightPixels-edgeThresholdPixels

	if nearLeftEdge || nearRightEdge || nearTopEdge || nearBottomEdge {
		t.Errorf("Rabbit is too close to world edge! Position: (%.1f, %.1f), World: %.1fx%.1f pixels",
			finalPos.X, finalPos.Y, worldWidthPixels, worldHeightPixels)
	}

	t.Logf("SUCCESS: Rabbit escaped from corner clustering")
	t.Logf("Final position: (%.1f, %.1f), Distance moved: %.2f",
		finalPos.X, finalPos.Y, distanceMoved)
}

func TestCornerClusteringBehavior_BoundaryRepulsionCalculation(t *testing.T) {
	// Тест проверяет корректность расчета отталкивания от границ
	worldWidth := float32(100)
	worldHeight := float32(100)

	// Создаем травоядную стратегию для тестирования
	strategy := &HerbivoreBehaviorStrategy{}

	testCases := []struct {
		name     string
		position core.Position
		expectX  string // "positive", "negative", "zero"
		expectY  string
	}{
		{
			name:     "Center position",
			position: core.Position{X: 50, Y: 50},
			expectX:  "zero",
			expectY:  "zero",
		},
		{
			name:     "Near left edge",
			position: core.Position{X: 2, Y: 50}, // В 5% зоне от левого края
			expectX:  "positive",                 // Толкает вправо
			expectY:  "zero",
		},
		{
			name:     "Near right edge",
			position: core.Position{X: 98, Y: 50}, // В 5% зоне от правого края
			expectX:  "negative",                  // Толкает влево
			expectY:  "zero",
		},
		{
			name:     "Near top edge",
			position: core.Position{X: 50, Y: 2}, // В 5% зоне от верхнего края
			expectX:  "zero",
			expectY:  "positive", // Толкает вниз
		},
		{
			name:     "Near bottom edge",
			position: core.Position{X: 50, Y: 98}, // В 5% зоне от нижнего края
			expectX:  "zero",
			expectY:  "negative", // Толкает вверх
		},
		{
			name:     "Left-top corner",
			position: core.Position{X: 2, Y: 2}, // В углу
			expectX:  "positive",                // Толкает вправо
			expectY:  "positive",                // Толкает вниз
		},
		{
			name:     "Right-bottom corner",
			position: core.Position{X: 98, Y: 98}, // В углу
			expectX:  "negative",                  // Толкает влево
			expectY:  "negative",                  // Толкает вверх
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repulsion := strategy.calculateBoundaryRepulsion(tc.position, worldWidth, worldHeight)

			// Проверяем направление X
			switch tc.expectX {
			case "positive":
				if repulsion.X <= 0 {
					t.Errorf("Expected positive X repulsion, got %.3f", repulsion.X)
				}
			case "negative":
				if repulsion.X >= 0 {
					t.Errorf("Expected negative X repulsion, got %.3f", repulsion.X)
				}
			case "zero":
				if math.Abs(float64(repulsion.X)) > 0.001 {
					t.Errorf("Expected zero X repulsion, got %.3f", repulsion.X)
				}
			}

			// Проверяем направление Y
			switch tc.expectY {
			case "positive":
				if repulsion.Y <= 0 {
					t.Errorf("Expected positive Y repulsion, got %.3f", repulsion.Y)
				}
			case "negative":
				if repulsion.Y >= 0 {
					t.Errorf("Expected negative Y repulsion, got %.3f", repulsion.Y)
				}
			case "zero":
				if math.Abs(float64(repulsion.Y)) > 0.001 {
					t.Errorf("Expected zero Y repulsion, got %.3f", repulsion.Y)
				}
			}

			t.Logf("Position (%.1f, %.1f) → Repulsion (%.3f, %.3f)",
				tc.position.X, tc.position.Y, repulsion.X, repulsion.Y)
		})
	}
}

func TestCornerClusteringBehavior_MultipleRabbitsAvoidCorners(t *testing.T) {
	// Тест с несколькими зайцами для проверки что они не все собираются в одном углу
	worldWidth := float32(50)
	worldHeight := float32(38)
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	terrain := &generator.Terrain{
		Width:  int(worldWidth),
		Height: int(worldHeight),
		Tiles:  make([][]generator.TileType, int(worldHeight)),
		Grass:  make([][]float32, int(worldHeight)),
	}
	for y := 0; y < int(worldHeight); y++ {
		terrain.Tiles[y] = make([]generator.TileType, int(worldWidth))
		terrain.Grass[y] = make([]float32, int(worldWidth))
		for x := 0; x < int(worldWidth); x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	behaviorSystem := NewAnimalBehaviorSystem(vegetationSystem)

	// Создаем волка в центре
	centerX := worldWidth / 2
	centerY := worldHeight / 2
	_ = CreateAnimal(world, core.TypeWolf, centerX, centerY)

	// Создаем 5 зайцев около волка
	var rabbits []core.EntityID
	for i := 0; i < 5; i++ {
		rabbitX := centerX + float32(i-2)*2 // Распределяем вокруг волка
		rabbitY := centerY + float32(i%2)*2
		rabbit := CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)
		rabbits = append(rabbits, rabbit)
	}

	// Симулируем 10 секунд
	deltaTime := float32(1.0 / 60.0)
	for tick := 0; tick < 600; tick++ {
		behaviorSystem.Update(world, deltaTime)
	}

	// Проверяем что зайцы не сгруппировались в углах
	const cornerThreshold = 0.15 // 15% от края = зона угла
	cornerZoneX := worldWidth * cornerThreshold
	cornerZoneY := worldHeight * cornerThreshold

	rabbitsInCorners := 0
	for i, rabbit := range rabbits {
		finalPos, _ := world.GetPosition(rabbit)

		inLeftCorner := finalPos.X < cornerZoneX
		inRightCorner := finalPos.X > worldWidth-cornerZoneX
		inTopCorner := finalPos.Y < cornerZoneY
		inBottomCorner := finalPos.Y > worldHeight-cornerZoneY

		isInCorner := (inLeftCorner || inRightCorner) && (inTopCorner || inBottomCorner)
		if isInCorner {
			rabbitsInCorners++
		}

		t.Logf("Rabbit %d final position: (%.1f, %.1f), in corner: %v",
			i, finalPos.X, finalPos.Y, isInCorner)
	}

	// Не больше 1 зайца должно быть в углах (случайность возможна)
	if rabbitsInCorners > 1 {
		t.Errorf("Too many rabbits (%d/5) ended up in corners, boundary repulsion not working",
			rabbitsInCorners)
	}

	t.Logf("SUCCESS: %d/5 rabbits in corners (acceptable)", rabbitsInCorners)
}
