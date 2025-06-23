package core

// Файл с методами делегирования World к специализированным менеджерам
// Применяем паттерн Facade для скрытия сложности композиции

// ===== ДЕЛЕГИРОВАНИЕ К COMPONENT MANAGER =====

// HasComponent проверяет наличие компонента у сущности
func (w *World) HasComponent(entity EntityID, component ComponentMask) bool {
	return w.componentManager.HasComponent(entity, component)
}

// HasComponents проверяет наличие всех указанных компонентов у сущности
func (w *World) HasComponents(entity EntityID, mask ComponentMask) bool {
	return w.componentManager.HasComponents(entity, mask)
}

// Position component delegation
func (w *World) AddPosition(entity EntityID, position Position) bool {
	w.componentManager.AddPosition(entity, position)
	// При создании новой сущности автоматически добавляем в пространственную систему
	// (это логично, так как новая позиция должна быть известна пространственной системе)
	w.updateSpatialEntity(entity, position.X, position.Y)
	return true
}

func (w *World) GetPosition(entity EntityID) (Position, bool) {
	return w.componentManager.GetPosition(entity)
}

func (w *World) SetPosition(entity EntityID, position Position) bool {
	// ИСПРАВЛЕНИЕ: Убираем побочный эффект - только обновляем компонент
	// Пространственная система должна обновляется явно через UpdateSpatialPosition
	return w.componentManager.SetPosition(entity, position)
}

// UpdateSpatialPosition обновляет позицию сущности в пространственной системе
// Должен вызываться системами после изменения Position компонента
func (w *World) UpdateSpatialPosition(entity EntityID, position Position) {
	w.updateSpatialEntity(entity, position.X, position.Y)
}

func (w *World) RemovePosition(entity EntityID) bool {
	if w.componentManager.RemovePosition(entity) {
		// Удаляем из пространственной системы
		w.removeSpatialEntity(entity)
		return true
	}
	return false
}

// Velocity component delegation
func (w *World) AddVelocity(entity EntityID, velocity Velocity) bool {
	w.componentManager.AddVelocity(entity, velocity)
	return true
}

func (w *World) GetVelocity(entity EntityID) (Velocity, bool) {
	return w.componentManager.GetVelocity(entity)
}

func (w *World) SetVelocity(entity EntityID, velocity Velocity) bool {
	return w.componentManager.SetVelocity(entity, velocity)
}

func (w *World) RemoveVelocity(entity EntityID) bool {
	return w.componentManager.RemoveVelocity(entity)
}

// Health component delegation
func (w *World) AddHealth(entity EntityID, health Health) bool {
	w.componentManager.AddHealth(entity, health)
	return true
}

func (w *World) GetHealth(entity EntityID) (Health, bool) {
	return w.componentManager.GetHealth(entity)
}

func (w *World) SetHealth(entity EntityID, health Health) bool {
	return w.componentManager.SetHealth(entity, health)
}

func (w *World) RemoveHealth(entity EntityID) bool {
	return w.componentManager.RemoveHealth(entity)
}

// Hunger component delegation
func (w *World) AddHunger(entity EntityID, hunger Hunger) bool {
	w.componentManager.AddHunger(entity, hunger)
	return true
}

func (w *World) GetHunger(entity EntityID) (Hunger, bool) {
	return w.componentManager.GetHunger(entity)
}

func (w *World) SetHunger(entity EntityID, hunger Hunger) bool {
	return w.componentManager.SetHunger(entity, hunger)
}

func (w *World) RemoveHunger(entity EntityID) bool {
	return w.componentManager.RemoveHunger(entity)
}

// AnimalType component delegation
func (w *World) AddAnimalType(entity EntityID, animalType AnimalType) bool {
	w.componentManager.AddAnimalType(entity, animalType)
	return true
}

func (w *World) GetAnimalType(entity EntityID) (AnimalType, bool) {
	return w.componentManager.GetAnimalType(entity)
}

func (w *World) SetAnimalType(entity EntityID, animalType AnimalType) bool {
	return w.componentManager.SetAnimalType(entity, animalType)
}

