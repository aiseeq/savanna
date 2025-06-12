package physics

import (
	"math"
)

// Circle представляет круг для коллизий
type Circle struct {
	Center Vec2
	Radius float32
}

// NewCircle создает новый круг
func NewCircle(center Vec2, radius float32) Circle {
	return Circle{Center: center, Radius: radius}
}

// Rectangle представляет прямоугольник для коллизий
type Rectangle struct {
	Min, Max Vec2 // Углы прямоугольника (левый нижний и правый верхний)
}

// NewRectangle создает новый прямоугольник
func NewRectangle(min, max Vec2) Rectangle {
	return Rectangle{Min: min, Max: max}
}

// NewRectangleFromCenter создает прямоугольник из центра и размеров
func NewRectangleFromCenter(center Vec2, width, height float32) Rectangle {
	half := Vec2{X: width / 2, Y: height / 2}
	return Rectangle{
		Min: center.Sub(half),
		Max: center.Add(half),
	}
}

// CircleCircleCollision проверяет коллизию между двумя кругами
func CircleCircleCollision(c1, c2 Circle) bool {
	distance := c1.Center.Distance(c2.Center)
	return distance <= (c1.Radius + c2.Radius)
}

// CircleCircleCollisionDetails возвращает детали коллизии между кругами
type CollisionDetails struct {
	Colliding    bool
	Penetration  float32 // Глубина проникновения
	Normal       Vec2    // Нормаль коллизии (направлена от объекта 1 к объекту 2)
	ContactPoint Vec2    // Точка контакта
}

// CircleCircleCollisionWithDetails проверяет коллизию кругов с деталями
func CircleCircleCollisionWithDetails(c1, c2 Circle) CollisionDetails {
	centerDistance := c2.Center.Sub(c1.Center) // Направление от c1 к c2
	distance := centerDistance.Length()
	radiusSum := c1.Radius + c2.Radius

	if distance > radiusSum {
		return CollisionDetails{Colliding: false}
	}

	// Обработка случая когда круги в одной точке
	if distance == 0 {
		return CollisionDetails{
			Colliding:    true,
			Penetration:  radiusSum,
			Normal:       Vec2{X: 1, Y: 0}, // Произвольное направление
			ContactPoint: c1.Center,
		}
	}

	normal := centerDistance.Normalize()
	penetration := radiusSum - distance
	contactPoint := c1.Center.Add(normal.Mul(c1.Radius - penetration/2))

	return CollisionDetails{
		Colliding:    true,
		Penetration:  penetration,
		Normal:       normal,
		ContactPoint: contactPoint,
	}
}

// CircleRectangleCollision проверяет коллизию между кругом и прямоугольником
func CircleRectangleCollision(circle Circle, rect Rectangle) bool {
	// Находим ближайшую точку на прямоугольнике к центру круга
	closestPoint := Vec2{
		X: float32(math.Max(float64(rect.Min.X), math.Min(float64(circle.Center.X), float64(rect.Max.X)))),
		Y: float32(math.Max(float64(rect.Min.Y), math.Min(float64(circle.Center.Y), float64(rect.Max.Y)))),
	}

	// Проверяем расстояние от центра круга до ближайшей точки
	distance := circle.Center.Distance(closestPoint)
	return distance <= circle.Radius
}

