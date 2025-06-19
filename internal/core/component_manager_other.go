package core

import "github.com/aiseeq/savanna/internal/constants"

// AnimalType component management

// AddAnimalType добавляет компонент AnimalType к сущности
func (cm *ComponentManager) AddAnimalType(entity EntityID, animalType AnimalType) {
	cm.types[entity] = animalType

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasType[index] |= 1 << bit
}

// GetAnimalType возвращает компонент AnimalType сущности
func (cm *ComponentManager) GetAnimalType(entity EntityID) (AnimalType, bool) {
	if !cm.HasComponent(entity, MaskAnimalType) {
		return AnimalType(0), false
	}
	return cm.types[entity], true
}

// SetAnimalType обновляет компонент AnimalType сущности
func (cm *ComponentManager) SetAnimalType(entity EntityID, animalType AnimalType) bool {
	if !cm.HasComponent(entity, MaskAnimalType) {
		return false
	}
	cm.types[entity] = animalType
	return true
}

// RemoveAnimalType удаляет компонент AnimalType у сущности
func (cm *ComponentManager) RemoveAnimalType(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskAnimalType) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasType[index] &= ^(1 << bit)
	cm.types[entity] = AnimalType(0)

	return true
}

// Size component management

// AddSize добавляет компонент Size к сущности
func (cm *ComponentManager) AddSize(entity EntityID, size Size) {
	cm.sizes[entity] = size

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasSize[index] |= 1 << bit
}

// GetSize возвращает компонент Size сущности
func (cm *ComponentManager) GetSize(entity EntityID) (Size, bool) {
	if !cm.HasComponent(entity, MaskSize) {
		return Size{}, false
	}
	return cm.sizes[entity], true
}

// SetSize обновляет компонент Size сущности
func (cm *ComponentManager) SetSize(entity EntityID, size Size) bool {
	if !cm.HasComponent(entity, MaskSize) {
		return false
	}
	cm.sizes[entity] = size
	return true
}

// RemoveSize удаляет компонент Size у сущности
func (cm *ComponentManager) RemoveSize(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskSize) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasSize[index] &= ^(1 << bit)
	cm.sizes[entity] = Size{}

	return true
}

// Speed component management

// AddSpeed добавляет компонент Speed к сущности
func (cm *ComponentManager) AddSpeed(entity EntityID, speed Speed) {
	cm.speeds[entity] = speed

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasSpeed[index] |= 1 << bit
}

// GetSpeed возвращает компонент Speed сущности
func (cm *ComponentManager) GetSpeed(entity EntityID) (Speed, bool) {
	if !cm.HasComponent(entity, MaskSpeed) {
		return Speed{}, false
	}
	return cm.speeds[entity], true
}

// SetSpeed обновляет компонент Speed сущности
func (cm *ComponentManager) SetSpeed(entity EntityID, speed Speed) bool {
	if !cm.HasComponent(entity, MaskSpeed) {
		return false
	}
	cm.speeds[entity] = speed
	return true
}

// RemoveSpeed удаляет компонент Speed у сущности
func (cm *ComponentManager) RemoveSpeed(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskSpeed) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasSpeed[index] &= ^(1 << bit)
	cm.speeds[entity] = Speed{}

	return true
}

// Animation component management

// AddAnimation добавляет компонент Animation к сущности
func (cm *ComponentManager) AddAnimation(entity EntityID, animation Animation) {
	cm.animations[entity] = animation

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasAnimation[index] |= 1 << bit
}

// GetAnimation возвращает компонент Animation сущности
func (cm *ComponentManager) GetAnimation(entity EntityID) (Animation, bool) {
	if !cm.HasComponent(entity, MaskAnimation) {
		return Animation{}, false
	}
	return cm.animations[entity], true
}

// SetAnimation обновляет компонент Animation сущности
func (cm *ComponentManager) SetAnimation(entity EntityID, animation Animation) bool {
	if !cm.HasComponent(entity, MaskAnimation) {
		return false
	}
	cm.animations[entity] = animation
	return true
}

// RemoveAnimation удаляет компонент Animation у сущности
func (cm *ComponentManager) RemoveAnimation(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskAnimation) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasAnimation[index] &= ^(1 << bit)
	cm.animations[entity] = Animation{}

	return true
}

