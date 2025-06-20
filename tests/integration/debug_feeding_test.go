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

// TestDebugFeeding детальная диагностика почему не создается EatingState
//
//nolint:gocognit,revive,funlen // Комплексный диагностический тест создания EatingState
func TestDebugFeeding(t *testing.T) {
	t.Parallel()

	t.Logf("=== ДЕБАГ ПИТАНИЯ ===")

	// Создаём минимальный мир
	world := core.NewWorld(64, 64, 12345)

	// Создаём terrain с травой
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 2 // 2x2 тайла
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Устанавливаем траву (ИСПРАВЛЕНИЕ: сначала тип тайла, потом количество)
	terrain.SetTileType(1, 1, generator.TileGrass) // Устанавливаем тип "трава"
	terrain.SetGrassAmount(1, 1, 100.0)            // Много травы

	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	hungerSystem := simulation.NewHungerSystem()                           // 1. Только управление голодом
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem) // 2. Только поиск травы и создание EatingState
	hungerSpeedModifier := simulation.NewHungerSpeedModifierSystem()       // 3. Только влияние голода на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()             // 4. Только урон от голода

	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)

	// Создаём systemManager в правильном порядке
	systemManager := core.NewSystemManager()
	systemManager.AddSystem(vegetationSystem)              // 1. Рост травы
	systemManager.AddSystem(&adapters.HungerSystemAdapter{ // 2. Управление голодом
		System: hungerSystem,
	})
	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{ // 3. Создание EatingState
		System: grassSearchSystem,
	})
	systemManager.AddSystem(&adapters.GrassEatingSystemAdapter{System: grassEatingSystem}) // 4. Дискретное поедание травы
	// 5. Поведение (проверяет EatingState)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: behaviorSystem})
	systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{ // 6. Влияние голода на скорость
		System: hungerSpeedModifier,
	})
	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{ // 7. Урон от голода
		System: starvationDamage,
	})

	// Создаём зайца в центре тайла с травой
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 48, 48) // Центр тайла (1,1)

	// Делаем зайца голодным
	world.SetHunger(rabbit, core.Hunger{Value: 50.0}) // 50% < 90%
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetHunger(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("Голод зайца: %.1f%% (должен быть < %.1f%%)", hunger.Value, simulation.RabbitHungerThreshold)
	t.Logf("Трава в позиции: %.1f единиц (минимум %.1f)", grassAmount, simulation.MinGrassAmountToFind)

	// Проверяем все условия вручную
	if hunger.Value >= simulation.RabbitHungerThreshold {
		t.Errorf("❌ Заяц слишком сыт: %.1f%% >= %.1f%%", hunger.Value, simulation.RabbitHungerThreshold)
	} else {
		t.Logf("✅ Заяц голоден: %.1f%% < %.1f%%", hunger.Value, simulation.RabbitHungerThreshold)
	}

	if grassAmount < simulation.MinGrassAmountToFind {
		t.Errorf("❌ Недостаточно травы: %.1f < %.1f", grassAmount, simulation.MinGrassAmountToFind)
	} else {
		t.Logf("✅ Достаточно травы: %.1f >= %.1f", grassAmount, simulation.MinGrassAmountToFind)
	}

	deltaTime := float32(1.0 / 60.0)
	grassToEat := simulation.GrassPerEatingTick * deltaTime
	t.Logf("Должно съесть травы: %.6f за тик", grassToEat)

	// Симулируем ConsumeGrassAt
	consumedGrass := vegetationSystem.ConsumeGrassAt(pos.X, pos.Y, grassToEat)
	t.Logf("Реально съедено травы: %.6f", consumedGrass)

	if consumedGrass > 0 {
		t.Logf("✅ Трава съедена - EatingState должен создаться")

		// Проверяем нет ли уже EatingState
		hasEating := world.HasComponent(rabbit, core.MaskEatingState)
		t.Logf("EatingState до создания: %v", hasEating)

		if !hasEating {
			// Имитируем создание EatingState
			eatingState := core.EatingState{
				Target:          0,
				TargetType:      core.EatingTargetGrass, // Тип: поедание травы
				EatingProgress:  0.0,
				NutritionGained: 0.0,
			}
			world.AddEatingState(rabbit, eatingState)
			t.Logf("✅ EatingState создан вручную")
		}

		hasEatingAfter := world.HasComponent(rabbit, core.MaskEatingState)
		t.Logf("EatingState после создания: %v", hasEatingAfter)

	} else {
		t.Errorf("❌ Трава НЕ съедена - EatingState не создастся")
	}

	// Теперь симулируем полный тик FeedingSystem
	t.Logf("\n--- Полный тик FeedingSystem ---")

	// Восстанавливаем состояние
	terrain.SetGrassAmount(1, 1, 100.0)
	world.SetHunger(rabbit, core.Hunger{Value: 50.0})
	if world.HasComponent(rabbit, core.MaskEatingState) {
		world.RemoveEatingState(rabbit)
	}

	isEatingBefore := world.HasComponent(rabbit, core.MaskEatingState)
	grassBefore := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	t.Logf("ДО FeedingSystem: EatingState=%v, Трава=%.1f", isEatingBefore, grassBefore)

	// Устанавливаем анимацию поедания (необходимо для GrassEatingSystem)
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(animation.AnimEat), // AnimEat
		Frame:       0,
		Timer:       0.0,
		Playing:     true,
		FacingRight: true,
	})

	// Запускаем достаточное количество тиков для завершения кадра анимации
	// Кадр длится 0.25 сек, deltaTime = 0.017 сек => нужно ~15 тиков
	for i := 0; i < 20; i++ {
		// Обновляем таймер анимации
		if anim, hasAnim := world.GetAnimation(rabbit); hasAnim {
			anim.Timer += deltaTime
			world.SetAnimation(rabbit, anim)
		}

		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Проверяем потребилась ли трава
		grassCurrent := vegetationSystem.GetGrassAt(pos.X, pos.Y)
		if grassCurrent < grassBefore {
			t.Logf("✅ Трава потреблена на тике %d: %.1f -> %.1f", i, grassBefore, grassCurrent)
			break
		}
	}

	isEatingAfter := world.HasComponent(rabbit, core.MaskEatingState)
	grassAfter := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	grassConsumed := grassBefore - grassAfter
	t.Logf("ПОСЛЕ FeedingSystem: EatingState=%v, Трава=%.1f, Съедено=%.6f", isEatingAfter, grassAfter, grassConsumed)

	// ИСПРАВЛЕНИЕ: FeedingSystem только создает EatingState, а GrassEatingSystem съедает траву
	if isEatingAfter && grassConsumed == 0 {
		t.Logf("✅ ПРАВИЛЬНО: FeedingSystem создал EatingState без съедания травы")
		t.Logf("   Трава будет съедена GrassEatingSystem дискретно по кадрам анимации")
	} else if isEatingAfter && grassConsumed > 0 {
		t.Errorf("❌ БАГ: FeedingSystem съел траву (%.6f) вместо GrassEatingSystem!", grassConsumed)
	} else if !isEatingAfter {
		t.Errorf("❌ БАГ: FeedingSystem не создал EatingState!")
	}
}
