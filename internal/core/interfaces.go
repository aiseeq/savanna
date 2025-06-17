package core

import (
	"math/rand"
)

// Файл interfaces.go содержит специализированные интерфейсы для соблюдения ISP (Interface Segregation Principle)
// Каждая система зависит только от тех методов, которые ей действительно нужны

// ============== БАЗОВЫЕ ДОСТУПЫ К КОМПОНЕНТАМ ==============

// PositionAccess предоставляет доступ к позициям сущностей
type PositionAccess interface {
	GetPosition(EntityID) (Position, bool)
	SetPosition(EntityID, Position) bool
}

// MovementAccess предоставляет доступ к движению (включает позиции)
type MovementAccess interface {
	PositionAccess
	GetVelocity(EntityID) (Velocity, bool)
	SetVelocity(EntityID, Velocity) bool
	GetSpeed(EntityID) (Speed, bool)
	SetSpeed(EntityID, Speed) bool
}

// HealthAccess предоставляет доступ к здоровью
type HealthAccess interface {
	GetHealth(EntityID) (Health, bool)
	SetHealth(EntityID, Health) bool
}

// HungerAccess предоставляет доступ к голоду
type HungerAccess interface {
	GetHunger(EntityID) (Hunger, bool)
	SetHunger(EntityID, Hunger) bool
}

// SizeAccess предоставляет доступ к размерам (только чтение)
type SizeAccess interface {
	GetSize(EntityID) (Size, bool)
}

// AnimationAccess предоставляет доступ к анимациям
type AnimationAccess interface {
	GetAnimation(EntityID) (Animation, bool)
	SetAnimation(EntityID, Animation) bool
}

// BehaviorAccess предоставляет доступ к поведению
type BehaviorAccess interface {
	GetBehavior(EntityID) (Behavior, bool)
	SetBehavior(EntityID, Behavior) bool
}

// AnimalTypeAccess предоставляет доступ к типам животных (только чтение)
type AnimalTypeAccess interface {
	GetAnimalType(EntityID) (AnimalType, bool)
}

// AnimalConfigAccess предоставляет доступ к конфигурации животных
type AnimalConfigAccess interface {
	GetAnimalConfig(EntityID) (AnimalConfig, bool)
	SetAnimalConfig(EntityID, AnimalConfig) bool
}

// ============== ECS CORE ==============

// ECSCore предоставляет базовые ECS операции
type ECSCore interface {
	ForEachWith(ComponentMask, QueryFunc)
	HasComponent(EntityID, ComponentMask) bool
	IsAlive(EntityID) bool
}

// SpatialQueries предоставляет пространственные запросы
type SpatialQueries interface {
	QueryInRadius(x, y, radius float32) []EntityID
	FindNearestByType(x, y, radius float32, animalType AnimalType) (EntityID, bool)
}

// RandomAccess предоставляет доступ к генератору случайных чисел
type RandomAccess interface {
	GetRNG() *rand.Rand
}

// EntityManagement предоставляет управление сущностями
type EntityManagement interface {
	CreateEntity() EntityID
	DestroyEntity(EntityID) bool
	IsAlive(EntityID) bool
}

// ============== СОСТОЯНИЯ ДЕЙСТВИЙ ==============

// CombatStateAccess предоставляет доступ к боевым состояниям
type CombatStateAccess interface {
	// Attack states
	AddAttackState(EntityID, AttackState) bool
	GetAttackState(EntityID) (AttackState, bool)
	SetAttackState(EntityID, AttackState) bool
	RemoveAttackState(EntityID) bool

	// Damage effects
	AddDamageFlash(EntityID, DamageFlash) bool
	GetDamageFlash(EntityID) (DamageFlash, bool)
	SetDamageFlash(EntityID, DamageFlash) bool
	RemoveDamageFlash(EntityID) bool
}

// EatingStateAccess предоставляет доступ к состояниям поедания
type EatingStateAccess interface {
	AddEatingState(EntityID, EatingState) bool
	GetEatingState(EntityID) (EatingState, bool)
	SetEatingState(EntityID, EatingState) bool
	RemoveEatingState(EntityID) bool
}

