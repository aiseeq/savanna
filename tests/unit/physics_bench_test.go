package unit

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/physics"
)

// Бенчмарки для векторных операций

func BenchmarkVec2Add(b *testing.B) {
	v1 := physics.NewVec2(3.5, 4.2)
	v2 := physics.NewVec2(1.8, 2.7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v1.Add(v2)
	}
}

func BenchmarkVec2Sub(b *testing.B) {
	v1 := physics.NewVec2(3.5, 4.2)
	v2 := physics.NewVec2(1.8, 2.7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v1.Sub(v2)
	}
}

func BenchmarkVec2Mul(b *testing.B) {
	v := physics.NewVec2(3.5, 4.2)
	scalar := float32(2.3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Mul(scalar)
	}
}

func BenchmarkVec2Dot(b *testing.B) {
	v1 := physics.NewVec2(3.5, 4.2)
	v2 := physics.NewVec2(1.8, 2.7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v1.Dot(v2)
	}
}

func BenchmarkVec2Length(b *testing.B) {
	v := physics.NewVec2(3.5, 4.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Length()
	}
}

func BenchmarkVec2LengthSquared(b *testing.B) {
	v := physics.NewVec2(3.5, 4.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.LengthSquared()
	}
}

func BenchmarkVec2Distance(b *testing.B) {
	v1 := physics.NewVec2(3.5, 4.2)
	v2 := physics.NewVec2(1.8, 2.7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v1.Distance(v2)
	}
}

func BenchmarkVec2DistanceSquared(b *testing.B) {
	v1 := physics.NewVec2(3.5, 4.2)
	v2 := physics.NewVec2(1.8, 2.7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v1.DistanceSquared(v2)
	}
}

func BenchmarkVec2Normalize(b *testing.B) {
	v := physics.NewVec2(3.5, 4.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Normalize()
	}
}

func BenchmarkVec2Rotate(b *testing.B) {
	v := physics.NewVec2(3.5, 4.2)
	angle := float32(math.Pi / 4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Rotate(angle)
	}
}

// Бенчмарки для системы коллизий

func BenchmarkCircleCircleCollision(b *testing.B) {
	c1 := physics.NewCircle(physics.NewVec2(0, 0), 2)
	c2 := physics.NewCircle(physics.NewVec2(3, 0), 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.CircleCircleCollision(c1, c2)
	}
}

func BenchmarkCircleCircleCollisionWithDetails(b *testing.B) {
	c1 := physics.NewCircle(physics.NewVec2(0, 0), 2)
	c2 := physics.NewCircle(physics.NewVec2(3, 0), 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.CircleCircleCollisionWithDetails(c1, c2)
	}
}

func BenchmarkCircleRectangleCollision(b *testing.B) {
	circle := physics.NewCircle(physics.NewVec2(5, 2.5), 1)
	rect := physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(4, 4))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.CircleRectangleCollision(circle, rect)
	}
}

func BenchmarkCircleRectangleCollisionWithDetails(b *testing.B) {
	circle := physics.NewCircle(physics.NewVec2(5, 2.5), 1)
	rect := physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(4, 4))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.CircleRectangleCollisionWithDetails(circle, rect)
	}
}

func BenchmarkRectangleRectangleCollision(b *testing.B) {
	r1 := physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(3, 3))
	r2 := physics.NewRectangle(physics.NewVec2(2, 2), physics.NewVec2(5, 5))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.RectangleRectangleCollision(r1, r2)
	}
}

func BenchmarkPointInCircle(b *testing.B) {
	point := physics.NewVec2(5, 1.5)
	circle := physics.NewCircle(physics.NewVec2(0, 0), 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.PointInCircle(point, circle)
	}
}

func BenchmarkPointInRectangle(b *testing.B) {
	point := physics.NewVec2(5, 1.5)
	rect := physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(3, 3))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.PointInRectangle(point, rect)
	}
}

// Бенчмарки для пространственной сетки

func BenchmarkSpatialGridInsert(b *testing.B) {
	grid := physics.NewSpatialGrid(1000, 1000, 50)
	position := physics.NewVec2(0, 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := physics.EntityID(i % 10000) // Переиспользуем ID для избежания бесконечного роста
		grid.Insert(entityID, position, 1.0)
	}
}

func BenchmarkSpatialGridRemove(b *testing.B) {
	grid := physics.NewSpatialGrid(1000, 1000, 50)

	// Предварительно заполняем сетку
	entityCount := 1000
	for i := 0; i < entityCount; i++ {
		entityID := physics.EntityID(i)
		x := float32(i%100) * 10
		y := float32(i/100) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := physics.EntityID(i % entityCount)
		grid.Remove(entityID)
		// Повторно вставляем для поддержания стабильности теста
		x := float32((i%entityCount)%100) * 10
		y := float32((i%entityCount)/100) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 1.0)
	}
}

func BenchmarkSpatialGridUpdate(b *testing.B) {
	grid := physics.NewSpatialGrid(1000, 1000, 50)

	// Предварительно заполняем сетку
	entityCount := 1000
	for i := 0; i < entityCount; i++ {
		entityID := physics.EntityID(i)
		x := float32(i%100) * 10
		y := float32(i/100) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := physics.EntityID(i % entityCount)
		newX := float32((i * 3) % 1000)
		newY := float32((i * 7) % 1000)
		grid.Update(entityID, physics.NewVec2(newX, newY), 1.0)
	}
}

