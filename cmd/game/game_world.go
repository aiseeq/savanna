package main

import (
	"fmt"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Константы популяции животных для начального размещения
const (
	InitialRabbitCount = 20 // Начальное количество зайцев
	InitialWolfCount   = 3  // Начальное количество волков
)

// WorldStats содержит статистику мира (заменяет map[string]interface{})
type WorldStats struct {
	Rabbits      int `json:"rabbits"`      // Количество зайцев
	Wolves       int `json:"wolves"`       // Количество волков
	TotalAnimals int `json:"totalAnimals"` // Общее количество животных
}

// GameWorld управляет симуляцией мира и его системами
// Соблюдает SRP - единственная ответственность: симуляция экосистемы
type GameWorld struct {
	world            *core.World
	systemManager    *core.SystemManager
	animationManager *AnimationManager
	terrain          *generator.Terrain
}

// NewGameWorld создаёт новый игровой мир
func NewGameWorld(worldWidth, worldHeight int, seed int64, terrain *generator.Terrain) *GameWorld {
	world := core.NewWorld(float32(worldWidth), float32(worldHeight), seed)
	systemManager := core.NewSystemManager()
	animationManager := NewAnimationManager()

	gw := &GameWorld{
		world:            world,
		systemManager:    systemManager,
		animationManager: animationManager,
		terrain:          terrain,
	}

	// Инициализируем системы симуляции
	gw.initializeSystems(worldWidth, worldHeight)

	return gw
}

// GetWorld возвращает мир для доступа к данным
func (gw *GameWorld) GetWorld() *core.World {
	return gw.world
}

// GetTerrain возвращает ландшафт
func (gw *GameWorld) GetTerrain() *generator.Terrain {
	return gw.terrain
}

// REMOVED: Старые методы отрисовки больше не используются
// Новая изометрическая система отрисовки используется напрямую в main.go

// Update обновляет симуляцию мира
func (gw *GameWorld) Update(deltaTime float32) {
	gw.world.Update(deltaTime)

	// КРИТИЧЕСКИЙ ИСПРАВЛЕНИЕ: Анимации должны обновляться ПЕРЕД системами
	// чтобы GrassEatingSystem видел актуальные значения анимационных таймеров
	gw.animationManager.UpdateAnimalAnimations(gw.world, deltaTime)

	gw.systemManager.Update(gw.world, deltaTime)
}

// initializeSystems инициализирует все системы симуляции
func (gw *GameWorld) initializeSystems(worldWidth, worldHeight int) {
	// Создаём системы (SRP рефакторинг: разделённые специализированные системы)
	vegetationSystem := simulation.NewVegetationSystem(gw.terrain)

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	satiationSystem := simulation.NewSatiationSystem() // 1. Только управление сытостью
	// 2. Только поиск травы и создание EatingState (DIP: использует интерфейс)
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
	satiationSpeedModifier := simulation.NewSatiationSpeedModifierSystem() // 3. Только влияние сытости на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()             // 4. Только урон от истощения

	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem) // DIP: использует интерфейс VegetationProvider
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	// Используем реальные размеры мира
	movementSystem := simulation.NewMovementSystem(float32(worldWidth), float32(worldHeight))
	// Уже включает DamageSystem внутри
	combatSystem := simulation.NewCombatSystem()

	// Добавляем системы в правильном порядке (КРИТИЧЕСКИ ВАЖЕН ДЛЯ ПИТАНИЯ!)
	gw.systemManager.AddSystem(vegetationSystem)                 // 1. Рост травы
	gw.systemManager.AddSystem(&adapters.SatiationSystemAdapter{ // 2. Управление сытостью
		System: satiationSystem,
	})
	gw.systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{ // 3. Создание EatingState
		System: grassSearchSystem,
	})
	gw.systemManager.AddSystem(grassEatingSystem)               // 4. Дискретное поедание травы
	gw.systemManager.AddSystem(&adapters.BehaviorSystemAdapter{ // 5. Поведение (проверяет EatingState)
		System: animalBehaviorSystem,
	})
	gw.systemManager.AddSystem(&adapters.SatiationSpeedModifierSystemAdapter{ // 6. Влияние сытости на скорость
		System: satiationSpeedModifier,
	})
	gw.systemManager.AddSystem(&adapters.MovementSystemAdapter{ // 7. Движение (сбрасывает скорость едящих)
		System: movementSystem,
	})
	gw.systemManager.AddSystem(combatSystem)                            // 8. Система боя
	gw.systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{ // 9. Урон от истощения
		System: starvationDamage,
	})

	// Загружаем анимации для всех типов животных
	if err := gw.animationManager.LoadAnimationsFromConfig(); err != nil {
		// В GameWorld мы не можем вернуть ошибку, поэтому просто логируем
		// В реальном приложении здесь может быть более сложная обработка
		return
	}
}

// PopulateWorld заполняет мир животными используя PopulationGenerator
func (gw *GameWorld) PopulateWorld(cfg *config.Config) {
	// ИСПРАВЛЕНИЕ: Используем PopulationGenerator вместо случайного размещения
	popGen := generator.NewPopulationGenerator(cfg, gw.terrain)
	placements := popGen.Generate()

	worldWidth, worldHeight := gw.world.GetWorldDimensions()

	for _, placement := range placements {
		// PopulationGenerator уже возвращает координаты в пикселях
		// CreateAnimal также ожидает координаты в пикселях
		x := placement.X
		y := placement.Y

		// Проверяем границы размещения (в пикселях)
		if x < 0 || x > worldWidth*32 || y < 0 || y > worldHeight*32 {
			fmt.Printf("WARNING: Animal placed outside world bounds!\n")
		}

		simulation.CreateAnimal(gw.world, placement.Type, x, y)
	}

	errors := popGen.ValidatePlacement(placements)
	if len(errors) > 0 {
		fmt.Printf("Предупреждения размещения GUI: %v\n", errors)
	}

	// Сводка размещения животных (без детальной отладки)
	popStats := popGen.GetStats(placements)
	fmt.Printf("Размещено животных: %d зайцев, %d волков\n",
		popStats["rabbits"], popStats["wolves"])
}

// GetStats возвращает типизированную статистику мира
func (gw *GameWorld) GetStats() WorldStats {
	var stats WorldStats

	// Подсчитываем животных
	gw.world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, ok := gw.world.GetAnimalType(entity)
		if !ok {
			return
		}

		switch animalType {
		case core.TypeRabbit:
			stats.Rabbits++
		case core.TypeWolf:
			stats.Wolves++
		}
	})

	stats.TotalAnimals = stats.Rabbits + stats.Wolves
	return stats
}
