package integration

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// SimpleMovementSystem движение без отражений от границ для чистого тестирования
type SimpleMovementSystem struct{}

func (sms *SimpleMovementSystem) Update(world *core.World, deltaTime float32) {
	// Только обновляем позиции по скорости, без отражений
	world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
		pos, _ := world.GetPosition(entity)
		vel, _ := world.GetVelocity(entity)

		pos.X += vel.X * deltaTime
		pos.Y += vel.Y * deltaTime

		world.SetPosition(entity, pos)
	})
}

// TestWolfWithoutBoundaryInterference тест без влияния границ мира
func TestWolfWithoutBoundaryInterference(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		World: config.WorldConfig{Size: 50, Seed: 123}, // Больший мир
	}

	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	world := core.NewWorld(1600, 1600, 123) // Очень большой мир
	vegetationSystem := simulation.NewVegetationSystem(terrain)
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(vegetationSystem)
	simpleMovementSystem := &SimpleMovementSystem{} // Без отражений

	// Создаем животных в центре большого мира
	rabbit := simulation.CreateRabbit(world, 800, 800)
	wolf := simulation.CreateWolf(world, 780, 800) // 20 единиц от зайца

	// Делаем волка голодным
	world.SetHunger(wolf, core.Hunger{Value: 20.0})

	t.Logf("=== Тест без влияния границ ===")

	deltaTime := float32(1.0 / 60.0)

	for i := 0; i < 120; i++ { // 2 секунды
		// Фиксируем зайца на месте перед обновлением
		world.SetPosition(rabbit, core.Position{X: 800, Y: 800})
		world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})
		world.SetSpeed(rabbit, core.Speed{Base: 0, Current: 0})

		// Обновляем системы
		animalBehaviorSystem.Update(world, deltaTime)
		simpleMovementSystem.Update(world, deltaTime)
		world.Update(deltaTime)

		// Получаем позиции после обновления
		wolfPos, _ := world.GetPosition(wolf)
		rabbitPos, _ := world.GetPosition(rabbit)
		wolfVel, _ := world.GetVelocity(wolf)

		distance := math.Sqrt(float64((wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)))

		if i%20 == 0 {
			t.Logf("%.1fс: волк (%.1f,%.1f) vel(%.1f,%.1f) | дистанция %.1f",
				float32(i)/60.0, wolfPos.X, wolfPos.Y, wolfVel.X, wolfVel.Y, distance)
		}

		// Проверяем перепрыгивание
		if wolfPos.X > 810 && i > 30 { // Если волк далеко за зайцем
			t.Errorf("ПЕРЕПРЫГИВАНИЕ: Волк (%.1f,%.1f) ушел далеко за неподвижного зайца (800,800) на тике %d",
				wolfPos.X, wolfPos.Y, i)
			t.Errorf("  Текущая скорость волка: (%.1f, %.1f)", wolfVel.X, wolfVel.Y)
			break
		}

		// Если волк остановился рядом с зайцем - успех
		if distance < 2.0 && math.Abs(float64(wolfVel.X)) < 1.0 && math.Abs(float64(wolfVel.Y)) < 1.0 {
			t.Logf("УСПЕХ: Волк остановился рядом с зайцем на дистанции %.1f", distance)
			return
		}
	}
}