func (w *World) RemoveAnimalType(entity EntityID) bool {
	return w.componentManager.RemoveAnimalType(entity)
}

// Size component delegation
func (w *World) AddSize(entity EntityID, size Size) bool {
	w.componentManager.AddSize(entity, size)
	return true
}

func (w *World) GetSize(entity EntityID) (Size, bool) {
	return w.componentManager.GetSize(entity)
}

func (w *World) SetSize(entity EntityID, size Size) bool {
	return w.componentManager.SetSize(entity, size)
}

func (w *World) RemoveSize(entity EntityID) bool {
	return w.componentManager.RemoveSize(entity)
}

// Speed component delegation
func (w *World) AddSpeed(entity EntityID, speed Speed) bool {
	w.componentManager.AddSpeed(entity, speed)
	return true
}

func (w *World) GetSpeed(entity EntityID) (Speed, bool) {
	return w.componentManager.GetSpeed(entity)
}

func (w *World) SetSpeed(entity EntityID, speed Speed) bool {
	return w.componentManager.SetSpeed(entity, speed)
}

func (w *World) RemoveSpeed(entity EntityID) bool {
	return w.componentManager.RemoveSpeed(entity)
}

// Animation component delegation
func (w *World) AddAnimation(entity EntityID, animation Animation) bool {
	w.componentManager.AddAnimation(entity, animation)
	return true
}

func (w *World) GetAnimation(entity EntityID) (Animation, bool) {
	return w.componentManager.GetAnimation(entity)
}

func (w *World) SetAnimation(entity EntityID, animation Animation) bool {
	return w.componentManager.SetAnimation(entity, animation)
}

func (w *World) RemoveAnimation(entity EntityID) bool {
	return w.componentManager.RemoveAnimation(entity)
}

// DamageFlash component delegation
func (w *World) AddDamageFlash(entity EntityID, damageFlash DamageFlash) bool {
	w.componentManager.AddDamageFlash(entity, damageFlash)
	return true // Add методы в ComponentManager всегда успешны
}

func (w *World) GetDamageFlash(entity EntityID) (DamageFlash, bool) {
	return w.componentManager.GetDamageFlash(entity)
}

func (w *World) SetDamageFlash(entity EntityID, damageFlash DamageFlash) bool {
	return w.componentManager.SetDamageFlash(entity, damageFlash)
}

func (w *World) RemoveDamageFlash(entity EntityID) bool {
	return w.componentManager.RemoveDamageFlash(entity)
}

// Corpse component delegation
func (w *World) AddCorpse(entity EntityID, corpse Corpse) bool {
	w.componentManager.AddCorpse(entity, corpse)
	return true
}

func (w *World) GetCorpse(entity EntityID) (Corpse, bool) {
	return w.componentManager.GetCorpse(entity)
}

func (w *World) SetCorpse(entity EntityID, corpse Corpse) bool {
	return w.componentManager.SetCorpse(entity, corpse)
}

func (w *World) RemoveCorpse(entity EntityID) bool {
	return w.componentManager.RemoveCorpse(entity)
}

// Carrion component delegation
func (w *World) AddCarrion(entity EntityID, carrion Carrion) bool {
	w.componentManager.AddCarrion(entity, carrion)
	return true
}

func (w *World) GetCarrion(entity EntityID) (Carrion, bool) {
	return w.componentManager.GetCarrion(entity)
}

func (w *World) SetCarrion(entity EntityID, carrion Carrion) bool {
	return w.componentManager.SetCarrion(entity, carrion)
}

func (w *World) RemoveCarrion(entity EntityID) bool {
	return w.componentManager.RemoveCarrion(entity)
}

// EatingState component delegation
func (w *World) AddEatingState(entity EntityID, eatingState EatingState) bool {
	w.componentManager.AddEatingState(entity, eatingState)
	return true
}

func (w *World) GetEatingState(entity EntityID) (EatingState, bool) {
	return w.componentManager.GetEatingState(entity)
}

func (w *World) SetEatingState(entity EntityID, eatingState EatingState) bool {
	return w.componentManager.SetEatingState(entity, eatingState)
}

