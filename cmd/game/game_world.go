package main

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/constants"
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

// LoD Compliance: инкапсулированные методы для отрисовки
// Теперь Game не должен знать о внутренних объектах менеджеров

// DrawTerrain отрисовывает ландшафт (соблюдение LoD)
func (gw *GameWorld) DrawTerrain(screen *ebiten.Image, camera Camera, renderer TerrainRenderer) {
	// Инкапсулируем логику, Game не получает прямой доступ к terrain
	renderer.DrawTerrain(screen, camera, gw.terrain)
}

// DrawAnimals отрисовывает животных (соблюдение LoD)
func (gw *GameWorld) DrawAnimals(screen *ebiten.Image, camera Camera, renderer AnimalRenderer) {
	// Инкапсулируем логику, Game не получает прямой доступ к world
	renderer.DrawAnimals(screen, camera, gw.world)
}

// TerrainRenderer интерфейс для отрисовки ландшафта (LoD)
type TerrainRenderer interface {
	DrawTerrain(screen *ebiten.Image, camera Camera, terrain *generator.Terrain)
}

// AnimalRenderer интерфейс для отрисовки животных (LoD)
type AnimalRenderer interface {
	DrawAnimals(screen *ebiten.Image, camera Camera, world *core.World)
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
	// Создаём системы (SRP рефакторинг: разделённые специализированные системы)
	vegetationSystem := simulation.NewVegetationSystem(gw.terrain)

	// НОВЫЕ СИСТЕМЫ (следуют принципу SRP):
	hungerSystem := simulation.NewHungerSystem() // 1. Только управление голодом
	// 2. Только поиск травы и создание EatingState (DIP: использует интерфейс)
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
	hungerSpeedModifier := simulation.NewHungerSpeedModifierSystem() // 3. Только влияние голода на скорость
	starvationDamage := simulation.NewStarvationDamageSystem()       // 4. Только урон от голода

	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem) // DIP: использует интерфейс VegetationProvider
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	// Стандартные размеры мира
	movementSystem := simulation.NewMovementSystem(constants.DefaultWorldSizePixels, constants.DefaultWorldSizePixels)
	// Уже включает DamageSystem внутри
	combatSystem := simulation.NewCombatSystem()

	// Добавляем системы в правильном порядке (КРИТИЧЕСКИ ВАЖЕН ДЛЯ ПИТАНИЯ!)
	gw.systemManager.AddSystem(vegetationSystem)              // 1. Рост травы
	gw.systemManager.AddSystem(&adapters.HungerSystemAdapter{ // 2. Управление голодом
		System: hungerSystem,
	})
	gw.systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{ // 3. Создание EatingState
		System: grassSearchSystem,
	})
	gw.systemManager.AddSystem(grassEatingSystem)               // 4. Дискретное поедание травы
	gw.systemManager.AddSystem(&adapters.BehaviorSystemAdapter{ // 5. Поведение (проверяет EatingState)
		System: animalBehaviorSystem,
	})
	gw.systemManager.AddSystem(&adapters.HungerSpeedModifierSystemAdapter{ // 6. Влияние голода на скорость
		System: hungerSpeedModifier,
	})
	gw.systemManager.AddSystem(&adapters.MovementSystemAdapter{ // 7. Движение (сбрасывает скорость едящих)
		System: movementSystem,
	})
	gw.systemManager.AddSystem(combatSystem)                            // 8. Система боя
	gw.systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{ // 9. Урон от голода
		System: starvationDamage,
	})

	// Загружаем анимации для всех типов животных
	if err := gw.animationManager.LoadAnimationsFromConfig(); err != nil {
		// В GameWorld мы не можем вернуть ошибку, поэтому просто логируем
		// В реальном приложении здесь может быть более сложная обработка
		return
	}
}

// PopulateWorld заполняет мир животными (унифицированная система)
func (gw *GameWorld) PopulateWorld() {
	// Размещаем зайцев
	for i := 0; i < 20; i++ {
		x := gw.world.GetRNG().Float32() * constants.DefaultWorldSizePixels // Используем стандартные размеры мира
		y := gw.world.GetRNG().Float32() * constants.DefaultWorldSizePixels
		simulation.CreateAnimal(gw.world, core.TypeRabbit, x, y)
	}

	// Размещаем волков
	for i := 0; i < 3; i++ {
		x := gw.world.GetRNG().Float32() * constants.DefaultWorldSizePixels
		y := gw.world.GetRNG().Float32() * constants.DefaultWorldSizePixels
		simulation.CreateAnimal(gw.world, core.TypeWolf, x, y)
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