// CircleRectangleCollisionWithDetails проверяет коллизию круга и прямоугольника с деталями
func CircleRectangleCollisionWithDetails(circle Circle, rect Rectangle) CollisionDetails {
	// Находим ближайшую точку на прямоугольнике к центру круга
	closestPoint := Vec2{
		X: float32(math.Max(float64(rect.Min.X), math.Min(float64(circle.Center.X), float64(rect.Max.X)))),
		Y: float32(math.Max(float64(rect.Min.Y), math.Min(float64(circle.Center.Y), float64(rect.Max.Y)))),
	}

	centerToClosest := circle.Center.Sub(closestPoint)
	distance := centerToClosest.Length()

	if distance > circle.Radius {
		return CollisionDetails{Colliding: false}
	}

	// Если круг полностью внутри прямоугольника
	if circle.Center.X >= rect.Min.X && circle.Center.X <= rect.Max.X &&
		circle.Center.Y >= rect.Min.Y && circle.Center.Y <= rect.Max.Y {

		// Находим ближайшую сторону для выталкивания
		distToLeft := circle.Center.X - rect.Min.X
		distToRight := rect.Max.X - circle.Center.X
		distToBottom := circle.Center.Y - rect.Min.Y
		distToTop := rect.Max.Y - circle.Center.Y

		minDist := float32(math.Min(math.Min(float64(distToLeft), float64(distToRight)),
			math.Min(float64(distToBottom), float64(distToTop))))

		var normal Vec2
		var contactPoint Vec2
		penetration := circle.Radius + minDist

		switch minDist {
		case distToLeft:
			normal = Vec2{X: -1, Y: 0}
			contactPoint = Vec2{X: rect.Min.X, Y: circle.Center.Y}
		case distToRight:
			normal = Vec2{X: 1, Y: 0}
			contactPoint = Vec2{X: rect.Max.X, Y: circle.Center.Y}
		case distToBottom:
			normal = Vec2{X: 0, Y: -1}
			contactPoint = Vec2{X: circle.Center.X, Y: rect.Min.Y}
		default: // distToTop
			normal = Vec2{X: 0, Y: 1}
			contactPoint = Vec2{X: circle.Center.X, Y: rect.Max.Y}
		}

		return CollisionDetails{
			Colliding:    true,
			Penetration:  penetration,
			Normal:       normal,
			ContactPoint: contactPoint,
		}
	}

	// Обычная коллизия (круг касается границы)
	if distance == 0 {
		return CollisionDetails{
			Colliding:    true,
			Penetration:  circle.Radius,
			Normal:       Vec2{X: 0, Y: 1}, // Произвольное направление
			ContactPoint: closestPoint,
		}
	}

	normal := centerToClosest.Normalize()
	penetration := circle.Radius - distance

	return CollisionDetails{
		Colliding:    true,
		Penetration:  penetration,
		Normal:       normal,
		ContactPoint: closestPoint,
	}
}

// RectangleRectangleCollision проверяет коллизию между двумя прямоугольниками
func RectangleRectangleCollision(r1, r2 Rectangle) bool {
	return !(r1.Max.X < r2.Min.X || r1.Min.X > r2.Max.X ||
		r1.Max.Y < r2.Min.Y || r1.Min.Y > r2.Max.Y)
}

// PointInCircle проверяет находится ли точка внутри круга
func PointInCircle(point Vec2, circle Circle) bool {
	return point.Distance(circle.Center) <= circle.Radius
}

// PointInRectangle проверяет находится ли точка внутри прямоугольника
func PointInRectangle(point Vec2, rect Rectangle) bool {
	return point.X >= rect.Min.X && point.X <= rect.Max.X &&
		point.Y >= rect.Min.Y && point.Y <= rect.Max.Y
}

// Center возвращает центр прямоугольника
func (r Rectangle) Center() Vec2 {
	return Vec2{
		X: (r.Min.X + r.Max.X) / 2,
		Y: (r.Min.Y + r.Max.Y) / 2,
	}
}

// Width возвращает ширину прямоугольника
func (r Rectangle) Width() float32 {
	return r.Max.X - r.Min.X
}

// Height возвращает высоту прямоугольника
func (r Rectangle) Height() float32 {
	return r.Max.Y - r.Min.Y
}

// Area возвращает площадь прямоугольника
func (r Rectangle) Area() float32 {
	return r.Width() * r.Height()
}

// Intersect возвращает пересечение двух прямоугольников
func (r Rectangle) Intersect(other Rectangle) Rectangle {
	return Rectangle{
		Min: Vec2{
			X: float32(math.Max(float64(r.Min.X), float64(other.Min.X))),
			Y: float32(math.Max(float64(r.Min.Y), float64(other.Min.Y))),
		},
		Max: Vec2{
			X: float32(math.Min(float64(r.Max.X), float64(other.Max.X))),
			Y: float32(math.Min(float64(r.Max.Y), float64(other.Max.Y))),
		},
	}
}

// Contains проверяет содержит ли прямоугольник другой прямоугольник
func (r Rectangle) Contains(other Rectangle) bool {
	return r.Min.X <= other.Min.X && r.Min.Y <= other.Min.Y &&
		r.Max.X >= other.Max.X && r.Max.Y >= other.Max.Y
}
