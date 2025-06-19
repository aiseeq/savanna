package simulation

// Константы для создания животных - используются только при создании
// Системы работают через компоненты Size, Speed, AnimalType

// Параметры создания зайца (используют константы из game_balance.go)
const (
	RabbitRadius        = RabbitBaseRadius // Радиус коллизий зайца
	RabbitSpeed         = RabbitBaseSpeed  // Базовая скорость зайца
	RabbitInitialHunger = 80.0             // Начальная сытость зайца (80%)
)

// Параметры создания волка (используют константы из game_balance.go)
const (
	WolfRadius        = WolfBaseRadius                             // Радиус коллизий волка
	WolfSpeed         = WolfBaseSpeed                              // Базовая скорость волка
	WolfAttackRange   = WolfBaseRadius * WolfAttackRangeMultiplier // Дальность атаки волка
	WolfInitialHunger = 70.0                                       // Начальная сытость волка (70%)
)

// Дальности видения (используются при создании компонента Behavior)
const (
	VisionRangeRabbit = RabbitBaseRadius * RabbitVisionMultiplier // Дальность видения зайца
	VisionRangeWolf   = WolfBaseRadius * WolfVisionMultiplier     // Дальность видения волка
)

// Пороги поведения (используются системами через компонент AnimalType)
const (
	RabbitHungryThreshold = RabbitHungerThreshold // Заяц начинает есть при голоде < 90%
)

// Константы унаследованы из game_balance.go (убираем дублирование)
// Здесь остаются только алиасы для обратной совместимости

// Константы для обратной совместимости (используют game_balance.go)
const (
	// Алиасы системных констант
	HungerDecreaseRate = BaseHungerDecreaseRate
	HealthDamageRate   = BaseHealthDamageRate
	MinGrassToFind     = MinGrassAmountToFind
	MaxHungerValue     = MaxHungerLimit

	// Алиасы констант поедания травы (сохраняем оригинальные имена)
	GrassPerTick        = GrassPerEatingTick
	GrassNutritionRatio = GrassNutritionValue

	// Алиасы множителей скорости (сохраняем оригинальные имена)
	SpeedSearchingFood = SearchSpeedMultiplier
	SpeedWanderingFood = WanderingSpeedMultiplier
	SpeedContentWalk   = ContentSpeedMultiplier

	// Алиасы порогов влияния голода на скорость
	FastHungerThreshold = OverfedSpeedThreshold

	// Алиасы параметров случайного движения (сохраняем оригинальные имена)
	RandomWalkMinTime = RandomWalkTimeMin
	RandomWalkMaxTime = RandomWalkTimeMax
	RandomSpeedMin    = RandomSpeedMinMultiplier
	RandomSpeedMax    = RandomSpeedMaxMultiplier
)
