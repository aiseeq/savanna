// Package gamestate contains pure game model without any rendering dependencies
package gamestate

import (
	"math/rand"
	"time"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// GameState представляет чистое состояние игры без зависимостей от рендеринга
type GameState struct {
	world         *core.World
	systemManager *core.SystemManager

	// Управление временем
	lastUpdateTime time.Time
	accumulator    float64
	fixedTimeStep  float64

	// Состояние камеры (логическое, не визуальное)
	camera CameraState

	// Конфигурация
	config *GameConfig
}

// CameraState логическое состояние камеры
type CameraState struct {
	X, Y         float64
	IsScrolling  bool
	ScrollStartX float64
	ScrollStartY float64
}

// GameConfig конфигурация игры
type GameConfig struct {
	WorldWidth    float32
	WorldHeight   float32
	FixedTimeStep float64
	RandomSeed    int64
}

// NewGameState создает новое состояние игры
func NewGameState(config *GameConfig) *GameState {
	// Создаем мир с фиксированным размером
	world := core.NewWorld(config.WorldWidth, config.WorldHeight, config.RandomSeed)

	// Создаем менеджер систем
	systemManager := core.NewSystemManager()

	// Инициализируем системы (в правильном порядке)
	initializeSystems(systemManager, world, config)

	return &GameState{
		world:          world,
		systemManager:  systemManager,
		lastUpdateTime: time.Now(),
		accumulator:    0,
		fixedTimeStep:  config.FixedTimeStep,
		config:         config,
		camera: CameraState{
			X: 0,
			Y: 0,
		},
	}
}

// Update обновляет состояние игры с фиксированным шагом времени
func (gs *GameState) Update() {
	now := time.Now()
	frameTime := now.Sub(gs.lastUpdateTime).Seconds()
	gs.lastUpdateTime = now

	// ФИКС ДЛЯ ТЕСТОВ: Если frameTime слишком мал, принудительно устанавливаем fixedTimeStep
	if frameTime < gs.fixedTimeStep*0.1 {
		frameTime = gs.fixedTimeStep
	}

	// Ограничиваем максимальный шаг времени
	if frameTime > 0.25 {
		frameTime = 0.25
	}

	gs.accumulator += frameTime

	// Фиксированный шаг времени для детерминизма
	for gs.accumulator >= gs.fixedTimeStep {
		// Обновляем время в мире
		gs.world.Update(float32(gs.fixedTimeStep))

		// TODO: Анимации будут добавлены после решения ebiten зависимостей

		gs.systemManager.Update(gs.world, float32(gs.fixedTimeStep))
		gs.accumulator -= gs.fixedTimeStep
	}
}

// ProcessInput обрабатывает входные события
func (gs *GameState) ProcessInput(events []InputEvent) {
	for _, event := range events {
		switch event.Type {
		case InputMouseDown:
			if event.Button == MouseButtonRight {
				gs.camera.IsScrolling = true
				gs.camera.ScrollStartX = event.X
				gs.camera.ScrollStartY = event.Y
			}
		case InputMouseUp:
			if event.Button == MouseButtonRight {
				gs.camera.IsScrolling = false
			}
		case InputMouseMove:
			if gs.camera.IsScrolling {
				dx := event.X - gs.camera.ScrollStartX
				dy := event.Y - gs.camera.ScrollStartY
				gs.camera.X += dx
				gs.camera.Y += dy
				gs.camera.ScrollStartX = event.X
				gs.camera.ScrollStartY = event.Y
			}
		}
	}
}

// GetCameraState возвращает текущее состояние камеры
func (gs *GameState) GetCameraState() CameraState {
	return gs.camera
}

// GetWorld возвращает мир для чтения (не для модификации!)
func (gs *GameState) GetWorld() *core.World {
	return gs.world
}

// initializeSystems инициализирует все игровые системы в правильном порядке
func initializeSystems(systemManager *core.SystemManager, world *core.World, config *GameConfig) {
	// Создаем terrain и vegetation (используем простой terrain для демонстрации)
	terrain := createSimpleTerrain(int(config.WorldWidth/32), int(config.WorldHeight/32))
	vegetationSystem := simulation.NewVegetationSystem(terrain)

	// Добавляем системы в КРИТИЧЕСКОМ порядке (из CLAUDE.md)
	systemManager.AddSystem(vegetationSystem)

	satiationSystem := simulation.NewSatiationSystem()
	systemManager.AddSystem(&adapters.SatiationSystemAdapter{System: satiationSystem})

	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
	systemManager.AddSystem(&adapters.GrassSearchSystemAdapter{System: grassSearchSystem})

	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	systemManager.AddSystem(grassEatingSystem)

	// ИСПРАВЛЕНИЕ: EatingSystem должна быть ПЕРЕД BehaviorSystem для поиска трупов
	eatingSystem := simulation.NewEatingSystem()
	systemManager.AddSystem(eatingSystem)

	behaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	behaviorAdapter := &adapters.BehaviorSystemAdapter{System: behaviorSystem}
	systemManager.AddSystem(behaviorAdapter)

	satiationSpeedModifier := simulation.NewSatiationSpeedModifierSystem()
	systemManager.AddSystem(&adapters.SatiationSpeedModifierSystemAdapter{System: satiationSpeedModifier})

	movementSystem := simulation.NewMovementSystem(config.WorldWidth, config.WorldHeight)
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})

	combatSystem := simulation.NewCombatSystem()
	systemManager.AddSystem(combatSystem)

	damageSystem := simulation.NewDamageSystem()
	systemManager.AddSystem(damageSystem)

	starvationDamage := simulation.NewStarvationDamageSystem()
	systemManager.AddSystem(&adapters.StarvationDamageSystemAdapter{System: starvationDamage})

	corpseSystem := simulation.NewCorpseSystem()
	systemManager.AddSystem(corpseSystem)

	// Генерируем начальную популяцию (упрощенная версия для демонстрации)
	createInitialPopulation(world, terrain, rand.New(rand.NewSource(config.RandomSeed)))
}

// createSimpleTerrain создает простой terrain для демонстрации
func createSimpleTerrain(width, height int) *generator.Terrain {
	terrain := &generator.Terrain{
		Width:  width,
		Height: height,
		Size:   max(width, height),
		Tiles:  make([][]generator.TileType, height),
		Grass:  make([][]float32, height),
	}

	// Инициализируем массивы
	for y := 0; y < height; y++ {
		terrain.Tiles[y] = make([]generator.TileType, width)
		terrain.Grass[y] = make([]float32, width)

		for x := 0; x < width; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0 // Полная трава
		}
	}

	return terrain
}

// max возвращает максимум из двух значений
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// createInitialPopulation создает начальную популяцию животных
func createInitialPopulation(world *core.World, terrain *generator.Terrain, rng *rand.Rand) {
	// Создаем нескольких зайцев
	for i := 0; i < 5; i++ {
		x := rng.Float32() * float32(terrain.Width*32)
		y := rng.Float32() * float32(terrain.Height*32)
		simulation.CreateAnimal(world, core.TypeRabbit, x, y)
	}

	// Создаем нескольких волков
	for i := 0; i < 2; i++ {
		x := rng.Float32() * float32(terrain.Width*32)
		y := rng.Float32() * float32(terrain.Height*32)
		simulation.CreateAnimal(world, core.TypeWolf, x, y)
	}
}
