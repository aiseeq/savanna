package common

// Общие константы для тестов (устраняет DRY нарушения)
// Все магические числа вынесены в именованные константы

const (
	// Размеры мира для разных типов тестов
	SmallWorldSize  = 320.0  // Для unit тестов
	MediumWorldSize = 640.0  // Для integration тестов
	LargeWorldSize  = 1600.0 // Для E2E тестов

	// Временные константы
	StandardTPS       = 60
	StandardDeltaTime = 1.0 / 60.0

	// Семена для детерминированных тестов
	DefaultTestSeed     = 42
	AlternativeTestSeed = 12345

	// Пороги голода для тестов (соответствуют игровой логике)
	WolfAttackSatiationThreshold    = 60.0
	RabbitFeedingSatiationThreshold = 90.0

	// Стандартные значения здоровья
	RabbitMaxHealth = int16(50)
	WolfMaxHealth   = int16(100)

	// Временные константы для симуляций
	OneSecondTicks    = 60   // 60 тиков = 1 секунда при 60 TPS
	FiveSecondTicks   = 300  // 300 тиков = 5 секунд
	TenSecondTicks    = 600  // 600 тиков = 10 секунд
	ThirtySecondTicks = 1800 // 1800 тиков = 30 секунд

	// Дистанции для тестов
	CloseDistance  = 5.0   // Очень близко
	AttackDistance = 30.0  // Дистанция атаки
	VisionDistance = 100.0 // В пределах видимости
	FarDistance    = 500.0 // Далеко

	// Позиции для стандартных тестовых сценариев
	RabbitStartX = 300.0
	RabbitStartY = 300.0
	WolfStartX   = 310.0 // Рядом с зайцем
	WolfStartY   = 300.0

	// Проценты голода для типичных тестовых сценариев
	VeryHungryPercentage = 20.0 // Очень голодный
	HungryPercentage     = 50.0 // Голодный
	SatedPercentage      = 80.0 // Сытый
	FullPercentage       = 95.0 // Полный
)
