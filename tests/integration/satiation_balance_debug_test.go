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

// TestSatiationBalanceDebug проверяет точный баланс сытости во время еды
//
//nolint:revive // function-length: Детальный тест баланса сытости с множественными проверками
func TestSatiationBalanceDebug(t *testing.T) {
	t.Parallel()

	t.Logf("=== Отладка баланса сытости: потеря vs восстановление ===")

	cfg := config.LoadDefaultConfig()
	worldWidth := float32(cfg.World.Size * 32)
	worldHeight := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldWidth, worldHeight, 12345)

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	systemManager := core.NewSystemManager()

	// Создаём системы
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)

	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(adapters.NewFeedingSystemAdapter(vegetationSystem))
	systemManager.AddSystem(grassEatingSystem)

	// Создаём анимационную систему как в игре
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimIdle, 2, 2.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimWalk, 2, 8.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimRun, 2, 12.0, true, nil)
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)

	// Создаём анимационный resolver
	animationResolver := animation.NewAnimationResolver()

	// Создаём зайца
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, 200, 200)

	// Устанавливаем траву
	tileX := int(200 / 32)
	tileY := int(200 / 32)
	terrain.SetTileType(tileX, tileY, generator.TileGrass)
	terrain.SetGrassAmount(tileX, tileY, 100.0)

	// Делаем зайца голодным
	initialSatiation := float32(60.0)
	world.SetSatiation(rabbit, core.Satiation{Value: initialSatiation})
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	// Создаём EatingState принудительно
	world.AddEatingState(rabbit, core.EatingState{
		Target:          0,
		TargetType:      core.EatingTargetGrass, // Тип: поедание травы
		EatingProgress:  0.0,
		NutritionGained: 0.0,
	})

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	t.Logf("Начальное состояние:")
	t.Logf("  Сытость: %.2f%%", initialSatiation)
	t.Logf("  Скорость потери сытости: %.2f%%/сек", 2.0)
	t.Logf("  За кадр анимации (0.25с): %.2f%% потери", 2.0*0.25)

	// Симулируем ПОЛНЫЙ цикл анимации (30 тиков = 2 кадра)
	satiationHistory := make([]float32, 0, 31)
	frameHistory := make([]int, 0, 31)
	satiation, _ := world.GetSatiation(rabbit)
	satiationHistory = append(satiationHistory, satiation.Value)
	anim, _ := world.GetAnimation(rabbit)
	frameHistory = append(frameHistory, anim.Frame)

	for tick := 0; tick < 30; tick++ {
		// Порядок как в реальной игре
		world.Update(deltaTime)

		// Обновляем анимации ПРАВИЛЬНО через AnimationSystem
		animalType, _ := world.GetAnimalType(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		newAnimType := animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)

		// Меняем анимацию если нужно
		if anim.CurrentAnim != int(newAnimType) {
			anim.CurrentAnim = int(newAnimType)
			anim.Frame = 0
			anim.Timer = 0
			anim.Playing = true
			world.SetAnimation(rabbit, anim)
		}

		// КРИТИЧЕСКИ ВАЖНО: Используем AnimationSystem.Update для правильного обновления кадров!
		if anim.Playing {
			// Конвертируем в AnimationComponent как в AnimationManager
			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			// Обновляем через систему анимации
			animationSystem.Update(&animComponent, deltaTime)

			// Сохраняем обновлённое состояние
			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			anim.FacingRight = animComponent.FacingRight
			world.SetAnimation(rabbit, anim)
		}

		systemManager.Update(world, deltaTime)

		// Записываем сытость и кадр после каждого тика
		currentSatiation, _ := world.GetSatiation(rabbit)
		currentAnim, _ := world.GetAnimation(rabbit)
		satiationHistory = append(satiationHistory, currentSatiation.Value)
		frameHistory = append(frameHistory, currentAnim.Frame)

		t.Logf("Тик %2d: сытость=%.3f%%, таймер=%.3f, кадр=%d", tick, currentSatiation.Value, currentAnim.Timer, currentAnim.Frame)
	}

	// Анализируем изменения сытости
	t.Logf("\n=== АНАЛИЗ ИЗМЕНЕНИЙ ГОЛОДА ===")
	t.Logf("История сытости: %v", satiationHistory)

	// Ищем момент когда сытость изменился
	for i := 1; i < len(satiationHistory); i++ {
		change := satiationHistory[i] - satiationHistory[i-1]
		if change != 0 {
			t.Logf("Тик %d: изменение %.3f%% (%.3f -> %.3f)",
				i-1, change, satiationHistory[i-1], satiationHistory[i])
		}
	}

	finalSatiation := satiationHistory[len(satiationHistory)-1]
	totalChange := finalSatiation - initialSatiation

	t.Logf("\n=== ИТОГИ ===")
	t.Logf("Начальный сытость: %.3f%%", initialSatiation)
	t.Logf("Финальный сытость: %.3f%%", finalSatiation)
	t.Logf("Общее изменение: %.3f%% за 30 тиков (2 кадра анимации)", totalChange)
	t.Logf("Ожидаемая потеря: %.3f%% (2.0%% * 0.5сек)", 2.0*0.5)
	t.Logf("Ожидаемое восстановление: +4.0%% (2 травы * 2 * 1 кадр на цикл)")

	// Анализируем кадры
	t.Logf("\n=== АНАЛИЗ КАДРОВ ===")
	frameChanges := 0
	for i := 1; i < len(frameHistory); i++ {
		if frameHistory[i] != frameHistory[i-1] {
			frameChanges++
			t.Logf("Кадр изменился на тике %d: %d → %d", i-1, frameHistory[i-1], frameHistory[i])
		}
	}
	t.Logf("Всего смен кадров: %d (ожидается 2, но питание только при 0→1)", frameChanges)

	if totalChange > 0 {
		t.Logf("✅ Голод ВОССТАНАВЛИВАЕТСЯ (+%.3f%%)", totalChange)
	} else {
		t.Logf("❌ Голод ТЕРЯЕТСЯ (%.3f%%)", totalChange)
	}
}
