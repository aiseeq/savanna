package simulation

import "github.com/aiseeq/savanna/internal/core"

// AnimalConfigFactory интерфейс для создания конфигурации животных (Factory Pattern)
// Соблюдает принципы OCP и SRP
type AnimalConfigFactory interface {
	CreateConfig() core.AnimalConfig
}

// AnimalConfigRegistry реестр factory для создания конфигураций животных
// Соблюдает принципы OCP - новые типы животных добавляются без изменения существующего кода
type AnimalConfigRegistry struct {
	factories map[core.AnimalType]AnimalConfigFactory
}

// NewAnimalConfigRegistry создаёт новый реестр конфигураций
func NewAnimalConfigRegistry() *AnimalConfigRegistry {
	registry := &AnimalConfigRegistry{
		factories: make(map[core.AnimalType]AnimalConfigFactory),
	}

	// Регистрируем стандартные factory
	registry.RegisterFactory(core.TypeRabbit, NewRabbitConfigFactory())
	registry.RegisterFactory(core.TypeWolf, NewWolfConfigFactory())

	return registry
}

// RegisterFactory регистрирует factory для типа животного
// Соблюдает принцип OCP - новые типы добавляются без изменения кода
func (r *AnimalConfigRegistry) RegisterFactory(animalType core.AnimalType, factory AnimalConfigFactory) {
	r.factories[animalType] = factory
}

// CreateConfig создаёт конфигурацию животного по типу
// Замещает switch-конструкцию на полиморфизм (OCP)
func (r *AnimalConfigRegistry) CreateConfig(animalType core.AnimalType) core.AnimalConfig {
	if factory, exists := r.factories[animalType]; exists {
		return factory.CreateConfig()
	}

	// Fallback на базовую конфигурацию для неизвестных типов
	return NewDefaultConfigFactory().CreateConfig()
}

// Глобальный реестр для backward compatibility
var defaultRegistry = NewAnimalConfigRegistry()
