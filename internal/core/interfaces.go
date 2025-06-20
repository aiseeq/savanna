package core

import "math/rand"

// Упрощённая архитектура интерфейсов (применяем YAGNI)
// Сокращено с 25 интерфейсов до 3 базовых

// QueryFunc тип функции для итерации по сущностям
type QueryFunc func(EntityID)

// ===== СПЕЦИАЛИЗИРОВАННЫЕ ИНТЕРФЕЙСЫ (ISP - Interface Segregation Principle) =====

// EntityProvider интерфейс для управления сущностями
type EntityProvider interface {
	CreateEntity() EntityID
	DestroyEntity(EntityID) bool
	IsAlive(EntityID) bool
	GetEntityCount() int
}

// ComponentReader интерфейс для чтения компонентов
type ComponentReader interface {
	// Position
	GetPosition(EntityID) (Position, bool)
	// Velocity
	GetVelocity(EntityID) (Velocity, bool)
	// Health
	GetHealth(EntityID) (Health, bool)
	// Hunger
	GetHunger(EntityID) (Hunger, bool)
	// AnimalType
	GetAnimalType(EntityID) (AnimalType, bool)
	// Size
	GetSize(EntityID) (Size, bool)
	// Speed
	GetSpeed(EntityID) (Speed, bool)
	// Animation
	GetAnimation(EntityID) (Animation, bool)
	// Behavior
	GetBehavior(EntityID) (Behavior, bool)
	// EatingState
	GetEatingState(EntityID) (EatingState, bool)
	// AttackState
	GetAttackState(EntityID) (AttackState, bool)
	// DamageFlash
	GetDamageFlash(EntityID) (DamageFlash, bool)
	// Corpse
	GetCorpse(EntityID) (Corpse, bool)
	// Carrion
	GetCarrion(EntityID) (Carrion, bool)
	// AnimalConfig
	GetAnimalConfig(EntityID) (AnimalConfig, bool)
}

// ComponentWriter интерфейс для изменения компонентов
type ComponentWriter interface {
	// Position
	SetPosition(EntityID, Position) bool
	AddPosition(EntityID, Position) bool
	// Velocity
	SetVelocity(EntityID, Velocity) bool
	// Health
	SetHealth(EntityID, Health) bool
	// Hunger
	SetHunger(EntityID, Hunger) bool
	// AnimalType
	SetAnimalType(EntityID, AnimalType) bool
	// Speed
	SetSpeed(EntityID, Speed) bool
	// Animation
	SetAnimation(EntityID, Animation) bool
	// Behavior
	SetBehavior(EntityID, Behavior) bool
	// EatingState
	SetEatingState(EntityID, EatingState) bool
	AddEatingState(EntityID, EatingState) bool
	RemoveEatingState(EntityID) bool
	// AttackState
	SetAttackState(EntityID, AttackState) bool
	AddAttackState(EntityID, AttackState) bool
	RemoveAttackState(EntityID) bool
	// DamageFlash
	SetDamageFlash(EntityID, DamageFlash) bool
	AddDamageFlash(EntityID, DamageFlash) bool
	RemoveDamageFlash(EntityID) bool
	// Corpse
	SetCorpse(EntityID, Corpse) bool
	AddCorpse(EntityID, Corpse) bool
	RemoveCorpse(EntityID) bool
	// Carrion
	SetCarrion(EntityID, Carrion) bool
	AddCarrion(EntityID, Carrion) bool
	RemoveCarrion(EntityID) bool
	// AnimalConfig
	SetAnimalConfig(EntityID, AnimalConfig) bool
}

// QueryProvider интерфейс для ECS запросов
type QueryProvider interface {
	ForEachWith(ComponentMask, QueryFunc)
	HasComponent(EntityID, ComponentMask) bool
	GetEntitiesWith(ComponentMask) []EntityID
	CountEntitiesWith(ComponentMask) int
}

// SpatialQueries интерфейс для пространственных запросов
type SpatialQueries interface {
	QueryInRadius(x, y, radius float32) []EntityID
	FindNearestByType(x, y, radius float32, animalType AnimalType) (EntityID, bool)
	FindNearestAnimal(x, y, radius float32) (EntityID, bool)
}

// SpatialUpdater интерфейс для обновления пространственной системы
type SpatialUpdater interface {
	UpdateSpatialPosition(EntityID, Position)
}

// WorldInfo интерфейс для состояния мира
type WorldInfo interface {
	GetTime() float32
	GetDeltaTime() float32
	GetRNG() *rand.Rand
	GetWorldDimensions() (width, height float32)
}

// ECSAccess объединённый интерфейс для обратной совместимости
// Deprecated: Используйте специализированные интерфейсы согласно ISP
type ECSAccess interface {
	EntityProvider  // Управление сущностями
	ComponentReader // Чтение компонентов
	ComponentWriter // Изменение компонентов
	QueryProvider   // ECS запросы
	SpatialQueries  // Пространственные запросы
	WorldInfo       // Состояние мира
}

// ===== КОМПОЗИТНЫЕ ИНТЕРФЕЙСЫ ДЛЯ УДОБСТВА =====
// Улучшено для соблюдения PoLA - интуитивные имена и ясное назначение

// SimulationAccess интерфейс для большинства систем симуляции
// Предоставляет: чтение/запись компонентов, ECS запросы, пространственный поиск, время/RNG
type SimulationAccess interface {
	ComponentReader // Чтение компонентов (Position, Health, Hunger и т.д.)
	ComponentWriter // Изменение компонентов (SetHealth, AddEatingState и т.д.)
	QueryProvider   // ECS запросы (ForEachWith, HasComponent и т.д.)
	SpatialQueries  // Пространственные запросы (FindNearestByType и т.д.)
	WorldInfo       // Состояние мира (время, RNG, размеры мира)
}

