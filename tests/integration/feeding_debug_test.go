package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestFeedingDebug отлаживает почему заяц не ест
func TestFeedingDebug(t *testing.T) {
	t.Parallel()

	// Создаём мир
	world := core.NewWorld(TestWorldSize, TestWorldSize, 12345)

	// Создаём terrain и vegetation
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём анимационную систему для корректной работы GrassEatingSystem
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)

	// Создаём FeedingSystem и GrassEatingSystem
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)

	// Найдём место с хорошей травой
	var grassX, grassY float32 = 100, 100
	maxGrass := float32(0)
	for x := float32(50); x < 400; x += 50 {
		for y := float32(50); y < 400; y += 50 {
			grass := vegetationSystem.GetGrassAt(x, y)
			if grass > maxGrass {
				maxGrass = grass
				grassX, grassY = x, y
			}
		}
	}

	t.Logf("=== TDD: Отладка почему заяц не ест ===")
	t.Logf("Лучшее место с травой: (%.0f, %.0f) = %.1f единиц", grassX, grassY, maxGrass)

	// Создаём зайца и настраиваем его для еды
	rabbit := simulation.CreateRabbit(world, grassX, grassY)
	world.SetHunger(rabbit, core.Hunger{Value: 85.0})    // Голодный
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0}) // Стоит

	deltaTime := float32(1.0 / 60.0)

	// Проверяем начальное состояние
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	animalType, _ := world.GetAnimalType(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("\n--- Начальное состояние ---")
	t.Logf("Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("Голод зайца: %.1f%%", hunger.Value)
	t.Logf("Тип животного: %v", animalType)
	t.Logf("Трава в позиции: %.1f единиц", grassAmount)

	// Проверяем условия для поедания
	t.Logf("\n--- Проверка условий поедания ---")

	// 1. Проверяем что заяц - травоядное
	if animalType != core.TypeRabbit {
		t.Errorf("❌ Животное не заяц: %v", animalType)
		return
	} else {
		t.Logf("✅ Животное - заяц")
	}

	// 2. Проверяем голод
	if hunger.Value >= 100.0 {
		t.Errorf("❌ Заяц сыт (%.1f%% >= 100%%)", hunger.Value)
		return
	} else {
		t.Logf("✅ Заяц голоден (%.1f%% < 100%%)", hunger.Value)
	}

	// 3. Проверяем траву
	minGrassAmount := float32(10.0) // simulation.MinGrassAmount
	if grassAmount < minGrassAmount {
		t.Errorf("❌ Недостаточно травы (%.1f < %.1f)", grassAmount, minGrassAmount)
		return
	} else {
		t.Logf("✅ Достаточно травы (%.1f >= %.1f)", grassAmount, minGrassAmount)
	}

	t.Logf("\n--- Симуляция поедания ---")

	// Устанавливаем анимацию поедания (необходимо для GrassEatingSystem)
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(animation.AnimEat),
		Frame:       0,
		Timer:       0.0,
		Playing:     true,
		FacingRight: true,
	})

	// Обновляем обе системы питания
	// Кадр длится 0.25 сек, deltaTime = 0.017 сек => нужно ~15 тиков
	for i := 0; i < 20; i++ {
		t.Logf("\nТик %d:", i)

		// Состояние ДО обновления
		hunger, _ = world.GetHunger(rabbit)
		pos, _ = world.GetPosition(rabbit)
		grassBefore := vegetationSystem.GetGrassAt(pos.X, pos.Y)
		isEatingBefore := world.HasComponent(rabbit, core.MaskEatingState)

		t.Logf("  ДО: голод=%.1f%% трава=%.1f ест=%v", hunger.Value, grassBefore, isEatingBefore)

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

		// Обновляем обе системы питания
		feedingSystem.Update(world, deltaTime)
		grassEatingSystem.Update(world, deltaTime)

		// Состояние ПОСЛЕ обновления
		hunger, _ = world.GetHunger(rabbit)
		grassAfter := vegetationSystem.GetGrassAt(pos.X, pos.Y)
		isEatingAfter := world.HasComponent(rabbit, core.MaskEatingState)

		t.Logf("  ПОСЛЕ: голод=%.1f%% трава=%.1f ест=%v", hunger.Value, grassAfter, isEatingAfter)

		// Проверяем изменения
		grassConsumed := grassBefore - grassAfter
		if grassConsumed > 0 {
			t.Logf("  ✅ Съедено травы: %.2f единиц", grassConsumed)
			t.Logf("  ✅ EatingState создан: %v", isEatingAfter)
			// Получаем детали EatingState
			if eatingState, hasEating := world.GetEatingState(rabbit); hasEating {
				t.Logf("    Прогресс: %.2f, Питательность: %.2f",
					eatingState.EatingProgress, eatingState.NutritionGained)
			}
			return // Трава потреблена - тест успешен!
		} else {
			t.Logf("  ⏱️  Трава пока не потреблена (ожидаем завершения кадра анимации)")
		}

		if isEatingAfter {
			t.Logf("  ✅ EatingState создан - ожидаем потребления травы")
		} else {
			t.Logf("  ⏱️  EatingState пока не создан")
		}
	}

	t.Errorf("❌ Заяц не начал есть за 20 тиков")
}
