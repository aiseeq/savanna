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

// TestEatingSleepAnimationBug воспроизводит баг когда заяц показывает анимацию сна вместо поедания
func TestEatingSleepAnimationBug(t *testing.T) {
	t.Parallel()

	// Создаём мир точно как в реальной игре
	world := core.NewWorld(TestWorldSize, TestWorldSize, 12345)

	// Создаём terrain и vegetation
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём все системы в том же порядке что в GUI/headless
	systemManager := core.NewSystemManager()
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(TestWorldSize, TestWorldSize)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Добавляем системы в правильном порядке
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: behaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Создаём анимационные системы как в реальной игре
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()
	loader := animation.NewAnimationLoader()
	loader.LoadHeadlessAnimations(wolfAnimSystem, rabbitAnimSystem)

	// Создаём менеджер анимаций
	animationManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

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

	t.Logf("=== TDD: Баг анимации сна вместо поедания ===")
	t.Logf("Лучшее место с травой: (%.0f, %.0f) = %.1f единиц", grassX, grassY, maxGrass)

	// Создаём голодного зайца точно на траве
	rabbit := simulation.CreateRabbit(world, grassX, grassY)
	world.SetHunger(rabbit, core.Hunger{Value: 85.0})    // Голодный
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0}) // Стоит

	deltaTime := float32(1.0 / 60.0)

	// Симулируем много тиков чтобы поймать проблему
	for i := 0; i < 200; i++ {
		// Обновляем все системы как в реальной игре
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
		animationManager.Update(world, deltaTime)

		// Получаем состояние зайца
		pos, _ := world.GetPosition(rabbit)
		vel, _ := world.GetVelocity(rabbit)
		hunger, _ := world.GetHunger(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		isEating := world.HasComponent(rabbit, core.MaskEatingState)
		grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		// Логируем каждые 60 тиков (каждую секунду)
		if i%60 == 0 {
			t.Logf("%.1fс: голод=%.1f%% ест=%v трава=%.1f скорость=%.2f анимация=%s кадр=%d",
				float32(i)*deltaTime, hunger.Value, isEating, grassAmount, speed, animType.String(), anim.Frame)
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: если заяц ест, анимация должна быть Eat, НЕ Sleep*
		if isEating {
			if animType == animation.AnimSleepFalling || animType == animation.AnimSleepLoop || animType == animation.AnimSleepWaking {
				t.Errorf("❌ БАГ НАЙДЕН на тике %d: Заяц ест (EatingState=true) но показывает анимацию сна %s!",
					i, animType.String())
				t.Errorf("   Детали: голод=%.1f%% трава=%.1f скорость=%.2f кадр=%d",
					hunger.Value, grassAmount, speed, anim.Frame)

				// Дополнительная диагностика
				t.Logf("   Проверяем что у зайца зарегистрирована анимация Eat...")
				eatAnim := rabbitAnimSystem.GetAnimation(animation.AnimEat)
				if eatAnim == nil {
					t.Errorf("   ПРИЧИНА: Анимация AnimEat НЕ зарегистрирована для зайца!")
				} else {
					t.Logf("   Анимация AnimEat зарегистрирована: %d кадров, %.1f FPS, зацикленная=%v",
						eatAnim.Frames, eatAnim.FPS, eatAnim.Loop)
				}

				// Проверяем resolver
				resolver := animation.NewAnimationResolver()
				expectedAnim := resolver.ResolveAnimalAnimationType(world, rabbit, core.TypeRabbit)
				t.Logf("   AnimationResolver возвращает: %s (ожидаем: Eat)", expectedAnim.String())

				return
			}

			// Проверяем что анимация правильная
			if animType != animation.AnimEat {
				t.Errorf("❌ БАГ: Заяц ест но анимация %s вместо Eat", animType.String())
				return
			}

			t.Logf("✅ Заяц ест с правильной анимацией Eat на тике %d", i)
			return
		}

		// Если заяц наелся - тест завершён
		if hunger.Value >= 99.0 {
			t.Logf("✅ Заяц наелся до %.1f%% на тике %d, анимация: %s", hunger.Value, i, animType.String())
			return
		}
	}

	// Если дошли сюда - заяц так и не начал есть
	hunger, _ := world.GetHunger(rabbit)
	anim, _ := world.GetAnimation(rabbit)
	animType := animation.AnimationType(anim.CurrentAnim)
	t.Errorf("❌ Заяц не начал есть за 200 тиков. Голод: %.1f%%, анимация: %s", hunger.Value, animType.String())
}
