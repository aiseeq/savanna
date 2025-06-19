package core

import "math/rand"

// Упрощённая архитектура интерфейсов (применяем YAGNI)
// Сокращено с 25 интерфейсов до 3 базовых

// QueryFunc тип функции для итерации по сущностям
type QueryFunc func(EntityID)

// ===== ОСНОВНЫЕ ИНТЕРФЕЙСЫ (3 штуки вместо 25) =====

// ECSAccess базовый интерфейс для всех ECS операций
// Объединяет все необходимые методы в одном месте (соблюдает KISS)
type ECSAccess interface {
	// === УПРАВЛЕНИЕ СУЩНОСТЯМИ ===
	CreateEntity() EntityID
	DestroyEntity(EntityID) bool
	IsAlive(EntityID) bool
	GetEntityCount() int

	// === КОМПОНЕНТЫ ===
	// Position
	GetPosition(EntityID) (Position, bool)
	SetPosition(EntityID, Position) bool
	AddPosition(EntityID, Position) bool

	// Velocity
	GetVelocity(EntityID) (Velocity, bool)
	SetVelocity(EntityID, Velocity) bool

	// Health
	GetHealth(EntityID) (Health, bool)
	SetHealth(EntityID, Health) bool

	// Hunger
	GetHunger(EntityID) (Hunger, bool)
	SetHunger(EntityID, Hunger) bool

	// AnimalType
	GetAnimalType(EntityID) (AnimalType, bool)
	SetAnimalType(EntityID, AnimalType) bool

	// Size
	GetSize(EntityID) (Size, bool)

	// Speed
	GetSpeed(EntityID) (Speed, bool)
	SetSpeed(EntityID, Speed) bool

	// Animation
	GetAnimation(EntityID) (Animation, bool)
	SetAnimation(EntityID, Animation) bool

	// Behavior
	GetBehavior(EntityID) (Behavior, bool)
	SetBehavior(EntityID, Behavior) bool

	// EatingState
	GetEatingState(EntityID) (EatingState, bool)
	SetEatingState(EntityID, EatingState) bool
	AddEatingState(EntityID, EatingState) bool
	RemoveEatingState(EntityID) bool

	// AttackState
	GetAttackState(EntityID) (AttackState, bool)
	SetAttackState(EntityID, AttackState) bool
	AddAttackState(EntityID, AttackState) bool
	RemoveAttackState(EntityID) bool

	// DamageFlash
	GetDamageFlash(EntityID) (DamageFlash, bool)
	SetDamageFlash(EntityID, DamageFlash) bool
	AddDamageFlash(EntityID, DamageFlash) bool
	RemoveDamageFlash(EntityID) bool

	// Corpse
	GetCorpse(EntityID) (Corpse, bool)
	SetCorpse(EntityID, Corpse) bool
	AddCorpse(EntityID, Corpse) bool
	RemoveCorpse(EntityID) bool

	// Carrion
	GetCarrion(EntityID) (Carrion, bool)
	SetCarrion(EntityID, Carrion) bool
	AddCarrion(EntityID, Carrion) bool
	RemoveCarrion(EntityID) bool

	// AnimalConfig
	GetAnimalConfig(EntityID) (AnimalConfig, bool)
	SetAnimalConfig(EntityID, AnimalConfig) bool

	// === ECS ЗАПРОСЫ ===
	ForEachWith(ComponentMask, QueryFunc)
	HasComponent(EntityID, ComponentMask) bool
	GetEntitiesWith(ComponentMask) []EntityID
	CountEntitiesWith(ComponentMask) int

	// === ПРОСТРАНСТВЕННЫЕ ЗАПРОСЫ ===
	QueryInRadius(x, y, radius float32) []EntityID
	FindNearestByType(x, y, radius float32, animalType AnimalType) (EntityID, bool)
	FindNearestAnimal(x, y, radius float32) (EntityID, bool)

	// === ВРЕМЯ И СОСТОЯНИЕ ===
	GetTime() float32
	GetDeltaTime() float32
	GetRNG() *rand.Rand
	GetWorldDimensions() (width, height float32)
}

// SimulationAccess интерфейс для систем симуляции
// Алиас для ECSAccess для ясности (можно расширить в будущем если понадобится)
type SimulationAccess interface {
	ECSAccess
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

	// IsPassable проверяет можно ли пройти через тайл
	IsPassable(tileX, tileY int) bool
}

// ===== LEGACY АЛИАСЫ ДЛЯ ОБРАТНОЙ СОВМЕСТИМОСТИ =====
// Эти алиасы позволяют сохранить обратную совместимость с существующим кодом
// В будущем могут быть удалены после полного перехода на упрощённую архитектуру

type MovementSystemAccess = SimulationAccess
type FeedingSystemAccess = SimulationAccess
type BehaviorSystemAccess = SimulationAccess
type AttackSystemAccess = SimulationAccess
type EatingSystemAccess = SimulationAccess
type DamageSystemAccess = SimulationAccess
type CorpseSystemAccess = SimulationAccess
type WorldAccess = SimulationAccess
