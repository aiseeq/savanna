package mocks

import "github.com/aiseeq/savanna/internal/gamestate"

// MockInputProvider для тестов - позволяет записывать и воспроизводить события
type MockInputProvider struct {
	events []gamestate.InputEvent
	index  int
}

// NewMockInputProvider создает новый mock провайдер
func NewMockInputProvider(events []gamestate.InputEvent) *MockInputProvider {
	return &MockInputProvider{
		events: events,
		index:  0,
	}
}

// PollEvents возвращает следующее событие из записанного скрипта
func (m *MockInputProvider) PollEvents() []gamestate.InputEvent {
	if m.index >= len(m.events) {
		return nil
	}

	event := m.events[m.index]
	m.index++
	return []gamestate.InputEvent{event}
}

// Reset сбрасывает воспроизведение с начала
func (m *MockInputProvider) Reset() {
	m.index = 0
}