// ===== УЗКОСПЕЦИАЛИЗИРОВАННЫЕ ИНТЕРФЕЙСЫ (ISP УЛУЧШЕНИЯ) =====

// HungerSystemAccess специализированный интерфейс для системы голода
// Предоставляет: только компоненты голода и сытости
type HungerSystemAccess interface {
	// Чтение голода
	GetHunger(EntityID) (Hunger, bool)
	GetSize(EntityID) (Size, bool) // Для расчёта скорости голода крупных животных
	// Изменение голода
	SetHunger(EntityID, Hunger) bool
	// Итерация
	ForEachWith(ComponentMask, QueryFunc)
}

// GrassSearchSystemAccess специализированный интерфейс для поиска травы
// Предоставляет: позицию, голод, конфигурацию животного, создание EatingState
type GrassSearchSystemAccess interface {
	// Чтение состояния животного
	GetPosition(EntityID) (Position, bool)
	GetHunger(EntityID) (Hunger, bool)
	GetAnimalType(EntityID) (AnimalType, bool)
	GetAnimalConfig(EntityID) (AnimalConfig, bool)
	GetBehavior(EntityID) (Behavior, bool)
	GetSize(EntityID) (Size, bool)
	// Проверка состояний
	HasComponent(EntityID, ComponentMask) bool
	// Создание состояния поедания
	AddEatingState(EntityID, EatingState) bool
	// Итерация
	ForEachWith(ComponentMask, QueryFunc)
}

// StarvationDamageSystemAccess специализированный интерфейс для урона от голода
// Предоставляет: только голод и здоровье
type StarvationDamageSystemAccess interface {
	// Чтение состояния
	GetHunger(EntityID) (Hunger, bool)
	GetHealth(EntityID) (Health, bool)
	// Изменение здоровья
	SetHealth(EntityID, Health) bool
	// Итерация
	ForEachWith(ComponentMask, QueryFunc)
}

// HungerSpeedModifierSystemAccess специализированный интерфейс для влияния голода на скорость
// Предоставляет: только голод, здоровье и скорость
type HungerSpeedModifierSystemAccess interface {
	// Чтение состояния
	GetHunger(EntityID) (Hunger, bool)
	GetHealth(EntityID) (Health, bool)
	GetSpeed(EntityID) (Speed, bool)
	// Изменение скорости
	SetSpeed(EntityID, Speed) bool
	// Итерация
	ForEachWith(ComponentMask, QueryFunc)
}

// MovementSystemAccess специализированный интерфейс для системы движения
// Предоставляет: компоненты позиции/скорости, границы мира, пространственные обновления
type MovementSystemAccess interface {
	ComponentReader // Position, Velocity, EatingState (блокирует движение)
	ComponentWriter // Изменение Position, Velocity после перемещения
	QueryProvider   // ForEachWith для итерации по движущимся сущностям
	SpatialQueries  // QueryInRadius для проверки коллизий
	SpatialUpdater  // UpdateSpatialPosition после движения
	WorldInfo       // GetWorldDimensions для проверки границ мира
}

// BehaviorSystemAccess специализированный интерфейс для системы поведения
// Предоставляет: состояние животных, поиск целей, изменение поведения
type BehaviorSystemAccess interface {
	EntityProvider  // IsAlive для проверки валидности целей
	ComponentReader // AnimalType, Behavior, Position, Hunger для принятия решений
	ComponentWriter // Velocity, Behavior для изменения действий животного
	QueryProvider   // ForEachWith для обработки всех животных
	SpatialQueries  // FindNearestByType для поиска пищи/хищников
	WorldInfo       // GetRNG для случайных решений
}

// CombatSystemAccess специализированный интерфейс для боевой системы
// Предоставляет: здоровье, атаки, создание трупов, визуальные эффекты
type CombatSystemAccess interface {
	EntityProvider  // CreateEntity для создания трупов при смерти
	ComponentReader // Health, Position, AttackState для обработки атак
	ComponentWriter // SetHealth, AddDamageFlash для урона и эффектов
	QueryProvider   // ForEachWith для обработки боевых состояний
}

// VegetationProvider узкоспециализированный интерфейс для работы с растительностью
// Единственный оставшийся специализированный интерфейс (реально используется)
type VegetationProvider interface {
	// FindNearestGrass находит ближайшую траву
	FindNearestGrass(worldX, worldY, searchRadius, minAmount float32) (grassX, grassY float32, found bool)

	// UpdateGrassAt обновляет количество травы в указанной точке
	UpdateGrassAt(worldX, worldY, delta float32)

	// GetGrassAt возвращает количество травы в указанной точке
	GetGrassAt(worldX, worldY float32) float32

	// ConsumeGrassAt потребляет траву в указанной точке и возвращает фактически съеденное количество
	ConsumeGrassAt(worldX, worldY, amount float32) float32

	// IsPassable проверяет можно ли пройти через тайл
	IsPassable(tileX, tileY int) bool
}

// ===== ПРИНЦИПЫ SOLID: УСТРАНЕНИЕ НАРУШЕНИЙ LSP =====
// Удалены алиасы интерфейсов которые создавали ложную замещаемость.
// Теперь системы используют прямые специализированные интерфейсы:
//
// ✅ FeedingSystem: core.SimulationAccess
// ✅ GrassEatingSystem: core.SimulationAccess
// ✅ CombatSystem: core.CombatSystemAccess
// ✅ CorpseSystem: core.SimulationAccess
// ✅ AttackSystem: core.CombatSystemAccess
//
// Это соблюдает принцип Liskov Substitution Principle - каждый интерфейс
// может быть заменён на свой базовый тип без нарушения функциональности.
