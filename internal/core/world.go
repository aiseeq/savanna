package core

import (
	"math/rand"
	"time"
	
	"github.com/aiseeq/savanna/internal/physics"
)

// World главная структура мира симуляции
// Использует Structure of Arrays (SOA) для оптимальной производительности
type World struct {
	// Управление сущностями
	entities EntityManager
	
	// Компоненты - индексируются по EntityID
	positions  [MAX_ENTITIES]Position
	velocities [MAX_ENTITIES]Velocity
	healths    [MAX_ENTITIES]Health
	hungers    [MAX_ENTITIES]Hunger
	ages       [MAX_ENTITIES]Age
	types      [MAX_ENTITIES]AnimalType
	sizes      [MAX_ENTITIES]Size
	speeds     [MAX_ENTITIES]Speed
	
	// Битовые маски для быстрой проверки наличия компонентов
	// Каждый uint64 хранит 64 бита, поэтому нужно MAX_ENTITIES/64 элементов
	hasPosition  [MAX_ENTITIES/64 + 1]uint64
	hasVelocity  [MAX_ENTITIES/64 + 1]uint64
	hasHealth    [MAX_ENTITIES/64 + 1]uint64
	hasHunger    [MAX_ENTITIES/64 + 1]uint64
	hasAge       [MAX_ENTITIES/64 + 1]uint64
	hasType      [MAX_ENTITIES/64 + 1]uint64
	hasSize      [MAX_ENTITIES/64 + 1]uint64
	hasSpeed     [MAX_ENTITIES/64 + 1]uint64
	
	// Физическая система
	spatialGrid *physics.SpatialGrid
	
	// Время симуляции
	time       float32     // Общее время симуляции в секундах
	deltaTime  float32     // Время с последнего обновления
	timeScale  float32     // Масштаб времени (1.0 = нормальная скорость)
	
	// Детерминированный генератор случайных чисел
	rng *rand.Rand
	
	// Размеры мира
	worldWidth  float32
	worldHeight float32
	
	// Буферы для переиспользования (предотвращение аллокаций)
	queryBuffer    []EntityID
	entitiesBuffer []EntityID
}

// NewWorld создаёт новый мир симуляции
func NewWorld(worldWidth, worldHeight float32, seed int64) *World {
	world := &World{
		entities:    *NewEntityManager(),
		spatialGrid: physics.NewSpatialGrid(worldWidth, worldHeight, 32.0), // 32 пикселя на ячейку
		time:        0,
		deltaTime:   0,
		timeScale:   1.0,
		rng:         rand.New(rand.NewSource(seed)),
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
		queryBuffer:    make([]EntityID, 0, 100),
		entitiesBuffer: make([]EntityID, 0, MAX_ENTITIES),
	}
	
	return world
}

// GetTime возвращает текущее время симуляции
func (w *World) GetTime() float32 {
	return w.time
}

// GetDeltaTime возвращает время с последнего обновления
func (w *World) GetDeltaTime() float32 {
	return w.deltaTime
}

// SetTimeScale устанавливает масштаб времени
func (w *World) SetTimeScale(scale float32) {
	if scale >= 0 {
		w.timeScale = scale
	}
}

// GetTimeScale возвращает текущий масштаб времени
func (w *World) GetTimeScale() float32 {
	return w.timeScale
}

// Update обновляет мир на один кадр
func (w *World) Update(realDeltaTime float32) {
	w.deltaTime = realDeltaTime * w.timeScale
	w.time += w.deltaTime
}

// CreateEntity создаёт новую сущность
func (w *World) CreateEntity() EntityID {
	return w.entities.CreateEntity()
}

// DestroyEntity уничтожает сущность и все её компоненты
func (w *World) DestroyEntity(entity EntityID) bool {
	if !w.entities.IsAlive(entity) {
		return false
	}
	
	// Удаляем из пространственной сетки если есть позиция
	if w.HasComponent(entity, MaskPosition) {
		w.spatialGrid.Remove(physics.EntityID(entity))
	}
	
	// Очищаем все компоненты
	w.removeAllComponents(entity)
	
	// Уничтожаем сущность
	return w.entities.DestroyEntity(entity)
}

// IsAlive проверяет, существует ли сущность
func (w *World) IsAlive(entity EntityID) bool {
	return w.entities.IsAlive(entity)
}

// GetEntityCount возвращает количество живых сущностей
func (w *World) GetEntityCount() int {
	return w.entities.Count()
}

// GetWorldDimensions возвращает размеры мира
func (w *World) GetWorldDimensions() (float32, float32) {
	return w.worldWidth, w.worldHeight
}

// GetSpatialGrid возвращает пространственную сетку для прямого доступа
func (w *World) GetSpatialGrid() *physics.SpatialGrid {
	return w.spatialGrid
}

// GetRNG возвращает генератор случайных чисел
func (w *World) GetRNG() *rand.Rand {
	return w.rng
}

// Clear очищает весь мир (для тестов и перезапуска)
func (w *World) Clear() {
	w.entities.Clear()
	w.spatialGrid.Clear()
	w.time = 0
	w.deltaTime = 0
	
	// Очищаем все битовые маски
	for i := range w.hasPosition {
		w.hasPosition[i] = 0
		w.hasVelocity[i] = 0
		w.hasHealth[i] = 0
		w.hasHunger[i] = 0
		w.hasAge[i] = 0
		w.hasType[i] = 0
		w.hasSize[i] = 0
		w.hasSpeed[i] = 0
	}
}

// Вспомогательные методы для работы с битовыми масками

// setBitMask устанавливает бит в маске
func setBitMask(mask []uint64, entity EntityID) {
	wordIndex := entity / 64
	bitIndex := entity % 64
	if wordIndex < uint16(len(mask)) {
		mask[wordIndex] |= 1 << bitIndex
	}
}

// clearBitMask очищает бит в маске
func clearBitMask(mask []uint64, entity EntityID) {
	wordIndex := entity / 64
	bitIndex := entity % 64
	if wordIndex < uint16(len(mask)) {
		mask[wordIndex] &^= 1 << bitIndex
	}
}

// testBitMask проверяет бит в маске
func testBitMask(mask []uint64, entity EntityID) bool {
	wordIndex := entity / 64
	bitIndex := entity % 64
	if wordIndex < uint16(len(mask)) {
		return mask[wordIndex]&(1<<bitIndex) != 0
	}
	return false
}

// removeAllComponents удаляет все компоненты у сущности
func (w *World) removeAllComponents(entity EntityID) {
	clearBitMask(w.hasPosition[:], entity)
	clearBitMask(w.hasVelocity[:], entity)
	clearBitMask(w.hasHealth[:], entity)
	clearBitMask(w.hasHunger[:], entity)
	clearBitMask(w.hasAge[:], entity)
	clearBitMask(w.hasType[:], entity)
	clearBitMask(w.hasSize[:], entity)
	clearBitMask(w.hasSpeed[:], entity)
}