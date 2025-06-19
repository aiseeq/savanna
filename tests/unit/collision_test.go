package unit

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/physics"
)

func TestNewCircle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		center   physics.Vec2
		radius   float32
		expected physics.Circle
	}{
		{"origin circle", physics.NewVec2(0, 0), 5, physics.Circle{Center: physics.NewVec2(0, 0), Radius: 5}},
		{"offset circle", physics.NewVec2(0, 20), 3.5, physics.Circle{Center: physics.NewVec2(0, 20), Radius: 3.5}},
		{"unit circle", physics.NewVec2(1, 1), 1, physics.Circle{Center: physics.NewVec2(1, 1), Radius: 1}},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.NewCircle(tt.center, tt.radius)
			if result != tt.expected {
				t.Errorf("NewCircle(%v, %f) = %v, expected %v", tt.center, tt.radius, result, tt.expected)
			}
		})
	}
}

func TestNewRectangle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		min, max physics.Vec2
		expected physics.Rectangle
	}{
		{
			"unit rectangle",
			physics.NewVec2(0, 0), physics.NewVec2(1, 1),
			physics.Rectangle{Min: physics.NewVec2(0, 0), Max: physics.NewVec2(1, 1)},
		},
		{
			"offset rectangle",
			physics.NewVec2(5, 10), physics.NewVec2(5, 25),
			physics.Rectangle{Min: physics.NewVec2(5, 10), Max: physics.NewVec2(5, 25)},
		},
		{
			"negative coords",
			physics.NewVec2(5, -5), physics.NewVec2(5, 5),
			physics.Rectangle{Min: physics.NewVec2(5, -5), Max: physics.NewVec2(5, 5)},
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.NewRectangle(tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("NewRectangle(%v, %v) = %v, expected %v", tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestNewRectangleFromCenter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		center        physics.Vec2
		width, height float32
		expectedMin   physics.Vec2
		expectedMax   physics.Vec2
	}{
		{"origin centered", physics.NewVec2(0, 0), 4, 6, physics.NewVec2(-2, -3), physics.NewVec2(2, 3)},
		{"offset centered", physics.NewVec2(10, 20), 8, 10, physics.NewVec2(6, 15), physics.NewVec2(14, 25)},
		{"unit square", physics.NewVec2(5, 5), 2, 2, physics.NewVec2(4, 4), physics.NewVec2(6, 6)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.NewRectangleFromCenter(tt.center, tt.width, tt.height)
			if !vectorsAlmostEqual(result.Min, tt.expectedMin) || !vectorsAlmostEqual(result.Max, tt.expectedMax) {
				t.Errorf("NewRectangleFromCenter(%v, %f, %f) = {%v, %v}, expected {%v, %v}",
					tt.center, tt.width, tt.height, result.Min, result.Max, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestCircleCircleCollision(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		c1, c2   physics.Circle
		expected bool
	}{
		{"no collision", physics.NewCircle(physics.NewVec2(0, 0), 1), physics.NewCircle(physics.NewVec2(3, 0), 1), false},
		{"touching", physics.NewCircle(physics.NewVec2(0, 0), 1), physics.NewCircle(physics.NewVec2(2, 0), 1), true},
		{"overlapping", physics.NewCircle(physics.NewVec2(0, 0), 2), physics.NewCircle(physics.NewVec2(1, 0), 2), true},
		{"one inside other", physics.NewCircle(physics.NewVec2(0, 0), 5), physics.NewCircle(physics.NewVec2(1, 1), 1), true},
		{"same position", physics.NewCircle(physics.NewVec2(5, 5), 2), physics.NewCircle(physics.NewVec2(5, 5), 3), true},
		{
			"diagonal touching",
			physics.NewCircle(physics.NewVec2(0, 0), 1),
			physics.NewCircle(physics.NewVec2(float32(math.Sqrt(2)), float32(math.Sqrt(2))), 1),
			true,
		},
		{
			"diagonal not touching",
			physics.NewCircle(physics.NewVec2(0, 0), 1),
			physics.NewCircle(physics.NewVec2(3, 3), 1),
			false,
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.CircleCircleCollision(tt.c1, tt.c2)
			if result != tt.expected {
				t.Errorf("CircleCircleCollision(%v, %v) = %t, expected %t", tt.c1, tt.c2, result, tt.expected)
			}

			// Проверяем коммутативность
			result2 := physics.CircleCircleCollision(tt.c2, tt.c1)
			if result2 != tt.expected {
				t.Errorf("CircleCircleCollision is not commutative: (%v, %v) = %t, (%v, %v) = %t",
					tt.c1, tt.c2, result, tt.c2, tt.c1, result2)
			}
		})
	}
}

//nolint:gocognit // Комплексный unit тест коллизий круг-круг
func TestCircleCircleCollisionWithDetails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		c1, c2              physics.Circle
		expectedColliding   bool
		expectedPenetration float32
		expectedNormal      physics.Vec2
	}{
		{
			"no collision",
			physics.NewCircle(physics.NewVec2(0, 0), 1),
			physics.NewCircle(physics.NewVec2(3, 0), 1),
			false, 0, physics.NewVec2(0, 0),
		},
		{
			"touching circles",
			physics.NewCircle(physics.NewVec2(0, 0), 1),
			physics.NewCircle(physics.NewVec2(2, 0), 1),
			true, 0, physics.NewVec2(1, 0),
		},
		{
			"overlapping horizontal",
			physics.NewCircle(physics.NewVec2(0, 0), 2),
			physics.NewCircle(physics.NewVec2(2, 0), 2),
			true, 2, physics.NewVec2(1, 0),
		},
		{
			"overlapping vertical",
			physics.NewCircle(physics.NewVec2(0, 0), 1.5),
			physics.NewCircle(physics.NewVec2(0, 2), 1.5),
			true, 1, physics.NewVec2(0, 1),
		},
		{
			"same position different sizes",
			physics.NewCircle(physics.NewVec2(5, 5), 2),
			physics.NewCircle(physics.NewVec2(5, 5), 3),
			true, 5, physics.NewVec2(1, 0), // Произвольная нормаль для совпадающих центров
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.CircleCircleCollisionWithDetails(tt.c1, tt.c2)

			if result.Colliding != tt.expectedColliding {
				t.Errorf("Expected colliding %t, got %t", tt.expectedColliding, result.Colliding)
			}

			if tt.expectedColliding {
				if !almostEqual(result.Penetration, tt.expectedPenetration) {
					t.Errorf("Expected penetration %f, got %f", tt.expectedPenetration, result.Penetration)
				}

				// Для случая совпадающих центров нормаль может быть любой единичной
				if tt.c1.Center == tt.c2.Center {
					if !almostEqual(result.Normal.Length(), 1.0) {
						t.Errorf("Expected unit normal for same center case, got %v with length %f",
							result.Normal, result.Normal.Length())
					}
				} else if !vectorsAlmostEqual(result.Normal, tt.expectedNormal) {
					t.Errorf("Expected normal %v, got %v", tt.expectedNormal, result.Normal)
				}
			}
		})
	}
}

func TestCircleRectangleCollision(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		circle   physics.Circle
		rect     physics.Rectangle
		expected bool
	}{
		{
			"no collision - far away",
			physics.NewCircle(physics.NewVec2(0, 10), 1),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			false,
		},
		{
			"collision - circle inside rectangle",
			physics.NewCircle(physics.NewVec2(1, 1), 0.5),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			true,
		},
		{
			"collision - circle overlaps corner",
			physics.NewCircle(physics.NewVec2(2.5, 2.5), 1),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			true,
		},
		{
			"collision - circle touches edge",
			physics.NewCircle(physics.NewVec2(3, 1), 1),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			true,
		},
		{
			"no collision - circle near corner",
			physics.NewCircle(physics.NewVec2(3, 3), 0.5),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			false,
		},
		{
			"collision - large circle encompasses rectangle",
			physics.NewCircle(physics.NewVec2(1, 1), 5),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			true,
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.CircleRectangleCollision(tt.circle, tt.rect)
			if result != tt.expected {
				t.Errorf("CircleRectangleCollision(%v, %v) = %t, expected %t", tt.circle, tt.rect, result, tt.expected)
			}
		})
	}
}

