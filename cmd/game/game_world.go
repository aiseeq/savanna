package main

import (
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

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
	gw.initializeSystems()

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

// Update обновляет симуляцию мира
func (gw *GameWorld) Update(deltaTime float32) {
	gw.world.Update(deltaTime)

	// КРИТИЧЕСКИЙ ИСПРАВЛЕНИЕ: Анимации должны обновляться ПЕРЕД системами
	// чтобы GrassEatingSystem видел актуальные значения анимационных таймеров
	gw.animationManager.UpdateAnimalAnimations(gw.world, deltaTime)

	gw.systemManager.Update(gw.world, deltaTime)
}

// initializeSystems инициализирует все системы симуляции
func (gw *GameWorld) initializeSystems() {
	// Создаём системы
	vegetationSystem := simulation.NewVegetationSystem(gw.terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := simulation.NewMovementSystem(1600.0, 1600.0) // Фиксированные размеры мира
	combatSystem := simulation.NewCombatSystem()                   // Уже включает DamageSystem внутри

	// Добавляем системы в правильном порядке
	// Используем адаптеры для систем с ISP интерфейсами
	gw.systemManager.AddSystem(vegetationSystem)
	gw.systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})         // Создаёт EatingState для травы
	gw.systemManager.AddSystem(grassEatingSystem)                                             // Дискретное поедание травы по кадрам анимации
	gw.systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem}) // Проверяет EatingState и не мешает еде
	gw.systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})       // Сбрасывает скорость едящих
	gw.systemManager.AddSystem(combatSystem)                                                  // Система боя (включает DamageSystem)

	// Загружаем анимации для всех типов животных
	if err := gw.animationManager.LoadAnimationsFromConfig(); err != nil {
		// В GameWorld мы не можем вернуть ошибку, поэтому просто логируем
		// В реальном приложении здесь может быть более сложная обработка
		return
	}
}

// PopulateWorld заполняет мир животными
func (gw *GameWorld) PopulateWorld() {
	// Создаём фабрику животных
	animalFactory := simulation.NewAnimalFactory()

	// Размещаем зайцев
	for i := 0; i < 20; i++ {
		x := gw.world.GetRNG().Float32() * 1600.0 // Используем размеры мира
		y := gw.world.GetRNG().Float32() * 1600.0
		animalFactory.CreateRabbit(gw.world, x, y)
	}

	// Размещаем волков
	for i := 0; i < 3; i++ {
		x := gw.world.GetRNG().Float32() * 1600.0
		y := gw.world.GetRNG().Float32() * 1600.0
		animalFactory.CreateWolf(gw.world, x, y)
	}
}

// GetStats возвращает статистику мира
func (gw *GameWorld) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Подсчитываем животных
	rabbitCount := 0
	wolfCount := 0

	gw.world.ForEachWith(core.MaskAnimalType, func(entity core.EntityID) {
		animalType, ok := gw.world.GetAnimalType(entity)
		if !ok {
			return
		}

		switch animalType {
		case core.TypeRabbit:
			rabbitCount++
		case core.TypeWolf:
			wolfCount++
		}
	})

	stats["rabbits"] = rabbitCount
	stats["wolves"] = wolfCount
	stats["total_animals"] = rabbitCount + wolfCount

	return stats
}
