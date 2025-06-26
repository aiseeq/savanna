package vec2

import "math"

// Vec2 представляет 2D вектор или точку
type Vec2 struct {
	X, Y float32
}

// New создает новый вектор
func New(x, y float32) Vec2 {
	return Vec2{X: x, Y: y}
}

// Zero возвращает нулевой вектор
func Zero() Vec2 {
	return Vec2{X: 0, Y: 0}
}

// Add складывает два вектора
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

// Sub вычитает векторы
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

// Scale умножает вектор на скаляр
func (v Vec2) Scale(s float32) Vec2 {
	return Vec2{
		X: v.X * s,
		Y: v.Y * s,
	}
}

// Length возвращает длину вектора
func (v Vec2) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

// LengthSquared возвращает квадрат длины (быстрее для сравнений)
func (v Vec2) LengthSquared() float32 {
	return v.X*v.X + v.Y*v.Y
}

// Normalize возвращает нормализованный вектор (длина = 1)
func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if length == 0 {
		return Vec2{X: 0, Y: 0}
	}
	return Vec2{
		X: v.X / length,
		Y: v.Y / length,
	}
}

// Distance возвращает расстояние между двумя точками
func (v Vec2) Distance(other Vec2) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

// DistanceSquared возвращает квадрат расстояния (быстрее для сравнений)
func (v Vec2) DistanceSquared(other Vec2) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return dx*dx + dy*dy
}

// Dot возвращает скалярное произведение векторов
func (v Vec2) Dot(other Vec2) float32 {
	return v.X*other.X + v.Y*other.Y
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

// Lerp линейная интерполяция между векторами
func (v Vec2) Lerp(other Vec2, t float32) Vec2 {
	return Vec2{
		X: v.X + (other.X-v.X)*t,
		Y: v.Y + (other.Y-v.Y)*t,
	}
}

// Clamp ограничивает вектор по длине
func (v Vec2) Clamp(maxLength float32) Vec2 {
	length := v.Length()
	if length <= maxLength {
		return v
	}
	return v.Scale(maxLength / length)
}

// Equal проверяет равенство векторов с учетом погрешности
func (v Vec2) Equal(other Vec2, epsilon float32) bool {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return dx*dx+dy*dy <= epsilon*epsilon
}
