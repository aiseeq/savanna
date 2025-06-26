package simulation

import "github.com/aiseeq/savanna/internal/core"

// SatiationConfigFactory интерфейс для определения начального голода животных (Factory Pattern)
// Соблюдает принципы OCP и SRP
type SatiationConfigFactory interface {
	GetInitialSatiation() float32
}

// SatiationConfigRegistry реестр factory для начального голода животных
// Соблюдает принципы OCP - новые типы животных добавляются без изменения существующего кода
type SatiationConfigRegistry struct {
	factories map[core.AnimalType]SatiationConfigFactory
}

// NewSatiationConfigRegistry создаёт новый реестр конфигураций начального голода
func NewSatiationConfigRegistry() *SatiationConfigRegistry {
	registry := &SatiationConfigRegistry{
		factories: make(map[core.AnimalType]SatiationConfigFactory),
	}

	// Регистрируем стандартные factory
	registry.RegisterFactory(core.TypeRabbit, NewRabbitSatiationFactory())
	registry.RegisterFactory(core.TypeWolf, NewWolfSatiationFactory())

	return registry
}

// RegisterFactory регистрирует factory для типа животного
func (r *SatiationConfigRegistry) RegisterFactory(animalType core.AnimalType, factory SatiationConfigFactory) {
	r.factories[animalType] = factory
}

// GetInitialSatiation возвращает начальный голод для типа животного
// Замещает switch-конструкцию на полиморфизм (OCP)
func (r *SatiationConfigRegistry) GetInitialSatiation(animalType core.AnimalType) float32 {
	if factory, exists := r.factories[animalType]; exists {
		return factory.GetInitialSatiation()
	}

	// Fallback на базовое значение для неизвестных типов
	return DefaultInitialSatiation
}

// Конкретные factory для начального голода

// RabbitSatiationFactory определяет начальный голод зайца
type RabbitSatiationFactory struct{}

func NewRabbitSatiationFactory() *RabbitSatiationFactory {
	return &RabbitSatiationFactory{}
}

func (f *RabbitSatiationFactory) GetInitialSatiation() float32 {
	return RabbitInitialSatiation
}

// WolfSatiationFactory определяет начальный голод волка
type WolfSatiationFactory struct{}

func NewWolfSatiationFactory() *WolfSatiationFactory {
	return &WolfSatiationFactory{}
}

func (f *WolfSatiationFactory) GetInitialSatiation() float32 {
	return WolfInitialSatiation
}

// Глобальный реестр для backward compatibility
var defaultSatiationRegistry = NewSatiationConfigRegistry()

// GetInitialSatiationForAnimal возвращает начальный голод для типа животного
// РЕФАКТОРИНГ OCP: Заменён switch на Factory Pattern
func GetInitialSatiationForAnimal(animalType core.AnimalType) float32 {
	return defaultSatiationRegistry.GetInitialSatiation(animalType)
}