func BenchmarkSpatialGridQueryRange(b *testing.B) {
	grid := physics.NewSpatialGrid(1000, 1000, 50)

	// Предварительно заполняем сетку
	entityCount := 1000
	for i := 0; i < entityCount; i++ {
		entityID := physics.EntityID(i)
		x := float32(i%100) * 10
		y := float32(i/100) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centerX := float32((i*13)%900) + 50
		centerY := float32((i*17)%900) + 50
		_ = grid.QueryRange(centerX-25, centerY-25, centerX+25, centerY+25)
	}
}

func BenchmarkSpatialGridQueryRadius(b *testing.B) {
	grid := physics.NewSpatialGrid(1000, 1000, 50)

	// Предварительно заполняем сетку
	entityCount := 1000
	for i := 0; i < entityCount; i++ {
		entityID := physics.EntityID(i)
		x := float32(i%100) * 10
		y := float32(i/100) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centerX := float32((i*13)%900) + 50
		centerY := float32((i*17)%900) + 50
		center := physics.NewVec2(centerX, centerY)
		_ = grid.QueryRadius(center, 50.0)
	}
}

func BenchmarkSpatialGridQueryNearest(b *testing.B) {
	grid := physics.NewSpatialGrid(1000, 1000, 50)

	// Предварительно заполняем сетку
	entityCount := 1000
	for i := 0; i < entityCount; i++ {
		entityID := physics.EntityID(i)
		x := float32(i%100) * 10
		y := float32(i/100) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centerX := float32((i*13)%900) + 50
		centerY := float32((i*17)%900) + 50
		center := physics.NewVec2(centerX, centerY)
		_, _ = grid.QueryNearest(center, 100.0)
	}
}

// Комплексные бенчмарки

func BenchmarkSpatialGridMixed1000Entities(b *testing.B) {
	grid := physics.NewSpatialGrid(1000, 1000, 50)

	// Предварительно заполняем сетку
	entityCount := 1000
	for i := 0; i < entityCount; i++ {
		entityID := physics.EntityID(i)
		x := float32(i%100) * 10
		y := float32(i/100) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		operation := i % 4
		entityID := physics.EntityID((i + 1000) % entityCount)

		switch operation {
		case 0: // Insert
			x := float32((i * 7) % 1000)
			y := float32((i * 11) % 1000)
			grid.Insert(entityID+1000, physics.NewVec2(x, y), 1.0)
		case 1: // Update
			x := float32((i * 3) % 1000)
			y := float32((i * 5) % 1000)
			grid.Update(entityID, physics.NewVec2(x, y), 1.0)
		case 2: // Query radius
			centerX := float32((i*13)%900) + 50
			centerY := float32((i*17)%900) + 50
			center := physics.NewVec2(centerX, centerY)
			_ = grid.QueryRadius(center, 75.0)
		case 3: // Query nearest
			centerX := float32((i*19)%900) + 50
			centerY := float32((i*23)%900) + 50
			center := physics.NewVec2(centerX, centerY)
			_, _ = grid.QueryNearest(center, 100.0)
		}
	}
}

// Бенчмарк для большой нагрузки
func BenchmarkSpatialGridLargeScale(b *testing.B) {
	grid := physics.NewSpatialGrid(2000, 2000, 100)

	// Заполняем большую сетку
	entityCount := 5000
	for i := 0; i < entityCount; i++ {
		entityID := physics.EntityID(i)
		x := float32(i%200) * 10
		y := float32(i/200) * 10
		grid.Insert(entityID, physics.NewVec2(x, y), 2.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centerX := float32((i*29)%1800) + 100
		centerY := float32((i*31)%1800) + 100
		center := physics.NewVec2(centerX, centerY)
		_ = grid.QueryRadius(center, 150.0)
	}
}

// Специальный бенчмарк для критического пути игрового цикла
func BenchmarkGameLoopCriticalPath(b *testing.B) {
	// Симулирует типичные операции за один кадр игры
	grid := physics.NewSpatialGrid(1600, 1600, 80) // 50x50 тайлов * 32 пикселя = 1600x1600

	// Инициализируем 100 животных
	animalCount := 100
	animals := make([]struct {
		id       physics.EntityID
		position physics.Vec2
		velocity physics.Vec2
		radius   float32
	}, animalCount)

	for i := 0; i < animalCount; i++ {
		animals[i].id = physics.EntityID(i)
		animals[i].position = physics.NewVec2(
			float32((i*17)%1500)+50,
			float32((i*19)%1500)+50,
		)
		animals[i].velocity = physics.NewVec2(
			float32((i%7)-3)*2,
			float32((i%5)-2)*2,
		)
		animals[i].radius = 5.0
		grid.Insert(animals[i].id, animals[i].position, animals[i].radius)
	}

	b.ResetTimer()

	// Симулируем один игровой цикл
	for frame := 0; frame < b.N; frame++ {
		// Обновляем позиции всех животных
		for i := range animals {
			// Движение
			animals[i].position = animals[i].position.Add(animals[i].velocity.Mul(1.0 / 60.0))

			// Ограничиваем границами мира
			if animals[i].position.X < 0 || animals[i].position.X > 1600 {
				animals[i].velocity.X = -animals[i].velocity.X
			}
			if animals[i].position.Y < 0 || animals[i].position.Y > 1600 {
				animals[i].velocity.Y = -animals[i].velocity.Y
			}

			// Обновляем в пространственной сетке
			grid.Update(animals[i].id, animals[i].position, animals[i].radius)

			// Ищем соседей для взаимодействий
			_ = grid.QueryRadius(animals[i].position, 50.0)
		}
	}
}
