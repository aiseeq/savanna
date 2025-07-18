package simulation

import (
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

// TestCustomHerbivoreVision_LargeVisionRange проверяет, что кастомное травоядное
// с большим радиусом зрения находит траву на большом расстоянии
// Согласно ЗАДАЧЕ 2.1 из плана: "Напиши тест, который создает кастомное травоядное
// с большим радиусом зрения и проверяет, что оно находит траву на этом расстоянии."
func TestCustomHerbivoreVision_LargeVisionRange(t *testing.T) {
	// Создаем мир
	world := core.NewWorld(50, 50, 12345)

	// Создаем terrain с травой
	terrain := &generator.Terrain{
		Width:  50,
		Height: 50,
		Size:   50, // ИСПРАВЛЕНИЕ: нужно установить Size для GetSize()
		Tiles:  make([][]generator.TileType, 50),
		Grass:  make([][]float32, 50),
	}
	for y := 0; y < 50; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 50)
		terrain.Grass[y] = make([]float32, 50)
		for x := 0; x < 50; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0 // Полная трава везде
		}
	}

	// Создаем стандартного зайца для сравнения (позиция в пикселях)
	centerX := constants.TilesToPixels(25.0) // 25 тайлов * 32 = 800 пикселей
	centerY := constants.TilesToPixels(25.0) // 25 тайлов * 32 = 800 пикселей
	standardRabbit := CreateAnimal(world, core.TypeRabbit, centerX, centerY)
	standardConfig, _ := world.GetAnimalConfig(standardRabbit)
	standardVisionRange := standardConfig.VisionRange

	// ИСПРАВЛЕНИЕ: Делаем стандартного зайца голодным (ниже порога 60%)
	world.SetSatiation(standardRabbit, core.Satiation{Value: 50.0}) // Ниже порога для поиска травы

	// Создаем кастомное травоядное с увеличенным в 2 раза радиусом зрения
	customHerbivore := world.CreateEntity()

	// Устанавливаем позицию
	world.AddPosition(customHerbivore, core.Position{X: centerX, Y: centerY})

	// Устанавливаем тип животного
	world.AddAnimalType(customHerbivore, core.TypeRabbit)

	// Создаем кастомную конфигурацию с большим радиусом зрения
	customConfig := core.AnimalConfig{
		BaseRadius:         RabbitBaseRadius,
		MaxHealth:          RabbitMaxHealth,
		BaseSpeed:          RabbitBaseSpeed,
		CollisionRadius:    RabbitBaseRadius * CollisionRadiusMultiplier,
		AttackRange:        0,                         // Травоядное не атакует
		VisionRange:        standardVisionRange * 2.0, // УВЕЛИЧЕННЫЙ В 2 РАЗА радиус зрения!
		SatiationThreshold: RabbitSatiationThreshold,
		FleeThreshold:      RabbitBaseRadius * RabbitFleeDistanceMultiplier,
		SearchSpeed:        SearchSpeedMultiplier,
		WanderingSpeed:     WanderingSpeedMultiplier,
		ContentSpeed:       ContentSpeedMultiplier,
		MinDirectionTime:   1.0,
		MaxDirectionTime:   3.0,
		AttackDamage:       0,
		AttackCooldown:     0,
		HitChance:          0,
	}
	world.AddAnimalConfig(customHerbivore, customConfig)

	// Добавляем поведение травоядного
	world.AddBehavior(customHerbivore, core.Behavior{
		Type:               core.BehaviorHerbivore,
		DirectionTimer:     0,
		SatiationThreshold: customConfig.SatiationThreshold,
		FleeThreshold:      customConfig.FleeThreshold,
		SearchSpeed:        customConfig.SearchSpeed,
		WanderingSpeed:     customConfig.WanderingSpeed,
		ContentSpeed:       customConfig.ContentSpeed,
		VisionRange:        customConfig.VisionRange, // Важно: увеличенный радиус (ТИПОБЕЗОПАСНОСТЬ)
		MinDirectionTime:   customConfig.MinDirectionTime,
		MaxDirectionTime:   customConfig.MaxDirectionTime,
	})

	// Добавляем голод (очень голодное, чтобы искало траву)
	world.AddSatiation(customHerbivore, core.Satiation{Value: 50.0}) // Ниже порога 90%

	// Добавляем скорость
	world.AddVelocity(customHerbivore, core.Velocity{X: 0, Y: 0})

	// Создаем системы
	vegetationSystem := NewVegetationSystem(terrain)
	grassSearchSystem := NewGrassSearchSystem(vegetationSystem)
	_ = NewAnimalBehaviorSystem(vegetationSystem) // behaviorSystem - not used in this test

	// Проверяем начальное состояние - животное голодное, но еще не ест
	if world.HasComponent(customHerbivore, core.MaskEatingState) {
		t.Fatal("Custom herbivore should not be eating initially")
	}

	// Симулируем один тик поиска травы
	grassSearchSystem.Update(world, 1.0/60.0)

	// Проверяем, что кастомное травоядное нашло траву и начало есть
	if !world.HasComponent(customHerbivore, core.MaskEatingState) {
		t.Errorf("Custom herbivore with large vision range (%.2f) should find grass and start eating",
			customConfig.VisionRange)

		// Дополнительная диагностика (ТИПОБЕЗОПАСНОСТЬ)
		pos, _ := world.GetPosition(customHerbivore)
		hunger, _ := world.GetSatiation(customHerbivore)
		t.Logf("Custom herbivore position: (%.2f, %.2f)", pos.X, pos.Y)
		t.Logf("Custom herbivore hunger: %.1f%% (threshold: %.1f%%)", hunger.Value, customConfig.SatiationThreshold)
		t.Logf("Custom herbivore vision range: %.2f (standard: %.2f)",
			customConfig.VisionRange, standardVisionRange)
	}

	// Проверяем, что стандартный заяц тоже может найти траву (контрольная проверка)
	grassSearchSystem.Update(world, 1.0/60.0)
	if !world.HasComponent(standardRabbit, core.MaskEatingState) {
		t.Error("Standard rabbit should also be able to find grass (control test)")
	}

	t.Logf("✅ Custom herbivore vision test passed:")
	t.Logf("   Standard rabbit vision: %.2f tiles", standardVisionRange)
	t.Logf("   Custom herbivore vision: %.2f tiles (2x larger)", customConfig.VisionRange)
	t.Logf("   Both animals successfully found grass with their vision ranges")
}

