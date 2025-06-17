package core

import "github.com/aiseeq/savanna/internal/physics"

// Методы для работы с компонентами в World

// componentAccessor представляет generic интерфейс для доступа к компонентам
// Устраняет дублирование 45+ функций с идентичной логикой валидации
type componentAccessor[T any] struct {
	data    []T           // Массив компонентов
	hasMask []uint64      // Битовая маска
	mask    ComponentMask // Маска компонента
}

// addComponent generic функция для добавления компонента (устраняет дублирование)
func addComponent[T any](w *World, entity EntityID, component T, accessor componentAccessor[T]) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}

	accessor.data[entity] = component
	setBitMask(accessor.hasMask, entity)
	return true
}

// getComponent generic функция для получения компонента (устраняет дублирование)
func getComponent[T any](w *World, entity EntityID, accessor componentAccessor[T]) (T, bool) {
	var zero T
	if !w.entities.IsAlive(entity) || !testBitMask(accessor.hasMask, entity) {
		return zero, false
	}

	return accessor.data[entity], true
}

// setComponent generic функция для установки компонента (устраняет дублирование)
func setComponent[T any](w *World, entity EntityID, component T, accessor componentAccessor[T]) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(accessor.hasMask, entity) {
		return false
	}

	accessor.data[entity] = component
	return true
}

// removeComponent generic функция для удаления компонента (устраняет дублирование)
func removeComponent[T any](w *World, entity EntityID, accessor componentAccessor[T]) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(accessor.hasMask, entity) {
		return false
	}

	clearBitMask(accessor.hasMask, entity)
	return true
}

// Accessor'ы для каждого типа компонента (устраняет дублирование)
// Position и Velocity accessor'ы не используются т.к. эти компоненты
// имеют специальную логику для пространственной сетки

func (w *World) healthAccessor() componentAccessor[Health] {
	return componentAccessor[Health]{
		data:    w.healths[:],
		hasMask: w.hasHealth[:],
		mask:    MaskHealth,
	}
}

func (w *World) hungerAccessor() componentAccessor[Hunger] {
	return componentAccessor[Hunger]{
		data:    w.hungers[:],
		hasMask: w.hasHunger[:],
		mask:    MaskHunger,
	}
}

func (w *World) ageAccessor() componentAccessor[Age] {
	return componentAccessor[Age]{
		data:    w.ages[:],
		hasMask: w.hasAge[:],
		mask:    MaskAge,
	}
}

func (w *World) animalTypeAccessor() componentAccessor[AnimalType] {
	return componentAccessor[AnimalType]{
		data:    w.types[:],
		hasMask: w.hasType[:],
		mask:    MaskAnimalType,
	}
}

func (w *World) sizeAccessor() componentAccessor[Size] {
	return componentAccessor[Size]{
		data:    w.sizes[:],
		hasMask: w.hasSize[:],
		mask:    MaskSize,
	}
}

func (w *World) speedAccessor() componentAccessor[Speed] {
	return componentAccessor[Speed]{
		data:    w.speeds[:],
		hasMask: w.hasSpeed[:],
		mask:    MaskSpeed,
	}
}

func (w *World) animationAccessor() componentAccessor[Animation] {
	return componentAccessor[Animation]{
		data:    w.animations[:],
		hasMask: w.hasAnimation[:],
		mask:    MaskAnimation,
	}
}

func (w *World) damageFlashAccessor() componentAccessor[DamageFlash] {
	return componentAccessor[DamageFlash]{
		data:    w.damageFlashes[:],
		hasMask: w.hasDamageFlash[:],
		mask:    MaskDamageFlash,
	}
}

func (w *World) corpseAccessor() componentAccessor[Corpse] {
	return componentAccessor[Corpse]{
		data:    w.corpses[:],
		hasMask: w.hasCorpse[:],
		mask:    MaskCorpse,
	}
}

func (w *World) carrionAccessor() componentAccessor[Carrion] {
	return componentAccessor[Carrion]{
		data:    w.carrions[:],
		hasMask: w.hasCarrion[:],
		mask:    MaskCarrion,
	}
}

func (w *World) eatingStateAccessor() componentAccessor[EatingState] {
	return componentAccessor[EatingState]{
		data:    w.eatingStates[:],
		hasMask: w.hasEatingState[:],
		mask:    MaskEatingState,
	}
}