func (w *World) RemoveEatingState(entity EntityID) bool {
	return w.componentManager.RemoveEatingState(entity)
}

// AttackState component delegation
func (w *World) AddAttackState(entity EntityID, attackState AttackState) bool {
	w.componentManager.AddAttackState(entity, attackState)
	return true
}

func (w *World) GetAttackState(entity EntityID) (AttackState, bool) {
	return w.componentManager.GetAttackState(entity)
}

func (w *World) SetAttackState(entity EntityID, attackState AttackState) bool {
	return w.componentManager.SetAttackState(entity, attackState)
}

func (w *World) RemoveAttackState(entity EntityID) bool {
	return w.componentManager.RemoveAttackState(entity)
}

// Behavior component delegation
func (w *World) AddBehavior(entity EntityID, behavior Behavior) bool {
	w.componentManager.AddBehavior(entity, behavior)
	return true
}

func (w *World) GetBehavior(entity EntityID) (Behavior, bool) {
	return w.componentManager.GetBehavior(entity)
}

func (w *World) SetBehavior(entity EntityID, behavior Behavior) bool {
	return w.componentManager.SetBehavior(entity, behavior)
}

func (w *World) RemoveBehavior(entity EntityID) bool {
	return w.componentManager.RemoveBehavior(entity)
}

// AnimalConfig component delegation
func (w *World) AddAnimalConfig(entity EntityID, config AnimalConfig) bool {
	w.componentManager.AddAnimalConfig(entity, config)
	return true
}

func (w *World) GetAnimalConfig(entity EntityID) (AnimalConfig, bool) {
	return w.componentManager.GetAnimalConfig(entity)
}

func (w *World) SetAnimalConfig(entity EntityID, config AnimalConfig) bool {
	return w.componentManager.SetAnimalConfig(entity, config)
}

func (w *World) RemoveAnimalConfig(entity EntityID) bool {
	return w.componentManager.RemoveAnimalConfig(entity)
}

// ===== ДЕЛЕГИРОВАНИЕ К QUERY MANAGER =====

// ForEach вызывает функцию для каждой активной сущности
func (w *World) ForEach(fn func(EntityID)) {
	w.queryManager.ForEach(fn)
}

// ForEachWith вызывает функцию для каждой сущности с указанными компонентами
func (w *World) ForEachWith(mask ComponentMask, fn QueryFunc) {
	w.queryManager.ForEachWith(mask, fn)
}

// GetEntitiesWith возвращает слайс сущностей с указанными компонентами
func (w *World) GetEntitiesWith(mask ComponentMask) []EntityID {
	return w.queryManager.GetEntitiesWith(mask)
}

// CountEntitiesWith подсчитывает количество сущностей с указанными компонентами
func (w *World) CountEntitiesWith(mask ComponentMask) int {
	return w.queryManager.CountEntitiesWith(mask)
}

// FindFirst находит первую сущность с указанными компонентами
func (w *World) FindFirst(mask ComponentMask) (EntityID, bool) {
	return w.queryManager.FindFirst(mask)
}

// ForEachWithBreak вызывает функцию для каждой сущности с указанными компонентами
func (w *World) ForEachWithBreak(mask ComponentMask, fn func(EntityID) bool) {
	w.queryManager.ForEachWithBreak(mask, fn)
}

// FilterEntities фильтрует сущности по пользовательскому предикату
func (w *World) FilterEntities(mask ComponentMask, predicate func(EntityID) bool) []EntityID {
	return w.queryManager.FilterEntities(mask, predicate)
}

// ===== ДОПОЛНИТЕЛЬНЫЕ МЕТОДЫ ДЛЯ СОВМЕСТИМОСТИ С ИНТЕРФЕЙСАМИ =====

