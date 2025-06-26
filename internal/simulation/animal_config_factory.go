package simulation

import "github.com/aiseeq/savanna/internal/core"

// CreateAnimalConfig создаёт конфигурацию животного по типу
// РЕФАКТОРИНГ OCP: Заменён switch на Factory Pattern для соблюдения принципа открытости/закрытости
func CreateAnimalConfig(animalType core.AnimalType) core.AnimalConfig {
	return defaultRegistry.CreateConfig(animalType)
}

// RegisterAnimalConfigFactory позволяет регистрировать новые типы животных
// Соблюдает принцип OCP - новые типы добавляются без изменения существующего кода
func RegisterAnimalConfigFactory(animalType core.AnimalType, factory AnimalConfigFactory) {
	defaultRegistry.RegisterFactory(animalType, factory)
}
