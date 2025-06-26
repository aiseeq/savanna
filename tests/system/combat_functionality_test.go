package system

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

// CombatTestWorld содержит все необходимые компоненты для тестирования боевой системы
type CombatTestWorld struct {
	world                 *core.World
	systemManager         *core.SystemManager
	wolfAnimationSystem   *animation.AnimationSystem
	rabbitAnimationSystem *animation.AnimationSystem
	animationManager      *animation.AnimationManager
	wolves                []core.EntityID
	rabbits               []core.EntityID
}

// setupCombatWorld создаёт полную симуляцию для тестирования боевой системы
func setupCombatWorld(t *testing.T, cfg *config.Config) *CombatTestWorld {
	t.Log("Создаём полную симуляцию для тестирования боевой системы...")

	// Создаём мир точно как в реальной игре
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * 32) //nolint:gomnd // Размер тайла в пикселях
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	// КРИТИЧЕСКИ ВАЖНО: анимационные системы для работы боевой системы
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Загружаем анимации для тестирования
	loader := animation.NewAnimationLoader()
	emptyImg := ebiten.NewImage(128, 64)
	loader.LoadAnimations(wolfAnimationSystem, rabbitAnimationSystem, emptyImg, emptyImg)

	// КРИТИЧЕСКИ ВАЖНО: AnimationManager для обновления анимаций в боевой системе
	animationManager := animation.NewAnimationManager(wolfAnimationSystem, rabbitAnimationSystem)

	setupSystems(systemManager, animationManager, worldSizePixels, terrain)

	wolves, rabbits := createTestAnimals(t, world, cfg, terrain)

	return &CombatTestWorld{
		world:                 world,
		systemManager:         systemManager,
		wolfAnimationSystem:   wolfAnimationSystem,
		rabbitAnimationSystem: rabbitAnimationSystem,
		animationManager:      animationManager,
		wolves:                wolves,
		rabbits:               rabbits,
	}
}

// setupSystems инициализирует все системы как в реальной игре
func setupSystems(
	systemManager *core.SystemManager,
	animationManager *animation.AnimationManager,
	worldSizePixels float32,
	terrain *generator.Terrain,
) {
	// Инициализируем все системы как в реальной игре
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()

	// ВРЕМЕННО отключаем VegetationSystem и FeedingSystem для изоляции тестирования боевой системы
	_ = vegetationSystem // Убираем предупреждение
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	movementSystem := simulation.NewMovementSystem(worldSizePixels, worldSizePixels)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(combatSystem)
	systemManager.AddSystem(simulation.NewCorpseSystem()) // ВАЖНО: для превращения в трупы

	// КРИТИЧЕСКИ ВАЖНО: Добавляем AnimationManager как систему для обновления анимаций
	systemManager.AddSystem(animationManager)
}

// createTestAnimals размещает животных точно как в реальной игре
func createTestAnimals(
	t *testing.T, world *core.World, cfg *config.Config, terrain *generator.Terrain,
) (rabbits, wolves []core.EntityID) {
	// Размещаем животных ТОЧНО как в реальной игре
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			rabbit := simulation.CreateAnimal(world, core.TypeRabbit, placement.X, placement.Y)
			rabbits = append(rabbits, rabbit)
		case core.TypeWolf:
			wolf := simulation.CreateAnimal(world, core.TypeWolf, placement.X, placement.Y)
			// КРИТИЧЕСКИ ВАЖНО: делаем волков голодными для тестирования боевой системы
			world.SetSatiation(wolf, core.Satiation{Value: 10.0}) //nolint:gomnd // 10% - очень голодные

			// КРИТИЧЕСКИ ВАЖНО: увеличиваем порог охоты для надёжного тестирования
			behavior, _ := world.GetBehavior(wolf)
			behavior.SatiationThreshold = 90.0 //nolint:gomnd // Охотится пока сытость < 90%
			world.SetBehavior(wolf, behavior)

			wolves = append(wolves, wolf)
		}
	}

	t.Logf("Размещено: %d зайцев, %d волков", len(rabbits), len(wolves))
	logWolfPositions(t, world, wolves, rabbits)

	return wolves, rabbits
}

