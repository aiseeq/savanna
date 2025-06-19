package core

// QueryManager управляет запросами к сущностям
// Соблюдает Single Responsibility Principle - только запросы и итерации
type QueryManager struct {
	componentManager *ComponentManager
	entityManager    *EntityManager

	// Буферы для переиспользования (предотвращение аллокаций)
	queryBuffer []EntityID
}

// NewQueryManager создаёт новый менеджер запросов
func NewQueryManager(componentManager *ComponentManager, entityManager *EntityManager) *QueryManager {
	return &QueryManager{
		componentManager: componentManager,
		entityManager:    entityManager,
		queryBuffer:      make([]EntityID, 0, 100),
	}
}

// ForEach вызывает функцию для каждой активной сущности
func (qm *QueryManager) ForEach(fn func(EntityID)) {
	for entity := EntityID(0); entity < EntityID(qm.entityManager.nextID); entity++ {
		if qm.entityManager.IsAlive(entity) {
			fn(entity)
		}
	}
}

// ForEachWith вызывает функцию для каждой сущности с указанными компонентами
func (qm *QueryManager) ForEachWith(mask ComponentMask, fn QueryFunc) {
	for entity := EntityID(0); entity < EntityID(qm.entityManager.nextID); entity++ {
		if qm.entityManager.IsAlive(entity) && qm.componentManager.HasComponents(entity, mask) {
			fn(entity)
		}
	}
}

// GetEntitiesWith возвращает слайс сущностей с указанными компонентами
func (qm *QueryManager) GetEntitiesWith(mask ComponentMask) []EntityID {
	// Очищаем буфер для переиспользования
	qm.queryBuffer = qm.queryBuffer[:0]

	for entity := EntityID(0); entity < EntityID(qm.entityManager.nextID); entity++ {
		if qm.entityManager.IsAlive(entity) && qm.componentManager.HasComponents(entity, mask) {
			qm.queryBuffer = append(qm.queryBuffer, entity)
		}
	}

	// Возвращаем копию чтобы избежать проблем при изменении буфера
	result := make([]EntityID, len(qm.queryBuffer))
	copy(result, qm.queryBuffer)
	return result
}

// CountEntitiesWith подсчитывает количество сущностей с указанными компонентами
func (qm *QueryManager) CountEntitiesWith(mask ComponentMask) int {
	count := 0
	for entity := EntityID(0); entity < EntityID(qm.entityManager.nextID); entity++ {
		if qm.entityManager.IsAlive(entity) && qm.componentManager.HasComponents(entity, mask) {
			count++
		}
	}
	return count
}

// FindFirst находит первую сущность с указанными компонентами
func (qm *QueryManager) FindFirst(mask ComponentMask) (EntityID, bool) {
	for entity := EntityID(0); entity < EntityID(qm.entityManager.nextID); entity++ {
		if qm.entityManager.IsAlive(entity) && qm.componentManager.HasComponents(entity, mask) {
			return entity, true
		}
	}
	return EntityID(0), false
}

// ForEachWithBreak вызывает функцию для каждой сущности с указанными компонентами
// Функция может вернуть false для прерывания итерации
func (qm *QueryManager) ForEachWithBreak(mask ComponentMask, fn func(EntityID) bool) {
	for entity := EntityID(0); entity < EntityID(qm.entityManager.nextID); entity++ {
		if qm.entityManager.IsAlive(entity) && qm.componentManager.HasComponents(entity, mask) {
			if !fn(entity) {
				break
			}
		}
	}
}

// FilterEntities фильтрует сущности по пользовательскому предикату
func (qm *QueryManager) FilterEntities(mask ComponentMask, predicate func(EntityID) bool) []EntityID {
	// Очищаем буфер для переиспользования
	qm.queryBuffer = qm.queryBuffer[:0]

	for entity := EntityID(0); entity < EntityID(qm.entityManager.nextID); entity++ {
		if qm.entityManager.IsAlive(entity) && qm.componentManager.HasComponents(entity, mask) {
			if predicate(entity) {
				qm.queryBuffer = append(qm.queryBuffer, entity)
			}
		}
	}

	// Возвращаем копию чтобы избежать проблем при изменении буфера
	result := make([]EntityID, len(qm.queryBuffer))
	copy(result, qm.queryBuffer)
	return result
}

// GetEntityCount возвращает общее количество активных сущностей
func (qm *QueryManager) GetEntityCount() int {
	return qm.entityManager.Count()
}
