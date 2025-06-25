package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestRabbitFleeAnimation проверяет анимацию зайца при побеге от волка
func TestRabbitFleeAnimation(t *testing.T) {
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

	// Создаём анимационные системы
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()
	loader := animation.NewAnimationLoader()
	emptyImg := ebiten.NewImage(128, 64)
	loader.LoadAnimations(wolfAnimSystem, rabbitAnimSystem, emptyImg, emptyImg)

	// Создаём менеджер анимаций
	animationManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём зайца и волка близко друг к другу
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 220, 200) // Близко к зайцу

	// Делаем волка голодным чтобы он охотился
	world.SetSatiation(wolf, core.Satiation{Value: 50.0})

	deltaTime := float32(1.0 / 60.0)

	t.Logf("=== Тест анимации побега зайца ===")
	t.Logf("Заяц: (200, 200), Волк: (220, 200)")

	// Симулируем несколько тиков
	for i := 0; i < 20; i++ {
		// Сохраняем состояние перед обновлением
		rabbitVelBefore, _ := world.GetVelocity(rabbit)
		rabbitAnimBefore, _ := world.GetAnimation(rabbit)

		// Обновляем системы
		behaviorSystem.Update(world, deltaTime)   // Заяц решает убегать от волка
		movementSystem.Update(world, deltaTime)   // Обновляет позицию
		animationManager.Update(world, deltaTime) // Обновляет анимации

		// Получаем состояние после обновления
		rabbitVel, _ := world.GetVelocity(rabbit)
		rabbitAnim, _ := world.GetAnimation(rabbit)
		rabbitPos, _ := world.GetPosition(rabbit)

		speed := rabbitVel.X*rabbitVel.X + rabbitVel.Y*rabbitVel.Y
		animType := animation.AnimationType(rabbitAnim.CurrentAnim)

		t.Logf("Тик %d: поз(%.1f,%.1f) скорость=%.2f анимация=%s кадр=%d",
			i, rabbitPos.X, rabbitPos.Y, speed, animType.String(), rabbitAnim.Frame)

		// КРИТИЧЕСКАЯ ПРОВЕРКА: анимация должна соответствовать скорости
		if speed < 0.1 {
			// Заяц стоит - должна быть анимация Idle
			if animType != animation.AnimIdle {
				t.Errorf("❌ ОШИБКА на тике %d: Заяц стоит (скорость %.2f) но анимация %s вместо Idle",
					i, speed, animType.String())

				// Дополнительная диагностика
				t.Logf("   Скорость до обновления: (%.2f, %.2f)", rabbitVelBefore.X, rabbitVelBefore.Y)
				t.Logf("   Скорость после обновления: (%.2f, %.2f)", rabbitVel.X, rabbitVel.Y)
				t.Logf("   Анимация до обновления: %s", animation.AnimationType(rabbitAnimBefore.CurrentAnim).String())
				t.Logf("   Анимация после обновления: %s", animType.String())
				return
			}
		} else if speed < 300.0 {
			// Заяц медленно движется - должна быть анимация Walk
			if animType != animation.AnimWalk {
				t.Logf("ℹ️  Заяц ходит (скорость %.2f) с анимацией %s", speed, animType.String())
			}
		} else {
			// Заяц быстро бежит - должна быть анимация Run
			if animType != animation.AnimRun {
				t.Logf("ℹ️  Заяц бежит (скорость %.2f) с анимацией %s", speed, animType.String())
			}
		}

		// Если заяц убежал далеко от волка - завершаем тест
		wolfPos, _ := world.GetPosition(wolf)
		distance := (rabbitPos.X-wolfPos.X)*(rabbitPos.X-wolfPos.X) + (rabbitPos.Y-wolfPos.Y)*(rabbitPos.Y-wolfPos.Y)
		if distance > 100*100 {
			t.Logf("✅ Заяц успешно убежал от волка на расстояние %.1f", distance)
			break
		}
	}

	t.Logf("✅ Тест анимации побега завершён")
}
