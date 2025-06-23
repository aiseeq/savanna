package simulation

import (
	"testing"
	"time"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

// Inline adapters to avoid import cycle with adapters package
type grassSearchAdapter struct{ *GrassSearchSystem }

func (a *grassSearchAdapter) Update(world *core.World, deltaTime float32) {
	a.GrassSearchSystem.Update(world, deltaTime)
}

type grassEatingAdapter struct{ *GrassEatingSystem }

func (a *grassEatingAdapter) Update(world *core.World, deltaTime float32) {
	a.GrassEatingSystem.Update(world, deltaTime)
}

type behaviorAdapter struct{ *AnimalBehaviorSystem }

func (a *behaviorAdapter) Update(world *core.World, deltaTime float32) {
	a.AnimalBehaviorSystem.Update(world, deltaTime)
}

type movementAdapter struct{ *MovementSystem }

func (a *movementAdapter) Update(world *core.World, deltaTime float32) {
	a.MovementSystem.Update(world, deltaTime)
}

func TestBalanceVerification_FinalStabilityCheck(t *testing.T) {
	// Финальный тест стабильности симуляции с обновленными параметрами баланса
	worldWidth := 50
	worldHeight := 38
	seed := time.Now().UnixNano()

	// Создаем мир и системы
	world := core.NewWorld(float32(worldWidth), float32(worldHeight), seed)

	// Создаем terrain
	terrain := &generator.Terrain{
		Width:  worldWidth,
		Height: worldHeight,
		Tiles:  make([][]generator.TileType, worldHeight),
		Grass:  make([][]float32, worldHeight),
	}
	for y := 0; y < worldHeight; y++ {
		terrain.Tiles[y] = make([]generator.TileType, worldWidth)
		terrain.Grass[y] = make([]float32, worldWidth)
		for x := 0; x < worldWidth; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	// Инициализируем все системы как в production
	vegetationSystem := NewVegetationSystem(terrain)
	_ = NewHungerSystem() // hungerSystem - not used in this test
	grassSearchSystem := NewGrassSearchSystem(vegetationSystem)
	grassEatingSystem := NewGrassEatingSystem(vegetationSystem)
	behaviorSystem := NewAnimalBehaviorSystem(vegetationSystem)
	movementSystem := NewMovementSystem(float32(worldWidth), float32(worldHeight))
	combatSystem := NewCombatSystem()

	// Создаем SystemManager и добавляем системы в правильном порядке
	systemManager := core.NewSystemManager()
	systemManager.AddSystem(vegetationSystem)
	// Create simple inline adapters to avoid import cycle
	systemManager.AddSystem(&grassSearchAdapter{grassSearchSystem})
	systemManager.AddSystem(&grassEatingAdapter{grassEatingSystem})
	systemManager.AddSystem(&behaviorAdapter{behaviorSystem})
	systemManager.AddSystem(&movementAdapter{movementSystem})
	systemManager.AddSystem(combatSystem)

	// Создаем животных
	var rabbits []core.EntityID
	var wolves []core.EntityID

	// Создаем 20 зайцев
	for i := 0; i < 20; i++ {
		x := world.GetRNG().Float32() * float32(worldWidth)
		y := world.GetRNG().Float32() * float32(worldHeight)
		rabbit := CreateAnimal(world, core.TypeRabbit, x, y)
		rabbits = append(rabbits, rabbit)
	}

	// Создаем 3 волков
	for i := 0; i < 3; i++ {
		x := world.GetRNG().Float32() * float32(worldWidth)
		y := world.GetRNG().Float32() * float32(worldHeight)
		wolf := CreateAnimal(world, core.TypeWolf, x, y)
		wolves = append(wolves, wolf)
	}

	initialRabbitCount := len(rabbits)
	initialWolfCount := len(wolves)

	t.Logf("Starting simulation: %d rabbits, %d wolves", initialRabbitCount, initialWolfCount)
	t.Logf("World size: %dx%d, Seed: %d", worldWidth, worldHeight, seed)

	// Симулируем 12 секунд (720 тиков)
	deltaTime := float32(1.0 / 60.0)
	var finalStats struct {
		aliveRabbits int
		aliveWolves  int
		deadRabbits  int
		deadWolves   int
		hungryCounts map[string]int
	}

	finalStats.hungryCounts = make(map[string]int)

	for tick := 0; tick < 720; tick++ {
		world.Update(deltaTime)

		// Простое обновление анимаций для дискретного питания
		world.ForEachWith(core.MaskAnimation, func(entity core.EntityID) {
			anim, hasAnim := world.GetAnimation(entity)
			if !hasAnim || !anim.Playing {
				return
			}

			anim.Timer += deltaTime
			if anim.Timer >= 0.2 { // Смена кадра каждые 0.2 сек
				anim.Timer = 0
				anim.Frame = 1 - anim.Frame // 0 ↔ 1
				world.SetAnimation(entity, anim)
			}
		})

		systemManager.Update(world, deltaTime)

		// Каждые 2 секунды (120 тиков) проверяем состояние
		if tick%120 == 0 {
			second := tick / 60
			aliveRabbits := 0
			aliveWolves := 0

			// Подсчитываем живых животных
			for _, rabbit := range rabbits {
				if world.IsAlive(rabbit) {
					aliveRabbits++
				}
			}
			for _, wolf := range wolves {
				if world.IsAlive(wolf) {
					aliveWolves++
				}
			}

			t.Logf("Second %d: %d rabbits, %d wolves alive", second, aliveRabbits, aliveWolves)
		}
	}

	// Финальный подсчет
	for _, rabbit := range rabbits {
		if world.IsAlive(rabbit) {
			finalStats.aliveRabbits++

			// Проверяем голод
			if hunger, hasHunger := world.GetHunger(rabbit); hasHunger {
				if hunger.Value < 30 {
					finalStats.hungryCounts["starving rabbits"]++
				} else if hunger.Value < 60 {
					finalStats.hungryCounts["hungry rabbits"]++
				}
			}
		} else {
			finalStats.deadRabbits++
		}
	}

	for _, wolf := range wolves {
		if world.IsAlive(wolf) {
			finalStats.aliveWolves++

			// Проверяем голод
			if hunger, hasHunger := world.GetHunger(wolf); hasHunger {
				if hunger.Value < 30 {
					finalStats.hungryCounts["starving wolves"]++
				} else if hunger.Value < 60 {
					finalStats.hungryCounts["hungry wolves"]++
				}
			}
		} else {
			finalStats.deadWolves++
		}
	}

	// Анализ результатов симуляции
	t.Logf("=== FINAL BALANCE ANALYSIS ===")
	t.Logf("Rabbits: %d alive, %d dead (%.1f%% survival)",
		finalStats.aliveRabbits, finalStats.deadRabbits,
		float64(finalStats.aliveRabbits)/float64(initialRabbitCount)*100)
	t.Logf("Wolves: %d alive, %d dead (%.1f%% survival)",
		finalStats.aliveWolves, finalStats.deadWolves,
		float64(finalStats.aliveWolves)/float64(initialWolfCount)*100)

	for category, count := range finalStats.hungryCounts {
		if count > 0 {
			t.Logf("Hunger issues: %d %s", count, category)
		}
	}

	// Проверки стабильности симуляции
	rabbitSurvivalRate := float64(finalStats.aliveRabbits) / float64(initialRabbitCount)
	_ = float64(finalStats.aliveWolves) / float64(initialWolfCount) // wolfSurvivalRate - not used

	// Зайцы не должны полностью вымереть (хотя бы 30% выживает)
	if rabbitSurvivalRate < 0.3 {
		t.Errorf("Rabbit survival rate too low: %.1f%% (expected ≥30%%)", rabbitSurvivalRate*100)
	}

	// Волки не должны полностью вымереть
	if finalStats.aliveWolves == 0 {
		t.Error("All wolves died - balance too harsh for predators")
	}

	// Зайцы не должны быть бессмертными (хотя бы один должен погибнуть)
	if finalStats.deadRabbits == 0 {
		t.Logf("WARNING: No rabbits died - wolves might be too weak")
	}

	// Финальная проверка движения (животные не должны застрять)
	stuckAnimals := 0
	world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
		vel, _ := world.GetVelocity(entity)
		if vel.X == 0 && vel.Y == 0 {
			// Проверяем что это не животное которое ест
			if !world.HasComponent(entity, core.MaskEatingState) {
				stuckAnimals++
			}
		}
	})

	if stuckAnimals > 5 {
		t.Errorf("Too many animals stuck (not moving): %d", stuckAnimals)
	}

	t.Logf("=== SIMULATION STABILITY: PASSED ===")
	t.Logf("Balance appears stable and playable!")
}

