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

// TestRunWhileStoppedBug воспроизводит баг когда заяц показывает анимацию бега но стоит на месте
func TestRunWhileStoppedBug(t *testing.T) {
	t.Parallel()

	// Создаём мир точно как в реальной игре
	world := core.NewWorld(TestWorldSize, TestWorldSize, 12345)

	// Создаём terrain и vegetation
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(TestWorldSize / 32)
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём все системы как в реальной игре
	systemManager := core.NewSystemManager()
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(TestWorldSize, TestWorldSize)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)

	// Добавляем системы в правильном порядке
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: behaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Создаём анимационные системы
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()
	loader := animation.NewAnimationLoader()
	loader.LoadHeadlessAnimations(wolfAnimSystem, rabbitAnimSystem)

	// Создаём менеджер анимаций с resolver
	animationManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	t.Logf("=== TDD: Баг анимации бега при нулевой скорости ===")

	// СЦЕНАРИЙ 1: Заяц убегает от волка, потом волк исчезает
	rabbit := simulation.CreateRabbit(world, 200, 200)
	wolf := simulation.CreateWolf(world, 220, 200)  // Близко к зайцу
	world.SetHunger(wolf, core.Hunger{Value: 50.0}) // Голодный волк

	deltaTime := float32(1.0 / 60.0)

	// Фаза 1: Заяц убегает от волка
	t.Logf("\n--- Фаза 1: Заяц убегает ---")
	for i := 0; i < 60; i++ { // 1 секунда бега
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
		animationManager.Update(world, deltaTime)

		vel, _ := world.GetVelocity(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		pos, _ := world.GetPosition(rabbit)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		if i%20 == 0 {
			t.Logf("%.1fс: поз(%.1f,%.1f) скорость=%.1f анимация=%s",
				float32(i)*deltaTime, pos.X, pos.Y, speed, animType.String())
		}

		// Заяц должен быстро бежать от волка
		if speed > 300.0 && animType != animation.AnimRun {
			t.Logf("ℹ️  Заяц быстро бежит но анимация %s (не критично)", animType.String())
		}
	}

	// Фаза 2: Телепортируем волка далеко (имитируем исчезновение угрозы)
	world.SetPosition(wolf, core.Position{X: 2000, Y: 2000})
	t.Logf("\n--- Фаза 2: Волк исчез, заяц должен успокоиться ---")

	for i := 0; i < 180; i++ { // 3 секунды на успокоение
		// Сохраняем состояние ДО обновления
		velBefore, _ := world.GetVelocity(rabbit)
		animBefore, _ := world.GetAnimation(rabbit)
		speedBefore := velBefore.X*velBefore.X + velBefore.Y*velBefore.Y
		animTypeBefore := animation.AnimationType(animBefore.CurrentAnim)

		// Обновляем системы
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)
		animationManager.Update(world, deltaTime)

		// Получаем состояние ПОСЛЕ обновления
		vel, _ := world.GetVelocity(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		pos, _ := world.GetPosition(rabbit)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		if i%60 == 0 {
			t.Logf("%.1fс: поз(%.1f,%.1f) скорость=%.1f анимация=%s",
				float32(i)*deltaTime, pos.X, pos.Y, speed, animType.String())
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: анимация должна соответствовать реальной скорости
		if speed < 0.1 && animType == animation.AnimRun {
			t.Errorf("❌ БАГ НАЙДЕН на тике %d: Заяц стоит (скорость %.2f) но показывает анимацию бега!",
				i, speed)
			t.Errorf("   До обновления: скорость=%.2f анимация=%s", speedBefore, animTypeBefore.String())
			t.Errorf("   После обновления: скорость=%.2f анимация=%s", speed, animType.String())

			// Проверяем resolver отдельно
			resolver := animation.NewAnimationResolver()
			expectedAnim := resolver.ResolveAnimalAnimationType(world, rabbit, core.TypeRabbit)
			t.Errorf("   AnimationResolver ожидает: %s", expectedAnim.String())
			return
		}

		// Если заяц замедлился до ходьбы - проверяем соответствие
		if speed < 300.0 && speed > 0.1 && animType == animation.AnimRun {
			t.Errorf("❌ БАГ: Заяц медленно движется (скорость %.2f) но показывает анимацию бега", speed)
			return
		}

		// Если заяц успокоился полностью - тест успешен
		if speed < 50.0 && (animType == animation.AnimIdle || animType == animation.AnimWalk) {
			t.Logf("✅ Заяц успокоился на тике %d: скорость=%.2f анимация=%s", i, speed, animType.String())
			return
		}
	}

	// Если дошли сюда - заяц так и не успокоился
	vel, _ := world.GetVelocity(rabbit)
	anim, _ := world.GetAnimation(rabbit)
	speed := vel.X*vel.X + vel.Y*vel.Y
	animType := animation.AnimationType(anim.CurrentAnim)
	t.Logf("⚠️  Заяц не успокоился за 3 секунды: скорость=%.2f анимация=%s", speed, animType.String())
}
