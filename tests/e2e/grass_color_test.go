package e2e

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

// TestGrassColorChangeE2E проверяет что цвет травы меняется при её поедании
//
//nolint:gocognit,revive // E2E тест изменения цвета травы
func TestGrassColorChangeE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка изменения цвета травы ===")

	// Создаём минимальный мир
	worldWidth := float32(200.0)
	world := core.NewWorld(worldWidth, worldWidth, 12345)

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 3
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Находим тайл с травой или создаем его
	grassTileX, grassTileY := 1, 1

	// Ищем подходящий тайл или устанавливаем тип травы
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			tileType := terrain.GetTileType(x, y)
			if tileType == generator.TileGrass || tileType == generator.TileWetland {
				grassTileX, grassTileY = x, y
				goto found
			}
		}
	}

	// Если не нашли подходящий тайл, устанавливаем тип травы принудительно
	terrain.SetTileType(grassTileX, grassTileY, generator.TileGrass)

found:
	// Размещаем максимальное количество травы
	terrain.SetGrassAmount(grassTileX, grassTileY, 100.0)

	// Размещаем зайца в центре тайла с травой
	rabbitX := float32(grassTileX*32 + 16) // Центр тайла
	rabbitY := float32(grassTileY*32 + 16)

	t.Logf("Тайл с травой: (%d,%d), тип=%d", grassTileX, grassTileY, terrain.GetTileType(grassTileX, grassTileY))
	t.Logf("Заяц будет размещен на: (%.1f, %.1f)", rabbitX, rabbitY)

	// Создаём vegetation систему для доступа к траве
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём анимационную систему ТОЧНО как в игре
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimWalk, 2, 8.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimRun, 2, 12.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)

	// ИСПРАВЛЕНИЕ: Используем terrain из этого теста, а не создаем новый
	systemManager := common.CreateTestSystemManagerWithTerrain(worldWidth, terrain)

	// Создаём зайца в центре тайла с травой
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)
	world.SetSatiation(rabbit, core.Satiation{Value: 10.0}) // Очень голодный
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	// Устанавливаем анимацию поедания
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(animation.AnimEat),
		Frame:       0,
		Timer:       0.0,
		Playing:     true,
		FacingRight: true,
	})

	// Проверяем начальные значения
	initialGrass := vegetationSystem.GetGrassAt(rabbitX, rabbitY)
	initialColor := calculateGrassColor(initialGrass)
	t.Logf("Начальная трава: %.1f единиц", initialGrass)
	t.Logf("Начальный цвет: R=%d, G=%d, B=%d", initialColor.R, initialColor.G, initialColor.B)

	// Симулируем поедание
	deltaTime := float32(1.0 / 60.0)
	for tick := 0; tick < 120; tick++ { // 2 секунды
		// Обновляем анимацию ТОЧНО как в игре (processAnimationUpdate)
		if anim, hasAnim := world.GetAnimation(rabbit); hasAnim && anim.Playing {
			// Создаём компонент для системы анимации (как в AnimationManager)
			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			// Обновляем через систему анимации (как в игре)
			animationSystem.Update(&animComponent, deltaTime)

			// Сохраняем обновлённое состояние (как в игре)
			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			anim.FacingRight = animComponent.FacingRight
			world.SetAnimation(rabbit, anim)
		}

		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Проверяем каждую секунду
		if tick%60 == 59 {
			currentGrass := vegetationSystem.GetGrassAt(rabbitX, rabbitY)
			currentColor := calculateGrassColor(currentGrass)
			hunger, _ := world.GetSatiation(rabbit)
			_, isEating := world.GetEatingState(rabbit)

			pos, _ := world.GetPosition(rabbit)
			config, _ := world.GetAnimalConfig(rabbit)

			// Проверяем что система поиска травы может найти траву
			_, _, foundGrass := vegetationSystem.FindNearestGrass(pos.X, pos.Y, 30.0, 10.0)

			t.Logf("Через %d сек: трава=%.1f единиц, голод=%.1f%%, ест=%v", (tick+1)/60, currentGrass, hunger.Value, isEating)
			t.Logf("  Позиция зайца: (%.1f, %.1f), порог голода=%.1f%%, найдена трава=%v", pos.X, pos.Y, config.SatiationThreshold, foundGrass)
			t.Logf("  Цвет: R=%d, G=%d, B=%d", currentColor.R, currentColor.G, currentColor.B)

			// Проверяем что трава потреблена и цвет изменился
			if currentGrass < initialGrass {
				t.Logf("✅ Трава потреблена: %.1f -> %.1f", initialGrass, currentGrass)

				// Проверяем изменение цвета
				if currentColor.G < initialColor.G {
					t.Logf("✅ Цвет потемнел: G%d -> G%d", initialColor.G, currentColor.G)
					colorDiff := int(initialColor.G) - int(currentColor.G)
					t.Logf("✅ Разница в зелёном канале: %d единиц", colorDiff)

					if colorDiff >= 10 {
						t.Logf("✅ УСПЕХ: Изменение цвета заметно (>= 10 единиц)")
						return
					} else {
						t.Logf("⚠️  Изменение цвета слабо заметно (< 10 единиц)")
					}
				} else {
					t.Errorf("❌ Цвет не изменился: G%d == G%d", initialColor.G, currentColor.G)
				}
			} else {
				t.Errorf("❌ Трава не потреблена: %.1f >= %.1f", currentGrass, initialGrass)
			}
		}
	}

	finalGrass := vegetationSystem.GetGrassAt(rabbitX, rabbitY)
	finalColor := calculateGrassColor(finalGrass)
	t.Logf("Финальная трава: %.1f единиц", finalGrass)
	t.Logf("Финальный цвет: R=%d, G=%d, B=%d", finalColor.R, finalColor.G, finalColor.B)

	if finalGrass >= initialGrass {
		t.Errorf("❌ Трава не была съедена за 2 секунды")
	}
}

// RGBAColor простая структура для цвета
type RGBAColor struct {
	R, G, B, A uint8
}

// calculateGrassColor вычисляет цвет травы как в реальной игре
func calculateGrassColor(grassAmount float32) RGBAColor {
	// Копируем логику из drawTerrain в main.go
	green := uint8(50 + grassAmount*2) // 50-250
	if green > 255 {
		green = 255
	}
	return RGBAColor{
		R: 20,
		G: green,
		B: 20,
		A: 255,
	}
}
