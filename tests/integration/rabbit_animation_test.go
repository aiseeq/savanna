package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestRabbitAnimations проверяет правильность анимаций зайца
func TestRabbitAnimations(t *testing.T) {
	t.Parallel()

	// Создаём мир
	world := core.NewWorld(TestWorldSize, TestWorldSize, 12345)

	// Создаём terrain и vegetation
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём системы
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(TestWorldSize, TestWorldSize)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Создаём анимационные системы
	rabbitAnimSystem := animation.NewAnimationSystem()
	loader := animation.NewAnimationLoader()
	loader.LoadHeadlessAnimations(animation.NewAnimationSystem(), rabbitAnimSystem)

	// Создаём менеджер анимаций
	animationManager := animation.NewAnimationManager(animation.NewAnimationSystem(), rabbitAnimSystem)

	// Найдём место с травой
	var grassX, grassY float32 = 100, 100
	for x := float32(50); x < 400; x += 50 {
		for y := float32(50); y < 400; y += 50 {
			if vegetationSystem.GetGrassAt(x, y) > 10.0 {
				grassX, grassY = x, y
				break
			}
		}
		if vegetationSystem.GetGrassAt(grassX, grassY) > 10.0 {
			break
		}
	}

	// Создаём голодного зайца
	rabbit := simulation.CreateRabbit(world, grassX, grassY)
	world.SetHunger(rabbit, core.Hunger{Value: 80.0})    // Голодный
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0}) // Стоит

	deltaTime := float32(1.0 / 60.0)

	t.Logf("=== Тест анимаций зайца ===")
	t.Logf("Начальная позиция: (%.0f, %.0f), голод: 80%%", grassX, grassY)

	// Симулируем несколько тиков
	for i := 0; i < 10; i++ {
		// Обновляем в правильном порядке
		feedingSystem.Update(world, deltaTime)    // Создаёт EatingState
		behaviorSystem.Update(world, deltaTime)   // Устанавливает скорость
		movementSystem.Update(world, deltaTime)   // Сбрасывает скорость едящих
		animationManager.Update(world, deltaTime) // Обновляет анимации

		// Получаем состояние
		vel, _ := world.GetVelocity(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		isEating := world.HasComponent(rabbit, core.MaskEatingState)

		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		t.Logf("Тик %d: скорость=%.2f ест=%v анимация=%s кадр=%d",
			i, speed, isEating, animType.String(), anim.Frame)

		// КРИТИЧЕСКИЕ ПРОВЕРКИ
		if isEating {
			// Если заяц ест, должна быть анимация Eat
			if animType != animation.AnimEat {
				t.Errorf("❌ ОШИБКА: Заяц ест но анимация %s вместо Eat", animType.String())
			}

			// Если заяц ест, скорость должна быть 0
			if speed > 0.1 {
				t.Errorf("❌ ОШИБКА: Заяц ест но движется со скоростью %.2f", speed)
			}
		} else {
			// Если заяц не ест и не движется, должна быть анимация Idle
			if speed < 0.1 && animType != animation.AnimIdle {
				t.Errorf("❌ ОШИБКА: Заяц стоит (скорость %.2f) но анимация %s вместо Idle", speed, animType.String())
			}

			// Если заяц быстро движется, должна быть анимация Run
			if speed > 100.0 && animType != animation.AnimRun {
				t.Errorf("❌ ОШИБКА: Заяц быстро бежит (скорость %.2f) но анимация %s вместо Run", speed, animType.String())
			}
		}

		// Если заяц начал есть - завершаем тест
		if isEating {
			t.Logf("✅ Заяц начал есть на тике %d с правильной анимацией %s", i, animType.String())
			return
		}
	}

	t.Logf("ℹ️  Заяц не начал есть за 10 тиков")
}