func (w *World) attackStateAccessor() componentAccessor[AttackState] {
	return componentAccessor[AttackState]{
		data:    w.attackStates[:],
		hasMask: w.hasAttackState[:],
		mask:    MaskAttackState,
	}
}

// HasComponent проверяет наличие компонента у сущности
func (w *World) HasComponent(entity EntityID, component ComponentMask) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}

	switch component {
	case MaskPosition:
		return testBitMask(w.hasPosition[:], entity)
	case MaskVelocity:
		return testBitMask(w.hasVelocity[:], entity)
	case MaskHealth:
		return testBitMask(w.hasHealth[:], entity)
	case MaskHunger:
		return testBitMask(w.hasHunger[:], entity)
	case MaskAge:
		return testBitMask(w.hasAge[:], entity)
	case MaskAnimalType:
		return testBitMask(w.hasType[:], entity)
	case MaskSize:
		return testBitMask(w.hasSize[:], entity)
	case MaskSpeed:
		return testBitMask(w.hasSpeed[:], entity)
	case MaskAnimation:
		return testBitMask(w.hasAnimation[:], entity)
	case MaskDamageFlash:
		return testBitMask(w.hasDamageFlash[:], entity)
	case MaskCorpse:
		return testBitMask(w.hasCorpse[:], entity)
	case MaskCarrion:
		return testBitMask(w.hasCarrion[:], entity)
	case MaskEatingState:
		return testBitMask(w.hasEatingState[:], entity)
	case MaskAttackState:
		return testBitMask(w.hasAttackState[:], entity)
	case MaskBehavior:
		return testBitMask(w.hasBehavior[:], entity)
	case MaskAnimalConfig:
		return testBitMask(w.hasAnimalConfig[:], entity)
	}
	return false
}

// HasComponents проверяет наличие всех указанных компонентов
func (w *World) HasComponents(entity EntityID, mask ComponentMask) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}

	// Lookup таблица для проверки компонентов (устраняет дублирование кода)
	componentChecks := []struct {
		mask    ComponentMask
		hasMask []uint64
	}{
		{MaskPosition, w.hasPosition[:]},
		{MaskVelocity, w.hasVelocity[:]},
		{MaskHealth, w.hasHealth[:]},
		{MaskHunger, w.hasHunger[:]},
		{MaskAge, w.hasAge[:]},
		{MaskAnimalType, w.hasType[:]},
		{MaskSize, w.hasSize[:]},
		{MaskSpeed, w.hasSpeed[:]},
		{MaskAnimation, w.hasAnimation[:]},
		{MaskDamageFlash, w.hasDamageFlash[:]},
		{MaskCorpse, w.hasCorpse[:]},
		{MaskCarrion, w.hasCarrion[:]},
		{MaskEatingState, w.hasEatingState[:]},
		{MaskAttackState, w.hasAttackState[:]},
		{MaskBehavior, w.hasBehavior[:]},
		{MaskAnimalConfig, w.hasAnimalConfig[:]},
	}

	// Проверяем каждый компонент через цикл
	for _, check := range componentChecks {
		if mask&check.mask != 0 && !testBitMask(check.hasMask, entity) {
			return false
		}
	}

	return true
}

// Position Component

// AddPosition добавляет компонент Position
func (w *World) AddPosition(entity EntityID, position Position) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}

	w.positions[entity] = position
	setBitMask(w.hasPosition[:], entity)

	// Обновляем пространственную сетку если есть размер
	if w.HasComponent(entity, MaskSize) {
		size := w.sizes[entity]
		w.spatialProvider.UpdateEntity(uint32(entity),
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}

	return true
}

// GetPosition получает компонент Position
func (w *World) GetPosition(entity EntityID) (Position, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasPosition[:], entity) {
		return Position{}, false
	}
	return w.positions[entity], true
}

// SetPosition изменяет позицию сущности
func (w *World) SetPosition(entity EntityID, position Position) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasPosition[:], entity) {
		return false
	}

	w.positions[entity] = position

	// Обновляем пространственную сетку если есть размер
	if w.HasComponent(entity, MaskSize) {
		size := w.sizes[entity]
		w.spatialProvider.UpdateEntity(uint32(entity),
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}

	return true
}

// RemovePosition удаляет компонент Position
func (w *World) RemovePosition(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasPosition[:], entity) {
		return false
	}

	clearBitMask(w.hasPosition[:], entity)
	w.spatialProvider.RemoveEntity(uint32(entity))
	return true
}