// logWolfPositions выводит отладочную информацию о позициях волков
func logWolfPositions(t *testing.T, world *core.World, wolves, rabbits []core.EntityID) {
	// DEBUG: проверяем расстояния между волками и зайцами
	for i, wolf := range wolves {
		wolfPos, _ := world.GetPosition(wolf)
		wolfSatiation, _ := world.GetSatiation(wolf)

		minDistanceToRabbit := float32(999999) //nolint:gomnd // Большое число для поиска минимума
		for _, rabbit := range rabbits {
			rabbitPos, _ := world.GetPosition(rabbit)
			// ТИПОБЕЗОПАСНОСТЬ: позиции уже float32
			dx := wolfPos.X - rabbitPos.X
			dy := wolfPos.Y - rabbitPos.Y
			distance := dx*dx + dy*dy
			if distance < minDistanceToRabbit {
				minDistanceToRabbit = distance
			}
		}

		t.Logf("DEBUG: Волк %d - позиция (%.1f, %.1f), сытость %.1f%%, ближайший заяц: %.1f пикселей",
			i, wolfPos.X, wolfPos.Y, wolfSatiation.Value, float32(math.Sqrt(float64(minDistanceToRabbit))))
	}
}

// updateAnimalAnimations обновляет анимации животных (КРИТИЧЕСКИ ВАЖНО!)
func (ctw *CombatTestWorld) updateAnimalAnimations() {
	ctw.world.ForEachWith(core.MaskAnimalType|core.MaskAnimation, func(entity core.EntityID) {
		animalType, ok := ctw.world.GetAnimalType(entity)
		if !ok {
			return
		}

		anim, hasAnim := ctw.world.GetAnimation(entity)
		if !hasAnim {
			return
		}

		var newAnimType animation.AnimationType
		var animSystem *animation.AnimationSystem

		switch animalType {
		case core.TypeWolf:
			newAnimType = getWolfAnimationTypeForTest(ctw.world, entity)
			animSystem = ctw.wolfAnimationSystem
		case core.TypeRabbit:
			newAnimType = getRabbitAnimationTypeForTest(ctw.world, entity)
			animSystem = ctw.rabbitAnimationSystem
		default:
			return
		}

		// НЕ прерываем анимацию ATTACK пока она играет!
		if anim.CurrentAnim != int(newAnimType) {
			if anim.CurrentAnim == int(animation.AnimAttack) && anim.Playing {
				// Анимация атаки должна доиграться до конца
			} else {
				anim.CurrentAnim = int(newAnimType)
				anim.Frame = 0 //nolint:gomnd // Сброс кадра
				anim.Timer = 0 //nolint:gomnd // Сброс таймера
				anim.Playing = true
				ctw.world.SetAnimation(entity, anim)
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

		deltaTime := float32(1.0 / 60.0) //nolint:gomnd // Фиксированный timestep
		animSystem.Update(&animComponent, deltaTime)

		anim.CurrentAnim = int(animComponent.CurrentAnim)
		anim.Frame = animComponent.Frame
		anim.Timer = animComponent.Timer
		anim.Playing = animComponent.Playing
		anim.FacingRight = animComponent.FacingRight

		ctw.world.SetAnimation(entity, anim)
	})
}

// testCombatFunctionality проверяет что боевая система реально работает
func testCombatFunctionality(t *testing.T, cfg *config.Config) {
	ctw := setupCombatWorld(t, cfg)

	initialWolves := len(ctw.wolves)
	initialRabbits := len(ctw.rabbits)

	if initialWolves == 0 || initialRabbits == 0 {
		t.Skip("❌ Нет волков или зайцев для тестирования боевой системы")
		return
	}

	runCombatSimulation(t, ctw, initialWolves, initialRabbits)
}

// runCombatSimulation запускает симуляцию боя и проверяет результаты
func runCombatSimulation(t *testing.T, ctw *CombatTestWorld, initialWolves, initialRabbits int) {
	maxTicks := 3600                 //nolint:gomnd // 60 секунд симуляции при 60 FPS
	deltaTick := float32(1.0 / 60.0) //nolint:gomnd // Фиксированный timestep

	for tick := 0; tick < maxTicks; tick++ {
		// КРИТИЧЕСКИ ВАЖНО: обновляем анимации ПЕРЕД системами!
		ctw.updateAnimalAnimations()

		// Обновляем все системы
		ctw.systemManager.Update(ctw.world, deltaTick)

		// Логируем прогресс каждые 10 секунд
		if tick%(10*60) == 0 { //nolint:gomnd // Каждые 10 секунд
			logCombatProgress(t, ctw.world, tick, initialWolves, initialRabbits)
		}

		// Проверяем критерии завершения боя
		if checkCombatCompletion(t, ctw.world, initialWolves, initialRabbits) {
			break
		}
	}

	// Финальные проверки
	validateCombatResults(t, ctw.world, initialWolves, initialRabbits)
}

// logCombatProgress выводит информацию о прогрессе боя
func logCombatProgress(t *testing.T, world *core.World, tick, initialWolves, initialRabbits int) {
	currentWolves := world.CountEntitiesWith(core.MaskAnimalType)
	currentRabbits := 0
	currentCorpses := world.CountEntitiesWith(core.MaskCorpse)

	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if animalType, ok := world.GetAnimalType(entity); ok && animalType == core.TypeRabbit {
			currentRabbits++
		}
	})

	seconds := tick / 60 //nolint:gomnd // Конвертация тиков в секунды
	t.Logf("⏱️  %d сек: волки %d/%d, зайцы %d/%d, трупы %d",
		seconds, currentWolves-currentRabbits, initialWolves, currentRabbits, initialRabbits, currentCorpses)
}

