package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestAnimationVelocitySync проверяет синхронизацию анимации и скорости
//
//nolint:revive // function-length: Комплексный тест синхронизации анимаций
func TestAnimationVelocitySync(t *testing.T) {
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
	_ = simulation.NewMovementSystem(TestWorldSize, TestWorldSize) // Не используется в этом тесте

	// Создаём анимационные системы
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()
	loader := animation.NewAnimationLoader()
	loader.LoadHeadlessAnimations(wolfAnimSystem, rabbitAnimSystem)

	// Создаём менеджер анимаций
	animationManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	t.Logf("=== Тест синхронизации анимации и скорости ===")

	// СЦЕНАРИЙ 1: Заяц стоит на месте
	t.Logf("\n--- Сценарий 1: Заяц стоит на месте ---")
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 100, 100)
	world.SetVelocity(rabbit1, core.Velocity{X: 0, Y: 0}) // Принудительно стоит

	deltaTime := float32(1.0 / 60.0)

	for i := 0; i < 5; i++ {
		animationManager.Update(world, deltaTime)

		vel, _ := world.GetVelocity(rabbit1)
		anim, _ := world.GetAnimation(rabbit1)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		t.Logf("Тик %d: скорость=%.2f анимация=%s", i, speed, animType.String())

		if speed < 0.1 && animType != animation.AnimIdle {
			t.Errorf("❌ ОШИБКА: Заяц стоит (скорость %.2f) но анимация %s", speed, animType.String())
		}
	}

	// СЦЕНАРИЙ 2: Заяц быстро бежит, потом резко останавливается
	t.Logf("\n--- Сценарий 2: Заяц бежит, потом останавливается ---")
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)

	// Заставляем зайца быстро бежать
	world.SetVelocity(rabbit2, core.Velocity{X: 20, Y: 0}) // Быстрый бег

	for i := 0; i < 3; i++ {
		animationManager.Update(world, deltaTime)

		vel, _ := world.GetVelocity(rabbit2)
		anim, _ := world.GetAnimation(rabbit2)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		t.Logf("Тик %d (бег): скорость=%.2f анимация=%s", i, speed, animType.String())
	}

	// Резко останавливаем зайца
	world.SetVelocity(rabbit2, core.Velocity{X: 0, Y: 0})

	for i := 0; i < 5; i++ {
		animationManager.Update(world, deltaTime)

		vel, _ := world.GetVelocity(rabbit2)
		anim, _ := world.GetAnimation(rabbit2)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		t.Logf("Тик %d (стоп): скорость=%.2f анимация=%s", i, speed, animType.String())

		// КРИТИЧЕСКАЯ ПРОВЕРКА: анимация должна измениться на Idle сразу после остановки
		if speed < 0.1 && animType != animation.AnimIdle {
			t.Errorf("❌ ОШИБКА на тике %d: Заяц остановился (скорость %.2f) но всё ещё показывает анимацию %s",
				i, speed, animType.String())
		}
	}

	// СЦЕНАРИЙ 3: Заяц убегает от волка, потом волк уходит
	t.Logf("\n--- Сценарий 3: Заяц убегает от волка, потом волк уходит ---")
	rabbit3 := simulation.CreateAnimal(world, core.TypeRabbit, 300, 300)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 310, 300) // Близко к зайцу
	world.SetHunger(wolf, core.Hunger{Value: 50.0})                 // Голодный волк

	// Симулируем поведение
	for i := 0; i < 5; i++ {
		behaviorSystem.Update(world, deltaTime) // Заяц убегает
		animationManager.Update(world, deltaTime)

		vel, _ := world.GetVelocity(rabbit3)
		anim, _ := world.GetAnimation(rabbit3)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		t.Logf("Тик %d (убегает): скорость=%.2f анимация=%s", i, speed, animType.String())
	}

	// Убираем волка (телепортируем далеко)
	world.SetPosition(wolf, core.Position{X: 1000, Y: 1000})

	// Заяц должен успокоиться
	for i := 0; i < 10; i++ {
		behaviorSystem.Update(world, deltaTime) // Заяц перестаёт убегать
		animationManager.Update(world, deltaTime)

		vel, _ := world.GetVelocity(rabbit3)
		anim, _ := world.GetAnimation(rabbit3)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animType := animation.AnimationType(anim.CurrentAnim)

		t.Logf("Тик %d (спокоен): скорость=%.2f анимация=%s", i, speed, animType.String())

		// Если заяц успокоился - анимация должна быть Idle или медленная Walk
		if speed < 50.0 && animType == animation.AnimRun {
			t.Errorf("❌ ОШИБКА на тике %d: Заяц спокоен (скорость %.2f) но показывает анимацию бега",
				i, speed)
		}
	}

	t.Logf("✅ Тест синхронизации завершён")
}
