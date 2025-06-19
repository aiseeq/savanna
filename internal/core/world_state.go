package core

import "math/rand"

// WorldState управляет состоянием мира (время, размеры, RNG)
// Соблюдает Single Responsibility Principle - только состояние мира
type WorldState struct {
	// Время симуляции
	time      float32 // Общее время симуляции в секундах
	deltaTime float32 // Время с последнего обновления
	timeScale float32 // Масштаб времени (1.0 = нормальная скорость)

	// Детерминированный генератор случайных чисел
	rng *rand.Rand

	// Размеры мира
	worldWidth  float32
	worldHeight float32

	// Пространственная система запросов (абстракция через интерфейс - соблюдает DIP)
	spatialProvider SpatialQueryProvider

	// Буферы для переиспользования (предотвращение аллокаций)
	entitiesBuffer []EntityID
}

// NewWorldState создаёт новое состояние мира
func NewWorldState(worldWidth, worldHeight float32, seed int64) *WorldState {
	return &WorldState{
		time:            0,
		deltaTime:       0,
		timeScale:       1.0,
		rng:             rand.New(rand.NewSource(seed)),
		worldWidth:      worldWidth,
		worldHeight:     worldHeight,
		spatialProvider: NewSpatialGridAdapter(worldWidth, worldHeight),
		entitiesBuffer:  make([]EntityID, 0, MaxEntities),
	}
}

// GetTime возвращает текущее время симуляции
func (ws *WorldState) GetTime() float32 {
	return ws.time
}

// GetDeltaTime возвращает время с последнего обновления
func (ws *WorldState) GetDeltaTime() float32 {
	return ws.deltaTime
}

// SetTimeScale устанавливает масштаб времени
func (ws *WorldState) SetTimeScale(scale float32) {
	ws.timeScale = scale
}

// GetTimeScale возвращает текущий масштаб времени
func (ws *WorldState) GetTimeScale() float32 {
	return ws.timeScale
}

// Update обновляет время симуляции
func (ws *WorldState) Update(deltaTime float32) {
	ws.deltaTime = deltaTime * ws.timeScale
	ws.time += ws.deltaTime
}

// GetRNG возвращает детерминированный генератор случайных чисел
func (ws *WorldState) GetRNG() *rand.Rand {
	return ws.rng
}

// GetWorldWidth возвращает ширину мира
func (ws *WorldState) GetWorldWidth() float32 {
	return ws.worldWidth
}

// GetWorldHeight возвращает высоту мира
func (ws *WorldState) GetWorldHeight() float32 {
	return ws.worldHeight
}

// GetSpatialProvider возвращает пространственную систему запросов
func (ws *WorldState) GetSpatialProvider() SpatialQueryProvider {
	return ws.spatialProvider
}

// GetEntitiesBuffer возвращает буфер для сущностей (для переиспользования)
func (ws *WorldState) GetEntitiesBuffer() []EntityID {
	// Очищаем буфер для переиспользования
	ws.entitiesBuffer = ws.entitiesBuffer[:0]
	return ws.entitiesBuffer
}
