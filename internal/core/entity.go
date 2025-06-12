package core

// EntityID уникальный идентификатор сущности
// Используем uint16 для поддержки до 65535 сущностей
type EntityID uint16

// MAX_ENTITIES максимальное количество сущностей в мире
const MAX_ENTITIES = 1000

// INVALID_ENTITY специальное значение для несуществующей сущности
const INVALID_ENTITY EntityID = 0

// EntityManager управляет созданием, удалением и переиспользованием ID сущностей
type EntityManager struct {
	nextID  EntityID           // Следующий доступный ID
	freeIDs []EntityID         // Освобождённые ID для переиспользования
	alive   [MAX_ENTITIES]bool // Битовая маска живых сущностей
	count   int                // Количество живых сущностей
}

// NewEntityManager создаёт новый менеджер сущностей
func NewEntityManager() *EntityManager {
	return &EntityManager{
		nextID:  1,                                   // Начинаем с 1, т.к. 0 это INVALID_ENTITY
		freeIDs: make([]EntityID, 0, MAX_ENTITIES/4), // Предварительно выделяем память
		count:   0,
	}
}

// CreateEntity создаёт новую сущность и возвращает её ID
func (em *EntityManager) CreateEntity() EntityID {
	var id EntityID

	// Если есть освобождённые ID, переиспользуем их
	if len(em.freeIDs) > 0 {
		id = em.freeIDs[len(em.freeIDs)-1]
		em.freeIDs = em.freeIDs[:len(em.freeIDs)-1]
	} else {
		// Иначе используем следующий доступный ID
		if em.nextID >= MAX_ENTITIES {
			return INVALID_ENTITY // Достигли лимита сущностей
		}
		id = em.nextID
		em.nextID++
	}

	em.alive[id] = true
	em.count++
	return id
}

// DestroyEntity уничтожает сущность и освобождает её ID для переиспользования
func (em *EntityManager) DestroyEntity(id EntityID) bool {
	if id == INVALID_ENTITY || id >= MAX_ENTITIES || !em.alive[id] {
		return false // Сущность не существует
	}

	em.alive[id] = false
	em.count--

	// Добавляем ID в список свободных для переиспользования
	em.freeIDs = append(em.freeIDs, id)

	return true
}

// IsAlive проверяет, существует ли сущность
func (em *EntityManager) IsAlive(id EntityID) bool {
	if id == INVALID_ENTITY || id >= MAX_ENTITIES {
		return false
	}
	return em.alive[id]
}

// Count возвращает количество живых сущностей
func (em *EntityManager) Count() int {
	return em.count
}

// GetAliveEntities возвращает слайс всех живых сущностей
// ВАЖНО: результат нужно использовать немедленно, он может быть перезаписан
func (em *EntityManager) GetAliveEntities(buffer []EntityID) []EntityID {
	if buffer == nil {
		buffer = make([]EntityID, 0, em.count)
	}
	buffer = buffer[:0] // Сбрасываем длину, но сохраняем capacity

	for id := EntityID(1); id < EntityID(len(em.alive)) && len(buffer) < em.count; id++ {
		if em.alive[id] {
			buffer = append(buffer, id)
		}
	}

	return buffer
}

// Clear очищает менеджер сущностей (для тестов и перезапуска)
func (em *EntityManager) Clear() {
	em.nextID = 1
	em.freeIDs = em.freeIDs[:0]
	em.count = 0
	for i := range em.alive {
		em.alive[i] = false
	}
}
