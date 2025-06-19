package core

import "github.com/aiseeq/savanna/internal/constants"

// ComponentManager управляет компонентами сущностей
// Соблюдает Single Responsibility Principle - только управление компонентами
type ComponentManager struct {
	// Компоненты - индексируются по EntityID (Structure of Arrays для производительности)
	positions     [MaxEntities]Position
	velocities    [MaxEntities]Velocity
	healths       [MaxEntities]Health
	hungers       [MaxEntities]Hunger
	types         [MaxEntities]AnimalType
	sizes         [MaxEntities]Size
	speeds        [MaxEntities]Speed
	animations    [MaxEntities]Animation
	damageFlashes [MaxEntities]DamageFlash
	corpses       [MaxEntities]Corpse
	carrions      [MaxEntities]Carrion
	eatingStates  [MaxEntities]EatingState
	attackStates  [MaxEntities]AttackState
	behaviors     [MaxEntities]Behavior
	animalConfigs [MaxEntities]AnimalConfig

	// Битовые маски для быстрой проверки наличия компонентов
	hasPosition     [MaxEntities/64 + 1]uint64
	hasVelocity     [MaxEntities/64 + 1]uint64
	hasHealth       [MaxEntities/64 + 1]uint64
	hasHunger       [MaxEntities/64 + 1]uint64
	hasType         [MaxEntities/64 + 1]uint64
	hasSize         [MaxEntities/64 + 1]uint64
	hasSpeed        [MaxEntities/64 + 1]uint64
	hasAnimation    [MaxEntities/64 + 1]uint64
	hasDamageFlash  [MaxEntities/64 + 1]uint64
	hasCorpse       [MaxEntities/64 + 1]uint64
	hasCarrion      [MaxEntities/64 + 1]uint64
	hasEatingState  [MaxEntities/64 + 1]uint64
	hasAttackState  [MaxEntities/64 + 1]uint64
	hasBehavior     [MaxEntities/64 + 1]uint64
	hasAnimalConfig [MaxEntities/64 + 1]uint64
}

// NewComponentManager создаёт новый менеджер компонентов
func NewComponentManager() *ComponentManager {
	return &ComponentManager{}
}

// HasComponent проверяет наличие компонента у сущности
//
//nolint:gocyclo // Оптимальный switch для производительности ECS
func (cm *ComponentManager) HasComponent(entity EntityID, component ComponentMask) bool {
	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64

	switch component {
	case MaskPosition:
		return cm.hasPosition[index]&(1<<bit) != 0
	case MaskVelocity:
		return cm.hasVelocity[index]&(1<<bit) != 0
	case MaskHealth:
		return cm.hasHealth[index]&(1<<bit) != 0
	case MaskHunger:
		return cm.hasHunger[index]&(1<<bit) != 0
	case MaskAnimalType:
		return cm.hasType[index]&(1<<bit) != 0
	case MaskSize:
		return cm.hasSize[index]&(1<<bit) != 0
	case MaskSpeed:
		return cm.hasSpeed[index]&(1<<bit) != 0
	case MaskAnimation:
		return cm.hasAnimation[index]&(1<<bit) != 0
	case MaskDamageFlash:
		return cm.hasDamageFlash[index]&(1<<bit) != 0
	case MaskCorpse:
		return cm.hasCorpse[index]&(1<<bit) != 0
	case MaskCarrion:
		return cm.hasCarrion[index]&(1<<bit) != 0
	case MaskEatingState:
		return cm.hasEatingState[index]&(1<<bit) != 0
	case MaskAttackState:
		return cm.hasAttackState[index]&(1<<bit) != 0
	case MaskBehavior:
		return cm.hasBehavior[index]&(1<<bit) != 0
	case MaskAnimalConfig:
		return cm.hasAnimalConfig[index]&(1<<bit) != 0
	default:
		return false
	}
}

// HasComponents проверяет наличие всех указанных компонентов у сущности
func (cm *ComponentManager) HasComponents(entity EntityID, mask ComponentMask) bool {
	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	entityMask := uint64(1 << bit)

	requiredComponents := []struct {
		mask ComponentMask
		bits *[MaxEntities/64 + 1]uint64
	}{
		{MaskPosition, &cm.hasPosition},
		{MaskVelocity, &cm.hasVelocity},
		{MaskHealth, &cm.hasHealth},
		{MaskHunger, &cm.hasHunger},
		{MaskAnimalType, &cm.hasType},
		{MaskSize, &cm.hasSize},
		{MaskSpeed, &cm.hasSpeed},
		{MaskAnimation, &cm.hasAnimation},
		{MaskDamageFlash, &cm.hasDamageFlash},
		{MaskCorpse, &cm.hasCorpse},
		{MaskCarrion, &cm.hasCarrion},
		{MaskEatingState, &cm.hasEatingState},
		{MaskAttackState, &cm.hasAttackState},
		{MaskBehavior, &cm.hasBehavior},
		{MaskAnimalConfig, &cm.hasAnimalConfig},
	}

	for _, comp := range requiredComponents {
		if mask&comp.mask != 0 {
			if comp.bits[index]&entityMask == 0 {
				return false
			}
		}
	}

	return true
}

// ClearAllComponents удаляет все компоненты у сущности (для DestroyEntity)
func (cm *ComponentManager) ClearAllComponents(entity EntityID) {
	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	clearMask := ^(uint64(1) << bit) // Инвертированная маска для очистки бита

	// Очищаем все битовые маски
	cm.hasPosition[index] &= clearMask
	cm.hasVelocity[index] &= clearMask
	cm.hasHealth[index] &= clearMask
	cm.hasHunger[index] &= clearMask
	cm.hasType[index] &= clearMask
	cm.hasSize[index] &= clearMask
	cm.hasSpeed[index] &= clearMask
	cm.hasAnimation[index] &= clearMask
	cm.hasDamageFlash[index] &= clearMask
	cm.hasCorpse[index] &= clearMask
	cm.hasCarrion[index] &= clearMask
	cm.hasEatingState[index] &= clearMask
	cm.hasAttackState[index] &= clearMask
	cm.hasBehavior[index] &= clearMask
	cm.hasAnimalConfig[index] &= clearMask

	// Очищаем данные компонентов (обнуляем для предотвращения утечек памяти)
	cm.positions[entity] = Position{}
	cm.velocities[entity] = Velocity{}
	cm.healths[entity] = Health{}
	cm.hungers[entity] = Hunger{}
	cm.types[entity] = AnimalType(0)
	cm.sizes[entity] = Size{}
	cm.speeds[entity] = Speed{}
	cm.animations[entity] = Animation{}
	cm.damageFlashes[entity] = DamageFlash{}
	cm.corpses[entity] = Corpse{}
	cm.carrions[entity] = Carrion{}
	cm.eatingStates[entity] = EatingState{}
	cm.attackStates[entity] = AttackState{}
	cm.behaviors[entity] = Behavior{}
	cm.animalConfigs[entity] = AnimalConfig{}
}
