package unit

import (
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/physics"
)

const epsilon = 0.0001

// Helper функция для сравнения float32 с погрешностью
func almostEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) < epsilon
}

// Helper функция для сравнения векторов с погрешностью
func vectorsAlmostEqual(v1, v2 physics.Vec2) bool {
	return almostEqual(v1.X, v2.X) && almostEqual(v1.Y, v2.Y)
}

func TestNewVec2(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		x, y     float32
		expected physics.Vec2
	}{
		{"zero vector", 0, 0, physics.NewVec2(0, 0)},
		{"positive values", 3, 4, physics.NewVec2(3, 4)},
		{"negative values", -2, -5, physics.NewVec2(-2, -5)},
		{"mixed values", 1.5, -2.5, physics.NewVec2(1.5, -2.5)},
	}

	for _, tt := range tests {
		tt := tt // Копируем переменную цикла для избежания loop closure
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := physics.NewVec2(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("NewVec2(%f, %f) = %v, expected %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestZero(t *testing.T) {
	t.Parallel()
	result := physics.Zero()
	expected := physics.NewVec2(0, 0)
	if result != expected {
		t.Errorf("Zero() = %v, expected %v", result, expected)
	}
}

func TestVec2Add(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		expected physics.Vec2
	}{
		{"zero vectors", physics.NewVec2(0, 0), physics.NewVec2(0, 0), physics.NewVec2(0, 0)},
		{"positive vectors", physics.NewVec2(1, 2), physics.NewVec2(3, 4), physics.NewVec2(4, 6)},
		{"negative vectors", physics.NewVec2(1, -2), physics.NewVec2(3, -4), physics.NewVec2(4, -6)},
		{"mixed vectors", physics.NewVec2(1, -2), physics.NewVec2(3, 4), physics.NewVec2(4, 2)},
		{"one zero", physics.NewVec2(5, 3), physics.NewVec2(0, 0), physics.NewVec2(5, 3)},
	}

	for _, tt := range tests {
		tt := tt // Копируем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.Add(tt.v2)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Add(%v) = %v, expected %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestVec2Sub(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		expected physics.Vec2
	}{
		{"zero vectors", physics.NewVec2(0, 0), physics.NewVec2(0, 0), physics.NewVec2(0, 0)},
		{"positive vectors", physics.NewVec2(5, 7), physics.NewVec2(2, 3), physics.NewVec2(3, 4)},
		{"negative result", physics.NewVec2(1, 2), physics.NewVec2(3, 4), physics.NewVec2(-2, -2)},
		{"mixed vectors", physics.NewVec2(1, 2), physics.NewVec2(3, -4), physics.NewVec2(-2, 6)},
		{"subtract from zero", physics.NewVec2(0, 0), physics.NewVec2(3, 4), physics.NewVec2(-3, -4)},
	}

	for _, tt := range tests {
		tt := tt // Копируем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.Sub(tt.v2)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Sub(%v) = %v, expected %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestVec2Mul(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		scalar   float32
		expected physics.Vec2
	}{
		{"zero vector", physics.NewVec2(0, 0), 5, physics.NewVec2(0, 0)},
		{"multiply by zero", physics.NewVec2(3, 4), 0, physics.NewVec2(0, 0)},
		{"multiply by one", physics.NewVec2(3, 4), 1, physics.NewVec2(3, 4)},
		{"multiply by positive", physics.NewVec2(2, 3), 2.5, physics.NewVec2(5, 7.5)},
		{"multiply by negative", physics.NewVec2(2, -3), -2, physics.NewVec2(-4, 6)},
		{"fractional scalar", physics.NewVec2(6, 8), 0.5, physics.NewVec2(3, 4)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Mul(tt.scalar)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Mul(%f) = %v, expected %v", tt.v, tt.scalar, result, tt.expected)
			}
		})
	}
}

func TestVec2Div(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		scalar   float32
		expected physics.Vec2
	}{
		{"divide by one", physics.NewVec2(3, 4), 1, physics.NewVec2(3, 4)},
		{"divide by positive", physics.NewVec2(6, 8), 2, physics.NewVec2(3, 4)},
		{"divide by negative", physics.NewVec2(6, -8), -2, physics.NewVec2(-3, 4)},
		{"divide by fraction", physics.NewVec2(3, 4), 0.5, physics.NewVec2(6, 8)},
		{"divide by zero", physics.NewVec2(3, 4), 0, physics.NewVec2(3, 4)}, // Должно вернуть исходный вектор
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Div(tt.scalar)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Div(%f) = %v, expected %v", tt.v, tt.scalar, result, tt.expected)
			}
		})
	}
}