//nolint:gocognit // Комплексный unit тест коллизий круг-прямоугольник
func TestCircleRectangleCollisionWithDetails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		circle            physics.Circle
		rect              physics.Rectangle
		expectedColliding bool
	}{
		{
			"no collision",
			physics.NewCircle(physics.NewVec2(0, 10), 1),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			false,
		},
		{
			"circle inside rectangle",
			physics.NewCircle(physics.NewVec2(1, 1), 0.3),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			true,
		},
		{
			"circle overlaps edge",
			physics.NewCircle(physics.NewVec2(3, 1), 1.5),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			true,
		},
		{
			"circle overlaps corner",
			physics.NewCircle(physics.NewVec2(2.5, 2.5), 1),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			true,
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.CircleRectangleCollisionWithDetails(tt.circle, tt.rect)

			if result.Colliding != tt.expectedColliding {
				t.Errorf("Expected colliding %t, got %t", tt.expectedColliding, result.Colliding)
			}

			if tt.expectedColliding {
				// Проверяем что нормаль - единичный вектор
				normalLength := result.Normal.Length()
				if !almostEqual(normalLength, 1.0) && !almostEqual(normalLength, 0.0) {
					t.Errorf("Expected unit normal, got %v with length %f", result.Normal, normalLength)
				}

				// Проверяем что проникновение положительное
				if result.Penetration < 0 {
					t.Errorf("Expected positive penetration, got %f", result.Penetration)
				}
			}
		})
	}
}

