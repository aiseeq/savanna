package chaos

import (
	"math/rand"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Chaos Engineering для тестирования устойчивости экосистемы
// Вводим случайные сбои и проверяем что система остаётся стабильной

type ChaosScenario struct {
	world             *core.World
	terrain           *generator.Terrain
	vegetationSystem  *simulation.VegetationSystem
	grassSearchSystem *simulation.GrassSearchSystem
	grassEatingSystem *simulation.GrassEatingSystem
	satiationSystem   *simulation.SatiationSystem
	animals           []core.EntityID
	rand              *rand.Rand
	t                 *testing.T
}

func newChaosScenario(t *testing.T, seed int64) *ChaosScenario {
	world := core.NewWorld(640, 640, seed)

	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 15
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	// Заполняем случайной травой
	rng := rand.New(rand.NewSource(seed))
	for x := 0; x < 15; x++ {
		for y := 0; y < 15; y++ {
			if rng.Float32() < 0.7 { // 70% тайлов с травой
				terrain.SetTileType(x, y, generator.TileGrass)
				terrain.SetGrassAmount(x, y, rng.Float32()*100)
			}
		}
	}

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	grassSearchSystem := simulation.NewGrassSearchSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	satiationSystem := simulation.NewSatiationSystem()

	return &ChaosScenario{
		world:             world,
		terrain:           terrain,
		vegetationSystem:  vegetationSystem,
		grassSearchSystem: grassSearchSystem,
		grassEatingSystem: grassEatingSystem,
		satiationSystem:   satiationSystem,
		animals:           make([]core.EntityID, 0),
		rand:              rng,
		t:                 t,
	}
}

func (cs *ChaosScenario) addRandomAnimals(count int) {
	for i := 0; i < count; i++ {
		animalType := core.TypeRabbit
		if cs.rand.Float32() < 0.2 { // 20% волков
			animalType = core.TypeWolf
		}

		x := cs.rand.Float32() * 480 // Случайная позиция
		y := cs.rand.Float32() * 480
		hunger := cs.rand.Float32() * 100 // Случайный голод

		animal := simulation.CreateAnimal(cs.world, animalType, x, y)
		cs.world.SetSatiation(animal, core.Satiation{Value: hunger})
		cs.animals = append(cs.animals, animal)
	}
}

// Chaos методы - вводят случайные сбои

func (cs *ChaosScenario) randomlyKillAnimals(probability float32) {
	for _, animal := range cs.animals {
		if cs.rand.Float32() < probability {
			cs.world.DestroyEntity(animal)
		}
	}
}

func (cs *ChaosScenario) randomlyCorruptGrass(probability float32) {
	for x := 0; x < 15; x++ {
		for y := 0; y < 15; y++ {
			if cs.rand.Float32() < probability {
				// Случайно "портим" траву
				if cs.rand.Float32() < 0.5 {
					// Убираем всю траву
					cs.terrain.SetGrassAmount(x, y, 0)
				} else {
					// Или делаем тайл непроходимым
					cs.terrain.SetTileType(x, y, generator.TileWater)
				}
			}
		}
	}
}

func (cs *ChaosScenario) randomlyDamageAnimals(probability float32) {
	for _, animal := range cs.animals {
		if cs.rand.Float32() < probability {
			if health, hasHealth := cs.world.GetHealth(animal); hasHealth {
				// Случайный урон 10-50 хитов
				damage := int16(cs.rand.Float32()*40 + 10) // Приводим к int16
				newHealth := health.Current - damage
				if newHealth < 0 {
					newHealth = 0
				}
				cs.world.SetHealth(animal, core.Health{
					Current: newHealth,
					Max:     health.Max,
				})
			}
		}
	}
}

func (cs *ChaosScenario) randomlyModifyHunger(probability float32) {
	for _, animal := range cs.animals {
		if cs.rand.Float32() < probability {
			// Случайно изменяем голод
			newHunger := cs.rand.Float32() * 100
			cs.world.SetSatiation(animal, core.Satiation{Value: newHunger})
		}
	}
}

func (cs *ChaosScenario) simulateWithChaos(ticks int, chaosLevel float32) {
	deltaTime := float32(1.0 / 60.0)

	for tick := 0; tick < ticks; tick++ {
		// Нормальное обновление систем
		cs.satiationSystem.Update(cs.world, deltaTime)
		cs.grassSearchSystem.Update(cs.world, deltaTime)
		cs.grassEatingSystem.Update(cs.world, deltaTime)

		// Chaos events - случайные сбои
		if cs.rand.Float32() < chaosLevel {
			switch cs.rand.Intn(4) {
			case 0:
				cs.randomlyKillAnimals(0.05) // Убиваем 5% животных
			case 1:
				cs.randomlyCorruptGrass(0.1) // Портим 10% травы
			case 2:
				cs.randomlyDamageAnimals(0.1) // Раним 10% животных
			case 3:
				cs.randomlyModifyHunger(0.1) // Изменяем голод 10% животных
			}
		}

		// Проверяем стабильность каждые 60 тиков (1 секунда)
		if tick%60 == 0 {
			cs.checkSystemStability()
		}
	}
}

func (cs *ChaosScenario) checkSystemStability() {
	// Проверяем что система не сломалась

	// 1. Подсчитываем живых животных
	aliveCount := 0
	for _, animal := range cs.animals {
		if cs.world.IsAlive(animal) {
			aliveCount++
		}
	}

	// 2. Проверяем что есть хотя бы некоторые живые животные
	if aliveCount == 0 && len(cs.animals) > 0 {
		cs.t.Logf("WARNING: All animals died during chaos simulation")
	}

	// 3. Проверяем что системы не паникуют на повреждённых данных
	for _, animal := range cs.animals {
		if cs.world.IsAlive(animal) {
			// Проверяем что компоненты в валидном состоянии
			if hunger, hasHunger := cs.world.GetSatiation(animal); hasHunger {
				if hunger.Value < 0 || hunger.Value > 150 { // Разумные границы
					cs.t.Errorf("Animal %d has invalid hunger: %.2f", animal, hunger.Value)
				}
			}

			if health, hasHealth := cs.world.GetHealth(animal); hasHealth {
				if health.Current < 0 || health.Current > health.Max*2 {
					cs.t.Errorf("Animal %d has invalid health: %d/%d",
						animal, health.Current, health.Max)
				}
			}
		}
	}

	// 4. Проверяем состояние травы
	totalGrass := float32(0)
	for x := 0; x < 15; x++ {
		for y := 0; y < 15; y++ {
			grass := cs.terrain.GetGrassAmount(x, y)
			if grass < 0 {
				cs.t.Errorf("Negative grass amount at (%d,%d): %.2f", x, y, grass)
			}
			totalGrass += grass
		}
	}

	// Должна остаться хотя бы некоторая трава
	if totalGrass == 0 {
		cs.t.Logf("WARNING: No grass left in ecosystem")
	}
}

// Chaos тесты

func TestEcosystemSurvivesLowChaos(t *testing.T) {
	t.Parallel()

	scenario := newChaosScenario(t, 42)
	scenario.addRandomAnimals(10) // 10 случайных животных

	// Низкий уровень хаоса (5% шанс сбоя за тик)
	scenario.simulateWithChaos(600, 0.05) // 10 секунд симуляции

	// Система должна остаться стабильной
	scenario.checkSystemStability()
}

func TestEcosystemSurvivesModerateChaos(t *testing.T) {
	t.Parallel()

	scenario := newChaosScenario(t, 123)
	scenario.addRandomAnimals(15)

	// Умеренный уровень хаоса (15% шанс сбоя за тик)
	scenario.simulateWithChaos(300, 0.15) // 5 секунд симуляции

	scenario.checkSystemStability()
}

func TestFeedingSystemRobustness(t *testing.T) {
	t.Parallel()

	scenario := newChaosScenario(t, 456)
	scenario.addRandomAnimals(8)

	// Специфичные сбои для системы питания
	for i := 0; i < 100; i++ {
		scenario.satiationSystem.Update(scenario.world, 1.0/60.0)
		scenario.grassSearchSystem.Update(scenario.world, 1.0/60.0)
		scenario.grassEatingSystem.Update(scenario.world, 1.0/60.0)

		// Каждые 10 тиков вводим сбой
		if i%10 == 0 {
			if scenario.rand.Float32() < 0.8 { // 80% шанс
				// Случайно повреждаем траву рядом с едящими животными
				for _, animal := range scenario.animals {
					if scenario.world.IsAlive(animal) &&
						scenario.world.HasComponent(animal, core.MaskEatingState) {

						pos, _ := scenario.world.GetPosition(animal)
						tileX := int(pos.X / 32)
						tileY := int(pos.Y / 32)

						// "Портим" траву под едящим животным
						scenario.terrain.SetGrassAmount(tileX, tileY, 0)
					}
				}
			}
		}
	}

	// Система питания должна обработать исчезновение травы gracefully
	scenario.checkSystemStability()
}

func TestMemoryLeakResistance(t *testing.T) {
	t.Parallel()

	scenario := newChaosScenario(t, 789)

	// Симулируем создание и уничтожение множества животных
	for cycle := 0; cycle < 50; cycle++ {
		// Создаём животных
		scenario.addRandomAnimals(5)

		// Несколько тиков симуляции
		for tick := 0; tick < 20; tick++ {
			scenario.satiationSystem.Update(scenario.world, 1.0/60.0)
			scenario.grassSearchSystem.Update(scenario.world, 1.0/60.0)
			scenario.grassEatingSystem.Update(scenario.world, 1.0/60.0)
		}

		// Убиваем всех животных
		scenario.randomlyKillAnimals(1.0) // 100% смертность

		// Очищаем список (имитируем освобождение ресурсов)
		scenario.animals = scenario.animals[:0]
	}

	// Система должна остаться стабильной после множественных создания/уничтожения
	entityCount := scenario.world.GetEntityCount()
	if entityCount > 100 { // Не должно быть слишком много "мёртвых" сущностей
		t.Logf("WARNING: High entity count after cleanup: %d", entityCount)
	}
}

func TestConcurrentAccessSimulation(t *testing.T) {
	t.Parallel()

	scenario := newChaosScenario(t, 101112)
	scenario.addRandomAnimals(6)

	// Симулируем concurrent доступ через быстрые изменения состояния
	for i := 0; i < 200; i++ {
		// Быстрые обновления систем
		scenario.satiationSystem.Update(scenario.world, 1.0/120.0) // Двойная скорость
		scenario.grassSearchSystem.Update(scenario.world, 1.0/120.0)
		scenario.grassEatingSystem.Update(scenario.world, 1.0/120.0)

		// Случайные изменения состояния во время обновлений
		if scenario.rand.Float32() < 0.3 {
			scenario.randomlyModifyHunger(0.5)
		}

		// Проверяем стабильность каждые 50 итераций
		if i%50 == 0 {
			scenario.checkSystemStability()
		}
	}
}
