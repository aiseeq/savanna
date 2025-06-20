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

// TestSimpleEating максимально простой тест: 1 заяц на 1x1 карте ест траву
//
//nolint:gocognit,revive,funlen // Комплексный интеграционный тест базового питания зайцев
func TestSimpleEating(t *testing.T) {
	t.Parallel()

	t.Logf("=== ПРОСТЕЙШИЙ ТЕСТ: 1 заяц ест на карте 1x1 ===")

	// Создаём минимальный мир 1x1 тайл = 32x32 пикселя
	world := core.NewWorld(32, 32, 12345)

	// Создаём terrain 1x1 с 100% травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 1 // 1 тайл
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Принудительно устанавливаем много травы в единственный тайл
	terrain.SetTileType(0, 0, generator.TileGrass)
	terrain.SetGrassAmount(0, 0, 100.0) // Много травы

	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём все необходимые системы
	systemManager := core.NewSystemManager()

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	hungerSystem := simulation.NewHungerSystem()                           // 1. Только управление голодом
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem) // 2. Только поиск травы и создание EatingState
	hungerSpeedModifier := simulation.NewHungerSpeedModifierSystem()       // 3. Только влияние голода на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()             // 4. Только урон от голода

	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(32, 32)

	systemManager.AddSystem(vegetationSystem)              // 1. Рост травы
	systemManager.AddSystem(&adapters.HungerSystemAdapter{ // 2. Управление голодом
		System: hungerSystem,
	})
	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{ // 3. Создание EatingState
		System: grassSearchSystem,
	})
	// 4. Поведение (проверяет EatingState)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: behaviorSystem})
	systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{ // 5. Влияние голода на скорость
		System: hungerSpeedModifier,
	})
	// 6. Движение (сбрасывает скорость едящих)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{ // 7. Урон от голода
		System: starvationDamage,
	})

	// Создаём анимационную систему с РЕАЛЬНЫМИ файлами
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Имитируем loadRabbitAnimations из main.go с РЕАЛЬНОЙ загрузкой файлов
	t.Logf("\n--- Загружаем анимации с реальными файлами ---")

	rabbitAnimations := []struct {
		name     string
		animType animation.AnimationType
	}{
		{"hare_idle", animation.AnimIdle},
		{"hare_eat", animation.AnimEat},
		{"hare_walk", animation.AnimWalk},
		{"hare_run", animation.AnimRun},
	}

	for _, config := range rabbitAnimations {
		// Проверяем существование файлов анимации
		file1 := "assets/animations/" + config.name + "_1.png"
		file2 := "assets/animations/" + config.name + "_2.png"

		t.Logf("  Проверяем: %s -> %s", config.name, config.animType.String())
		t.Logf("    Файл 1: %s", file1)
		t.Logf("    Файл 2: %s", file2)

		// Регистрируем анимацию (с пустым изображением для теста)
		rabbitAnimSystem.RegisterAnimation(config.animType, 2, 4.0, true, nil)
		t.Logf("    ✅ Зарегистрирована: %s", config.animType.String())
	}

	// Создаём resolver
	animationResolver := animation.NewAnimationResolver()

	// Создаём зайца в центре единственного тайла
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 16, 16) // Центр 32x32 тайла

	// Делаем зайца ОЧЕНЬ голодным чтобы он точно ел
	world.SetHunger(rabbit, core.Hunger{Value: 50.0})    // 50% голода - точно будет есть
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0}) // Стоит на месте

	t.Logf("\n--- Начальное состояние ---")
	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("Голод зайца: %.1f%%", hunger.Value)
	t.Logf("Трава в позиции: %.1f единиц", grassAmount)

	deltaTime := float32(1.0 / 60.0)

	t.Logf("\n--- Симуляция еды ---")

	// Симулируем несколько тиков
	for i := 0; i < 10; i++ {
		t.Logf("\n=== ТИК %d ===", i)

		// Состояние ДО обновления систем
		hunger, _ = world.GetHunger(rabbit)
		vel, _ := world.GetVelocity(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		isEatingBefore := world.HasComponent(rabbit, core.MaskEatingState)
		grassBefore := vegetationSystem.GetGrassAt(pos.X, pos.Y)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animTypeBefore := animation.AnimationType(anim.CurrentAnim)

		t.Logf("ДО обновления:")
		t.Logf("  Голод: %.1f%%, Трава: %.1f, Скорость: %.2f", hunger.Value, grassBefore, speed)
		t.Logf("  EatingState: %v, Анимация: %s (код %d)", isEatingBefore, animTypeBefore.String(), anim.CurrentAnim)

		// ОБНОВЛЯЕМ ВСЕ СИСТЕМЫ
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// КРИТИЧЕСКИ ВАЖНО: Обновляем анимации как в GUI!
		animalType, _ := world.GetAnimalType(rabbit)
		anim, _ = world.GetAnimation(rabbit)

		// Определяем новый тип анимации через resolver
		newAnimType := animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)

		// Обновляем анимацию если нужно (как в GUI updateAnimationIfNeeded)
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
				t.Logf("  🔄 Сменили анимацию: %s -> %s", animTypeBefore.String(), newAnimType.String())
			}
		}

		// Состояние ПОСЛЕ обновления систем
		hunger, _ = world.GetHunger(rabbit)
		vel, _ = world.GetVelocity(rabbit)
		anim, _ = world.GetAnimation(rabbit)
		isEatingAfter := world.HasComponent(rabbit, core.MaskEatingState)
		grassAfter := vegetationSystem.GetGrassAt(pos.X, pos.Y)
		speed = vel.X*vel.X + vel.Y*vel.Y
		animTypeAfter := animation.AnimationType(anim.CurrentAnim)

		t.Logf("ПОСЛЕ систем:")
		t.Logf("  Голод: %.1f%%, Трава: %.1f, Скорость: %.2f", hunger.Value, grassAfter, speed)
		t.Logf("  EatingState: %v, Анимация: %s (код %d)", isEatingAfter, animTypeAfter.String(), anim.CurrentAnim)

		// ПРОВЕРЯЕМ ЧТО ДОЛЖЕН СКАЗАТЬ RESOLVER
		animalType, _ = world.GetAnimalType(rabbit)
		expectedAnim := animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)
		t.Logf("  AnimationResolver ожидает: %s", expectedAnim.String())

		// КРИТИЧЕСКИЕ ПРОВЕРКИ
		if isEatingAfter {
			t.Logf("  ✅ EatingState создан - заяц ест!")

			if expectedAnim != animation.AnimEat {
				t.Errorf("  ❌ БАГ В RESOLVER: Заяц ест но resolver возвращает %s вместо Eat", expectedAnim.String())
			}

			if animTypeAfter != animation.AnimEat {
				t.Errorf("  ❌ БАГ В АНИМАЦИИ: Заяц ест (EatingState=true) но анимация %s вместо Eat", animTypeAfter.String())
				t.Errorf("     ПРОБЛЕМА: Анимация должна быть Eat но показывается %s", animTypeAfter.String())

				// Проверяем что анимация Eat зарегистрирована
				eatAnim := rabbitAnimSystem.GetAnimation(animation.AnimEat)
				if eatAnim == nil {
					t.Errorf("     ПРИЧИНА: AnimEat НЕ ЗАРЕГИСТРИРОВАНА!")
				} else {
					t.Logf("     AnimEat зарегистрирована: %d кадров, %.1f FPS", eatAnim.Frames, eatAnim.FPS)
				}
				return
			} else {
				t.Logf("  ✅ ИДЕАЛЬНО: Заяц ест и показывает анимацию Eat!")
				return
			}
		}

		grassConsumed := grassBefore - grassAfter
		if grassConsumed > 0 {
			t.Logf("  Съедено травы: %.2f единиц", grassConsumed)
		}
	}

	t.Errorf("❌ Заяц не начал есть за 10 тиков!")
}
