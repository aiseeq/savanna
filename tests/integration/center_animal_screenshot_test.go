package integration

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/rendering"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestCenterAnimalScreenshot проверяет логику центрирования камеры на животном (без GUI)
func TestCenterAnimalScreenshot(t *testing.T) {
	t.Log("Тестирование логики центрирования камеры на животном...")

	// Создаем детерминированную симуляцию
	seed := int64(12345) // Фиксированный seed для воспроизводимости
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = seed
	cfg.Population.Rabbits = 5 // Несколько зайцев
	cfg.Population.Wolves = 1  // Один волк

	// Генерируем ландшафт
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.GenerateRectangular(50, 38)

	// Создаем мир
	worldWidthTiles := float32(terrain.Width)
	worldHeightTiles := float32(terrain.Height)
	world := core.NewWorld(worldWidthTiles, worldHeightTiles, seed)

	// Создаем системы
	systemManager := core.NewSystemManager()

	// Создаем и добавляем системы в правильном порядке
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	satiationSystem := simulation.NewSatiationSystem()
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	satiationSpeedModifier := simulation.NewSatiationSpeedModifierSystem()
	movementSystem := simulation.NewMovementSystem(worldWidthTiles, worldHeightTiles)
	combatSystem := simulation.NewCombatSystem()
	starvationDamage := simulation.NewStarvationDamageSystem()

	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.HungerSystemAdapter{System: satiationSystem})
	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{System: grassSearchSystem})
	systemManager.AddSystem(grassEatingSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{System: satiationSpeedModifier})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(combatSystem)
	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{System: starvationDamage})

	// Создаем анимационную систему
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()
	loader := animation.NewAnimationLoader()
	emptyImg := ebiten.NewImage(128, 64)
	loader.LoadAnimations(wolfAnimationSystem, rabbitAnimationSystem, emptyImg, emptyImg)

	// Размещаем животных
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	for _, placement := range placements {
		// Преобразуем координаты из пикселей в тайлы
		tileX := placement.X / 32.0
		tileY := placement.Y / 32.0
		simulation.CreateAnimal(world, placement.Type, tileX, tileY)
	}

	t.Logf("Размещено животных: %d", len(placements))

	// Симулируем некоторое время для развития экосистемы
	deltaTime := float32(1.0 / 60)
	for i := 0; i < 300; i++ { // 5 секунд симуляции
		world.Update(deltaTime)

		// Обновляем анимации
		animationManager := animation.NewAnimationManager(wolfAnimationSystem, rabbitAnimationSystem)
		animationManager.UpdateAllAnimations(world, deltaTime)

		// Обновляем системы
		systemManager.Update(world, deltaTime)
	}

	// Создаем камеру для тестирования логики центрирования
	camera := rendering.NewCamera(float32(terrain.Width), float32(terrain.Height))

	// Тестируем логику центрирования камеры на первом животном
	centerCameraOnFirstAnimal(camera, world)

	// Проверяем что камера была правильно настроена
	if camera.GetZoom() != 4.0 {
		t.Errorf("Ожидался zoom 4.0, получен %.1f", camera.GetZoom())
	}

	// Проверяем что первое животное находится в центре экрана
	var foundAnimal bool
	world.ForEachWith(core.MaskPosition|core.MaskAnimalType, func(entity core.EntityID) {
		if !foundAnimal {
			if pos, hasPos := world.GetPosition(entity); hasPos {
				// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
				screenX, screenY := camera.WorldToScreen(pos.X, pos.Y)

				// Центр экрана 1024x768
				expectedCenterX := float32(512)
				expectedCenterY := float32(384)

				tolerance := float32(5.0) // Разрешаем небольшую погрешность

				if abs(screenX-expectedCenterX) > tolerance || abs(screenY-expectedCenterY) > tolerance {
					t.Errorf("Животное не в центре экрана: screen(%.1f,%.1f), ожидалось(%.1f,%.1f)",
						screenX, screenY, expectedCenterX, expectedCenterY)
				} else {
					t.Logf("✅ Животное правильно центрировано: screen(%.1f,%.1f)", screenX, screenY)
				}
				foundAnimal = true
			}
		}
	})

	if !foundAnimal {
		t.Error("Не найдено ни одного животного для центрирования")
	}

	t.Log("✅ Логика центрирования камеры работает корректно")
}

// centerCameraOnFirstAnimal центрирует камеру на первом найденном животном
func centerCameraOnFirstAnimal(camera *rendering.Camera, world *core.World) {
	var targetPos core.Position
	var targetEntity core.EntityID
	found := false

	// Находим первое животное
	world.ForEachWith(core.MaskPosition|core.MaskAnimalType, func(entity core.EntityID) {
		if !found {
			if pos, hasPos := world.GetPosition(entity); hasPos {
				targetEntity = entity
				targetPos = pos
				found = true
			}
		}
	})

	if found {
		// Устанавливаем максимальный zoom (4x)
		camera.SetZoom(4.0)

		// Центр экрана (1024x768)
		screenCenterX := float32(512)
		screenCenterY := float32(384)

		// Преобразуем мировые координаты животного в экранные БЕЗ камеры
		// Используем базовую изометрическую проекцию
		// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для вычислений
		baseScreenX := (targetPos.X - targetPos.Y) * 32 / 2 // TileWidth = 32
		baseScreenY := (targetPos.X + targetPos.Y) * 16 / 2 // TileHeight = 16

		// Применяем zoom
		zoomedScreenX := baseScreenX * 4.0
		zoomedScreenY := baseScreenY * 4.0

		// Вычисляем нужное смещение камеры для центрирования
		cameraX := zoomedScreenX - screenCenterX
		cameraY := zoomedScreenY - screenCenterY

		camera.SetPosition(cameraX, cameraY)

		// Проверяем результат
		// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
		resultScreenX, resultScreenY := camera.WorldToScreen(targetPos.X, targetPos.Y)

		// Получаем информацию о животном
		animalTypeStr := "неизвестное"
		if animalType, hasType := world.GetAnimalType(targetEntity); hasType {
			switch animalType {
			case core.TypeRabbit:
				animalTypeStr = "заяц"
			case core.TypeWolf:
				animalTypeStr = "волк"
			}
		}

		fmt.Printf("Камера центрирована: entity=%d (%s), world(%.1f,%.1f), screen(%.1f,%.1f), zoom=4.0x\n",
			targetEntity, animalTypeStr, targetPos.X, targetPos.Y, resultScreenX, resultScreenY)
	}
}
