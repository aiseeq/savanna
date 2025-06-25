package unit

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
)

// BenchmarkEntityCreation бенчмарк создания сущностей
func BenchmarkEntityCreation(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Ограничиваем количество сущностей, чтобы не превысить MaxEntities
		if i%500 == 0 && i > 0 {
			world.Clear() // Очищаем каждые 500 итераций
		}

		entity := world.CreateEntity()
		if entity == core.InvalidEntity {
			b.Fatal("Failed to create entity")
		}
	}
}

// BenchmarkEntityDestruction бенчмарк уничтожения сущностей
func BenchmarkEntityDestruction(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Предварительно создаем сущности
	entities := make([]core.EntityID, b.N)
	for i := 0; i < b.N; i++ {
		entities[i] = world.CreateEntity()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		world.DestroyEntity(entities[i])
	}
}

// BenchmarkComponentAdd бенчмарк добавления компонентов
func BenchmarkComponentAdd(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем сущности заранее
	entities := make([]core.EntityID, b.N)
	for i := 0; i < b.N; i++ {
		entities[i] = world.CreateEntity()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		entity := entities[i]
		world.AddPosition(entity, core.Position{X: float32(i), Y: float32(i)})
	}
}

// BenchmarkComponentGet бенчмарк получения компонентов
func BenchmarkComponentGet(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем сущности с компонентами
	entities := make([]core.EntityID, 1000)
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		world.AddPosition(entity, core.Position{X: float32(i), Y: float32(i)})
		entities[i] = entity
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		entity := entities[i%1000]
		_, _ = world.GetPosition(entity)
	}
}

// BenchmarkComponentHas бенчмарк проверки наличия компонентов
func BenchmarkComponentHas(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем сущности с компонентами
	entities := make([]core.EntityID, 1000)
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		if i%2 == 0 {
			world.AddPosition(entity, core.Position{X: float32(i), Y: float32(i)})
		}
		entities[i] = entity
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		entity := entities[i%1000]
		_ = world.HasComponent(entity, core.MaskPosition)
	}
}

// BenchmarkForEachWith бенчмарк итерации по сущностям
func BenchmarkForEachWith(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем 1000 сущностей с разными компонентами
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		world.AddPosition(entity, core.Position{X: float32(i), Y: float32(i)})

		if i%2 == 0 {
			world.AddVelocity(entity, core.Velocity{X: 1, Y: 1})
		}

		if i%3 == 0 {
			world.AddHealth(entity, core.Health{Current: 100, Max: 100})
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		count := 0
		world.ForEachWith(core.MaskPosition, func(entity core.EntityID) {
			count++
		})
	}
}

// BenchmarkQuery1000Entities бенчмарк запросов по 1000 сущностям
func BenchmarkQuery1000Entities(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем 1000 сущностей
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		world.AddPosition(entity, core.Position{X: float32(i % 100), Y: float32(i / 100)})
		world.AddVelocity(entity, core.Velocity{X: 1, Y: 0})

		if i%10 == 0 {
			world.AddAnimalType(entity, core.TypeRabbit)
		} else if i%10 == 1 {
			world.AddAnimalType(entity, core.TypeWolf)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = world.QueryEntitiesWith(core.MaskPosition | core.MaskVelocity)
	}
}

// BenchmarkSpatialQuery бенчмарк пространственных запросов
func BenchmarkSpatialQuery(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем 1000 сущностей распределенных по миру
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		x := float32(i%100) * 10
		y := float32(i/100) * 10

		world.AddPosition(entity, core.Position{X: x, Y: y})
		world.AddSize(entity, core.Size{Radius: 5, AttackRange: 0})
		world.AddAnimalType(entity, core.TypeRabbit)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Ищем в радиусе 50 от центра (500, 500)
		_ = world.QueryInRadius(500, 500, 50)
	}
}

