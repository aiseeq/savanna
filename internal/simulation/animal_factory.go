package simulation

import (
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
)

// CreateAnimal создает животное используя унифицированную систему AnimalConfig
// Заменяет дублированную логику между AnimalCreationConfig и core.AnimalConfig
func CreateAnimal(world *core.World, animalType core.AnimalType, x, y float32) core.EntityID {
	config := CreateAnimalConfig(animalType)
	return createEntityFromConfig(world, config, x, y)
}

// createEntityFromConfig создает сущность из конфигурации (единая точка создания)
func createEntityFromConfig(world *core.World, config core.AnimalConfig, x, y float32) core.EntityID {
	entity := world.CreateEntity()

	// Добавляем базовые компоненты
	world.AddPosition(entity, core.Position{X: x, Y: y})
	world.AddVelocity(entity, core.Velocity{X: 0, Y: 0})
	world.AddHealth(entity, core.Health{Current: config.MaxHealth, Max: config.MaxHealth})

	// Используем InitialHunger из конфигурации
	initialHunger := float32(70.0) //nolint:gomnd // Дефолтное значение голода для новых животных
	if animalType := getAnimalTypeFromConfig(config); animalType == core.TypeRabbit {
		initialHunger = 80.0 //nolint:gomnd // Начальный голод зайца (80%)
	}
	world.AddHunger(entity, core.Hunger{Value: initialHunger})

	// Добавляем AnimalConfig компонент
	world.AddAnimalConfig(entity, config)

	// Тип животного определяется из анимации или конфигурации
	world.AddAnimalType(entity, getAnimalTypeFromConfig(config))

	// Размеры из конфигурации
	world.AddSize(entity, core.Size{
		Radius:      config.CollisionRadius,
		AttackRange: config.AttackRange,
	})

	// Скорость из конфигурации
	world.AddSpeed(entity, core.Speed{
		Current: config.BaseSpeed,
		Base:    config.BaseSpeed,
	})

	// Поведение из конфигурации
	behaviorType := getBehaviorTypeFromConfig(config)
	world.AddBehavior(entity, core.Behavior{
		Type:             behaviorType,
		DirectionTimer:   0,
		HungerThreshold:  config.HungerThreshold,
		FleeThreshold:    config.FleeThreshold,
		SearchSpeed:      config.SearchSpeed,
		WanderingSpeed:   config.WanderingSpeed,
		ContentSpeed:     config.ContentSpeed,
		VisionRange:      config.VisionRange,
		MinDirectionTime: config.MinDirectionTime,
		MaxDirectionTime: config.MaxDirectionTime,
	})

	// Анимация
	world.AddAnimation(entity, core.Animation{
		CurrentAnim: int(animation.AnimIdle),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	return entity
}

// getAnimalTypeFromConfig определяет тип животного из конфигурации
func getAnimalTypeFromConfig(config core.AnimalConfig) core.AnimalType {
	// Определяем по характерным параметрам (AttackRange > 0 = хищник)
	if config.AttackRange > 0 {
		return core.TypeWolf
	}
	return core.TypeRabbit
}

// getBehaviorTypeFromConfig определяет тип поведения из конфигурации
func getBehaviorTypeFromConfig(config core.AnimalConfig) core.BehaviorType {
	// Определяем по характерным параметрам (AttackRange > 0 = хищник)
	if config.AttackRange > 0 {
		return core.BehaviorPredator
	}
	return core.BehaviorHerbivore
}
