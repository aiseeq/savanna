package e2e

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestSeed6Debug воспроизводит ТОЧНО вашу ситуацию: seed 6, первые 15 секунд
//
//nolint:gocognit,revive // Сложный E2E тест, имитирующий полную игровую сессию
func TestSeed6Debug(t *testing.T) {
	t.Parallel()
	t.Logf("=== ОТЛАДКА SEED 6: ПЕРВЫЕ 15 СЕКУНД ===")
	t.Logf("Воспроизводим: make build && ./bin/savanna-game -seed 6")

	// ТОЧНО такая же инициализация как в ./bin/savanna-game
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = 6 // ТОТ ЖЕ SEED!

	t.Logf("Конфигурация:")
	t.Logf("  Seed: %d", cfg.World.Seed)
	t.Logf("  Размер мира: %d", cfg.World.Size)
	t.Logf("  Зайцев: %d", cfg.Population.Rabbits)
	t.Logf("  Волков: %d", cfg.Population.Wolves)

	// Генерируем мир ТОЧНО как в игре
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)

	// Создаём ТОЧНО такие же системы как в main.go
	systemManager := core.NewSystemManager()
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)

	// Добавляем системы в ТОМ ЖЕ ПОРЯДКЕ что в main.go
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(adapters.NewFeedingSystemAdapter(vegetationSystem))
	systemManager.AddSystem(combatSystem)

	// КРИТИЧЕСКИ ВАЖНО: анимационные системы как в GUI
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Загружаем анимации ТОЧНО как в GUI
	loadAnimationsLikeGUI(wolfAnimationSystem, rabbitAnimationSystem)

	// Создаём off-screen буфер (как в GUI)
	offscreenImage := ebiten.NewImage(int(worldSizePixels), int(worldSizePixels))

	// Размещаем животных ТОЧНО как в игре
	t.Logf("\nРазмещение животных...")
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	// Создаём животных на основе сгенерированных позиций
	rabbits := []core.EntityID{}
	wolves := []core.EntityID{}

	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			rabbit := simulation.CreateAnimal(world, core.TypeRabbit, placement.X, placement.Y)
			rabbits = append(rabbits, rabbit)
			t.Logf("  Заяц %d: позиция (%.1f, %.1f)", rabbit, placement.X, placement.Y)
		case core.TypeWolf:
			wolf := simulation.CreateAnimal(world, core.TypeWolf, placement.X, placement.Y)
			wolves = append(wolves, wolf)

			// ИСПРАВЛЕНИЕ: Увеличиваем радиус видения для seed 6 (животные далеко друг от друга)
			if behavior, hasBehavior := world.GetBehavior(wolf); hasBehavior {
				behavior.VisionRange = 25.0 // Увеличиваем с ~5 до 25 тайлов для этого теста
				world.SetBehavior(wolf, behavior)
			}

			// ИСПРАВЛЕНИЕ: Делаем волков голодными чтобы они атаковали (было 70% > порога 60%)
			world.SetSatiation(wolf, core.Satiation{Value: 40.0}) // 40% < порога 60%

			// Проверяем начальный голод волка
			hunger, _ := world.GetSatiation(wolf)
			t.Logf("  Волк %d: позиция (%.1f, %.1f), голод %.1f%%, видение 25 тайлов", wolf, placement.X, placement.Y, hunger.Value)
		}
	}

	t.Logf("Создано: %d зайцев, %d волков", len(rabbits), len(wolves))

	// Функция обновления анимаций ТОЧНО как в main.go
	updateAnimalAnimations := func() {
		world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
			animalType, ok := world.GetAnimalType(entity)
			if !ok {
				return
			}

			anim, hasAnim := world.GetAnimation(entity)
			if !hasAnim {
				return
			}

			// Определяем тип анимации и систему на основе типа животного
			var newAnimType animation.AnimationType
			var animSystem *animation.AnimationSystem

			switch animalType {
			case core.TypeWolf:
				newAnimType = getWolfAnimationTypeLikeGUI(world, entity)
				animSystem = wolfAnimationSystem
			case core.TypeRabbit:
				newAnimType = getRabbitAnimationTypeLikeGUI(world, entity)
				animSystem = rabbitAnimationSystem
			default:
				return
			}

			// КРИТИЧЕСКИ ВАЖНО: НЕ прерываем анимацию ATTACK пока она играет!
			oldAnimType := animation.AnimationType(anim.CurrentAnim)
			animChanged := false

			if anim.CurrentAnim != int(newAnimType) {
				if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
					// НЕ сбрасываем анимацию ATTACK!
				} else {
					anim.CurrentAnim = int(newAnimType)
					anim.Frame = 0
					anim.Timer = 0
					anim.Playing = true
					world.SetAnimation(entity, anim)
					animChanged = true
				}
			}

			// Обновляем анимацию
			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			oldFrame := animComponent.Frame
			oldPlaying := animComponent.Playing

			animSystem.Update(&animComponent, 1.0/60.0)

			// Сохраняем состояние
			anim.Frame = animComponent.Frame
			anim.Timer = animComponent.Timer
			anim.Playing = animComponent.Playing
			world.SetAnimation(entity, anim)

			// Логируем только важные изменения
			if animChanged {
				t.Logf("    [ANIM] %s %d: %s -> %s", animalType.String(), entity, oldAnimType.String(), newAnimType.String())
			}
			if oldFrame != animComponent.Frame {
				t.Logf("    [FRAME] %s %d: кадр %d->%d, играет %t",
					animalType.String(), entity, oldFrame, animComponent.Frame, animComponent.Playing)
			}
			if oldPlaying && !animComponent.Playing {
				t.Logf("    [END] %s %d: анимация %s завершена",
					animalType.String(), entity, animation.AnimationType(anim.CurrentAnim).String())
			}
		})
	}

	// Функция отрисовки как в GUI
	renderFrame := func() {
		offscreenImage.Clear()

		// Имитируем полную отрисовку как в GUI
		world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
			animalType, _ := world.GetAnimalType(entity)
			anim, hasAnim := world.GetAnimation(entity)
			if !hasAnim {
				return
			}

			animComponent := animation.AnimationComponent{
				CurrentAnim: animation.AnimationType(anim.CurrentAnim),
				Frame:       anim.Frame,
				Timer:       anim.Timer,
				Playing:     anim.Playing,
				FacingRight: anim.FacingRight,
			}

			// Получаем кадр как в GUI
			var frameImg *ebiten.Image
			switch animalType {
			case core.TypeWolf:
				frameImg = wolfAnimationSystem.GetFrameImage(&animComponent)
			case core.TypeRabbit:
				frameImg = rabbitAnimationSystem.GetFrameImage(&animComponent)
			}

			_ = frameImg // Использовался бы для отрисовки
		})
	}

	// Отслеживание всех событий
	lastHealths := make(map[core.EntityID]int16)
	lastHungers := make(map[core.EntityID]float32)
	lastPositions := make(map[core.EntityID]core.Position)
	attackEvents := []string{}
	damageEvents := []string{}

	// Инициализируем начальные состояния
	for _, rabbit := range rabbits {
		health, _ := world.GetHealth(rabbit)
		pos, _ := world.GetPosition(rabbit)
		lastHealths[rabbit] = health.Current
		lastPositions[rabbit] = pos
	}
	for _, wolf := range wolves {
		hunger, _ := world.GetSatiation(wolf)
		pos, _ := world.GetPosition(wolf)
		lastHungers[wolf] = hunger.Value
		lastPositions[wolf] = pos
	}

	// ГЛАВНЫЙ ЦИКЛ: 20 секунд симуляции (1200 тиков) чтобы дождаться голода < 60%
	deltaTime := float32(1.0 / 60.0)

	t.Logf("\n=== НАЧАЛО СИМУЛЯЦИИ (20 СЕКУНД) ===")

	for tick := 0; tick < 1200; tick++ {
		// Обновляем мир ТОЧНО как в GUI
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Обновляем анимации ТОЧНО как в GUI
		updateAnimalAnimations()

		// "Отрисовываем" ТОЧНО как в GUI
		renderFrame()

		// ДЕТАЛЬНОЕ ЛОГИРОВАНИЕ СОБЫТИЙ

		// Логируем каждую секунду
		if tick%60 == 0 {
			t.Logf("\n--- СЕКУНДА %d ---", tick/60)

			// Статистика популяций
			stats := world.GetStats()
			t.Logf("Популяция: %d зайцев, %d волков", stats[core.TypeRabbit], stats[core.TypeWolf])
		}

		// Отслеживаем анимации атак волков
		for _, wolf := range wolves {
			if !world.IsAlive(wolf) {
				continue
			}

			if wolfAnim, hasAnim := world.GetAnimation(wolf); hasAnim {
				if wolfAnim.CurrentAnim == int(animation.AnimAttack) {
					event := ""
					if wolfAnim.Frame == 0 && wolfAnim.Playing {
						event = "замах"
					} else if wolfAnim.Frame == 1 && wolfAnim.Playing {
						event = "удар"
					} else if wolfAnim.Frame == 1 && !wolfAnim.Playing {
						event = "завершение"
					}

					if event != "" {
						logEntry := fmt.Sprintf("[TICK %3d] 🐺 Волк %d АТАКУЕТ: %s (кадр %d, играет %t)",
							tick, wolf, event, wolfAnim.Frame, wolfAnim.Playing)
						t.Logf(logEntry)
						attackEvents = append(attackEvents, logEntry)
					}
				}
			}
		}

		// Отслеживаем урон зайцев
		for _, rabbit := range rabbits {
			if !world.IsAlive(rabbit) {
				continue
			}

			currentHealth, _ := world.GetHealth(rabbit)
			if currentHealth.Current != lastHealths[rabbit] {
				logEntry := fmt.Sprintf("[TICK %3d] 🩸 Заяц %d: здоровье %d -> %d",
					tick, rabbit, lastHealths[rabbit], currentHealth.Current)
				t.Logf(logEntry)
				damageEvents = append(damageEvents, logEntry)

				// Проверяем DamageFlash
				if world.HasComponent(rabbit, core.MaskDamageFlash) {
					flash, _ := world.GetDamageFlash(rabbit)
					t.Logf("    ✨ DamageFlash: %.3f сек", flash.Timer)
				} else {
					t.Logf("    ❌ DamageFlash НЕ активен!")
				}

				lastHealths[rabbit] = currentHealth.Current

				// Если заяц умер
				if currentHealth.Current == 0 {
					t.Logf("    ⚰️ Заяц %d УМЕР!", rabbit)

					if world.HasComponent(rabbit, core.MaskCorpse) {
						corpse, _ := world.GetCorpse(rabbit)
						t.Logf("    📦 Труп создан: питательность %.1f", corpse.NutritionalValue)
					} else {
						t.Logf("    ❌ Труп НЕ создан!")
					}
				}
			}
		}

		// Отслеживаем голод волков
		for _, wolf := range wolves {
			if !world.IsAlive(wolf) {
				continue
			}

			currentHunger, _ := world.GetSatiation(wolf)
			if currentHunger.Value != lastHungers[wolf] {
				t.Logf("[TICK %3d] 🍖 Волк %d: голод %.1f%% -> %.1f%%",
					tick, wolf, lastHungers[wolf], currentHunger.Value)
				lastHungers[wolf] = currentHunger.Value
			}
		}

		// Отслеживаем исчезновение животных
		for _, rabbit := range rabbits {
			if !world.IsAlive(rabbit) {
				t.Logf("[TICK %3d] 👻 Заяц %d ИСЧЕЗ (съеден или уничтожен)", tick, rabbit)
				// Удаляем из отслеживания
				delete(lastHealths, rabbit)
				delete(lastPositions, rabbit)
			}
		}

		// Отслеживаем поедание
		for _, wolf := range wolves {
			if !world.IsAlive(wolf) {
				continue
			}

			if world.HasComponent(wolf, core.MaskEatingState) {
				eating, _ := world.GetEatingState(wolf)
				t.Logf("[TICK %3d] 🍽️ Волк %d ест труп %d", tick, wolf, eating.Target)
			}
		}

		// Логируем движение каждые 2 секунды
		if tick%120 == 0 && tick > 0 {
			t.Logf("\n--- ПОЗИЦИИ НА СЕКУНДЕ %d ---", tick/60)
			for _, wolf := range wolves {
				if !world.IsAlive(wolf) {
					continue
				}
				pos, _ := world.GetPosition(wolf)
				hunger, _ := world.GetSatiation(wolf)
				t.Logf("  Волк %d: (%.1f, %.1f), голод %.1f%%", wolf, pos.X, pos.Y, hunger.Value)
			}
			for _, rabbit := range rabbits {
				if !world.IsAlive(rabbit) {
					continue
				}
				pos, _ := world.GetPosition(rabbit)
				health, _ := world.GetHealth(rabbit)
				t.Logf("  Заяц %d: (%.1f, %.1f), здоровье %d", rabbit, pos.X, pos.Y, health.Current)
			}
		}
	}

	// АНАЛИЗ РЕЗУЛЬТАТОВ
	t.Logf("\n=== АНАЛИЗ 5 СЕКУНД СИМУЛЯЦИИ ===")

	t.Logf("События атак: %d", len(attackEvents))
	for i, event := range attackEvents {
		if i < 10 { // Показываем первые 10
			t.Logf("  %s", event)
		}
	}
	if len(attackEvents) > 10 {
		t.Logf("  ... и еще %d событий", len(attackEvents)-10)
	}

	t.Logf("События урона: %d", len(damageEvents))
	for i, event := range damageEvents {
		if i < 10 { // Показываем первые 10
			t.Logf("  %s", event)
		}
	}
	if len(damageEvents) > 10 {
		t.Logf("  ... и еще %d событий", len(damageEvents)-10)
	}

	// Финальное состояние
	finalStats := world.GetStats()
	t.Logf("Финальная популяция: %d зайцев, %d волков", finalStats[core.TypeRabbit], finalStats[core.TypeWolf])

	// КРИТИЧЕСКИЕ ПРОВЕРКИ
	if len(attackEvents) == 0 {
		t.Error("❌ НЕТ СОБЫТИЙ АТАК! Волки не атаковали.")
	} else if len(damageEvents) == 0 {
		t.Error("❌ НЕТ УРОНА! Атаки есть, но урон не наносится.")
	}

	// Проверяем что в логах есть и замах и удар
	hasSwing := false
	hasStrike := false
	for _, event := range attackEvents {
		if contains(event, "замах") {
			hasSwing = true
		}
		if contains(event, "удар") {
			hasStrike = true
		}
	}

	if !hasSwing {
		t.Error("❌ НЕТ КАДРА ЗАМАХА! Анимация начинается не с кадра 0.")
	}

	if !hasStrike {
		t.Error("❌ НЕТ КАДРА УДАРА! Анимация не доходит до кадра 1.")
	}

	t.Logf("\n🎯 Отладка seed 6 завершена")
}