// DamageFlash component management

// AddDamageFlash добавляет компонент DamageFlash к сущности
func (cm *ComponentManager) AddDamageFlash(entity EntityID, damageFlash DamageFlash) {
	cm.damageFlashes[entity] = damageFlash

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasDamageFlash[index] |= 1 << bit
}

// GetDamageFlash возвращает компонент DamageFlash сущности
func (cm *ComponentManager) GetDamageFlash(entity EntityID) (DamageFlash, bool) {
	if !cm.HasComponent(entity, MaskDamageFlash) {
		return DamageFlash{}, false
	}
	return cm.damageFlashes[entity], true
}

// SetDamageFlash обновляет компонент DamageFlash сущности
func (cm *ComponentManager) SetDamageFlash(entity EntityID, damageFlash DamageFlash) bool {
	if !cm.HasComponent(entity, MaskDamageFlash) {
		return false
	}
	cm.damageFlashes[entity] = damageFlash
	return true
}

// RemoveDamageFlash удаляет компонент DamageFlash у сущности
func (cm *ComponentManager) RemoveDamageFlash(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskDamageFlash) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasDamageFlash[index] &= ^(1 << bit)
	cm.damageFlashes[entity] = DamageFlash{}

	return true
}

// Corpse component management

// AddCorpse добавляет компонент Corpse к сущности
func (cm *ComponentManager) AddCorpse(entity EntityID, corpse Corpse) {
	cm.corpses[entity] = corpse

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasCorpse[index] |= 1 << bit
}

// GetCorpse возвращает компонент Corpse сущности
func (cm *ComponentManager) GetCorpse(entity EntityID) (Corpse, bool) {
	if !cm.HasComponent(entity, MaskCorpse) {
		return Corpse{}, false
	}
	return cm.corpses[entity], true
}

// SetCorpse обновляет компонент Corpse сущности
func (cm *ComponentManager) SetCorpse(entity EntityID, corpse Corpse) bool {
	if !cm.HasComponent(entity, MaskCorpse) {
		return false
	}
	cm.corpses[entity] = corpse
	return true
}

// RemoveCorpse удаляет компонент Corpse у сущности
func (cm *ComponentManager) RemoveCorpse(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskCorpse) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasCorpse[index] &= ^(1 << bit)
	cm.corpses[entity] = Corpse{}

	return true
}

// Carrion component management

// AddCarrion добавляет компонент Carrion к сущности
func (cm *ComponentManager) AddCarrion(entity EntityID, carrion Carrion) {
	cm.carrions[entity] = carrion

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasCarrion[index] |= 1 << bit
}

// GetCarrion возвращает компонент Carrion сущности
func (cm *ComponentManager) GetCarrion(entity EntityID) (Carrion, bool) {
	if !cm.HasComponent(entity, MaskCarrion) {
		return Carrion{}, false
	}
	return cm.carrions[entity], true
}

// SetCarrion обновляет компонент Carrion сущности
func (cm *ComponentManager) SetCarrion(entity EntityID, carrion Carrion) bool {
	if !cm.HasComponent(entity, MaskCarrion) {
		return false
	}
	cm.carrions[entity] = carrion
	return true
}

// RemoveCarrion удаляет компонент Carrion у сущности
func (cm *ComponentManager) RemoveCarrion(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskCarrion) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasCarrion[index] &= ^(1 << bit)
	cm.carrions[entity] = Carrion{}

	return true
}

// EatingState component management

// AddEatingState добавляет компонент EatingState к сущности
func (cm *ComponentManager) AddEatingState(entity EntityID, eatingState EatingState) {
	cm.eatingStates[entity] = eatingState

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasEatingState[index] |= 1 << bit
}

// GetEatingState возвращает компонент EatingState сущности
func (cm *ComponentManager) GetEatingState(entity EntityID) (EatingState, bool) {
	if !cm.HasComponent(entity, MaskEatingState) {
		return EatingState{}, false
	}
	return cm.eatingStates[entity], true
}