// Velocity Component

// AddVelocity добавляет компонент Velocity
func (w *World) AddVelocity(entity EntityID, velocity Velocity) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}

	w.velocities[entity] = velocity
	setBitMask(w.hasVelocity[:], entity)
	return true
}

// GetVelocity получает компонент Velocity
func (w *World) GetVelocity(entity EntityID) (Velocity, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasVelocity[:], entity) {
		return Velocity{}, false
	}
	return w.velocities[entity], true
}

// SetVelocity изменяет скорость сущности
func (w *World) SetVelocity(entity EntityID, velocity Velocity) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasVelocity[:], entity) {
		return false
	}

	w.velocities[entity] = velocity
	return true
}

// RemoveVelocity удаляет компонент Velocity
func (w *World) RemoveVelocity(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasVelocity[:], entity) {
		return false
	}

	clearBitMask(w.hasVelocity[:], entity)
	return true
}

// Health Component

// AddHealth добавляет компонент Health
func (w *World) AddHealth(entity EntityID, health Health) bool {
	return addComponent(w, entity, health, w.healthAccessor())
}

// GetHealth получает компонент Health
func (w *World) GetHealth(entity EntityID) (Health, bool) {
	return getComponent(w, entity, w.healthAccessor())
}

// SetHealth изменяет здоровье сущности
func (w *World) SetHealth(entity EntityID, health Health) bool {
	return setComponent(w, entity, health, w.healthAccessor())
}

// RemoveHealth удаляет компонент Health
func (w *World) RemoveHealth(entity EntityID) bool {
	return removeComponent(w, entity, w.healthAccessor())
}

// Hunger Component

// AddHunger добавляет компонент Hunger
func (w *World) AddHunger(entity EntityID, hunger Hunger) bool {
	return addComponent(w, entity, hunger, w.hungerAccessor())
}

// GetHunger получает компонент Hunger
func (w *World) GetHunger(entity EntityID) (Hunger, bool) {
	return getComponent(w, entity, w.hungerAccessor())
}

// SetHunger изменяет голод сущности
func (w *World) SetHunger(entity EntityID, hunger Hunger) bool {
	return setComponent(w, entity, hunger, w.hungerAccessor())
}

// RemoveHunger удаляет компонент Hunger
func (w *World) RemoveHunger(entity EntityID) bool {
	return removeComponent(w, entity, w.hungerAccessor())
}

// Age Component

// AddAge добавляет компонент Age
func (w *World) AddAge(entity EntityID, age Age) bool {
	return addComponent(w, entity, age, w.ageAccessor())
}

// GetAge получает компонент Age
func (w *World) GetAge(entity EntityID) (Age, bool) {
	return getComponent(w, entity, w.ageAccessor())
}

// SetAge изменяет возраст сущности
func (w *World) SetAge(entity EntityID, age Age) bool {
	return setComponent(w, entity, age, w.ageAccessor())
}

// RemoveAge удаляет компонент Age
func (w *World) RemoveAge(entity EntityID) bool {
	return removeComponent(w, entity, w.ageAccessor())
}

// AnimalType Component

// AddAnimalType добавляет компонент AnimalType
func (w *World) AddAnimalType(entity EntityID, animalType AnimalType) bool {
	return addComponent(w, entity, animalType, w.animalTypeAccessor())
}

// GetAnimalType получает компонент AnimalType
func (w *World) GetAnimalType(entity EntityID) (AnimalType, bool) {
	return getComponent(w, entity, w.animalTypeAccessor())
}

// SetAnimalType изменяет тип животного сущности
func (w *World) SetAnimalType(entity EntityID, animalType AnimalType) bool {
	return setComponent(w, entity, animalType, w.animalTypeAccessor())
}

// RemoveAnimalType удаляет компонент AnimalType
func (w *World) RemoveAnimalType(entity EntityID) bool {
	return removeComponent(w, entity, w.animalTypeAccessor())
}

// Size Component

