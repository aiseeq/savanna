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

// TestRabbitFeedingBalance проверяет что зайцы НЕ наедаются слишком быстро
//
//nolint:gocognit,revive,funlen // Комплексный интеграционный тест баланса системы питания
func TestRabbitFeedingBalance(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка баланса питания зайцев ===")
	t.Logf("ЦЕЛЬ: Убедиться что зайцы НЕ наедаются мгновенно")
	t.Logf("ПРОБЛЕМА: Пользователь сообщает что зайцы очень быстро наедаются")

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
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)

	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})
	systemManager.AddSystem(grassEatingSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: behaviorSystem})

	// Создаём анимационную систему как в реальной игре
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)
	animationResolver := animation.NewAnimationResolver()

	// Создаём зайца и траву
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)
	tileX := int(200 / 32)
	tileY := int(200 / 32)
	terrain.SetGrassAmount(tileX, tileY, 100.0)

	// Делаем зайца голодным
	initialHunger := float32(50.0) // 50% - голодный но не критично
	world.SetHunger(rabbit, core.Hunger{Value: initialHunger})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	deltaTime := float32(1.0 / 60.0)
	// ИСПРАВЛЕНИЕ: Заяц теперь ест до полного насыщения, а не до RabbitHungryThreshold
	satietyThreshold := float32(simulation.MaxHungerValue - simulation.SatietyTolerance) // 99.9%

	t.Logf("Начальное состояние:")
	t.Logf("  Голод зайца: %.1f%%", initialHunger)
	t.Logf("  Трава: %.1f единиц", terrain.GetGrassAmount(tileX, tileY))
	t.Logf("  Порог сытости: %.1f%% (заяц должен прекратить есть)", satietyThreshold)

	// КРИТИЧЕСКИЙ ТЕСТ: Симулируем 30 секунд (1800 тиков) - после уменьшения скорости поедания
	t.Logf("\n=== ТЕСТИРУЕМ СКОРОСТЬ НАСЫЩЕНИЯ ===")
	t.Logf("Ожидание: заяц НЕ должен наедаться мгновенно")

	maxTicks := 1800 // 30 секунд при 60 FPS (после снижения скорости поедания нужно ещё больше времени)
	eatingStarted := false
	ticksToSatiation := -1

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем системы
		world.Update(deltaTime)

		// Обновляем анимации правильно
		animalType, _ := world.GetAnimalType(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		newAnimType := animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)

		if anim.CurrentAnim != int(newAnimType) {
			anim.CurrentAnim = int(newAnimType)
			anim.Frame = 0
			anim.Timer = 0
			anim.Playing = true
			world.SetAnimation(rabbit, anim)
		}

		if anim.Playing {
			// Конвертируем в AnimationComponent как в реальной игре
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
			world.SetAnimation(rabbit, anim)
		}

		systemManager.Update(world, deltaTime)

		// Проверяем состояние каждые 30 тиков (0.5 сек)
		if tick%30 == 0 {
			currentHunger, _ := world.GetHunger(rabbit)
			isEating := world.HasComponent(rabbit, core.MaskEatingState)
			currentAnimType := animation.AnimationType(anim.CurrentAnim)

			t.Logf("%.1fs: голод=%.1f%%, ест=%v, анимация=%s",
				float32(tick)/60.0, currentHunger.Value, isEating, currentAnimType.String())

			// Отслеживаем начало поедания
			if isEating && !eatingStarted {
				eatingStarted = true
				t.Logf("✅ Заяц начал есть на %.1f секунде", float32(tick)/60.0)
			}

			// Проверяем насыщение - заяц прекратил есть при высоком голоде
			if !isEating && eatingStarted && currentHunger.Value >= 95.0 {
				ticksToSatiation = tick
				t.Logf("✅ Заяц насытился на %.1f секунде (%.1f%%) и прекратил есть",
					float32(tick)/60.0, currentHunger.Value)
				break
			}
		}
	}

	// АНАЛИЗ РЕЗУЛЬТАТОВ
	t.Logf("\n=== АНАЛИЗ БАЛАНСА ===")

	finalHunger, _ := world.GetHunger(rabbit)
	t.Logf("Финальный голод: %.1f%%", finalHunger.Value)

	if !eatingStarted {
		t.Errorf("❌ ПРОБЛЕМА: Заяц НЕ начал есть за 5 секунд")
		return
	}

	if ticksToSatiation == -1 {
		t.Errorf("❌ ПРОБЛЕМА: Заяц НЕ насытился за 30 секунд")
		t.Errorf("   Возможно баланс слишком медленный")
		return
	}

	timeToSatiation := float32(ticksToSatiation) / 60.0

	// КРИТЕРИИ БАЛАНСА
	if timeToSatiation < 1.0 {
		t.Errorf("❌ БАГ: Заяц насытился СЛИШКОМ БЫСТРО за %.1f сек", timeToSatiation)
		t.Errorf("   Это подтверждает жалобу пользователя!")
		t.Errorf("   ОЖИДАЕТСЯ: минимум 2-3 секунды для насыщения")
	} else if timeToSatiation < 2.0 {
		t.Logf("⚠️  ВНИМАНИЕ: Заяц насытился довольно быстро за %.1f сек", timeToSatiation)
		t.Logf("   Это может объяснять жалобу пользователя")
	} else {
		t.Logf("✅ БАЛАНС OK: Заяц насытился за %.1f сек (разумно)", timeToSatiation)
	}
}