func TestVec2Dot(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		expected float32
	}{
		{"zero vectors", physics.NewVec2(0, 0), physics.NewVec2(0, 0), 0},
		{"unit vectors x", physics.NewVec2(1, 0), physics.NewVec2(1, 0), 1},
		{"unit vectors y", physics.NewVec2(0, 1), physics.NewVec2(0, 1), 1},
		{"perpendicular vectors", physics.NewVec2(1, 0), physics.NewVec2(0, 1), 0},
		{"3-4-5 triangle", physics.NewVec2(3, 4), physics.NewVec2(3, 4), 25},
		{"opposite vectors", physics.NewVec2(1, 0), physics.NewVec2(-1, 0), -1},
		{"general case", physics.NewVec2(2, 3), physics.NewVec2(4, 5), 23},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.Dot(tt.v2)
			if !almostEqual(result, tt.expected) {
				t.Errorf("%v.Dot(%v) = %f, expected %f", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestVec2Length(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		expected float32
	}{
		{"zero vector", physics.NewVec2(0, 0), 0},
		{"unit vector x", physics.NewVec2(1, 0), 1},
		{"unit vector y", physics.NewVec2(0, 1), 1},
		{"3-4-5 triangle", physics.NewVec2(3, 4), 5},
		{"negative values", physics.NewVec2(3, -4), 5},
		{"sqrt(2)", physics.NewVec2(1, 1), float32(math.Sqrt(2))},
		{"sqrt(8)", physics.NewVec2(2, 2), float32(math.Sqrt(8))},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Length()
			if !almostEqual(result, tt.expected) {
				t.Errorf("%v.Length() = %f, expected %f", tt.v, result, tt.expected)
			}
		})
	}
}

func TestVec2LengthSquared(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		expected float32
	}{
		{"zero vector", physics.NewVec2(0, 0), 0},
		{"unit vector x", physics.NewVec2(1, 0), 1},
		{"unit vector y", physics.NewVec2(0, 1), 1},
		{"3-4-5 triangle", physics.NewVec2(3, 4), 25},
		{"negative values", physics.NewVec2(3, -4), 25},
		{"2,2 vector", physics.NewVec2(2, 2), 8},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.LengthSquared()
			if !almostEqual(result, tt.expected) {
				t.Errorf("%v.LengthSquared() = %f, expected %f", tt.v, result, tt.expected)
			}
		})
	}
}

