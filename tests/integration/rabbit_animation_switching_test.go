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

// TestRabbitAnimationSwitching проверяет проблему переключения idle 0 ↔ eat 0
//
//nolint:gocognit,revive,funlen // Комплексный тест переключения анимаций зайца
func TestRabbitAnimationSwitching(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка переключения анимаций idle 0 ↔ eat 0 ===")
	t.Logf("ЦЕЛЬ: Воспроизвести проблему которую видит пользователь")
	t.Logf("ПРОБЛЕМА: Зайцы переключаются между idle 0 и eat 0 вместо eat 0 → eat 1")

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

	// Создаём анимационную систему ТОЧНО как в реальной игре
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
	world.SetHunger(rabbit, core.Hunger{Value: 60.0})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	deltaTime := float32(1.0 / 60.0)

	t.Logf("Начальное состояние:")
	t.Logf("  Голод зайца: 60.0%%")
	t.Logf("  Трава: 100.0 единиц")

	t.Logf("\n=== ОТСЛЕЖИВАНИЕ АНИМАЦИЙ ===")

	// Отслеживаем анимации для выявления паттерна
	var prevAnimType animation.AnimationType
	animationSwitches := 0
	idleToEatSwitches := 0
	eatToIdleSwitches := 0

	for tick := 0; tick < 120; tick++ { // 2 секунды
		// Обновляем системы
		world.Update(deltaTime)

		// Обновляем анимации правильно как в реальной игре
		animalType, _ := world.GetAnimalType(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		newAnimType := animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)

		// Проверяем смену анимации
		animationChanged := false
		if anim.CurrentAnim != int(newAnimType) {
			animationChanged = true
			prevAnimType = animation.AnimationType(anim.CurrentAnim)

			anim.CurrentAnim = int(newAnimType)
			anim.Frame = 0
			anim.Timer = 0
			anim.Playing = true
			world.SetAnimation(rabbit, anim)

			animationSwitches++

			// Подсчитываем типы переключений
			if prevAnimType == animation.AnimIdle && newAnimType == animation.AnimEat {
				idleToEatSwitches++
			} else if prevAnimType == animation.AnimEat && newAnimType == animation.AnimIdle {
				eatToIdleSwitches++
			}
		}

		// Обновляем кадры анимации
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

		// Логируем каждые 15 тиков (0.25 сек - один кадр анимации)
		if tick%15 == 0 || animationChanged {
			currentHunger, _ := world.GetHunger(rabbit)
			isEating := world.HasComponent(rabbit, core.MaskEatingState)
			currentAnimType := animation.AnimationType(anim.CurrentAnim)

			marker := ""
			if animationChanged {
				marker = " ← СМЕНА!"
			}

			t.Logf("Тик %3d: анимация=%s кадр=%d, ест=%v, голод=%.1f%%%s",
				tick, currentAnimType.String(), anim.Frame, isEating, currentHunger.Value, marker)
		}

		// Проверяем проблемный паттерн: eat кадр 0 постоянно
		if animation.AnimationType(anim.CurrentAnim) == animation.AnimEat && anim.Frame == 0 && tick > 30 {
			// Проверяем что eat анимация не прогрессирует к кадру 1
			if tick%30 == 0 { // Каждые 0.5 сек
				t.Logf("⚠️  ПРОБЛЕМА: Eat анимация застряла на кадре 0 на тике %d", tick)
			}
		}
	}

	// АНАЛИЗ РЕЗУЛЬТАТОВ
	t.Logf("\n=== АНАЛИЗ ПЕРЕКЛЮЧЕНИЙ АНИМАЦИЙ ===")
	t.Logf("Всего смен анимаций: %d", animationSwitches)
	t.Logf("Idle → Eat: %d переключений", idleToEatSwitches)
	t.Logf("Eat → Idle: %d переключений", eatToIdleSwitches)

	// ДИАГНОСТИКА ПРОБЛЕМЫ
	if idleToEatSwitches > 3 && eatToIdleSwitches > 3 {
		t.Errorf("❌ БАГ ПОДТВЕРЖДЁН: Частые переключения Idle ↔ Eat")
		t.Errorf("   Это объясняет жалобу пользователя!")
		t.Errorf("   Заяц должен стабильно оставаться в анимации Eat")
	} else if animationSwitches > 10 {
		t.Errorf("❌ БАГ: Слишком много смен анимаций (%d)", animationSwitches)
	} else {
		t.Logf("✅ Анимации работают стабильно")
	}
}
