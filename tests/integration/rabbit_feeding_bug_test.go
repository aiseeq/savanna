package integration

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestRabbitFeedingBugE2E проверяет что зайцы правильно насыщаются до 90% как в игре
func TestRabbitFeedingBugE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка бага насыщения зайцев ===")
	t.Logf("Проблема: в игре видно что зайцы едят но не наедаются")
	t.Logf("Ожидание: заяц должен наедаться до 90%% и прекращать есть")

	// Создаём ТОЧНО такую же конфигурацию как в GUI игре
	cfg := config.LoadDefaultConfig()

	// Создаём мир такого же размера как в игре
	worldWidth := float32(cfg.World.Size * 32)  // Как в игре
	worldHeight := float32(cfg.World.Size * 32) // Как в игре
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	// Создаём terrain ТОЧНО как в игре
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаём все системы ТОЧНО как в GUI игре (game_world.go)
	systemManager := core.NewSystemManager()

	// VegetationSystem (как в игре)
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	systemManager.AddSystem(vegetationSystem)

	// FeedingSystem (как в игре)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})

	// GrassEatingSystem (как в игре)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.GrassEatingSystemAdapter{System: grassEatingSystem})

	// AnimalBehaviorSystem (как в игре)
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: behaviorSystem})

	// MovementSystem (как в игре)
	movementSystem := simulation.NewMovementSystem(worldWidth, worldHeight)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Создаём анимационную систему ТОЧНО как в игре
	animationSystem := animation.NewAnimationSystem()

	// Регистрируем анимации ТОЧНО как в GUI (loadRabbitAnimations)
	animationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimWalk, 2, 8.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimRun, 2, 12.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)

	// Animation resolver ТОЧНО как в игре
	animationResolver := animation.NewAnimationResolver()

	// Создаём зайца с хорошими условиями для еды
	rabbit := simulation.CreateRabbit(world, 160, 160) // В центре карты

	// Делаем зайца голодным но не критично (как в игре)
	initialHunger := float32(60.0) // 60% - будет искать еду но не умрёт
	world.SetHunger(rabbit, core.Hunger{Value: initialHunger})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0}) // Стоит на месте

	// Размещаем траву прямо под зайцем
	tileX := int(160 / 32)
	tileY := int(160 / 32)
	terrain.SetGrassAmount(tileX, tileY, 100.0) // Много травы

	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("Начальное состояние (как в игре):")
	t.Logf("  Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Голод зайца: %.1f%%", hunger.Value)
	t.Logf("  Трава в позиции: %.1f единиц", grassAmount)
	t.Logf("  SatietyThreshold: %.1f%% (заяц должен прекратить есть на этом уровне)", simulation.MaxHungerValue-simulation.SatietyTolerance)

	// Проверяем начальные условия
	if grassAmount < 50.0 {
		t.Errorf("❌ Недостаточно травы для теста: %.1f < 50.0", grassAmount)
		return
	}

	satietyThreshold := float32(simulation.MaxHungerValue - simulation.SatietyTolerance)
	if hunger.Value >= satietyThreshold {
		t.Errorf("❌ Заяц слишком сыт для теста: %.1f%% >= %.1f%%", hunger.Value, satietyThreshold)
		return
	}

	deltaTime := float32(1.0 / 60.0) // 60 FPS как в игре
	maxTicks := 1200                 // 20 секунд симуляции (после уменьшения скорости поедания)

	t.Logf("\nНачинаем симуляцию ТОЧНО как в GUI игре...")

	eatingStarted := false
	maxHungerReached := float32(0.0)

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем системы ТОЧНО как в игре (новый правильный порядок)
		world.Update(deltaTime)

		// ИСПРАВЛЕНИЕ: Анимации должны обновляться ПЕРЕД системами
		// Ручное обновление анимаций (имитируем AnimationManager)
		animalType, _ := world.GetAnimalType(rabbit)
		anim, _ := world.GetAnimation(rabbit)

		// Определяем новый тип анимации через resolver (ТОЧНО как в игре)
		newAnimType := animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)

		// Обновляем анимацию если нужно (updateAnimationIfNeeded как в игре)
		if anim.CurrentAnim != int(newAnimType) {
			// НЕ прерываем анимацию ATTACK (как в игре)
			if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
				// Не меняем
			} else {
				// Обычная смена анимации (как в игре)
				anim.CurrentAnim = int(newAnimType)
				anim.Frame = 0
				anim.Timer = 0
				anim.Playing = true
				world.SetAnimation(rabbit, anim)
			}
		}

		// Обновляем анимацию ТОЧНО как в игре (processAnimationUpdate)
		if anim.Playing {
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

		// ТЕПЕРЬ обновляем системы ПОСЛЕ анимаций (новый правильный порядок)
		systemManager.Update(world, deltaTime)

		// Проверяем состояние каждые 15 тиков (кадр анимации поедания)
		if tick%15 == 0 {
			currentHunger, _ := world.GetHunger(rabbit)
			currentGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)
			isEating := world.HasComponent(rabbit, core.MaskEatingState)
			currentAnimType := animation.AnimationType(anim.CurrentAnim)

			// Отладочная информация о состоянии EatingState
			var eatingInfo string
			if eatingState, hasEatingState := world.GetEatingState(rabbit); hasEatingState {
				eatingInfo = fmt.Sprintf("(прогресс:%.2f, питательность:%.2f)", eatingState.EatingProgress, eatingState.NutritionGained)
			} else {
				eatingInfo = "(нет состояния)"
			}

			// Получаем информацию о таймере анимации
			currentAnim, _ := world.GetAnimation(rabbit)
			frameTarget := float32(1.0 / 4.0) // 0.25 секунды
			shouldTrigger := currentAnim.Timer >= frameTarget

			t.Logf("Тик %d (%.3fs): голод=%.1f%%, трава=%.1f, ест=%v %s",
				tick, float32(tick)/60.0, currentHunger.Value, currentGrass, isEating, eatingInfo)
			t.Logf("  Анимация: %s, таймер=%.3f, цель=%.3f, кадр_готов=%v",
				currentAnimType.String(), currentAnim.Timer, frameTarget, shouldTrigger)

			// Отслеживаем максимальный голод
			if currentHunger.Value > maxHungerReached {
				maxHungerReached = currentHunger.Value
			}

			// Проверяем что заяц начал есть
			if isEating && !eatingStarted {
				eatingStarted = true
				t.Logf("✅ Заяц начал есть на тике %d", tick)
			}

			// КРИТИЧЕСКАЯ ПРОВЕРКА: заяц НЕ должен превышать максимально возможное значение голода
			maxPossibleHunger := float32(simulation.MaxHungerValue)
			if currentHunger.Value > maxPossibleHunger {
				t.Errorf("❌ БАГ: Заяц превысил максимальное значение голода!")
				t.Errorf("   Текущий голод: %.1f%% > %.1f%% (максимум)", currentHunger.Value, maxPossibleHunger)
				t.Errorf("   Это нарушает ограничения системы голода")
				return
			}

			// Проверяем что заяц прекратил есть при достижении полного насыщения
			// ИСПРАВЛЕНИЕ: Теперь заяц ест до 99.9%, а не до 90%
			if currentHunger.Value >= satietyThreshold && !isEating && eatingStarted {
				t.Logf("✅ ПРАВИЛЬНО: Заяц прекратил есть при голоде %.1f%% (порог %.1f%%)",
					currentHunger.Value, satietyThreshold)
				t.Logf("✅ Максимальный голод достигнут: %.1f%%", maxHungerReached)
				return
			}
		}

		// Проверяем здоровье зайца
		health, hasHealth := world.GetHealth(rabbit)
		if hasHealth && health.Current <= 0 {
			t.Errorf("❌ Заяц умер во время теста на тике %d", tick)
			return
		}
	}

	// Анализируем результат
	finalHunger, _ := world.GetHunger(rabbit)
	finalGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	isStillEating := world.HasComponent(rabbit, core.MaskEatingState)

	t.Errorf("❌ ТЕСТ НЕ ЗАВЕРШИЛСЯ за %d тиков", maxTicks)
	t.Errorf("   Начальный голод: %.1f%%", initialHunger)
	t.Errorf("   Финальный голод: %.1f%%", finalHunger.Value)
	t.Errorf("   Максимальный голод: %.1f%%", maxHungerReached)
	t.Errorf("   Ещё ест: %v", isStillEating)
	t.Errorf("   Трава: %.1f -> %.1f", grassAmount, finalGrass)
	t.Errorf("   Начал есть: %v", eatingStarted)

	if !eatingStarted {
		t.Errorf("   ПРОБЛЕМА: Заяц НИКОГДА не начал есть!")
	} else {
		// ИСПРАВЛЕНИЕ: Проверяем новый порог насыщения (99.9%)
		satietyThreshold := float32(simulation.MaxHungerValue - simulation.SatietyTolerance)
		if finalHunger.Value < satietyThreshold {
			t.Errorf("   ПРОБЛЕМА: Заяц не достиг порога насыщения %.1f%% (получил %.1f%%)",
				satietyThreshold, finalHunger.Value)
		} else {
			t.Errorf("   ПРОБЛЕМА: Заяц достиг %.1f%% но продолжает есть", finalHunger.Value)
		}
	}
}
