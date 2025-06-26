package gamestate

import (
	"encoding/json"
	"fmt"
	"os"
)

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
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.events); err != nil {
		return fmt.Errorf("failed to encode events: %w", err)
	}

	return nil
}

// LoadFromFile загружает события из JSON файла
func LoadEventsFromFile(filename string) ([]InputEvent, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var events []InputEvent
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}