func TestRectangleRectangleCollision(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		r1, r2   physics.Rectangle
		expected bool
	}{
		{
			"no collision - separated",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			physics.NewRectangle(physics.NewVec2(3, 3), physics.NewVec2(5, 5)),
			false,
		},
		{
			"collision - overlapping",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			physics.NewRectangle(physics.NewVec2(1, 1), physics.NewVec2(3, 3)),
			true,
		},
		{
			"collision - touching edge",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			physics.NewRectangle(physics.NewVec2(2, 0), physics.NewVec2(4, 2)),
			true,
		},
		{
			"collision - one inside other",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(4, 4)),
			physics.NewRectangle(physics.NewVec2(1, 1), physics.NewVec2(3, 3)),
			true,
		},
		{
			"no collision - adjacent vertically",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			physics.NewRectangle(physics.NewVec2(0, 3), physics.NewVec2(2, 5)),
			false,
		},
		{
			"no collision - adjacent horizontally",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			physics.NewRectangle(physics.NewVec2(3, 0), physics.NewVec2(5, 2)),
			false,
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.RectangleRectangleCollision(tt.r1, tt.r2)
			if result != tt.expected {
				t.Errorf("RectangleRectangleCollision(%v, %v) = %t, expected %t", tt.r1, tt.r2, result, tt.expected)
			}

			// Проверяем коммутативность
			result2 := physics.RectangleRectangleCollision(tt.r2, tt.r1)
			if result2 != tt.expected {
				t.Errorf("RectangleRectangleCollision is not commutative")
			}
		})
	}
}

func TestPointInCircle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		point    physics.Vec2
		circle   physics.Circle
		expected bool
	}{
		{"center point", physics.NewVec2(0, 0), physics.NewCircle(physics.NewVec2(0, 0), 5), true},
		{"point inside", physics.NewVec2(1, 1), physics.NewCircle(physics.NewVec2(0, 0), 2), true},
		{"point on circumference", physics.NewVec2(3, 0), physics.NewCircle(physics.NewVec2(0, 0), 3), true},
		{"point outside", physics.NewVec2(5, 0), physics.NewCircle(physics.NewVec2(0, 0), 3), false},
		{"negative coordinates", physics.NewVec2(2, -2), physics.NewCircle(physics.NewVec2(1, -1), 2), true},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.PointInCircle(tt.point, tt.circle)
			if result != tt.expected {
				t.Errorf("PointInCircle(%v, %v) = %t, expected %t", tt.point, tt.circle, result, tt.expected)
			}
		})
	}
}

func TestPointInRectangle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		point    physics.Vec2
		rect     physics.Rectangle
		expected bool
	}{
		{"center point", physics.NewVec2(1, 1), physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)), true},
		{"corner point", physics.NewVec2(0, 0), physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)), true},
		{"edge point", physics.NewVec2(1, 0), physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)), true},
		{"outside point", physics.NewVec2(3, 3), physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)), false},
		{
			"negative coordinates", physics.NewVec2(-1, -1),
			physics.NewRectangle(physics.NewVec2(-2, -2), physics.NewVec2(0, 0)), true,
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.PointInRectangle(tt.point, tt.rect)
			if result != tt.expected {
				t.Errorf("PointInRectangle(%v, %v) = %t, expected %t", tt.point, tt.rect, result, tt.expected)
			}
		})
	}
}