// checkCombatCompletion проверяет критерии завершения боя
func checkCombatCompletion(t *testing.T, world *core.World, _, _ int) bool {
	currentWolves := 0
	currentRabbits := 0
	currentCorpses := world.CountEntitiesWith(core.MaskCorpse)

	world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		if animalType, ok := world.GetAnimalType(entity); ok {
			switch animalType {
			case core.TypeWolf:
				currentWolves++
			case core.TypeRabbit:
				currentRabbits++
			}
		}
	})

	// Прерываем если есть успех
	if currentCorpses > 0 {
		t.Logf("✅ УСПЕХ! Найдены трупы: %d. Боевая система работает!",
			currentCorpses)
		return true
	}

	// Прерываем если все волки умерли
	if currentWolves == 0 {
		t.Error("❌ КРИТИЧЕСКАЯ ОШИБКА: Все волки умерли от голода несмотря на то " +
			"что убивали зайцев! Проблема с поеданием трупов!")
		return true
	}

	// Прерываем если все зайцы умерли
	if currentRabbits == 0 {
		t.Logf("⚠️  Все зайцы убиты. Тестирование боевой системы завершено.")
		return true
	}

	return false
}

// validateCombatResults выполняет финальную валидацию результатов боя
func validateCombatResults(t *testing.T, world *core.World, _, _ int) {
	finalCorpses := world.CountEntitiesWith(core.MaskCorpse)
	finalAnimals := world.CountEntitiesWith(core.MaskAnimalType)

	if finalCorpses > 0 {
		t.Logf("✅ Боевая система РАБОТАЕТ! Создано трупов: %d", finalCorpses)
	} else {
		t.Error("❌ Боевая система НЕ РАБОТАЕТ: ни одного трупа не создано")
	}

	t.Logf("📊 Финальная статистика: животных %d, трупов %d", finalAnimals, finalCorpses)
}

// getWolfAnimationTypeForTest определяет тип анимации для волка
func getWolfAnimationTypeForTest(world *core.World, entity core.EntityID) animation.AnimationType {
	if world.HasComponent(entity, core.MaskEatingState) {
		return animation.AnimEat
	}

	if isWolfAttackingForTest(world, entity) {
		return animation.AnimAttack
	}

	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	// ИСПРАВЛЕНИЕ: Используем правильные пороги для тайловой системы
	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 4.0 { // WolfWalkSpeedThreshold из animation/resolver.go
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

// getRabbitAnimationTypeForTest определяет тип анимации для зайца
func getRabbitAnimationTypeForTest(world *core.World, entity core.EntityID) animation.AnimationType {
	if world.HasComponent(entity, core.MaskCorpse) {
		return animation.AnimDeathDying
	}

	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return animation.AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	// ИСПРАВЛЕНИЕ: Используем правильные пороги для тайловой системы
	if speed < 0.1 {
		return animation.AnimIdle
	} else if speed < 2.25 { // RabbitWalkSpeedThreshold из animation/resolver.go
		return animation.AnimWalk
	} else {
		return animation.AnimRun
	}
}

// isWolfAttackingForTest проверяет атакует ли волк
func isWolfAttackingForTest(world *core.World, wolf core.EntityID) bool {
	satiation, hasSatiation := world.GetSatiation(wolf)
	if !hasSatiation || satiation.Value >= 60.0 {
		return false
	}

	pos, hasPos := world.GetPosition(wolf)
	if !hasPos {
		return false
	}

	// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
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

	// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для вычислений
	dx := pos.X - rabbitPos.X
	dy := pos.Y - rabbitPos.Y
	distance := dx*dx + dy*dy
	return distance <= 13.0*13.0 // Используем текущий радиус
}

// TestCombatSystem тестирует что боевая система реально работает
func TestCombatSystem(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 10 // Очень маленький мир для гарантированных встреч
	cfg.Population.Rabbits = 3
	cfg.Population.Wolves = 2
	cfg.World.Seed = 12345 // Детерминированный seed

	testCombatFunctionality(t, cfg)
}