func TestVec2Distance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		expected float32
	}{
		{"same points", physics.NewVec2(0, 0), physics.NewVec2(0, 0), 0},
		{"unit distance x", physics.NewVec2(0, 0), physics.NewVec2(1, 0), 1},
		{"unit distance y", physics.NewVec2(0, 0), physics.NewVec2(0, 1), 1},
		{"3-4-5 triangle", physics.NewVec2(0, 0), physics.NewVec2(3, 4), 5},
		{"negative coordinates", physics.NewVec2(-1, -1), physics.NewVec2(2, 3), 5},
		{"diagonal", physics.NewVec2(1, 1), physics.NewVec2(2, 2), float32(math.Sqrt(2))},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.Distance(tt.v2)
			if !almostEqual(result, tt.expected) {
				t.Errorf("%v.Distance(%v) = %f, expected %f", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestVec2DistanceSquared(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		expected float32
	}{
		{"same points", physics.NewVec2(0, 0), physics.NewVec2(0, 0), 0},
		{"unit distance x", physics.NewVec2(0, 0), physics.NewVec2(1, 0), 1},
		{"unit distance y", physics.NewVec2(0, 0), physics.NewVec2(0, 1), 1},
		{"3-4-5 triangle", physics.NewVec2(0, 0), physics.NewVec2(3, 4), 25},
		{"negative coordinates", physics.NewVec2(-1, -1), physics.NewVec2(2, 3), 25},
		{"diagonal", physics.NewVec2(1, 1), physics.NewVec2(2, 2), 2},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.DistanceSquared(tt.v2)
			if !almostEqual(result, tt.expected) {
				t.Errorf("%v.DistanceSquared(%v) = %f, expected %f", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestVec2Normalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		expected physics.Vec2
	}{
		{"zero vector", physics.NewVec2(0, 0), physics.NewVec2(0, 0)},
		{"unit vector x", physics.NewVec2(1, 0), physics.NewVec2(1, 0)},
		{"unit vector y", physics.NewVec2(0, 1), physics.NewVec2(0, 1)},
		{"3-4 vector", physics.NewVec2(3, 4), physics.NewVec2(0.6, 0.8)},
		{"negative vector", physics.NewVec2(3, -4), physics.NewVec2(0.6, -0.8)},
		{"diagonal", physics.NewVec2(1, 1), physics.NewVec2(float32(1.0/math.Sqrt(2)), float32(1.0/math.Sqrt(2)))},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Normalize()
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Normalize() = %v, expected %v", tt.v, result, tt.expected)
			}

			// Проверяем что длина нормализованного вектора равна 1 (кроме нулевого)
			if !tt.v.IsZero() {
				length := result.Length()
				if !almostEqual(length, 1.0) {
					t.Errorf("Normalized vector %v has length %f, expected 1.0", result, length)
				}
			}
		})
	}
}

func TestVec2Rotate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		angle    float32
		expected physics.Vec2
	}{
		{"zero vector", physics.NewVec2(0, 0), float32(math.Pi / 2), physics.NewVec2(0, 0)},
		{"90 degrees", physics.NewVec2(1, 0), float32(math.Pi / 2), physics.NewVec2(0, 1)},
		{"180 degrees", physics.NewVec2(1, 0), float32(math.Pi), physics.NewVec2(-1, 0)},
		{"270 degrees", physics.NewVec2(1, 0), float32(3 * math.Pi / 2), physics.NewVec2(0, -1)},
		{"360 degrees", physics.NewVec2(1, 0), float32(2 * math.Pi), physics.NewVec2(1, 0)},
		{"45 degrees", physics.NewVec2(1, 0), float32(math.Pi / 4), physics.NewVec2(float32(1.0/math.Sqrt(2)), float32(1.0/math.Sqrt(2)))},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Rotate(tt.angle)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Rotate(%f) = %v, expected %v", tt.v, tt.angle, result, tt.expected)
			}
		})
	}
}

func TestVec2Angle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		expected float32
	}{
		{"positive x", physics.NewVec2(1, 0), 0},
		{"positive y", physics.NewVec2(0, 1), float32(math.Pi / 2)},
		{"negative x", physics.NewVec2(-1, 0), float32(math.Pi)},
		{"negative y", physics.NewVec2(0, -1), float32(-math.Pi / 2)},
		{"45 degrees", physics.NewVec2(1, 1), float32(math.Pi / 4)},
		{"135 degrees", physics.NewVec2(-1, 1), float32(3 * math.Pi / 4)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Angle()
			if !almostEqual(result, tt.expected) {
				t.Errorf("%v.Angle() = %f, expected %f", tt.v, result, tt.expected)
			}
		})
	}
}

func TestVec2AngleTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		expected float32
	}{
		{"same point", physics.NewVec2(1, 1), physics.NewVec2(1, 1), 0},
		{"to right", physics.NewVec2(0, 0), physics.NewVec2(1, 0), 0},
		{"to top", physics.NewVec2(0, 0), physics.NewVec2(0, 1), float32(math.Pi / 2)},
		{"to left", physics.NewVec2(0, 0), physics.NewVec2(-1, 0), float32(math.Pi)},
		{"to bottom", physics.NewVec2(0, 0), physics.NewVec2(0, -1), float32(-math.Pi / 2)},
		{"45 degrees", physics.NewVec2(0, 0), physics.NewVec2(1, 1), float32(math.Pi / 4)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.AngleTo(tt.v2)
			if !almostEqual(result, tt.expected) {
				t.Errorf("%v.AngleTo(%v) = %f, expected %f", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestVec2Lerp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		t        float32
		expected physics.Vec2
	}{
		{"t=0", physics.NewVec2(0, 0), physics.NewVec2(4, 6), 0, physics.NewVec2(0, 0)},
		{"t=1", physics.NewVec2(0, 0), physics.NewVec2(4, 6), 1, physics.NewVec2(4, 6)},
		{"t=0.5", physics.NewVec2(0, 0), physics.NewVec2(4, 6), 0.5, physics.NewVec2(2, 3)},
		{"t=0.25", physics.NewVec2(2, 2), physics.NewVec2(6, 10), 0.25, physics.NewVec2(3, 4)},
		{"negative values", physics.NewVec2(-2, -4), physics.NewVec2(2, 4), 0.5, physics.NewVec2(0, 0)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.Lerp(tt.v2, tt.t)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Lerp(%v, %f) = %v, expected %v", tt.v1, tt.v2, tt.t, result, tt.expected)
			}
		})
	}
}