// FindNearestByType находит ближайшую сущность указанного типа в радиусе
// ВХОДНЫЕ ПАРАМЕТРЫ: x,y в пикселях (позиции животных), radius в пикселях
// ВНУТРЕННЯЯ ЛОГИКА: все расчеты в пикселях для совместимости с системой позиций
func (w *World) FindNearestByType(x, y, radius float32, animalType AnimalType) (EntityID, bool) {
	var nearestEntity EntityID
	nearestDistance := radius * radius // Используем квадрат расстояния для быстрого сравнения
	found := false

	// Ищем среди всех сущностей с позицией и типом
	w.queryManager.ForEachWith(MaskPosition|MaskAnimalType, func(entity EntityID) {
		// Получаем позицию и тип
		if pos, hasPos := w.componentManager.GetPosition(entity); hasPos {
			if entityType, hasType := w.componentManager.GetAnimalType(entity); hasType {
				if entityType == animalType {
					// Вычисляем квадрат расстояния
					dx := pos.X - x
					dy := pos.Y - y
					distanceSquared := dx*dx + dy*dy

					if distanceSquared < nearestDistance {
						nearestDistance = distanceSquared
						nearestEntity = entity
						found = true
					}
				}
			}
		}
	})

	return nearestEntity, found
}

// FindNearestByTypeInTiles находит ближайшую сущность указанного типа в радиусе (в тайлах)
// РЕФАКТОРИНГ: новая функция для работы с тайлами как базовой единицей измерения
// ВХОДНЫЕ ПАРАМЕТРЫ: x,y в пикселях (позиции животных), radiusInTiles в тайлах
// ВНУТРЕННЯЯ ЛОГИКА: конвертирует тайлы в пиксели и вызывает стандартную функцию
func (w *World) FindNearestByTypeInTiles(x, y, radiusInTiles float32, animalType AnimalType) (EntityID, bool) {
	// Импортируем константы для конвертации
	const TileSizePixels = 32.0 // Временно используем константу прямо здесь
	radiusInPixels := radiusInTiles * TileSizePixels
	return w.FindNearestByType(x, y, radiusInPixels, animalType)
}

// QueryInRadius возвращает сущности в указанном радиусе (для SpatialQueries интерфейса)
func (w *World) QueryInRadius(x, y, radius float32) []EntityID {
	// Используем метод-фасад для скрытия сложности пространственной системы (LoD)
	return w.querySpatialRadius(x, y, radius)
}

// GetStats возвращает статистику животных по типам
func (w *World) GetStats() map[AnimalType]int {
	stats := make(map[AnimalType]int)

	// Подсчитываем животных по типам
	w.queryManager.ForEachWith(MaskAnimalType, func(entity EntityID) {
		if animalType, ok := w.componentManager.GetAnimalType(entity); ok {
			stats[animalType]++
		}
	})

	return stats
}

// ===== LEGACY МЕТОДЫ ДЛЯ СОВМЕСТИМОСТИ С ТЕСТАМИ =====

// QueryEntitiesWith алиас для GetEntitiesWith (для совместимости с тестами)
func (w *World) QueryEntitiesWith(mask ComponentMask) []EntityID {
	return w.queryManager.GetEntitiesWith(mask)
}

// QueryByType находит сущности по типу животного
func (w *World) QueryByType(animalType AnimalType) []EntityID {
	return w.queryManager.FilterEntities(MaskAnimalType, func(entity EntityID) bool {
		if entityType, ok := w.componentManager.GetAnimalType(entity); ok {
			return entityType == animalType
		}
		return false
	})
}

// FindNearestAnimal находит ближайшее животное любого типа (для совместимости с тестами)
func (w *World) FindNearestAnimal(x, y, radius float32) (EntityID, bool) {
	var nearestEntity EntityID
	nearestDistance := radius * radius // Используем квадрат расстояния для быстрого сравнения
	found := false

	// Ищем среди всех сущностей с позицией и типом
	w.queryManager.ForEachWith(MaskPosition|MaskAnimalType, func(entity EntityID) {
		// Получаем позицию
		if pos, hasPos := w.componentManager.GetPosition(entity); hasPos {
			// Вычисляем квадрат расстояния
			dx := pos.X - x
			dy := pos.Y - y
			distanceSquared := dx*dx + dy*dy

			if distanceSquared < nearestDistance {
				nearestDistance = distanceSquared
				nearestEntity = entity
				found = true
			}
		}
	})

	return nearestEntity, found
}
