package core

// Системы запросов и итераторов для эффективной работы с сущностями

// QueryFunc функция обратного вызова для итераторов
type QueryFunc func(EntityID)

// ForEachWith итерирует по всем сущностям с указанными компонентами
func (w *World) ForEachWith(mask ComponentMask, fn QueryFunc) {
	w.entitiesBuffer = w.entities.GetAliveEntities(w.entitiesBuffer)
	
	for _, entity := range w.entitiesBuffer {
		if w.HasComponents(entity, mask) {
			fn(entity)
		}
	}
}

// QueryEntitiesWith возвращает слайс сущностей с указанными компонентами
func (w *World) QueryEntitiesWith(mask ComponentMask) []EntityID {
	w.queryBuffer = w.queryBuffer[:0] // Сбрасываем длину
	
	w.entitiesBuffer = w.entities.GetAliveEntities(w.entitiesBuffer)
	for _, entity := range w.entitiesBuffer {
		if w.HasComponents(entity, mask) {
			w.queryBuffer = append(w.queryBuffer, entity)
		}
	}
	
	return w.queryBuffer
}

// CountEntitiesWith подсчитывает количество сущностей с указанными компонентами
func (w *World) CountEntitiesWith(mask ComponentMask) int {
	count := 0
	w.entitiesBuffer = w.entities.GetAliveEntities(w.entitiesBuffer)
	
	for _, entity := range w.entitiesBuffer {
		if w.HasComponents(entity, mask) {
			count++
		}
	}
	
	return count
}

// FindFirstWith находит первую сущность с указанными компонентами
func (w *World) FindFirstWith(mask ComponentMask) (EntityID, bool) {
	w.entitiesBuffer = w.entities.GetAliveEntities(w.entitiesBuffer)
	
	for _, entity := range w.entitiesBuffer {
		if w.HasComponents(entity, mask) {
			return entity, true
		}
	}
	
	return INVALID_ENTITY, false
}

// ForEachMoving итерирует по всем движущимся сущностям (Position + Velocity)
func (w *World) ForEachMoving(fn QueryFunc) {
	w.ForEachWith(MovingEntities.Mask(), fn)
}

// ForEachLiving итерирует по всем живым сущностям
func (w *World) ForEachLiving(fn QueryFunc) {
	w.ForEachWith(LivingEntities.Mask(), fn)
}

// ForEachAnimal итерирует по всем животным
func (w *World) ForEachAnimal(fn QueryFunc) {
	w.ForEachWith(AnimalsEntities.Mask(), fn)
}

// QueryByType возвращает всех животных определённого типа
func (w *World) QueryByType(animalType AnimalType) []EntityID {
	w.queryBuffer = w.queryBuffer[:0]
	
	w.ForEachWith(MaskAnimalType, func(entity EntityID) {
		if w.types[entity] == animalType {
			w.queryBuffer = append(w.queryBuffer, entity)
		}
	})
	
	return w.queryBuffer
}

// QueryInRadius возвращает всех животных в указанном радиусе от позиции
func (w *World) QueryInRadius(centerX, centerY, radius float32) []EntityID {
	w.queryBuffer = w.queryBuffer[:0]
	
	// Используем пространственную сетку для быстрого поиска
	spatialResults := w.spatialGrid.QueryRadius(
		physics.Vec2{X: centerX, Y: centerY}, radius)
	
	for _, result := range spatialResults {
		entity := EntityID(result.ID)
		if w.entities.IsAlive(entity) {
			w.queryBuffer = append(w.queryBuffer, entity)
		}
	}
	
	return w.queryBuffer
}

// FindNearestAnimal находит ближайшее животное к указанной позиции
func (w *World) FindNearestAnimal(centerX, centerY, maxRadius float32) (EntityID, bool) {
	result, found := w.spatialGrid.QueryNearest(
		physics.Vec2{X: centerX, Y: centerY}, maxRadius)
	
	if found {
		entity := EntityID(result.ID)
		if w.entities.IsAlive(entity) {
			return entity, true
		}
	}
	
	return INVALID_ENTITY, false
}

// FindNearestByType находит ближайшее животное определённого типа
func (w *World) FindNearestByType(centerX, centerY, maxRadius float32, animalType AnimalType) (EntityID, bool) {
	// Получаем всех в радиусе
	nearby := w.QueryInRadius(centerX, centerY, maxRadius)
	
	var nearestEntity EntityID = INVALID_ENTITY
	var nearestDistanceSq float32 = maxRadius * maxRadius
	
	for _, entity := range nearby {
		if !w.HasComponent(entity, MaskAnimalType) {
			continue
		}
		
		if w.types[entity] != animalType {
			continue
		}
		
		if !w.HasComponent(entity, MaskPosition) {
			continue
		}
		
		pos := w.positions[entity]
		dx := pos.X - centerX
		dy := pos.Y - centerY
		distanceSq := dx*dx + dy*dy
		
		if distanceSq < nearestDistanceSq {
			nearestDistanceSq = distanceSq
			nearestEntity = entity
		}
	}
	
	return nearestEntity, nearestEntity != INVALID_ENTITY
}

// GetStats возвращает статистику по типам животных
func (w *World) GetStats() map[AnimalType]int {
	stats := make(map[AnimalType]int)
	
	w.ForEachWith(MaskAnimalType, func(entity EntityID) {
		animalType := w.types[entity]
		stats[animalType]++
	})
	
	return stats
}

// MovementQuery специализированный запрос для системы движения
type MovementQuery struct {
	Entity   EntityID
	Position *Position
	Velocity *Velocity
}

// QueryMovement возвращает все данные для системы движения
func (w *World) QueryMovement() []MovementQuery {
	var results []MovementQuery
	
	w.ForEachWith(MovingEntities.Mask(), func(entity EntityID) {
		results = append(results, MovementQuery{
			Entity:   entity,
			Position: &w.positions[entity],
			Velocity: &w.velocities[entity],
		})
	})
	
	return results
}

// AnimalQuery специализированный запрос для логики животных
type AnimalQuery struct {
	Entity     EntityID
	Position   *Position
	Velocity   *Velocity
	Health     *Health
	Hunger     *Hunger
	Age        *Age
	AnimalType AnimalType
	Size       *Size
	Speed      *Speed
}

// QueryAnimals возвращает все данные животных
func (w *World) QueryAnimals() []AnimalQuery {
	var results []AnimalQuery
	
	w.ForEachWith(AnimalsEntities.Mask(), func(entity EntityID) {
		results = append(results, AnimalQuery{
			Entity:     entity,
			Position:   &w.positions[entity],
			Velocity:   &w.velocities[entity],
			Health:     &w.healths[entity],
			Hunger:     &w.hungers[entity],
			Age:        &w.ages[entity],
			AnimalType: w.types[entity],
			Size:       &w.sizes[entity],
			Speed:      &w.speeds[entity],
		})
	})
	
	return results
}