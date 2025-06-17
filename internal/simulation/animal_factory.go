package simulation

import (
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
)

// AnimalCreationConfig временная конфигурация для создания животного
// DEPRECATED: будет заменена на core.AnimalConfig
type AnimalCreationConfig struct {
	Type             core.AnimalType
	MaxHealth        int16
	InitialHunger    float32
	Radius           float32
	AttackRange      float32
	Speed            float32
	BehaviorType     core.BehaviorType
	HungerThreshold  float32
	FleeThreshold    float32
	SearchSpeed      float32
	WanderingSpeed   float32
	ContentSpeed     float32
	VisionRange      float32
	MinDirectionTime float32
	MaxDirectionTime float32
}

// AnimalFactory фабрика для создания животных (устраняет нарушение God Object)
type AnimalFactory struct{}

// NewAnimalFactory создаёт новую фабрику животных
func NewAnimalFactory() *AnimalFactory {
	return &AnimalFactory{}
}

// createAnimal универсальная фабричная функция для создания животных (устраняет дублирование)
func (af *AnimalFactory) createAnimal(world *core.World, x, y float32, config AnimalCreationConfig) core.EntityID {
	entity := world.CreateEntity()

	world.AddPosition(entity, core.Position{X: x, Y: y})
	world.AddVelocity(entity, core.Velocity{X: 0, Y: 0})
	world.AddHealth(entity, core.Health{Current: config.MaxHealth, Max: config.MaxHealth})
	world.AddHunger(entity, core.Hunger{Value: config.InitialHunger})
	world.AddAge(entity, core.Age{Seconds: 0})
	world.AddAnimalType(entity, config.Type)
	world.AddSize(entity, core.Size{
		Radius:      config.Radius,
		AttackRange: config.AttackRange,
	})
	world.AddSpeed(entity, core.Speed{Base: config.Speed, Current: config.Speed})

	// НОВОЕ: Добавляем компонент AnimalConfig для SOLID-архитектуры
	animalConfig := CreateAnimalConfig(config.Type)
	world.AddAnimalConfig(entity, animalConfig)

	// Добавляем компонент анимации, начинаем с idle
	world.AddAnimation(entity, core.Animation{
		CurrentAnim: int(animation.AnimIdle),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Добавляем компонент поведения
	world.AddBehavior(entity, core.Behavior{
		Type:             config.BehaviorType,
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

	return entity
}

// CreateRabbit создаёт зайца в указанной позиции
func (af *AnimalFactory) CreateRabbit(world *core.World, x, y float32) core.EntityID {
	return af.createAnimal(world, x, y, AnimalCreationConfig{
		Type:             core.TypeRabbit,
		MaxHealth:        RabbitMaxHealth,
		InitialHunger:    RabbitInitialHunger,
		Radius:           RabbitRadius,
		AttackRange:      0, // Зайцы мирные
		Speed:            RabbitSpeed,
		BehaviorType:     core.BehaviorHerbivore,
		HungerThreshold:  RabbitHungryThreshold,
		FleeThreshold:    VisionRangeRabbit,
		SearchSpeed:      SpeedSearchingFood,
		WanderingSpeed:   SpeedWanderingFood,
		ContentSpeed:     SpeedContentWalk,
		VisionRange:      VisionRangeRabbit,
		MinDirectionTime: RandomWalkMinTime,
		MaxDirectionTime: RandomWalkMaxTime,
	})
}

// CreateWolf создаёт волка в указанной позиции
func (af *AnimalFactory) CreateWolf(world *core.World, x, y float32) core.EntityID {
	return af.createAnimal(world, x, y, AnimalCreationConfig{
		Type:             core.TypeWolf,
		MaxHealth:        WolfMaxHealth,
		InitialHunger:    WolfInitialHunger,
		Radius:           WolfRadius,
		AttackRange:      WolfAttackRange,
		Speed:            WolfSpeed,
		BehaviorType:     core.BehaviorPredator,
		HungerThreshold:  WolfHungerThreshold,
		FleeThreshold:    0, // Волки ни от кого не убегают
		SearchSpeed:      SpeedSearchingFood,
		WanderingSpeed:   SpeedWanderingFood,
		ContentSpeed:     SpeedContentWalk,
		VisionRange:      VisionRangeWolf,
		MinDirectionTime: RandomWalkMinTime,
		MaxDirectionTime: RandomWalkMaxTime,
	})
}

// Глобальные функции для обратной совместимости

// CreateRabbit создаёт зайца в указанной позиции (глобальная функция для совместимости)
func CreateRabbit(world *core.World, x, y float32) core.EntityID {
	factory := NewAnimalFactory()
	return factory.CreateRabbit(world, x, y)
}

// CreateWolf создаёт волка в указанной позиции (глобальная функция для совместимости)
func CreateWolf(world *core.World, x, y float32) core.EntityID {
	factory := NewAnimalFactory()
	return factory.CreateWolf(world, x, y)
}