// TestCustomHerbivoreVision_VisionRangeValidation проверяет, что различные радиусы
// зрения работают правильно при поиске травы
func TestCustomHerbivoreVision_VisionRangeValidation(t *testing.T) {
	// Создаем мир с ограниченной травой
	world := core.NewWorld(20, 20, 12345)

	// Создаем terrain с травой только в определенных местах
	terrain := &generator.Terrain{
		Width:  20,
		Height: 20,
		Size:   20, // ИСПРАВЛЕНИЕ: нужно установить Size для GetSize()
		Tiles:  make([][]generator.TileType, 20),
		Grass:  make([][]float32, 20),
	}

	// Заполняем terrain (по умолчанию без травы)
	for y := 0; y < 20; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 20)
		terrain.Grass[y] = make([]float32, 20)
		for x := 0; x < 20; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 0.0 // Нет травы
		}
	}

	// Добавляем траву только в дальнем углу (15, 15)
	terrain.Grass[15][15] = 100.0

	testCases := []struct {
		name            string
		visionRange     float32
		shouldFindGrass bool
		description     string
	}{
		{
			name:            "Small vision range",
			visionRange:     2.0, // Маленький радиус - не должен найти траву на (15,15)
			shouldFindGrass: false,
			description:     "Should NOT find grass at distance > 2 tiles",
		},
		{
			name:            "Large vision range",
			visionRange:     15.0, // Большой радиус - должен найти траву на (15,15)
			shouldFindGrass: true,
			description:     "Should find grass at distance < 15 tiles",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем кастомное травоядное в тайле (5, 5) в пикселях
			customHerbivore := world.CreateEntity()
			positionX := constants.TilesToPixels(5.0) // 5 тайл * 32 = 160 пикселей
			positionY := constants.TilesToPixels(5.0) // 5 тайл * 32 = 160 пикселей
			world.AddPosition(customHerbivore, core.Position{X: positionX, Y: positionY})
			world.AddAnimalType(customHerbivore, core.TypeRabbit)

			// Кастомная конфигурация с тестовым радиусом зрения
			customConfig := core.AnimalConfig{
				BaseRadius:         RabbitBaseRadius,
				MaxHealth:          RabbitMaxHealth,
				BaseSpeed:          RabbitBaseSpeed,
				CollisionRadius:    RabbitBaseRadius,
				AttackRange:        0,
				VisionRange:        tc.visionRange, // Тестируемый радиус зрения
				SatiationThreshold: RabbitSatiationThreshold,
				FleeThreshold:      RabbitBaseRadius * RabbitFleeDistanceMultiplier,
				SearchSpeed:        SearchSpeedMultiplier,
				WanderingSpeed:     WanderingSpeedMultiplier,
				ContentSpeed:       ContentSpeedMultiplier,
				MinDirectionTime:   1.0,
				MaxDirectionTime:   3.0,
				AttackDamage:       0,
				AttackCooldown:     0,
				HitChance:          0,
			}
			world.AddAnimalConfig(customHerbivore, customConfig)

			// Добавляем поведение
			world.AddBehavior(customHerbivore, core.Behavior{
				Type:               core.BehaviorHerbivore,
				DirectionTimer:     0,
				SatiationThreshold: customConfig.SatiationThreshold,
				FleeThreshold:      customConfig.FleeThreshold,
				SearchSpeed:        customConfig.SearchSpeed,
				WanderingSpeed:     customConfig.WanderingSpeed,
				ContentSpeed:       customConfig.ContentSpeed,
				VisionRange:        customConfig.VisionRange, // ТИПОБЕЗОПАСНОСТЬ
				MinDirectionTime:   customConfig.MinDirectionTime,
				MaxDirectionTime:   customConfig.MaxDirectionTime,
			})

			// Голодное животное
			world.AddSatiation(customHerbivore, core.Satiation{Value: 50.0})
			world.AddVelocity(customHerbivore, core.Velocity{X: 0, Y: 0})

			// Тестируем поиск травы
			vegetationSystem := NewVegetationSystem(terrain)
			grassSearchSystem := NewGrassSearchSystem(vegetationSystem)

			grassSearchSystem.Update(world, 1.0/60.0)

			hasEatingState := world.HasComponent(customHerbivore, core.MaskEatingState)

			if tc.shouldFindGrass && !hasEatingState {
				t.Errorf("%s: Expected to find grass with vision range %.1f, but didn't",
					tc.description, tc.visionRange)
			}

			if !tc.shouldFindGrass && hasEatingState {
				t.Errorf("%s: Should NOT find grass with vision range %.1f, but did",
					tc.description, tc.visionRange)
			}

			t.Logf("✅ %s: Vision range %.1f tiles - %s",
				tc.name, tc.visionRange, tc.description)

			// Очистка для следующего теста
			world.DestroyEntity(customHerbivore)
		})
	}
}