// SetEatingState обновляет компонент EatingState сущности
func (cm *ComponentManager) SetEatingState(entity EntityID, eatingState EatingState) bool {
	if !cm.HasComponent(entity, MaskEatingState) {
		return false
	}
	cm.eatingStates[entity] = eatingState
	return true
}

// RemoveEatingState удаляет компонент EatingState у сущности
func (cm *ComponentManager) RemoveEatingState(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskEatingState) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasEatingState[index] &= ^(1 << bit)
	cm.eatingStates[entity] = EatingState{}

	return true
}

// AttackState component management

// AddAttackState добавляет компонент AttackState к сущности
func (cm *ComponentManager) AddAttackState(entity EntityID, attackState AttackState) {
	cm.attackStates[entity] = attackState

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasAttackState[index] |= 1 << bit
}

// GetAttackState возвращает компонент AttackState сущности
func (cm *ComponentManager) GetAttackState(entity EntityID) (AttackState, bool) {
	if !cm.HasComponent(entity, MaskAttackState) {
		return AttackState{}, false
	}
	return cm.attackStates[entity], true
}

// SetAttackState обновляет компонент AttackState сущности
func (cm *ComponentManager) SetAttackState(entity EntityID, attackState AttackState) bool {
	if !cm.HasComponent(entity, MaskAttackState) {
		return false
	}
	cm.attackStates[entity] = attackState
	return true
}

// RemoveAttackState удаляет компонент AttackState у сущности
func (cm *ComponentManager) RemoveAttackState(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskAttackState) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasAttackState[index] &= ^(1 << bit)
	cm.attackStates[entity] = AttackState{}

	return true
}

// Behavior component management

// AddBehavior добавляет компонент Behavior к сущности
func (cm *ComponentManager) AddBehavior(entity EntityID, behavior Behavior) {
	cm.behaviors[entity] = behavior

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasBehavior[index] |= 1 << bit
}

// GetBehavior возвращает компонент Behavior сущности
func (cm *ComponentManager) GetBehavior(entity EntityID) (Behavior, bool) {
	if !cm.HasComponent(entity, MaskBehavior) {
		return Behavior{}, false
	}
	return cm.behaviors[entity], true
}

// SetBehavior обновляет компонент Behavior сущности
func (cm *ComponentManager) SetBehavior(entity EntityID, behavior Behavior) bool {
	if !cm.HasComponent(entity, MaskBehavior) {
		return false
	}
	cm.behaviors[entity] = behavior
	return true
}

// RemoveBehavior удаляет компонент Behavior у сущности
func (cm *ComponentManager) RemoveBehavior(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskBehavior) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasBehavior[index] &= ^(1 << bit)
	cm.behaviors[entity] = Behavior{}

	return true
}

// AnimalConfig component management

// AddAnimalConfig добавляет компонент AnimalConfig к сущности
func (cm *ComponentManager) AddAnimalConfig(entity EntityID, config AnimalConfig) {
	cm.animalConfigs[entity] = config

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasAnimalConfig[index] |= 1 << bit
}

// GetAnimalConfig возвращает компонент AnimalConfig сущности
func (cm *ComponentManager) GetAnimalConfig(entity EntityID) (AnimalConfig, bool) {
	if !cm.HasComponent(entity, MaskAnimalConfig) {
		return AnimalConfig{}, false
	}
	return cm.animalConfigs[entity], true
}

// SetAnimalConfig обновляет компонент AnimalConfig сущности
func (cm *ComponentManager) SetAnimalConfig(entity EntityID, config AnimalConfig) bool {
	if !cm.HasComponent(entity, MaskAnimalConfig) {
		return false
	}
	cm.animalConfigs[entity] = config
	return true
}

// RemoveAnimalConfig удаляет компонент AnimalConfig у сущности
func (cm *ComponentManager) RemoveAnimalConfig(entity EntityID) bool {
	if !cm.HasComponent(entity, MaskAnimalConfig) {
		return false
	}

	index := uint(entity) / constants.BitsPerUint64
	bit := uint(entity) % constants.BitsPerUint64
	cm.hasAnimalConfig[index] &= ^(1 << bit)
	cm.animalConfigs[entity] = AnimalConfig{}

	return true
}
