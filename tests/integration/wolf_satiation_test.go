package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestWolfFullSatiation проверяет что волки едят до полного насыщения (100%), а не до 60%
func TestWolfFullSatiation(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка полного насыщения волков ===")
	t.Logf("ПРОБЛЕМА: Волки наедаются только до 60%% и бросают пойманного зайца")
	t.Logf("ОЖИДАНИЕ: Волки должны наедаться до 100%% (если позволяет питательность зайца)")

	// Создаём мир как в реальной игре
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

	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(eatingSystem)

	// Создаём анимационную систему
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)
	animationResolver := animation.NewAnimationResolver()

	// Создаём волка и зайца
	wolf := simulation.CreateWolf(world, 200, 200)
	rabbit := simulation.CreateRabbit(world, 200, 200)

	// Делаем волка очень голодным
	initialWolfHunger := float32(30.0) // 30% - очень голоден, должен охотиться
	world.SetHunger(wolf, core.Hunger{Value: initialWolfHunger})

	// Убиваем зайца и создаём из него труп (имитируем результат боя)
	world.RemoveHealth(rabbit)
	corpse := core.Corpse{
		NutritionalValue: 200.0, // Достаточно чтобы волк наелся до 100%
		MaxNutritional:   200.0,
		DecayTimer:       300.0,
	}
	world.AddCorpse(rabbit, corpse)

	deltaTime := float32(1.0 / 60.0)
	maxTicks := 1800 // 30 секунд

	t.Logf("Начальное состояние:")
	t.Logf("  Голод волка: %.1f%%", initialWolfHunger)
	t.Logf("  Питательность трупа: %.1f единиц", corpse.NutritionalValue)
	t.Logf("  Порог сытости: %.1f%% (волк должен прекратить есть)", float32(simulation.MaxHungerLimit-simulation.SatietyTolerance))

	eatingStarted := false
	maxHungerReached := float32(0.0)

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем системы
		world.Update(deltaTime)

		// Обновляем анимации как в реальной игре
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

		// Проверяем состояние каждые 30 тиков (0.5 сек)
		if tick%30 == 0 {
			currentHunger, _ := world.GetHunger(wolf)
			isEating := world.HasComponent(wolf, core.MaskEatingState)
			currentAnimType := animation.AnimationType(anim.CurrentAnim)
			
			// Получаем текущую питательность трупа
			var currentNutrition float32
			if corpseData, hasCorpse := world.GetCorpse(rabbit); hasCorpse {
				currentNutrition = corpseData.NutritionalValue
			}

			t.Logf("%.1fs: голод=%.1f%%, ест=%v, анимация=%s, питательность=%.1f",
				float32(tick)/60.0, currentHunger.Value, isEating, currentAnimType.String(), currentNutrition)

			// Отслеживаем максимальный голод
			if currentHunger.Value > maxHungerReached {
				maxHungerReached = currentHunger.Value
			}

			// Отслеживаем начало поедания
			if isEating && !eatingStarted {
				eatingStarted = true
				t.Logf("✅ Волк начал есть на %.1f секунде", float32(tick)/60.0)
			}

			// Проверяем что волк прекратил есть при полном насыщении
			satietyThreshold := float32(simulation.MaxHungerLimit - simulation.SatietyTolerance)
			if !isEating && eatingStarted && currentHunger.Value >= satietyThreshold {
				t.Logf("✅ ПРАВИЛЬНО: Волк прекратил есть при голоде %.1f%% (порог %.1f%%)",
					currentHunger.Value, satietyThreshold)
				t.Logf("✅ Максимальный голод достигнут: %.1f%%", maxHungerReached)
				return
			}

			// ПРОВЕРКА БАГА: Волк НЕ должен прекращать есть на 60%
			if !isEating && eatingStarted && currentHunger.Value < 90.0 {
				t.Errorf("❌ БАГ: Волк прекратил есть слишком рано при %.1f%% голода!", currentHunger.Value)
				t.Errorf("   Это подтверждает жалобу пользователя!")
				t.Errorf("   ОЖИДАЕТСЯ: волк должен есть до ~100%%, а не до %.1f%%", currentHunger.Value)
				return
			}
		}

		// Проверяем здоровье волка
		health, hasHealth := world.GetHealth(wolf)
		if hasHealth && health.Current <= 0 {
			t.Errorf("❌ Волк умер во время теста на тике %d", tick)
			return
		}
	}

	// Анализируем результат
	finalHunger, _ := world.GetHunger(wolf)
	isStillEating := world.HasComponent(wolf, core.MaskEatingState)

	t.Errorf("❌ ТЕСТ НЕ ЗАВЕРШИЛСЯ за %d тиков", maxTicks)
	t.Errorf("   Начальный голод: %.1f%%", initialWolfHunger)
	t.Errorf("   Финальный голод: %.1f%%", finalHunger.Value)
	t.Errorf("   Максимальный голод: %.1f%%", maxHungerReached)
	t.Errorf("   Ещё ест: %v", isStillEating)
	t.Errorf("   Начал есть: %v", eatingStarted)

	if !eatingStarted {
		t.Errorf("   ПРОБЛЕМА: Волк НИКОГДА не начал есть!")
	} else {
		satietyThreshold := float32(simulation.MaxHungerLimit - simulation.SatietyTolerance)
		if finalHunger.Value < satietyThreshold {
			t.Errorf("   ПРОБЛЕМА: Волк не достиг порога насыщения %.1f%% (получил %.1f%%)",
				satietyThreshold, finalHunger.Value)
		} else {
			t.Errorf("   ПРОБЛЕМА: Волк достиг %.1f%% но продолжает есть", finalHunger.Value)
		}
	}
}