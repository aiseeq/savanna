package gamestate

// InputEventType тип входного события
type InputEventType int

const (
	InputMouseDown InputEventType = iota
	InputMouseUp
	InputMouseMove
	InputKeyDown
	InputKeyUp
)

// MouseButton кнопка мыши
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

// InputEvent представляет входное событие
type InputEvent struct {
	Type   InputEventType
	Button MouseButton // для мыши
	Key    int         // для клавиатуры
	X, Y   float64     // координаты мыши
}

// InputProvider интерфейс для получения входных событий
type InputProvider interface {
	PollEvents() []InputEvent
}

// MockInputProvider для тестов - позволяет записывать и воспроизводить события
type MockInputProvider struct {
	events []InputEvent
	index  int
}

// NewMockInputProvider создает новый mock провайдер
func NewMockInputProvider(events []InputEvent) *MockInputProvider {
	return &MockInputProvider{
		events: events,
		index:  0,
	}
}

// PollEvents возвращает следующее событие из записанного скрипта
func (m *MockInputProvider) PollEvents() []InputEvent {
	if m.index >= len(m.events) {
		return nil
	}

	event := m.events[m.index]
	m.index++
	return []InputEvent{event}
}

// Reset сбрасывает воспроизведение с начала
func (m *MockInputProvider) Reset() {
	m.index = 0
}

// EventRecorder записывает события для последующего воспроизведения
type EventRecorder struct {
	events []InputEvent
}

// NewEventRecorder создает новый рекордер
func NewEventRecorder() *EventRecorder {
	return &EventRecorder{
		events: make([]InputEvent, 0),
	}
}

// Record записывает событие
func (r *EventRecorder) Record(event InputEvent) {
	r.events = append(r.events, event)
}

// GetEvents возвращает записанные события
func (r *EventRecorder) GetEvents() []InputEvent {
	return r.events
}

// SaveToFile сохраняет события в JSON файл для воспроизведения в тестах
func (r *EventRecorder) SaveToFile(filename string) error {
	// TODO: Реализовать сохранение в JSON
	return nil
}

// LoadFromFile загружает события из JSON файла
func LoadEventsFromFile(filename string) ([]InputEvent, error) {
	// TODO: Реализовать загрузку из JSON
	return nil, nil
}