// BenchmarkMovementSystem бенчмарк симуляции системы движения
func BenchmarkMovementSystem(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем 1000 движущихся сущностей
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		world.AddPosition(entity, core.Position{X: float32(i % 100), Y: float32(i / 100)})
		world.AddVelocity(entity, core.Velocity{X: float32(i%10 - 5), Y: float32(i%7 - 3)})
	}

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Симулируем систему движения
		world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
			pos, _ := world.GetPosition(entity)
			vel, _ := world.GetVelocity(entity)

			// Обновляем позицию
			pos.X += vel.X * deltaTime
			pos.Y += vel.Y * deltaTime

			// Ограничиваем границами мира
			if pos.X < 0 {
				pos.X = 0
				vel.X = -vel.X
			} else if pos.X > 1000 {
				pos.X = 1000
				vel.X = -vel.X
			}

			if pos.Y < 0 {
				pos.Y = 0
				vel.Y = -vel.Y
			} else if pos.Y > 1000 {
				pos.Y = 1000
				vel.Y = -vel.Y
			}

			world.SetPosition(entity, pos)
			world.SetVelocity(entity, vel)
		})
	}
}

// BenchmarkFullGameLoop бенчмарк полного игрового цикла
func BenchmarkFullGameLoop(b *testing.B) {
	world := core.NewWorld(1000, 1000, 42)
	defer world.Clear()

	// Создаем экосистему: 800 зайцев и 200 волков
	for i := 0; i < 800; i++ {
		entity := world.CreateEntity()
		world.AddPosition(entity, core.Position{
			X: float32(world.GetRNG().Intn(1000)),
			Y: float32(world.GetRNG().Intn(1000)),
		})
		world.AddVelocity(entity, core.Velocity{X: 0, Y: 0})
		world.AddHealth(entity, core.Health{Current: 50, Max: 50})
		world.AddSatiation(entity, core.Satiation{Value: 80})
		world.AddAnimalType(entity, core.TypeRabbit)
		world.AddSize(entity, core.Size{Radius: 5, AttackRange: 0})
		world.AddSpeed(entity, core.Speed{Base: 20, Current: 20})
	}

	for i := 0; i < 200; i++ {
		entity := world.CreateEntity()
		world.AddPosition(entity, core.Position{
			X: float32(world.GetRNG().Intn(1000)),
			Y: float32(world.GetRNG().Intn(1000)),
		})
		world.AddVelocity(entity, core.Velocity{X: 0, Y: 0})
		world.AddHealth(entity, core.Health{Current: 100, Max: 100})
		world.AddSatiation(entity, core.Satiation{Value: 60})
		world.AddAnimalType(entity, core.TypeWolf)
		world.AddSize(entity, core.Size{Radius: 10, AttackRange: 0})
		world.AddSpeed(entity, core.Speed{Base: 30, Current: 30})
	}

	deltaTime := float32(1.0 / 60.0) // 60 FPS

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Обновляем время
		world.Update(deltaTime)

		// Система движения
		world.ForEachWith(core.MaskPosition|core.MaskVelocity, func(entity core.EntityID) {
			pos, _ := world.GetPosition(entity)
			vel, _ := world.GetVelocity(entity)

			pos.X += vel.X * deltaTime
			pos.Y += vel.Y * deltaTime

			world.SetPosition(entity, pos)
		})

		// Система голода
		world.ForEachWith(core.MaskSatiation, func(entity core.EntityID) {
			hunger, _ := world.GetSatiation(entity)
			hunger.Value -= deltaTime * 0.2 // Теряем 0.2% голода в секунду
			if hunger.Value < 0 {
				hunger.Value = 0
			}
			world.SetSatiation(entity, hunger)
		})

		// Подсчет статистики
		_ = world.GetStats()
	}
}

// BenchmarkEntityManager бенчмарки для EntityManager
func BenchmarkEntityManagerCreate1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		em := core.NewEntityManager()

		for j := 0; j < 1000; j++ {
			_ = em.CreateEntity()
		}
	}
}

func BenchmarkEntityManagerCreateDestroy(b *testing.B) {
	em := core.NewEntityManager()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		entity := em.CreateEntity()
		em.DestroyEntity(entity)
	}
}

// BenchmarkComponentMasks бенчмарки для битовых масок
func BenchmarkComponentMaskHas(b *testing.B) {
	mask := core.MaskPosition | core.MaskVelocity | core.MaskHealth

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = mask.HasComponent(core.MaskPosition)
		_ = mask.HasComponent(core.MaskVelocity)
		_ = mask.HasComponent(core.MaskSatiation)
	}
}

func BenchmarkComponentMaskOperations(b *testing.B) {
	mask := core.MaskPosition

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mask = mask.AddComponent(core.MaskVelocity)
		mask = mask.AddComponent(core.MaskHealth)
		mask = mask.RemoveComponent(core.MaskPosition)
		mask = mask.AddComponent(core.MaskPosition)
	}
}