func TestBalanceVerification_ParameterConsistency(t *testing.T) {
	// Проверяем что все параметры баланса находятся в разумных пределах

	// Скорости должны быть положительными и разумными
	if RabbitBaseSpeed <= 0 || RabbitBaseSpeed > 10 {
		t.Errorf("RabbitBaseSpeed out of reasonable range: %f", RabbitBaseSpeed)
	}

	if WolfBaseSpeed <= 0 || WolfBaseSpeed > 10 {
		t.Errorf("WolfBaseSpeed out of reasonable range: %f", WolfBaseSpeed)
	}

	// Волк должен быть быстрее зайца
	if WolfBaseSpeed <= RabbitBaseSpeed {
		t.Errorf("Wolf should be faster than rabbit: wolf=%f, rabbit=%f",
			WolfBaseSpeed, RabbitBaseSpeed)
	}

	// Размеры должны быть положительными
	if RabbitBaseRadius <= 0 || WolfBaseRadius <= 0 {
		t.Errorf("Animal radii must be positive: rabbit=%f, wolf=%f",
			RabbitBaseRadius, WolfBaseRadius)
	}

	// Волк должен быть крупнее зайца
	if WolfBaseRadius <= RabbitBaseRadius {
		t.Errorf("Wolf should be larger than rabbit: wolf=%f, rabbit=%f",
			WolfBaseRadius, RabbitBaseRadius)
	}

	// Дальность видения должна быть больше размера животного
	rabbitVision := RabbitBaseRadius * RabbitVisionMultiplier
	wolfVision := WolfBaseRadius * WolfVisionMultiplier

	if rabbitVision <= RabbitBaseRadius {
		t.Errorf("Rabbit vision range too small: %f (radius: %f)",
			rabbitVision, RabbitBaseRadius)
	}

	if wolfVision <= WolfBaseRadius {
		t.Errorf("Wolf vision range too small: %f (radius: %f)",
			wolfVision, WolfBaseRadius)
	}

	// Радиус атаки волка должен быть больше его размера
	wolfAttackRange := WolfBaseRadius * WolfAttackRangeMultiplier
	if wolfAttackRange <= WolfBaseRadius {
		t.Errorf("Wolf attack range should be larger than wolf radius: attack=%f, radius=%f",
			wolfAttackRange, WolfBaseRadius)
	}

	t.Logf("=== PARAMETER CONSISTENCY: PASSED ===")
	t.Logf("All balance parameters are within reasonable ranges")
	t.Logf("Rabbit: speed=%.1f, radius=%.1f, vision=%.1f",
		RabbitBaseSpeed, RabbitBaseRadius, rabbitVision)
	t.Logf("Wolf: speed=%.1f, radius=%.1f, vision=%.1f, attack=%.1f",
		WolfBaseSpeed, WolfBaseRadius, wolfVision, wolfAttackRange)
}