func TestRectangleCenter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		rect     physics.Rectangle
		expected physics.Vec2
	}{
		{"unit square", physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)), physics.NewVec2(1, 1)},
		{"offset rectangle", physics.NewRectangle(physics.NewVec2(5, 10), physics.NewVec2(15, 20)), physics.NewVec2(10, 15)},
		{"negative coordinates", physics.NewRectangle(physics.NewVec2(-4, -6), physics.NewVec2(4, 6)), physics.NewVec2(0, 0)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.rect.Center()
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Center() = %v, expected %v", tt.rect, result, tt.expected)
			}
		})
	}
}

func TestRectangleDimensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		rect           physics.Rectangle
		expectedWidth  float32
		expectedHeight float32
		expectedArea   float32
	}{
		{"unit square", physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)), 2, 2, 4},
		{"rectangle", physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(4, 3)), 4, 3, 12},
		{"offset rectangle", physics.NewRectangle(physics.NewVec2(5, 10), physics.NewVec2(15, 25)), 10, 15, 150},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			width := tt.rect.Width()
			height := tt.rect.Height()
			area := tt.rect.Area()

			if !almostEqual(width, tt.expectedWidth) {
				t.Errorf("Width: expected %f, got %f", tt.expectedWidth, width)
			}
			if !almostEqual(height, tt.expectedHeight) {
				t.Errorf("Height: expected %f, got %f", tt.expectedHeight, height)
			}
			if !almostEqual(area, tt.expectedArea) {
				t.Errorf("Area: expected %f, got %f", tt.expectedArea, area)
			}
		})
	}
}

func TestRectangleIntersect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		r1, r2   physics.Rectangle
		expected physics.Rectangle
	}{
		{
			"overlapping rectangles",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(4, 4)),
			physics.NewRectangle(physics.NewVec2(2, 2), physics.NewVec2(6, 6)),
			physics.NewRectangle(physics.NewVec2(2, 2), physics.NewVec2(4, 4)),
		},
		{
			"one inside other",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(10, 10)),
			physics.NewRectangle(physics.NewVec2(2, 3), physics.NewVec2(7, 8)),
			physics.NewRectangle(physics.NewVec2(2, 3), physics.NewVec2(7, 8)),
		},
		{
			"no intersection",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(2, 2)),
			physics.NewRectangle(physics.NewVec2(3, 3), physics.NewVec2(5, 5)),
			physics.NewRectangle(physics.NewVec2(3, 3), physics.NewVec2(2, 2)), // Инвертированный прямоугольник для несекущихся
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.r1.Intersect(tt.r2)
			if !vectorsAlmostEqual(result.Min, tt.expected.Min) || !vectorsAlmostEqual(result.Max, tt.expected.Max) {
				t.Errorf("%v.Intersect(%v) = %v, expected %v", tt.r1, tt.r2, result, tt.expected)
			}
		})
	}
}

func TestRectangleContains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		r1, r2   physics.Rectangle
		expected bool
	}{
		{
			"contains smaller rectangle",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(10, 10)),
			physics.NewRectangle(physics.NewVec2(2, 3), physics.NewVec2(7, 8)),
			true,
		},
		{
			"same rectangle",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(5, 5)),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(5, 5)),
			true,
		},
		{
			"does not contain",
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(5, 5)),
			physics.NewRectangle(physics.NewVec2(3, 3), physics.NewVec2(8, 8)),
			false,
		},
		{
			"smaller cannot contain larger",
			physics.NewRectangle(physics.NewVec2(2, 2), physics.NewVec2(4, 4)),
			physics.NewRectangle(physics.NewVec2(0, 0), physics.NewVec2(6, 6)),
			false,
		},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.r1.Contains(tt.r2)
			if result != tt.expected {
				t.Errorf("%v.Contains(%v) = %t, expected %t", tt.r1, tt.r2, result, tt.expected)
			}
		})
	}
}
