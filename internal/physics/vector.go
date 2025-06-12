package physics

import (
	"math"
)

// Vec2 представляет двумерный вектор
type Vec2 struct {
	X, Y float32
}

// NewVec2 создает новый вектор
func NewVec2(x, y float32) Vec2 {
	return Vec2{X: x, Y: y}
}

// Zero возвращает нулевой вектор
func Zero() Vec2 {
	return Vec2{X: 0, Y: 0}
}

// Add складывает два вектора
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub вычитает один вектор из другого
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Mul умножает вектор на скаляр
func (v Vec2) Mul(scalar float32) Vec2 {
	return Vec2{X: v.X * scalar, Y: v.Y * scalar}
}

// Div делит вектор на скаляр
func (v Vec2) Div(scalar float32) Vec2 {
	if scalar == 0 {
		return v
	}
	return Vec2{X: v.X / scalar, Y: v.Y / scalar}
}

// Dot вычисляет скалярное произведение
func (v Vec2) Dot(other Vec2) float32 {
	return v.X*other.X + v.Y*other.Y
}

// Length возвращает длину вектора
func (v Vec2) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

// LengthSquared возвращает квадрат длины вектора (быстрее для сравнений)
func (v Vec2) LengthSquared() float32 {
	return v.X*v.X + v.Y*v.Y
}

// Distance возвращает расстояние между двумя точками
func (v Vec2) Distance(other Vec2) float32 {
	return v.Sub(other).Length()
}

// DistanceSquared возвращает квадрат расстояния между двумя точками
func (v Vec2) DistanceSquared(other Vec2) float32 {
	return v.Sub(other).LengthSquared()
}

// Normalize нормализует вектор (приводит к единичной длине)
func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if length == 0 {
		return Vec2{X: 0, Y: 0}
	}
	return Vec2{X: v.X / length, Y: v.Y / length}
}

// Rotate поворачивает вектор на угол в радианах
func (v Vec2) Rotate(angle float32) Vec2 {
	cos := float32(math.Cos(float64(angle)))
	sin := float32(math.Sin(float64(angle)))
	return Vec2{
		X: v.X*cos - v.Y*sin,
		Y: v.X*sin + v.Y*cos,
	}
}

// Angle возвращает угол вектора в радианах
func (v Vec2) Angle() float32 {
	return float32(math.Atan2(float64(v.Y), float64(v.X)))
}

// AngleTo возвращает угол между двумя векторами
func (v Vec2) AngleTo(other Vec2) float32 {
	return float32(math.Atan2(float64(other.Y-v.Y), float64(other.X-v.X)))
}

// Lerp выполняет линейную интерполяцию между векторами
func (v Vec2) Lerp(other Vec2, t float32) Vec2 {
	return Vec2{
		X: v.X + (other.X-v.X)*t,
		Y: v.Y + (other.Y-v.Y)*t,
	}
}

// Equal проверяет равенство векторов с учетом погрешности
func (v Vec2) Equal(other Vec2, epsilon float32) bool {
	return float32(math.Abs(float64(v.X-other.X))) < epsilon &&
		float32(math.Abs(float64(v.Y-other.Y))) < epsilon
}

// IsZero проверяет является ли вектор нулевым
func (v Vec2) IsZero() bool {
	return v.X == 0 && v.Y == 0
}

// Clamp ограничивает длину вектора максимальным значением
func (v Vec2) Clamp(maxLength float32) Vec2 {
	if v.LengthSquared() <= maxLength*maxLength {
		return v
	}
	return v.Normalize().Mul(maxLength)
}

// Reflect отражает вектор относительно нормали
func (v Vec2) Reflect(normal Vec2) Vec2 {
	return v.Sub(normal.Mul(2 * v.Dot(normal)))
}
