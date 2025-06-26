package constants

// Общие константы для всего проекта (устраняет gomnd нарушения)
// Все магические числа заменены на именованные константы с объяснениями

const (
	// Битовые операции
	BitsPerUint64 = 64 // Количество бит в uint64 для битовых масок

	// Мировые константы
	DefaultWorldSizePixels = 1600.0 // Стандартный размер мира в пикселях (50 тайлов * 32 пикселя)
	TileSizePixels         = 32     // Размер одного тайла в пикселях

	// Константы анимации
	AnimationFrameZero  = 0     // Первый кадр анимации (windup)
	AnimationFrameOne   = 1     // Второй кадр анимации (strike/action)
	LargeAnimationTimer = 999.0 // Большой таймер для остановки анимации

	// Константы движения
	MovementThreshold          = 0.1  // Минимальная скорость для определения движения
	CriticalCollisionThreshold = 10.0 // Порог критической коллизии для логирования

	// Константы питания
	MaxNutritionalValue = 100.0 // Максимальная питательная ценность
	FullGrassAmount     = 100.0 // Полное количество травы на тайле

	// Константы поиска
	MaxSearchDistance = 999999.0 // Максимальная дистанция поиска (для инициализации)

	// Константы состояний
	NoTarget               = 0   // Отсутствие цели (поедание травы)
	InitialProgress        = 0.0 // Начальный прогресс поедания
	InitialNutrition       = 0.0 // Начальная полученная питательность
	SatietyTolerance       = 0.1 // Допуск для проверки сытости (99.9%)
	NutritionToHungerRatio = 1.0 // Коэффициент конвертации питательности в голод

	// Размеры спрайтов
	DefaultSpriteSize = 32 // Размер спрайта по умолчанию

	// Константы позиций (для начальной камеры и UI)
	DefaultCameraX = 400 // Начальная позиция камеры X
	DefaultCameraY = 300 // Начальная позиция камеры Y

	// Константы времени и FPS
	StandardFPS       = 60.0              // Стандартный FPS игры
	StandardDeltaTime = 1.0 / StandardFPS // Стандартный deltaTime

	// Масштабирование камеры
	CameraMoveSpeed  = 20.0 // Скорость движения камеры
	CameraZoomFactor = 1.1  // Коэффициент увеличения/уменьшения масштаба
	MinCameraZoom    = 0.2  // Минимальный зум камеры
	MaxCameraZoom    = 5.0  // Максимальный зум камеры

	// Масштабы времени
	TimeScaleMinimum = 0.1  // Минимальное замедление времени
	TimeScaleMaximum = 10.0 // Максимальное ускорение времени
	TimeScaleHalf    = 0.5  // Половинная скорость
	TimeScaleQuarter = 0.25 // Четвертная скорость
	TimeScaleDouble  = 2.0  // Двойная скорость
	TimeScaleQuad    = 4.0  // Четырехкратная скорость
	TimeScaleOcta    = 8.0  // Восьмикратная скорость

	// Масштабы спрайтов
	// ИСПРАВЛЕНО: Уменьшен размер волка в 1.75 раза для лучшего баланса
	RabbitSpriteScale = 0.067 // Масштаб спрайта зайца (1/15)
	WolfSpriteScale   = 0.114 // Масштаб спрайта волка (уменьшен с 0.2)

	// Константы шрифтов
	DefaultFontSize = 14 // Размер шрифта по умолчанию

	// Константы эффектов
	DamageFlashIntensityMultiplier = 5.0 // Множитель интенсивности вспышки урона (белый эффект)
)

// ===== ФУНКЦИИ КОНВЕРТАЦИИ КООРДИНАТ =====
// Унифицированные функции для конвертации между тайлами и пикселями
// Устраняют магические числа 32.0 по всему коду

// TilesToPixels конвертирует тайлы в пиксели
func TilesToPixels(tiles float32) float32 {
	return tiles * TileSizePixels
}

// PixelsToTiles конвертирует пиксели в тайлы
func PixelsToTiles(pixels float32) float32 {
	return pixels / TileSizePixels
}

// TilesToPixelsInt конвертирует тайлы в пиксели (int версия)
func TilesToPixelsInt(tiles int) int {
	return tiles * TileSizePixels
}

// PixelsToTilesInt конвертирует пиксели в тайлы (int версия)
func PixelsToTilesInt(pixels int) int {
	return pixels / TileSizePixels
}

// WorldSizePixels вычисляет размер мира в пикселях по размеру в тайлах
func WorldSizePixels(tileSizeX, tileSizeY int) (float32, float32) {
	return float32(tileSizeX * TileSizePixels), float32(tileSizeY * TileSizePixels)
}

// ===== ФУНКЦИИ КОНВЕРТАЦИИ РАЗМЕРОВ =====
// Size компонент хранит размеры в пикселях (для рендеринга)
// Но игровая логика работает в тайлах - нужна конвертация

// SizeRadiusToTiles конвертирует радиус из Size компонента (пиксели) в тайлы для игровой логики
func SizeRadiusToTiles(pixelRadius float32) float32 {
	return pixelRadius / TileSizePixels
}

// SizeAttackRangeToTiles конвертирует радиус атаки из Size компонента (пиксели) в тайлы
func SizeAttackRangeToTiles(pixelAttackRange float32) float32 {
	return pixelAttackRange / TileSizePixels
}

// ===== HELPER ФУНКЦИИ ДЛЯ УСТРАНЕНИЯ ДУБЛИРОВАНИЯ КОДА =====
// Эти функции устраняют нарушения DRY принципа в movement.go и других файлах

// Vec2ToPixels конвертирует 2D вектор из тайлов в пиксели
func Vec2ToPixels(x, y float32) (float32, float32) {
	return TilesToPixels(x), TilesToPixels(y)
}

// Vec2ToTiles конвертирует 2D вектор из пикселей в тайлы
func Vec2ToTiles(x, y float32) (float32, float32) {
	return PixelsToTiles(x), PixelsToTiles(y)
}

// VelocityToPixels конвертирует скорость из тайлов/сек в пиксели/сек
func VelocityToPixels(velX, velY float32) (float32, float32) {
	return TilesToPixels(velX), TilesToPixels(velY)
}

// PositionToTiles конвертирует позицию из пикселей в тайлы
func PositionToTiles(posX, posY float32) (float32, float32) {
	return PixelsToTiles(posX), PixelsToTiles(posY)
}

// WorldBoundsToPixels конвертирует размеры мира из тайлов в пиксели
func WorldBoundsToPixels(worldWidth, worldHeight float32) (float32, float32) {
	return TilesToPixels(worldWidth), TilesToPixels(worldHeight)
}