// AddSize добавляет компонент Size (с обновлением пространственной сетки)
func (w *World) AddSize(entity EntityID, size Size) bool {
	if !addComponent(w, entity, size, w.sizeAccessor()) {
		return false
	}

	// Обновляем пространственную систему если есть позиция
	if w.HasComponent(entity, MaskPosition) {
		position := w.positions[entity]
		w.spatialProvider.UpdateEntity(uint32(entity),
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}

	return true
}

// GetSize получает компонент Size
func (w *World) GetSize(entity EntityID) (Size, bool) {
	return getComponent(w, entity, w.sizeAccessor())
}

// SetSize изменяет размер сущности (с обновлением пространственной сетки)
func (w *World) SetSize(entity EntityID, size Size) bool {
	if !setComponent(w, entity, size, w.sizeAccessor()) {
		return false
	}

	// Обновляем пространственную систему если есть позиция
	if w.HasComponent(entity, MaskPosition) {
		position := w.positions[entity]
		w.spatialProvider.UpdateEntity(uint32(entity),
			physics.Vec2{X: position.X, Y: position.Y}, size.Radius)
	}

	return true
}

// RemoveSize удаляет компонент Size
func (w *World) RemoveSize(entity EntityID) bool {
	return removeComponent(w, entity, w.sizeAccessor())
}

// Speed Component

// AddSpeed добавляет компонент Speed
func (w *World) AddSpeed(entity EntityID, speed Speed) bool {
	return addComponent(w, entity, speed, w.speedAccessor())
}

// GetSpeed получает компонент Speed
func (w *World) GetSpeed(entity EntityID) (Speed, bool) {
	return getComponent(w, entity, w.speedAccessor())
}

// SetSpeed изменяет скорость сущности
func (w *World) SetSpeed(entity EntityID, speed Speed) bool {
	return setComponent(w, entity, speed, w.speedAccessor())
}

// RemoveSpeed удаляет компонент Speed
func (w *World) RemoveSpeed(entity EntityID) bool {
	return removeComponent(w, entity, w.speedAccessor())
}

// Animation Component

// AddAnimation добавляет компонент Animation
func (w *World) AddAnimation(entity EntityID, animation Animation) bool {
	return addComponent(w, entity, animation, w.animationAccessor())
}

// GetAnimation получает компонент Animation
func (w *World) GetAnimation(entity EntityID) (Animation, bool) {
	return getComponent(w, entity, w.animationAccessor())
}

// SetAnimation изменяет анимацию сущности
func (w *World) SetAnimation(entity EntityID, animation Animation) bool {
	return setComponent(w, entity, animation, w.animationAccessor())
}

// RemoveAnimation удаляет компонент Animation
func (w *World) RemoveAnimation(entity EntityID) bool {
	return removeComponent(w, entity, w.animationAccessor())
}

// DamageFlash Component

// AddDamageFlash добавляет компонент DamageFlash
func (w *World) AddDamageFlash(entity EntityID, damageFlash DamageFlash) bool {
	return addComponent(w, entity, damageFlash, w.damageFlashAccessor())
}

// GetDamageFlash получает компонент DamageFlash
func (w *World) GetDamageFlash(entity EntityID) (DamageFlash, bool) {
	return getComponent(w, entity, w.damageFlashAccessor())
}

// SetDamageFlash изменяет DamageFlash сущности
func (w *World) SetDamageFlash(entity EntityID, damageFlash DamageFlash) bool {
	return setComponent(w, entity, damageFlash, w.damageFlashAccessor())
}

// RemoveDamageFlash удаляет компонент DamageFlash
func (w *World) RemoveDamageFlash(entity EntityID) bool {
	return removeComponent(w, entity, w.damageFlashAccessor())
}

// Corpse Component

// AddCorpse добавляет компонент Corpse
func (w *World) AddCorpse(entity EntityID, corpse Corpse) bool {
	return addComponent(w, entity, corpse, w.corpseAccessor())
}

// GetCorpse получает компонент Corpse
func (w *World) GetCorpse(entity EntityID) (Corpse, bool) {
	return getComponent(w, entity, w.corpseAccessor())
}

// SetCorpse изменяет Corpse сущности
func (w *World) SetCorpse(entity EntityID, corpse Corpse) bool {
	return setComponent(w, entity, corpse, w.corpseAccessor())
}

// RemoveCorpse удаляет компонент Corpse
func (w *World) RemoveCorpse(entity EntityID) bool {
	return removeComponent(w, entity, w.corpseAccessor())
}

// Carrion Component

// AddCarrion добавляет компонент Carrion
func (w *World) AddCarrion(entity EntityID, carrion Carrion) bool {
	return addComponent(w, entity, carrion, w.carrionAccessor())
}

// GetCarrion получает компонент Carrion
func (w *World) GetCarrion(entity EntityID) (Carrion, bool) {
	return getComponent(w, entity, w.carrionAccessor())
}

// SetCarrion изменяет Carrion сущности
func (w *World) SetCarrion(entity EntityID, carrion Carrion) bool {
	return setComponent(w, entity, carrion, w.carrionAccessor())
}

// RemoveCarrion удаляет компонент Carrion
func (w *World) RemoveCarrion(entity EntityID) bool {
	return removeComponent(w, entity, w.carrionAccessor())
}

// EatingState Component

// AddEatingState добавляет компонент EatingState
func (w *World) AddEatingState(entity EntityID, eatingState EatingState) bool {
	return addComponent(w, entity, eatingState, w.eatingStateAccessor())
}

// GetEatingState получает компонент EatingState
func (w *World) GetEatingState(entity EntityID) (EatingState, bool) {
	return getComponent(w, entity, w.eatingStateAccessor())
}

// SetEatingState изменяет EatingState сущности
func (w *World) SetEatingState(entity EntityID, eatingState EatingState) bool {
	return setComponent(w, entity, eatingState, w.eatingStateAccessor())
}

// RemoveEatingState удаляет компонент EatingState
func (w *World) RemoveEatingState(entity EntityID) bool {
	return removeComponent(w, entity, w.eatingStateAccessor())
}

// AttackState Component

// AddAttackState добавляет компонент AttackState
func (w *World) AddAttackState(entity EntityID, attackState AttackState) bool {
	return addComponent(w, entity, attackState, w.attackStateAccessor())
}

// GetAttackState получает компонент AttackState
func (w *World) GetAttackState(entity EntityID) (AttackState, bool) {
	return getComponent(w, entity, w.attackStateAccessor())
}

// SetAttackState изменяет AttackState сущности
func (w *World) SetAttackState(entity EntityID, attackState AttackState) bool {
	return setComponent(w, entity, attackState, w.attackStateAccessor())
}

// RemoveAttackState удаляет компонент AttackState
func (w *World) RemoveAttackState(entity EntityID) bool {
	return removeComponent(w, entity, w.attackStateAccessor())
}

// Behavior Component

// AddBehavior добавляет компонент Behavior
func (w *World) AddBehavior(entity EntityID, behavior Behavior) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}

	setBitMask(w.hasBehavior[:], entity)
	w.behaviors[entity] = behavior
	return true
}

