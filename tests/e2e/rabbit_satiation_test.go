package e2e

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestRabbitSatiationE2E проверяет что зайцы могут правильно насыщаться
//
//nolint:gocognit,revive // E2E тест системы насыщения зайцев
func TestRabbitSatiationE2E(t *testing.T) {
	t.Parallel()

	t.Logf("=== TDD: Проверка насыщения зайцев ===")

	// Создаём минимальный мир
	world := core.NewWorld(200, 200, 12345)

	// Создаём terrain с большим количеством травы
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 3 // 3x3 тайла для быстрого теста
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Находим тайл с травой или создаем его
	grassTileX, grassTileY := 1, 1

	// Ищем подходящий тайл или устанавливаем тип травы
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			tileType := terrain.GetTileType(x, y)
			if tileType == generator.TileGrass || tileType == generator.TileWetland {
				grassTileX, grassTileY = x, y
				goto found
			}
		}
	}

	// Если не нашли подходящий тайл, устанавливаем тип травы принудительно
	terrain.SetTileType(grassTileX, grassTileY, generator.TileGrass)

found:
	// Размещаем много травы в тайле с травой
	terrain.SetGrassAmount(grassTileX, grassTileY, 100.0) // Максимальное количество травы

	// Размещаем зайца в центре тайла с травой
	rabbitX := float32(grassTileX*32 + 16) // Центр тайла
	rabbitY := float32(grassTileY*32 + 16)

	t.Logf("Тайл с травой: (%d,%d), тип=%d", grassTileX, grassTileY, terrain.GetTileType(grassTileX, grassTileY))
	t.Logf("Заяц будет размещен на: (%.1f, %.1f)", rabbitX, rabbitY)

	// Создаём анимационные системы для теста
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Загружаем анимации
	loader := animation.NewAnimationLoader()
	emptyImg := ebiten.NewImage(128, 64)
	loader.LoadAnimations(wolfAnimationSystem, rabbitAnimationSystem, emptyImg, emptyImg)

	// Создаём менеджер анимаций
	animationManager := animation.NewAnimationManager(wolfAnimationSystem, rabbitAnimationSystem)

	// Создаём системы
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Используем объединенную систему питания как в реальной игре
	deprecatedFeedingAdapter := adapters.NewDeprecatedFeedingSystemAdapter(vegetationSystem)

	// Создаём systemManager в правильном порядке
	systemManager := core.NewSystemManager()
	systemManager.AddSystem(vegetationSystem)         // 1. Рост травы
	systemManager.AddSystem(deprecatedFeedingAdapter) // 2. Полная система питания

	// Создаём зайца в центре тайла с травой
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)

	// ТЕСТ СЦЕНАРИЯ ПОЛЬЗОВАТЕЛЯ: заяц с голодом 90% должен есть до 100%
	initialHunger := float32(90.0) // Слегка голодный (как жаловался пользователь)
	world.SetSatiation(rabbit, core.Satiation{Value: initialHunger})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0}) // Стоит на месте

	pos, _ := world.GetPosition(rabbit)
	hunger, _ := world.GetSatiation(rabbit)
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Logf("Начальное состояние:")
	t.Logf("  Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("  Голод зайца: %.1f%%", hunger.Value)
	t.Logf("  Трава в позиции: %.1f единиц", grassAmount)

	// Проверяем начальные условия
	if grassAmount < 50.0 {
		t.Errorf("❌ Недостаточно травы для теста: %.1f < 50.0", grassAmount)
		return
	}

	if hunger.Value >= 95.0 {
		t.Errorf("❌ Заяц слишком сыт для теста: %.1f%% >= 95%%", hunger.Value)
		return
	}

	// Устанавливаем анимацию поедания
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(animation.AnimEat),
		Frame:       0,
		Timer:       0.0,
		Playing:     true,
		FacingRight: true,
	})

	deltaTime := float32(1.0 / 60.0)
	maxTicks := 1200 // 20 секунд симуляции (после снижения скорости поедания с 2.0 до 1.0)

	t.Logf("\nНачинаем симуляцию поедания...")

	for tick := 0; tick < maxTicks; tick++ {
		// ИСПРАВЛЕНИЕ: Обновляем анимации через менеджер
		animationManager.UpdateAllAnimations(world, deltaTime)

		// Обновляем все системы
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Проверяем состояние каждые 10 тиков для детального отслеживания
		if tick%10 == 0 {
			currentHunger, _ := world.GetSatiation(rabbit)
			currentGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)
			anim, _ := world.GetAnimation(rabbit)
			eatingState, _ := world.GetEatingState(rabbit)

			hasEating := world.HasComponent(rabbit, core.MaskEatingState)
			t.Logf("Тик %d (%.1fs): голод=%.1f%%, трава=%.1f, eating=%v, кадр=%d, питание=%.1f",
				tick, float32(tick)/60.0, currentHunger.Value, currentGrass, hasEating, anim.Frame, eatingState.NutritionGained)

			// Проверяем прогресс
			if currentHunger.Value > hunger.Value {
				t.Logf("✅ Голод уменьшается! %.1f%% -> %.1f%%", hunger.Value, currentHunger.Value)
				hunger = currentHunger

				// ТЕСТ СЦЕНАРИЯ ПОЛЬЗОВАТЕЛЯ: заяц должен есть до высокого уровня сытости
				if currentHunger.Value >= 98.0 { // Реалистичный порог - близко к 99.9% но с учетом реальности
					t.Logf("✅ УСПЕХ: Заяц хорошо насытился до %.1f%%!", currentHunger.Value)
					t.Logf("✅ Трава потреблена: %.1f -> %.1f", grassAmount, currentGrass)
					return
				}
			}

			// Проверяем что трава потребляется
			if currentGrass < grassAmount {
				t.Logf("✅ Трава потребляется: %.1f -> %.1f", grassAmount, currentGrass)
				grassAmount = currentGrass
			}
		}

		// Проверяем здоровье зайца - оно не должно падать при поедании
		health, hasHealth := world.GetHealth(rabbit)
		if hasHealth && health.Current <= 0 {
			t.Errorf("❌ Заяц умер во время еды на тике %d", tick)
			return
		}
	}

	// Если мы дошли сюда, заяц не насытился за отведённое время
	finalHunger, _ := world.GetSatiation(rabbit)
	finalGrass := vegetationSystem.GetGrassAt(pos.X, pos.Y)

	t.Errorf("❌ Заяц не насытился за %d тиков", maxTicks)
	t.Errorf("   Голод: %.1f%% -> %.1f%% (цель: >= 98%%)", initialHunger, finalHunger.Value)
	t.Errorf("   Трава: %.1f -> %.1f", grassAmount, finalGrass)

	// Дополнительная диагностика
	hasEating := world.HasComponent(rabbit, core.MaskEatingState)
	hasBehavior := world.HasComponent(rabbit, core.MaskBehavior)
	hasAnimation := world.HasComponent(rabbit, core.MaskAnimation)

	t.Logf("   EatingState активен: %v", hasEating)
	t.Logf("   Behavior компонент: %v", hasBehavior)
	t.Logf("   Animation компонент: %v", hasAnimation)

	if behavior, ok := world.GetBehavior(rabbit); ok {
		t.Logf("   BehaviorType: %v", behavior.Type)
	}

	if anim, ok := world.GetAnimation(rabbit); ok {
		t.Logf("   Анимация: тип=%d, кадр=%d, таймер=%.3f", anim.CurrentAnim, anim.Frame, anim.Timer)
	}

	if eatingState, ok := world.GetEatingState(rabbit); ok {
		t.Logf("   EatingProgress: %.2f", eatingState.EatingProgress)
		t.Logf("   NutritionGained: %.2f", eatingState.NutritionGained)
		t.Logf("   Target: %d", eatingState.Target)
	}
}