// CorpseAccess предоставляет доступ к трупам и падали
type CorpseAccess interface {
	// Corpses
	AddCorpse(EntityID, Corpse) bool
	GetCorpse(EntityID) (Corpse, bool)
	SetCorpse(EntityID, Corpse) bool
	RemoveCorpse(EntityID) bool

	// Carrion
	AddCarrion(EntityID, Carrion) bool
	GetCarrion(EntityID) (Carrion, bool)
	SetCarrion(EntityID, Carrion) bool
	RemoveCarrion(EntityID) bool
}

// ComponentRemoval предоставляет удаление компонентов
type ComponentRemoval interface {
	RemoveVelocity(EntityID) bool
	RemoveHunger(EntityID) bool
	RemoveSpeed(EntityID) bool
}

// ============== КОМПОЗИТНЫЕ ИНТЕРФЕЙСЫ ДЛЯ СИСТЕМ ==============

// MovementSystemAccess - всё что нужно MovementSystem
type MovementSystemAccess interface {
	MovementAccess
	SizeAccess
	ECSCore
	SpatialQueries
}

// FeedingSystemAccess - всё что нужно FeedingSystem
type FeedingSystemAccess interface {
	HungerAccess
	HealthAccess
	MovementAccess // для SetSpeed
	SizeAccess
	AnimalTypeAccess
	AnimalConfigAccess // для получения порогов голода
	BehaviorAccess     // для обработки травоядных
	EatingStateAccess
	ECSCore
}

// BehaviorSystemAccess - всё что нужно AnimalBehaviorSystem
type BehaviorSystemAccess interface {
	BehaviorAccess
	PositionAccess
	GetVelocity(EntityID) (Velocity, bool)
	SetVelocity(EntityID, Velocity) bool
	GetSpeed(EntityID) (Speed, bool)
	HungerAccess
	AnimalConfigAccess // для получения размеров животного
	AnimationAccess
	EatingStateAccess // Нужно для прерывания поедания при побеге
	SpatialQueries
	RandomAccess
	ECSCore
}

// AttackSystemAccess - всё что нужно AttackSystem
type AttackSystemAccess interface {
	PositionAccess
	SizeAccess
	BehaviorAccess
	HealthAccess
	HungerAccess
	CombatStateAccess
	AnimationAccess
	RandomAccess
	ECSCore
	EntityManagement
}

// EatingSystemAccess - всё что нужно EatingSystem
type EatingSystemAccess interface {
	BehaviorAccess
	PositionAccess
	HungerAccess
	EatingStateAccess
	CorpseAccess
	EntityManagement
	ECSCore
}

// DamageSystemAccess - всё что нужно DamageSystem
type DamageSystemAccess interface {
	GetDamageFlash(EntityID) (DamageFlash, bool)
	SetDamageFlash(EntityID, DamageFlash) bool
	RemoveDamageFlash(EntityID) bool
	ECSCore
}

// CorpseSystemAccess - всё что нужно CorpseSystem
type CorpseSystemAccess interface {
	HealthAccess
	AnimalTypeAccess
	CorpseAccess
	AnimationAccess
	ComponentRemoval
	EntityManagement
	ECSCore
}

// ============== ОБРАТНАЯ СОВМЕСТИМОСТЬ ==============

// WorldAccess предоставляет все методы World (для обратной совместимости)
// Постепенно системы должны переходить на специализированные интерфейсы
type WorldAccess interface {
	MovementAccess
	HealthAccess
	HungerAccess
	SizeAccess
	AnimationAccess
	BehaviorAccess
	AnimalTypeAccess
	AnimalConfigAccess // для новой SOLID-архитектуры
	ECSCore
	SpatialQueries
	RandomAccess
	EntityManagement
	CombatStateAccess
	EatingStateAccess
	CorpseAccess
	ComponentRemoval
}
