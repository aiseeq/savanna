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

// TestGUIAnimationBug проверяет точную логику GUI анимаций как в main.go
//
//nolint:gocognit,revive,funlen // Комплексный тест GUI анимационной логики
func TestGUIAnimationBug(t *testing.T) {
	t.Parallel()

	// Создаём мир точно как в GUI main.go
	world := core.NewWorld(TestWorldSize, TestWorldSize, 12345)

	// Создаём terrain и vegetation
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём системы точно как в GUI
	systemManager := core.NewSystemManager()
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем системы в том же порядке что в GUI
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{
		System: simulation.NewMovementSystem(TestWorldSize, TestWorldSize),
	})
	systemManager.AddSystem(combatSystem)

	// Создаём анимационные системы точно как в GUI
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Имитируем loadRabbitAnimations из main.go
	rabbitAnimations := []struct {
		name     string
		frames   int
		fps      float32
		loop     bool
		animType animation.AnimationType
	}{
		{"hare_idle", 2, 2.0, true, animation.AnimIdle},
		{"hare_walk", 2, 4.0, true, animation.AnimWalk},
		{"hare_run", 2, 12.0, true, animation.AnimRun},
		{"hare_attack", 2, 5.0, false, animation.AnimAttack},
		{"hare_eat", 2, 4.0, true, animation.AnimEat}, // ЭТО ДОЛЖНО РАБОТАТЬ!
		{"hare_dead", 2, 3.0, false, animation.AnimDeathDying},
	}

	// Регистрируем все анимации (с пустыми изображениями для теста)
	for _, config := range rabbitAnimations {
		rabbitAnimationSystem.RegisterAnimation(config.animType, config.frames, config.fps, config.loop, nil)
	}

	// Создаём resolver точно как в GUI
	animationResolver := animation.NewAnimationResolver()

	// Найдём место с травой
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

	t.Logf("=== TDD: Проверка реальной GUI логики анимаций ===")
	t.Logf("Лучшее место с травой: (%.0f, %.0f) = %.1f единиц", grassX, grassY, maxGrass)

	// Создаём голодного зайца
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, grassX, grassY)
	world.SetHunger(rabbit, core.Hunger{Value: 85.0})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	deltaTime := float32(1.0 / 60.0)

	// Имитируем точную логику GUI updateSingleAnimalAnimation
	t.Logf("\n--- Проверяем точную GUI логику ---")

	for i := 0; i < 10; i++ {
		// Обновляем все системы ТОЧНО как в GUI
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Получаем состояние зайца
		animalType, _ := world.GetAnimalType(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		vel, _ := world.GetVelocity(rabbit)
		hunger, _ := world.GetHunger(rabbit)
		isEating := world.HasComponent(rabbit, core.MaskEatingState)
		grassAmount := vegetationSystem.GetGrassAt(grassX, grassY)

		speed := vel.X*vel.X + vel.Y*vel.Y
		oldAnimType := animation.AnimationType(anim.CurrentAnim)

		t.Logf("Тик %d:", i)
		t.Logf("  Состояние: голод=%.1f%% ест=%v трава=%.1f скорость=%.2f",
			hunger.Value, isEating, grassAmount, speed)
		t.Logf("  Старая анимация: %s", oldAnimType.String())

		// КРИТИЧЕСКАЯ ЧАСТЬ: имитируем getAnimalAnimationType из GUI
		var newAnimType animation.AnimationType
		if animationResolver != nil {
			newAnimType = animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)
		} else {
			newAnimType = animation.AnimIdle
		}

		t.Logf("  AnimationResolver возвращает: %s", newAnimType.String())

		// Проверяем что resolver правильно работает
		if isEating && newAnimType != animation.AnimEat {
			t.Errorf("❌ БАГ НАЙДЕН: Заяц ест (EatingState=true) но resolver возвращает %s вместо Eat",
				newAnimType.String())

			// Дополнительная диагностика
			t.Logf("  Проверяем resolver напрямую...")
			resolver := animation.NewAnimationResolver()
			directResult := resolver.ResolveAnimalAnimationType(world, rabbit, core.TypeRabbit)
			t.Logf("  Прямой вызов resolver: %s", directResult.String())

			// Проверяем что анимация зарегистрирована
			eatAnim := rabbitAnimationSystem.GetAnimation(animation.AnimEat)
			if eatAnim == nil {
				t.Errorf("  ПРИЧИНА: AnimEat НЕ зарегистрирована!")
			} else {
				t.Logf("  AnimEat зарегистрирована: %d кадров, %.1f FPS", eatAnim.Frames, eatAnim.FPS)
			}
			return
		}

		// Имитируем updateAnimationIfNeeded
		if anim.CurrentAnim != int(newAnimType) {
			// НЕ прерываем анимацию ATTACK
			if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
				t.Logf("  Не меняем анимацию - Attack играет")
			} else {
				// Обычная смена анимации
				anim.CurrentAnim = int(newAnimType)
				anim.Frame = 0
				anim.Timer = 0
				anim.Playing = true
				world.SetAnimation(rabbit, anim)
				t.Logf("  Сменили анимацию: %s -> %s", oldAnimType.String(), newAnimType.String())
			}
		} else {
			t.Logf("  Анимация не изменилась: %s", newAnimType.String())
		}

		// Проверяем финальную анимацию
		finalAnim, _ := world.GetAnimation(rabbit)
		finalAnimType := animation.AnimationType(finalAnim.CurrentAnim)
		t.Logf("  Финальная анимация: %s", finalAnimType.String())

		if isEating && finalAnimType != animation.AnimEat {
			t.Errorf("❌ БАГ: Заяц ест но финальная анимация %s", finalAnimType.String())
			return
		}

		if isEating && finalAnimType == animation.AnimEat {
			t.Logf("✅ ПРАВИЛЬНО: Заяц ест и показывает анимацию Eat")
			return
		}
	}

	t.Logf("Тест завершён")
}
