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

// TestAnimationInvestigation ДЕТАЛЬНОЕ исследование проблемы с анимацией idle во время еды
//
//nolint:gocognit,revive,funlen // Детальное исследование анимационной системы
func TestAnimationInvestigation(t *testing.T) {
	t.Parallel()

	t.Logf("=== ДЕТАЛЬНОЕ ИССЛЕДОВАНИЕ ПРОБЛЕМЫ С АНИМАЦИЕЙ ===")

	// Создаём точную копию GUI настроек
	world := core.NewWorld(1600, 1600, 12345) // 50x50 тайлов как в игре

	// Создаём terrain точно как в GUI
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 50
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Принудительно устанавливаем траву в центр
	centerX, centerY := 25, 25
	t.Logf("Устанавливаем траву в тайл (%d, %d)", centerX, centerY)

	// Проверяем тип тайла ПЕРЕД установкой травы
	tileType := terrain.GetTileType(centerX, centerY)
	t.Logf("Тип тайла (%d, %d): %v", centerX, centerY, tileType)

	terrain.SetTileType(centerX, centerY, generator.TileGrass)
	terrain.SetGrassAmount(centerX, centerY, 100.0)

	// Проверяем что трава действительно установилась
	grassAfterSet := terrain.GetGrassAmount(centerX, centerY)
	t.Logf("Трава в тайле (%d, %d) после установки: %.1f", centerX, centerY, grassAfterSet)

	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Создаём ВСЕ системы точно как в GUI main.go
	systemManager := core.NewSystemManager()
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()
	movementSystem := simulation.NewMovementSystem(1600, 1600)

	// Добавляем системы в ТОМ ЖЕ порядке что в GUI
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(adapters.NewFeedingSystemAdapter(vegetationSystem)) // 1. Создаёт EatingState
	// 2. Дискретное поедание травы по кадрам анимации
	systemManager.AddSystem(grassEatingSystem)
	// 3. Проверяет EatingState и не мешает еде
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem}) // 4. Сбрасывает скорость едящих
	systemManager.AddSystem(combatSystem)                                            // 5. Система боя

	// Создаём анимационные системы ТОЧНО как в GUI
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Имитируем ТОЧНУЮ загрузку анимаций из GUI loadRabbitAnimations
	t.Logf("\n--- Загрузка анимаций зайца как в GUI ---")
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
		{"hare_eat", 2, 4.0, true, animation.AnimEat},
		{"hare_dead", 2, 3.0, false, animation.AnimDeathDying},
	}

	for _, config := range rabbitAnimations {
		rabbitAnimationSystem.RegisterAnimation(config.animType, config.frames, config.fps, config.loop, nil)
		t.Logf("  Зарегистрирована: %s (%d кадров, %.1f FPS, зацикленная=%v)",
			config.animType.String(), config.frames, config.fps, config.loop)
	}

	// Создаём resolver точно как в GUI
	animationResolver := animation.NewAnimationResolver()

	// Создаём зайца в центре где есть трава
	rabbitX, rabbitY := float32(centerX*32+16), float32(centerY*32+16) // Центр тайла
	rabbit := simulation.CreateAnimal(world, core.TypeRabbit, rabbitX, rabbitY)

	// Делаем зайца голодным чтобы он точно ел
	world.SetSatiation(rabbit, core.Satiation{Value: 70.0}) // 70% - точно будет есть (порог 90%)
	world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

	deltaTime := float32(1.0 / 60.0)

	t.Logf("\n--- Начальное состояние ---")
	pos, _ := world.GetPosition(rabbit)
	satiation, _ := world.GetSatiation(rabbit)
	// ТИПОБЕЗОПАСНОСТЬ: позиции уже float32
	grassAmount := vegetationSystem.GetGrassAt(pos.X, pos.Y)
	behavior, _ := world.GetBehavior(rabbit)

	t.Logf("Позиция зайца: (%.1f, %.1f)", pos.X, pos.Y)
	t.Logf("Сытость зайца: %.1f%% (порог: %.1f%%)", satiation.Value, behavior.SatiationThreshold)
	t.Logf("Трава в позиции: %.1f единиц", grassAmount)

	// КРИТИЧЕСКИ ВАЖНО: дебаг тайла
	tileX := int(pos.X / 32)
	tileY := int(pos.Y / 32)
	t.Logf("Заяц в тайле: (%d, %d), ожидаем (%d, %d)", tileX, tileY, centerX, centerY)
	tileType = terrain.GetTileType(tileX, tileY)
	grassInTile := terrain.GetGrassAmount(tileX, tileY)
	t.Logf("Тип тайла зайца: %v, трава в тайле: %.1f", tileType, grassInTile)

	t.Logf("\n--- ПОШАГОВАЯ СИМУЛЯЦИЯ GUI ЛОГИКИ ---")

	// Симулируем точно GUI updateSimulation + updateAnimalAnimations
	for i := 0; i < 20; i++ {
		t.Logf("\n=== ТИК %d ===", i)

		// === ЭТАП 1: Состояние ДО обновления ===
		pos, _ = world.GetPosition(rabbit)
		satiation, _ = world.GetSatiation(rabbit)
		anim, _ := world.GetAnimation(rabbit)
		var vel core.Velocity
		isEatingBefore := world.HasComponent(rabbit, core.MaskEatingState)
		// ТИПОБЕЗОПАСНОСТЬ: позиции уже float32
		grassBefore := vegetationSystem.GetGrassAt(pos.X, pos.Y)
		animTypeBefore := animation.AnimationType(anim.CurrentAnim)

		// ДЕБАГ: проверяем вычисление тайла
		tileX := int(pos.X / 32)
		tileY := int(pos.Y / 32)
		grassInTile := terrain.GetGrassAmount(tileX, tileY)

		t.Logf("ДО обновления:")
		t.Logf("  Позиция: (%.1f, %.1f), Тайл: (%d, %d)", pos.X, pos.Y, tileX, tileY)
		t.Logf("  Сытость: %.1f%% (порог %.1f%%), Трава через VegetationSystem: %.1f, Трава в terrain тайле: %.1f",
			satiation.Value, simulation.RabbitSatiationThreshold, grassBefore, grassInTile)
		t.Logf("  EatingState: %v, Анимация: %s (код %d, кадр %d)",
			isEatingBefore, animTypeBefore.String(), anim.CurrentAnim, anim.Frame)

		// === ЭТАП 2: Обновление мира и систем (как в GUI updateSimulation) ===
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// === ЭТАП 3: Состояние ПОСЛЕ систем ===
		satiation, _ = world.GetSatiation(rabbit)
		vel, _ = world.GetVelocity(rabbit)
		anim, _ = world.GetAnimation(rabbit)
		isEatingAfter := world.HasComponent(rabbit, core.MaskEatingState)
		// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
		grassAfter := vegetationSystem.GetGrassAt(pos.X, pos.Y)
		speed := vel.X*vel.X + vel.Y*vel.Y
		animTypeAfterSystems := animation.AnimationType(anim.CurrentAnim)

		t.Logf("ПОСЛЕ систем:")
		t.Logf("  Сытость: %.1f%%, Трава: %.1f, Скорость: %.2f", satiation.Value, grassAfter, speed)
		t.Logf("  EatingState: %v, Анимация: %s (код %d, кадр %d)",
			isEatingAfter, animTypeAfterSystems.String(), anim.CurrentAnim, anim.Frame)

		// === ЭТАП 4: Обновление анимаций (как в GUI updateAnimalAnimations) ===

		// 4.1 Проверяем что говорит resolver
		animalType, _ := world.GetAnimalType(rabbit)
		expectedAnim := animationResolver.ResolveAnimalAnimationType(world, rabbit, animalType)
		t.Logf("  AnimationResolver ожидает: %s", expectedAnim.String())

		// 4.2 Имитируем getAnimationContext
		var animSystem *animation.AnimationSystem
		switch animalType {
		case core.TypeRabbit:
			animSystem = rabbitAnimationSystem
		case core.TypeWolf:
			animSystem = wolfAnimationSystem
		}

		if animSystem == nil {
			t.Errorf("  ❌ AnimationSystem НЕ НАЙДЕНА для типа %v", animalType)
			continue
		}

		// 4.3 Имитируем updateAnimationIfNeeded
		newAnimType := expectedAnim
		if anim.CurrentAnim != int(newAnimType) {
			// НЕ прерываем анимацию ATTACK
			if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
				t.Logf("  Не меняем анимацию - Attack играет")
			} else {
				// Обычная смена анимации
				oldAnimType := animation.AnimationType(anim.CurrentAnim)
				anim.CurrentAnim = int(newAnimType)
				anim.Frame = 0
				anim.Timer = 0
				anim.Playing = true
				world.SetAnimation(rabbit, anim)
				t.Logf("  🔄 СМЕНИЛИ АНИМАЦИЮ: %s -> %s", oldAnimType.String(), newAnimType.String())
			}
		} else {
			t.Logf("  Анимация не изменилась: %s", newAnimType.String())
		}

		// 4.4 Имитируем updateAnimationDirection
		if vel, hasVel := world.GetVelocity(rabbit); hasVel {
			if vel.X > 0.1 {
				anim.FacingRight = true
			} else if vel.X < -0.1 {
				anim.FacingRight = false
			}
		}

		// 4.5 Имитируем processAnimationUpdate
		animComponent := animation.AnimationComponent{
			CurrentAnim: animation.AnimationType(anim.CurrentAnim),
			Frame:       anim.Frame,
			Timer:       anim.Timer,
			Playing:     anim.Playing,
			FacingRight: anim.FacingRight,
		}

		animSystem.Update(&animComponent, deltaTime)

		// Сохраняем обновленное состояние
		anim.Frame = animComponent.Frame
		anim.Timer = animComponent.Timer
		anim.Playing = animComponent.Playing
		anim.FacingRight = animComponent.FacingRight
		world.SetAnimation(rabbit, anim)

		// === ЭТАП 5: Финальное состояние ===
		anim, _ = world.GetAnimation(rabbit)
		animTypeFinal := animation.AnimationType(anim.CurrentAnim)

		t.Logf("ФИНАЛЬНОЕ состояние:")
		t.Logf("  Анимация: %s (код %d, кадр %d, таймер %.2f, играет %v)",
			animTypeFinal.String(), anim.CurrentAnim, anim.Frame, anim.Timer, anim.Playing)

		// === КРИТИЧЕСКИЕ ПРОВЕРКИ ===
		if isEatingAfter {
			if animTypeFinal != animation.AnimEat {
				t.Errorf("  ❌ БАГ ОБНАРУЖЕН: Заяц ест (EatingState=true) но анимация %s вместо Eat",
					animTypeFinal.String())
				t.Errorf("    Resolver ожидает: %s", expectedAnim.String())
				t.Errorf("    Анимация после систем: %s", animTypeAfterSystems.String())
				t.Errorf("    Финальная анимация: %s", animTypeFinal.String())

				// Диагностика
				eatAnim := rabbitAnimationSystem.GetAnimation(animation.AnimEat)
				if eatAnim == nil {
					t.Errorf("    ПРИЧИНА: AnimEat НЕ зарегистрирована!")
				} else {
					t.Logf("    AnimEat зарегистрирована: %d кадров, %.1f FPS", eatAnim.Frames, eatAnim.FPS)
				}
				return
			} else {
				t.Logf("  ✅ ПРАВИЛЬНО: Заяц ест и показывает анимацию Eat")
				return
			}
		}

		// Проверяем прогресс
		grassConsumed := grassBefore - grassAfter
		if grassConsumed > 0 {
			t.Logf("  Съедено травы: %.3f единиц", grassConsumed)
		}
	}

	t.Errorf("❌ Заяц не начал есть за 20 тиков")
}
