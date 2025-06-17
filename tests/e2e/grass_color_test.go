package e2e

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestGrassColorChangeE2E проверяет что цвет травы меняется при её поедании
func TestGrassColorChangeE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка изменения цвета травы ===")

	// Создаём минимальный мир
	world := core.NewWorld(200, 200, 12345)

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 3
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Размещаем максимальное количество травы
	terrain.SetGrassAmount(1, 1, 100.0)

	// Создаём системы питания
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)

	// Создаём анимационную систему ТОЧНО как в игре
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimWalk, 2, 8.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimRun, 2, 12.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)

	systemManager := core.NewSystemManager()
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})
	systemManager.AddSystem(&adapters.GrassEatingSystemAdapter{System: grassEatingSystem})

	// Создаём зайца
	rabbit := simulation.CreateRabbit(world, 48, 48)
	world.SetHunger(rabbit, core.Hunger{Value: 10.0}) // Очень голодный
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
	initialGrass := vegetationSystem.GetGrassAt(48, 48)
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
			currentGrass := vegetationSystem.GetGrassAt(48, 48)
			currentColor := calculateGrassColor(currentGrass)

			t.Logf("Через %d сек: трава=%.1f единиц", (tick+1)/60, currentGrass)
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

	finalGrass := vegetationSystem.GetGrassAt(48, 48)
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
