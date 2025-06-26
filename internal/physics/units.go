package physics

import "fmt"

// Типобезопасные единицы измерения для предотвращения ошибок конвертации

// Pixels представляет расстояние в пикселях (экранные координаты)
type Pixels float32

// Tiles представляет расстояние в тайлах (логические игровые единицы)
type Tiles float32

// TilesPerSecond представляет скорость в тайлах за секунду
type TilesPerSecond float32

// Константы конвертации
const (
	PixelsPerTile = 32.0 // 1 тайл = 32 пикселя в изометрии
)

// Конвертация Tiles -> Pixels
func (t Tiles) ToPixels() Pixels {
	return Pixels(float32(t) * PixelsPerTile)
}

// Конвертация Pixels -> Tiles
func (p Pixels) ToTiles() Tiles {
	return Tiles(float32(p) / PixelsPerTile)
}

// Конвертация TilesPerSecond -> скорость в пикселях за секунду
func (tps TilesPerSecond) ToPixelsPerSecond() float32 {
	return float32(tps) * PixelsPerTile
}

// Получение сырого float32 значения (для совместимости с существующим кодом)
func (p Pixels) Float32() float32           { return float32(p) }
func (t Tiles) Float32() float32            { return float32(t) }
func (tps TilesPerSecond) Float32() float32 { return float32(tps) }

// Создание из float32 (для миграции)
func NewPixels(val float32) Pixels                 { return Pixels(val) }
func NewTiles(val float32) Tiles                   { return Tiles(val) }
func NewTilesPerSecond(val float32) TilesPerSecond { return TilesPerSecond(val) }

// Строковые представления для отладки
func (p Pixels) String() string           { return fmt.Sprintf("%.1fpx", float32(p)) }
func (t Tiles) String() string            { return fmt.Sprintf("%.2ft", float32(t)) }
func (tps TilesPerSecond) String() string { return fmt.Sprintf("%.2ft/s", float32(tps)) }

// Математические операции для Tiles
func (t Tiles) Add(other Tiles) Tiles    { return Tiles(float32(t) + float32(other)) }
func (t Tiles) Sub(other Tiles) Tiles    { return Tiles(float32(t) - float32(other)) }
func (t Tiles) Mul(factor float32) Tiles { return Tiles(float32(t) * factor) }
func (t Tiles) Div(factor float32) Tiles { return Tiles(float32(t) / factor) }

// Математические операции для Pixels
func (p Pixels) Add(other Pixels) Pixels   { return Pixels(float32(p) + float32(other)) }
func (p Pixels) Sub(other Pixels) Pixels   { return Pixels(float32(p) - float32(other)) }
func (p Pixels) Mul(factor float32) Pixels { return Pixels(float32(p) * factor) }
func (p Pixels) Div(factor float32) Pixels { return Pixels(float32(p) / factor) }

// Сравнения для Tiles
func (t Tiles) Equals(other Tiles) bool         { return float32(t) == float32(other) }
func (t Tiles) LessThan(other Tiles) bool       { return float32(t) < float32(other) }
func (t Tiles) LessOrEqual(other Tiles) bool    { return float32(t) <= float32(other) }
func (t Tiles) GreaterThan(other Tiles) bool    { return float32(t) > float32(other) }
func (t Tiles) GreaterOrEqual(other Tiles) bool { return float32(t) >= float32(other) }

// Сравнения для Pixels
func (p Pixels) Equals(other Pixels) bool         { return float32(p) == float32(other) }
func (p Pixels) LessThan(other Pixels) bool       { return float32(p) < float32(other) }
func (p Pixels) LessOrEqual(other Pixels) bool    { return float32(p) <= float32(other) }
func (p Pixels) GreaterThan(other Pixels) bool    { return float32(p) > float32(other) }
func (p Pixels) GreaterOrEqual(other Pixels) bool { return float32(p) >= float32(other) }

// Сравнения для TilesPerSecond
func (tps TilesPerSecond) Equals(other TilesPerSecond) bool   { return float32(tps) == float32(other) }
func (tps TilesPerSecond) LessThan(other TilesPerSecond) bool { return float32(tps) < float32(other) }
func (tps TilesPerSecond) LessOrEqual(other TilesPerSecond) bool {
	return float32(tps) <= float32(other)
}
func (tps TilesPerSecond) GreaterThan(other TilesPerSecond) bool {
	return float32(tps) > float32(other)
}
func (tps TilesPerSecond) GreaterOrEqual(other TilesPerSecond) bool {
	return float32(tps) >= float32(other)
}

// Позиции с типобезопасностью
type PixelPosition struct {
	X, Y Pixels
}

type TilePosition struct {
	X, Y Tiles
}

// Конвертация позиций
func (pp PixelPosition) ToTiles() TilePosition {
	return TilePosition{
		X: pp.X.ToTiles(),
		Y: pp.Y.ToTiles(),
	}
}

func (tp TilePosition) ToPixels() PixelPosition {
	return PixelPosition{
		X: tp.X.ToPixels(),
		Y: tp.Y.ToPixels(),
	}
}

// Вектор скорости в тайлах
type TileVelocity struct {
	X, Y TilesPerSecond
}

// Конвертация в пиксели за секунду для применения к позициям
func (tv TileVelocity) ToPixelsPerSecond() (float32, float32) {
	return tv.X.ToPixelsPerSecond(), tv.Y.ToPixelsPerSecond()
}
