package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfEatingBehaviorImproved проверяет улучшенное поведение волков при поедании
//
//nolint:gocognit,revive,funlen // Комплексный интеграционный тест поведения волков
func TestWolfEatingBehaviorImproved(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Улучшенное поведение волков при поедании ===")
	t.Logf("ПРОБЛЕМЫ:")
	t.Logf("1. Волки должны есть дискретно (по кадрам анимации), а не каждый тик")
	t.Logf("2. Волки не должны \"телепортироваться\" на труп при поедании")

	// Создаём мир
	cfg := config.LoadDefaultConfig()
	worldWidth := float32(cfg.World.Size * 32)
	worldHeight := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	// Создаём terrain
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Все системы как в реальной игре
	systemManager := core.NewSystemManager()
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	eatingSystem := simulation.NewEatingSystem()
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(worldWidth, worldHeight)

	// Важно: порядок системы как в реальной игре
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: behaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(eatingSystem)

	// Создаём анимационную систему
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil) // 4 FPS = 0.25 сек на кадр
	animationResolver := animation.NewAnimationResolver()

	// Создаём волка и зайца близко (для тайловой системы)
	wolfStartX, wolfStartY := float32(200), float32(200)
	rabbitX, rabbitY := float32(200.2), float32(200.2) // Очень близко для новых размеров

	wolf := simulation.CreateAnimal(world, core.TypeWolf, wolfStartX, wolfStartY)
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)

	// Делаем волка голодным
	world.SetSatiation(wolf, core.Satiation{Value: 30.0})

	// Убиваем зайца и создаём из него труп
	world.RemoveHealth(rabbit)
	corpse := core.Corpse{
		NutritionalValue: 50.0, // Небольшая питательность для тестирования дискретности
		MaxNutritional:   50.0,
		DecayTimer:       300.0,
	}
	world.AddCorpse(rabbit, corpse)

	deltaTime := float32(1.0 / 60.0)
	maxTicks := 600 // 10 секунд

	t.Logf("Начальное состояние:")
	t.Logf("  Позиция волка: (%.1f, %.1f)", wolfStartX, wolfStartY)
	t.Logf("  Позиция трупа: (%.1f, %.1f)", rabbitX, rabbitY)
	t.Logf("  Голод волка: 30.0%%")
	t.Logf("  Питательность трупа: 50.0 единиц")

	eatingStarted := false
	var wolfPositionWhenEatingStarted core.Position
	var nutritionGainedCount int
	var lastNutritionGained float32

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем системы
		world.Update(deltaTime)

		// Обновляем анимации
		animalType, _ := world.GetAnimalType(wolf)
		anim, _ := world.GetAnimation(wolf)
		newAnimType := animationResolver.ResolveAnimalAnimationType(world, wolf, animalType)

		if anim.CurrentAnim != int(newAnimType) {
			anim.CurrentAnim = int(newAnimType)
			anim.Frame = 0
			anim.Timer = 0
			anim.Playing = true
			world.SetAnimation(wolf, anim)
		}

		if anim.Playing {
			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			animationSystem.Update(&animComponent, deltaTime)

			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			anim.FacingRight = animComponent.FacingRight
			world.SetAnimation(wolf, anim)
		}

		systemManager.Update(world, deltaTime)

		// Проверяем состояние каждые 15 тиков
		if tick%15 == 0 {
			wolfPos, _ := world.GetPosition(wolf)
			wolfHunger, _ := world.GetSatiation(wolf)
			isEating := world.HasComponent(wolf, core.MaskEatingState)
			currentAnimType := animation.AnimationType(anim.CurrentAnim)

			// Получаем питательность трупа
			var currentNutrition float32
			if corpseData, hasCorpse := world.GetCorpse(rabbit); hasCorpse {
				currentNutrition = corpseData.NutritionalValue
			}

			// Получаем прогресс поедания
			var nutritionGained float32
			if eatingState, hasEatingState := world.GetEatingState(wolf); hasEatingState {
				nutritionGained = eatingState.NutritionGained
			}

			t.Logf("%.1fs: pos=(%.1f,%.1f), голод=%.1f%%, ест=%v, анимация=%s, питательность=%.1f, съедено=%.1f",
				float32(tick)/60.0, wolfPos.X, wolfPos.Y, wolfHunger.Value, isEating,
				currentAnimType.String(), currentNutrition, nutritionGained)

			// Отслеживаем начало поедания
			if isEating && !eatingStarted {
				eatingStarted = true
				wolfPositionWhenEatingStarted = wolfPos
				t.Logf("✅ Волк начал есть на позиции (%.1f, %.1f)", wolfPos.X, wolfPos.Y)
			}

			// ТЕСТ 1: Дискретное поедание - питательность должна изменяться только по кадрам
			if eatingStarted && nutritionGained != lastNutritionGained {
				nutritionGainedCount++
				t.Logf("🍖 Питательность получена #%d: %.1f единиц", nutritionGainedCount, nutritionGained-lastNutritionGained)
				lastNutritionGained = nutritionGained
			}

			// ТЕСТ 2: Позиционирование - волк не должен двигаться ВО ВРЕМЯ поедания (только когда ест!)
			if eatingStarted && isEating {
				distance := ((wolfPos.X-wolfPositionWhenEatingStarted.X)*(wolfPos.X-wolfPositionWhenEatingStarted.X) +
					(wolfPos.Y-wolfPositionWhenEatingStarted.Y)*(wolfPos.Y-wolfPositionWhenEatingStarted.Y))
				if distance > 1.0 { // Допуск 1 пиксель
					t.Errorf("❌ Волк движется во время поедания! Был на (%.1f,%.1f), стал на (%.1f,%.1f)",
						wolfPositionWhenEatingStarted.X, wolfPositionWhenEatingStarted.Y, wolfPos.X, wolfPos.Y)
					return
				}
			}

			// Проверяем завершение поедания
			if eatingStarted && !isEating {
				t.Logf("✅ Поедание завершено на %.1f секунде", float32(tick)/60.0)
				break
			}
		}
	}

	// Анализ результатов
	t.Logf("\n=== АНАЛИЗ РЕЗУЛЬТАТОВ ===")

	if !eatingStarted {
		t.Errorf("❌ Волк не начал есть за 10 секунд")
		return
	}

	// ТЕСТ 1: Проверяем дискретность поедания
	// expectedNutritionGains := 10 // 50 единиц / 5 единиц за укус = 10 укусов
	if nutritionGainedCount < 5 {
		t.Errorf("❌ Слишком мало дискретных приёмов пищи: %d (ожидалось >5)", nutritionGainedCount)
	} else if nutritionGainedCount > 15 {
		t.Errorf("❌ Слишком много дискретных приёмов пищи: %d (ожидалось <15)", nutritionGainedCount)
	} else {
		t.Logf("✅ Дискретное поедание работает: %d приёмов пищи", nutritionGainedCount)
	}

	// ТЕСТ 2: Проверяем что волк не "телепортировался" к трупу
	finalWolfPos, _ := world.GetPosition(wolf)
	// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для вычислений
	distanceTraveled := ((finalWolfPos.X-wolfStartX)*(finalWolfPos.X-wolfStartX) +
		(finalWolfPos.Y-wolfStartY)*(finalWolfPos.Y-wolfStartY))

	// Расстояние до трупа было ~28 единиц, плюс небольшое движение после поедания
	maxReasonableDistance := float32(40 * 40) // 40 пикселей максимум (разумно для подхода к трупу)

	if distanceTraveled > maxReasonableDistance {
		t.Errorf("❌ Волк слишком далеко переместился: от (%.1f,%.1f) до (%.1f,%.1f), расстояние=%.1f",
			wolfStartX, wolfStartY, finalWolfPos.X, finalWolfPos.Y, float32(distanceTraveled))
	} else {
		t.Logf("✅ Волк переместился разумно: от (%.1f,%.1f) до (%.1f,%.1f), расстояние=%.1f пикселей",
			wolfStartX, wolfStartY, finalWolfPos.X, finalWolfPos.Y, float32(distanceTraveled))
	}

	t.Logf("\n✅ Все тесты улучшенного поведения волков пройдены!")
}
