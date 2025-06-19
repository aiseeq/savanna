package common

import (
	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// CreateTestSystemManager создает стандартный набор систем для интеграционных тестов
// Устраняет дублирование создания систем в 30+ тестах
func CreateTestSystemManager(worldSize float32) *core.SystemManager {
	systemManager := core.NewSystemManager()

	// Создаем vegetation систему (нужна для feeding)
	vegetationSystem := CreateTestVegetationSystem(worldSize)
	systemManager.AddSystem(vegetationSystem)

	// КРИТИЧЕСКИ ВАЖНО: Анимационная система для корректной работы атак
	animationAdapter := NewAnimationSystemAdapter()
	systemManager.AddSystem(animationAdapter)

	// Системы движения
	movementSystem := simulation.NewMovementSystem(worldSize, worldSize)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	// Системы поведения
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})

	// Системы питания
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.FeedingSystemAdapter{System: feedingSystem})

	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.GrassEatingSystemAdapter{System: grassEatingSystem})

	// Боевые системы - используем новую CombatSystem
	combatSystem := simulation.NewCombatSystem()
	systemManager.AddSystem(combatSystem)

	// Системы урона и трупов
	damageSystem := simulation.NewDamageSystem()
	systemManager.AddSystem(damageSystem)

	corpseSystem := simulation.NewCorpseSystem()
	systemManager.AddSystem(corpseSystem)

	eatingSystem := simulation.NewEatingSystem()
	systemManager.AddSystem(eatingSystem)

	return systemManager
}

// CreateMinimalSystemManager создает минимальный набор систем (для простых тестов)
func CreateMinimalSystemManager(worldSize float32) *core.SystemManager {
	systemManager := core.NewSystemManager()

	// Создаем vegetation систему для behavior системы
	vegetationSystem := CreateTestVegetationSystem(worldSize)
	systemManager.AddSystem(vegetationSystem)

	// КРИТИЧЕСКИ ВАЖНО: Анимационная система даже в минимальном наборе
	animationAdapter := NewAnimationSystemAdapter()
	systemManager.AddSystem(animationAdapter)

	// Только основные системы
	movementSystem := simulation.NewMovementSystem(worldSize, worldSize)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})

	return systemManager
}

// CreateCombatSystemManager создает системы для тестов боя
func CreateCombatSystemManager(worldSize float32) *core.SystemManager {
	systemManager := core.NewSystemManager()

	// КРИТИЧЕСКИ ВАЖНО: Анимационная система для тестов боя
	animationAdapter := NewAnimationSystemAdapter()
	systemManager.AddSystem(animationAdapter)

	// Боевые системы - используем новую CombatSystem
	combatSystem := simulation.NewCombatSystem()
	systemManager.AddSystem(combatSystem)

	// Системы урона и трупов
	damageSystem := simulation.NewDamageSystem()
	systemManager.AddSystem(damageSystem)

	corpseSystem := simulation.NewCorpseSystem()
	systemManager.AddSystem(corpseSystem)

	eatingSystem := simulation.NewEatingSystem()
	systemManager.AddSystem(eatingSystem)

	return systemManager
}

// CreateTestVegetationSystem создает систему растительности для тестов
// Устраняет дублирование создания vegetation в 35+ тестах
func CreateTestVegetationSystem(worldSize float32) *simulation.VegetationSystem {
	// Создаем конфигурацию для теста
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = int(worldSize / simulation.TileSizePixels) // Конвертируем пиксели в тайлы

	// Генерируем terrain
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Создаем vegetation систему
	return simulation.NewVegetationSystem(terrain)
}

// CreateMockVegetationSystem создает mock vegetation систему для unit тестов
func CreateMockVegetationSystem() *simulation.VegetationSystem {
	// Используем минимальный mock terrain
	terrain := NewMockTerrain(simulation.TestWorldTileSize) // Тестовый размер мира
	return simulation.NewVegetationSystem(terrain)
}

// MockTerrain простая реализация terrain для unit тестов
type MockTerrain struct {
	size int
}

// NewMockTerrain создает mock terrain заданного размера
func NewMockTerrain(size int) *MockTerrain {
	return &MockTerrain{size: size}
}

// GetSize возвращает размер terrain
func (mt *MockTerrain) GetSize() int {
	return mt.size
}

// GetTileType возвращает тип тайла (всегда трава для простоты)
func (mt *MockTerrain) GetTileType(x, y int) generator.TileType {
	if x >= 0 && x < mt.size && y >= 0 && y < mt.size {
		return generator.TileGrass
	}
	return generator.TileWater // За границами - вода
}

// GetGrassAmount возвращает фиксированное количество травы
func (mt *MockTerrain) GetGrassAmount(x, y int) float32 {
	if mt.GetTileType(x, y) != generator.TileGrass {
		return 0
	}
	return simulation.TestGrassAmount // Стандартное количество травы в тестах
}

// SetGrassAmount устанавливает количество травы (ничего не делает в mock)
func (mt *MockTerrain) SetGrassAmount(x, y int, amount float32) {
	// Mock - ничего не делаем
}