func TestVec2Equal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1, v2   physics.Vec2
		epsilon  float32
		expected bool
	}{
		{"exactly equal", physics.NewVec2(1, 2), physics.NewVec2(1, 2), 0.001, true},
		{"within epsilon", physics.NewVec2(1, 2), physics.NewVec2(1.0005, 2.0005), 0.001, true},
		{"outside epsilon", physics.NewVec2(1, 2), physics.NewVec2(2, 2.002), 0.001, false},
		{"zero vectors", physics.NewVec2(0, 0), physics.NewVec2(0, 0), 0.001, true},
		{"negative difference", physics.NewVec2(1, 2), physics.NewVec2(0.9995, 1.9995), 0.001, true},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v1.Equal(tt.v2, tt.epsilon)
			if result != tt.expected {
				t.Errorf("%v.Equal(%v, %f) = %t, expected %t", tt.v1, tt.v2, tt.epsilon, result, tt.expected)
			}
		})
	}
}

func TestVec2IsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		expected bool
	}{
		{"zero vector", physics.NewVec2(0, 0), true},
		{"non-zero x", physics.NewVec2(1, 0), false},
		{"non-zero y", physics.NewVec2(0, 1), false},
		{"non-zero both", physics.NewVec2(1, 1), false},
		{"negative zero", physics.NewVec2(0, -0), true},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.IsZero()
			if result != tt.expected {
				t.Errorf("%v.IsZero() = %t, expected %t", tt.v, result, tt.expected)
			}
		})
	}
}

func TestVec2Clamp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		v         physics.Vec2
		maxLength float32
		expected  physics.Vec2
	}{
		{"zero vector", physics.NewVec2(0, 0), 5, physics.NewVec2(0, 0)},
		{"within limit", physics.NewVec2(3, 4), 10, physics.NewVec2(3, 4)},
		{"exactly at limit", physics.NewVec2(3, 4), 5, physics.NewVec2(3, 4)},
		{"over limit", physics.NewVec2(6, 8), 5, physics.NewVec2(3, 4)},
		{"unit vector over limit", physics.NewVec2(1, 0), 0.5, physics.NewVec2(0.5, 0)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Clamp(tt.maxLength)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Clamp(%f) = %v, expected %v", tt.v, tt.maxLength, result, tt.expected)
			}

			// Проверяем что длина не превышает максимум
			length := result.Length()
			if length > tt.maxLength+epsilon {
				t.Errorf("Clamped vector %v has length %f, which exceeds maxLength %f", result, length, tt.maxLength)
			}
		})
	}
}

func TestVec2Reflect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v        physics.Vec2
		normal   physics.Vec2
		expected physics.Vec2
	}{
		{"horizontal surface", physics.NewVec2(1, -1), physics.NewVec2(0, 1), physics.NewVec2(1, 1)},
		{"vertical surface", physics.NewVec2(1, 1), physics.NewVec2(1, 0), physics.NewVec2(-1, 1)},
		{"45 degree surface", physics.NewVec2(1, 0), physics.NewVec2(float32(1.0/math.Sqrt(2)), float32(1.0/math.Sqrt(2))), physics.NewVec2(0, -1)},
		{"perpendicular hit", physics.NewVec2(0, -1), physics.NewVec2(0, 1), physics.NewVec2(0, 1)},
		{"parallel to surface", physics.NewVec2(1, 0), physics.NewVec2(0, 1), physics.NewVec2(1, 0)},
	}

	for _, tt := range tests {
		tt := tt // Захватываем переменную цикла
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.v.Reflect(tt.normal)
			if !vectorsAlmostEqual(result, tt.expected) {
				t.Errorf("%v.Reflect(%v) = %v, expected %v", tt.v, tt.normal, result, tt.expected)
			}
		})
	}
}