// GetBehavior получает компонент Behavior
func (w *World) GetBehavior(entity EntityID) (Behavior, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasBehavior[:], entity) {
		return Behavior{}, false
	}

	return w.behaviors[entity], true
}

// SetBehavior устанавливает компонент Behavior
func (w *World) SetBehavior(entity EntityID, behavior Behavior) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasBehavior[:], entity) {
		return false
	}

	w.behaviors[entity] = behavior
	return true
}

// RemoveBehavior удаляет компонент Behavior
func (w *World) RemoveBehavior(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasBehavior[:], entity) {
		return false
	}

	clearBitMask(w.hasBehavior[:], entity)
	return true
}

// AnimalConfig Component

// AddAnimalConfig добавляет компонент AnimalConfig
func (w *World) AddAnimalConfig(entity EntityID, config AnimalConfig) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}

	setBitMask(w.hasAnimalConfig[:], entity)
	w.animalConfigs[entity] = config
	return true
}

// GetAnimalConfig получает компонент AnimalConfig
func (w *World) GetAnimalConfig(entity EntityID) (AnimalConfig, bool) {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasAnimalConfig[:], entity) {
		return AnimalConfig{}, false
	}

	return w.animalConfigs[entity], true
}

// SetAnimalConfig устанавливает компонент AnimalConfig
func (w *World) SetAnimalConfig(entity EntityID, config AnimalConfig) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasAnimalConfig[:], entity) {
		return false
	}

	w.animalConfigs[entity] = config
	return true
}

// RemoveAnimalConfig удаляет компонент AnimalConfig
func (w *World) RemoveAnimalConfig(entity EntityID) bool {
	if !w.entities.IsAlive(entity) || !testBitMask(w.hasAnimalConfig[:], entity) {
		return false
	}

	clearBitMask(w.hasAnimalConfig[:], entity)
	return true
}
