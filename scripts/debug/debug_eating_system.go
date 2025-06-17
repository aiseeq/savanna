package main

import (
	"fmt"
	"time"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Константы диагностической системы (устраняет магические числа)
const (
	// Размеры мира и настройки симуляции
	WorldSeed      = 6  // Фиксированный seed для воспроизводимых результатов
	WorldSize      = 50 // Размер мира в тайлах
	TileSizePixels = 32 // Размер тайла в пикселях
	RabbitCount    = 30 // Количество зайцев в тесте
	WolfCount      = 3  // Количество волков в тесте

	// Параметры симуляции
	TicksPerSecond = 60                             // Частота обновления (TPS)
	SecondsToTest  = 12                             // Время теста до смерти волков
	MaxTicks       = SecondsToTest * TicksPerSecond // 720 тиков

	// Параметры анимации
	EmptyImageWidth  = 128 // Ширина пустого изображения для headless режима
	EmptyImageHeight = 64  // Высота пустого изображения для headless режима

	// Пороги скорости для определения типа анимации
	IdleSpeedThreshold  = 0.1   // Порог неподвижности (скорость < 0.1)
	WolfWalkThreshold   = 400.0 // Порог между ходьбой и бегом для волка
	RabbitWalkThreshold = 300.0 // Порог между ходьбой и бегом для зайца

	// Параметры поиска и атаки
	WolfSearchRadius   = 15.0 // Радиус поиска добычи волком
	WolfAttackDistance = 13.0 // Дистанция атаки волка
	WolfHungerTrigger  = 60.0 // Порог голода для начала охоты

	// Интервалы диагностики и производительности
	DiagnosticInterval = 60  // Интервал вывода диагностики (каждую секунду)
	SleepMicroseconds  = 100 // Пауза между тиками в микросекундах
)

func main() {
	fmt.Println("=== ДИАГНОСТИКА СИСТЕМЫ ПОЕДАНИЯ ТРУПОВ ===")

	// ТОЧНО такая же инициализация как в реальной headless игре
	cfg := config.LoadDefaultConfig()
	cfg.World.Seed = WorldSeed
	cfg.World.Size = WorldSize
	cfg.Population.Rabbits = RabbitCount
	cfg.Population.Wolves = WolfCount

	// Генерируем мир ТОЧНО как в headless
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	worldSizePixels := float32(cfg.World.Size * TileSizePixels)
	world := core.NewWorld(worldSizePixels, worldSizePixels, cfg.World.Seed)
	systemManager := core.NewSystemManager()

	// КРИТИЧЕСКИ ВАЖНО: создаём анимационные системы для headless режима
	wolfAnimationSystem := animation.NewAnimationSystem()
	rabbitAnimationSystem := animation.NewAnimationSystem()

	// Загружаем анимации через общий загрузчик (устраняет дублирование)
	loader := animation.NewAnimationLoader()
	loader.LoadHeadlessAnimations(wolfAnimationSystem, rabbitAnimationSystem)

	// Создаём системы с зависимостями ТОЧНО как в headless
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	combatSystem := simulation.NewCombatSystem()

	// Добавляем системы в правильном порядке ТОЧНО как в headless
	systemManager.AddSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: simulation.NewMovementSystem(worldSizePixels, worldSizePixels)})
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})
	systemManager.AddSystem(combatSystem)

	// Размещаем животных ТОЧНО как в headless
	popGen := generator.NewPopulationGenerator(cfg, terrain)
	placements := popGen.Generate()

	var wolves []core.EntityID
	var rabbits []core.EntityID

	for _, placement := range placements {
		switch placement.Type {
		case core.TypeRabbit:
			rabbit := simulation.CreateRabbit(world, placement.X, placement.Y)
			rabbits = append(rabbits, rabbit)
		case core.TypeWolf:
			wolf := simulation.CreateWolf(world, placement.X, placement.Y)
			wolves = append(wolves, wolf)
		}
	}

	fmt.Printf("Размещено: %d зайцев, %d волков\n", len(rabbits), len(wolves))

	// Создаём менеджер анимаций (устраняет дублирование логики)
	animationManager := animation.NewAnimationManager(wolfAnimationSystem, rabbitAnimationSystem)

	// Функция обновления анимаций через менеджер (устраняет дублирование)
	updateAnimalAnimations := func() {
		animationManager.UpdateAllAnimations(world, 1.0/TicksPerSecond)
	}

	// Диагностический цикл - ТОЧНО как в headless
	deltaTime := float32(1.0 / TicksPerSecond)
	maxTicks := MaxTicks

	for tick := 0; tick < maxTicks; tick++ {
		// Обновляем мир ТОЧНО как в headless
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// КРИТИЧЕСКИ ВАЖНО: обновляем анимации для корректной работы боевой системы
		updateAnimalAnimations()

		// Каждую секунду выводим диагностику
		if tick%DiagnosticInterval == 0 {
			second := tick / DiagnosticInterval

			// Подсчитываем живых животных и их состояния
			currentWolves := 0
			currentRabbits := 0
			attackingWolves := 0
			eatingWolves := 0
			corpses := 0

			world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
				animalType, ok := world.GetAnimalType(entity)
				if !ok {
					return
				}

				if animalType == core.TypeWolf {
					currentWolves++

					// Проверяем состояния волка
					if world.HasComponent(entity, core.MaskAttackState) {
						attackingWolves++
					}
					if world.HasComponent(entity, core.MaskEatingState) {
						eatingWolves++
					}
				} else if animalType == core.TypeRabbit {
					if world.HasComponent(entity, core.MaskCorpse) {
						corpses++
					} else {
						currentRabbits++
					}
				}
			})

			fmt.Printf("[%2ds] Живые зайцы: %2d, Трупы: %2d, Волки: %d (атакуют: %d, едят: %d)\n",
				second, currentRabbits, corpses, currentWolves, attackingWolves, eatingWolves)

			// Детальная информация о волках
			if currentWolves > 0 {
				fmt.Printf("      Детали волков:\n")
				wolfIndex := 0
				world.ForEachWith(core.MaskAnimalType|core.MaskHunger, func(entity core.EntityID) {
					animalType, ok := world.GetAnimalType(entity)
					if !ok || animalType != core.TypeWolf {
						return
					}

					hunger, _ := world.GetHunger(entity)
					health, _ := world.GetHealth(entity)

					status := "Живой"
					if health.Current <= 0 {
						status = "МЁРТВ"
					}

					// Детальная диагностика состояний
					attackState := world.HasComponent(entity, core.MaskAttackState)
					eatingState := world.HasComponent(entity, core.MaskEatingState)

					var eatingTarget core.EntityID = 0
					if eatingState {
						if eating, hasEating := world.GetEatingState(entity); hasEating {
							eatingTarget = eating.Target
						}
					}

					fmt.Printf("        Волк #%d: %s, голод %.1f%%, здоровье %d\n",
						wolfIndex+1, status, hunger.Value, health.Current)
					fmt.Printf("                  AttackState: %t, EatingState: %t (цель: %d)\n",
						attackState, eatingState, eatingTarget)
					wolfIndex++
				})
			}

			// Остановка если все волки мертвы
			if currentWolves == 0 {
				fmt.Printf("❌ ВСЕ ВОЛКИ МЕРТВЫ на %d секунде!\n", second)
				break
			}
		}

		time.Sleep(time.Microsecond * SleepMicroseconds)
	}

	fmt.Println("\n=== ДИАГНОСТИКА ЗАВЕРШЕНА ===")
}
