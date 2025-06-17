package core

import (
	"math/rand"

	"github.com/aiseeq/savanna/internal/physics"
)

// World главная структура мира симуляции
// Использует Structure of Arrays (SOA) для оптимальной производительности
type World struct {
	// Управление сущностями
	entities EntityManager

	// Компоненты - индексируются по EntityID
	positions     [MaxEntities]Position
	velocities    [MaxEntities]Velocity
	healths       [MaxEntities]Health
	hungers       [MaxEntities]Hunger
	ages          [MaxEntities]Age
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
	// Каждый uint64 хранит 64 бита, поэтому нужно MaxEntities/64 элементов
	hasPosition     [MaxEntities/64 + 1]uint64
	hasVelocity     [MaxEntities/64 + 1]uint64
	hasHealth       [MaxEntities/64 + 1]uint64
	hasHunger       [MaxEntities/64 + 1]uint64
	hasAge          [MaxEntities/64 + 1]uint64
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

	// Пространственная система запросов (абстракция через интерфейс - соблюдает DIP)
	spatialProvider SpatialQueryProvider

	// Время симуляции
	time      float32 // Общее время симуляции в секундах
	deltaTime float32 // Время с последнего обновления
	timeScale float32 // Масштаб времени (1.0 = нормальная скорость)

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
		entities: *NewEntityManager(),
		// Используем адаптер для SpatialGrid (соблюдает DIP)
		spatialProvider: NewSpatialGridAdapter(worldWidth, worldHeight),
		time:            0,
		deltaTime:       0,
		timeScale:       1.0,
		rng:             rand.New(rand.NewSource(seed)),
		worldWidth:      worldWidth,
		worldHeight:     worldHeight,
		queryBuffer:     make([]EntityID, 0, 100),
		entitiesBuffer:  make([]EntityID, 0, MaxEntities),
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
		w.spatialProvider.RemoveEntity(uint32(entity))
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
func (w *World) GetWorldDimensions() (width, height float32) {
	return w.worldWidth, w.worldHeight
}

// GetSpatialGrid возвращает пространственную сетку для прямого доступа
// DEPRECATED: нарушает DIP, будет удалена в будущем
func (w *World) GetSpatialGrid() *physics.SpatialGrid {
	// Приведение к конкретному типу через адаптер
	if adapter, ok := w.spatialProvider.(*SpatialGridAdapter); ok {
		return adapter.grid
	}
	return nil
}

// GetRNG возвращает генератор случайных чисел
func (w *World) GetRNG() *rand.Rand {
	return w.rng
}

// Clear очищает весь мир (для тестов и перезапуска)
func (w *World) Clear() {
	w.entities.Clear()
	w.spatialProvider.Clear()
	w.time = 0
	w.deltaTime = 0

	// Очищаем все битовые маски через слайс указателей (устраняет дублирование)
	allMasks := []*[MaxEntities/64 + 1]uint64{
		&w.hasPosition, &w.hasVelocity, &w.hasHealth, &w.hasHunger,
		&w.hasAge, &w.hasType, &w.hasSize, &w.hasSpeed,
		&w.hasAnimation, &w.hasDamageFlash, &w.hasCorpse, &w.hasEatingState,
		&w.hasAttackState, &w.hasBehavior, &w.hasAnimalConfig,
	}

	for _, mask := range allMasks {
		for i := range mask {
			mask[i] = 0
		}
	}
}

// Константы для битовых операций
const (
	BitsPerWord = 64 // Количество бит в uint64 слове
)

// Вспомогательные методы для работы с битовыми масками

// setBitMask устанавливает бит в маске
func setBitMask(mask []uint64, entity EntityID) {
	wordIndex := entity / BitsPerWord
	bitIndex := entity % BitsPerWord
	if int(wordIndex) < len(mask) {
		mask[wordIndex] |= 1 << bitIndex
	}
}

// clearBitMask очищает бит в маске
func clearBitMask(mask []uint64, entity EntityID) {
	wordIndex := entity / BitsPerWord
	bitIndex := entity % BitsPerWord
	if int(wordIndex) < len(mask) {
		mask[wordIndex] &^= 1 << bitIndex
	}
}

// testBitMask проверяет бит в маске
func testBitMask(mask []uint64, entity EntityID) bool {
	wordIndex := entity / BitsPerWord
	bitIndex := entity % BitsPerWord
	if int(wordIndex) < len(mask) {
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
	clearBitMask(w.hasAnimation[:], entity)
	clearBitMask(w.hasDamageFlash[:], entity)
	clearBitMask(w.hasCorpse[:], entity)
	clearBitMask(w.hasEatingState[:], entity)
	clearBitMask(w.hasAttackState[:], entity)
	clearBitMask(w.hasBehavior[:], entity)
	clearBitMask(w.hasAnimalConfig[:], entity)
}