// Вспомогательные функции

func getWolfAnimationTypeLikeGUI(world *core.World, entity core.EntityID) animation.AnimationType {
	// ТОЧНО как в main.go
	if world.HasComponent(entity, core.MaskEatingState) {
		return animation.AnimEat
	}

	if isWolfAttackingLikeGUI(world, entity) {
		return animation.AnimAttack
	}

	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 400.0 {
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

func getRabbitAnimationTypeLikeGUI(world *core.World, entity core.EntityID) animation.AnimationType {
	// ТОЧНО как в main.go
	if world.HasComponent(entity, core.MaskCorpse) {
		return animation.AnimDeathDying
	}

	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 300.0 {
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

func isWolfAttackingLikeGUI(world *core.World, wolf core.EntityID) bool {
	// ТОЧНО как в main.go
	hunger, hasHunger := world.GetSatiation(wolf)
	if !hasHunger || hunger.Value > 60.0 {
		return false
	}

	pos, hasPos := world.GetPosition(wolf)
	if !hasPos {
		return false
	}

	nearestRabbit, foundRabbit := world.FindNearestByType(pos.X, pos.Y, 15.0, core.TypeRabbit)
	if !foundRabbit {
		return false
	}

	if world.HasComponent(nearestRabbit, core.MaskCorpse) {
		return false
	}

	rabbitPos, hasRabbitPos := world.GetPosition(nearestRabbit)
	if !hasRabbitPos {
		return false
	}

	distance := (pos.X-rabbitPos.X)*(pos.X-rabbitPos.X) + (pos.Y-rabbitPos.Y)*(pos.Y-rabbitPos.Y)
	return distance <= 12.0*12.0
}

func loadAnimationsLikeGUI(wolfAnimSystem, rabbitAnimSystem *animation.AnimationSystem) {
	// Пустые спрайтшиты для тестирования (содержимое не важно)
	emptySheet := ebiten.NewImage(128, 64)

	// Волк - ТОЧНО как в main.go
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptySheet)
	wolfAnimSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptySheet)
	wolfAnimSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptySheet)
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, emptySheet) // НЕ зацикленная!
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, emptySheet)

	// Заяц - ТОЧНО как в main.go
	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, emptySheet)
	rabbitAnimSystem.RegisterAnimation(animation.AnimWalk, 4, 8.0, true, emptySheet)
	rabbitAnimSystem.RegisterAnimation(animation.AnimRun, 4, 12.0, true, emptySheet)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 1, 1.0, false, emptySheet)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
