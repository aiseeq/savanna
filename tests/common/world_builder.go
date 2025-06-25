package common

import (
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// AnimalSpec описывает параметры создаваемого животного
type AnimalSpec struct {
	X         float32 // Позиция X
	Y         float32 // Позиция Y
	Satiation float32 // Процент сытости (0-100)
	Health    int16   // Текущее здоровье
}

// TestEntities содержит ссылки на созданных в тесте животных
type TestEntities struct {
	Rabbits []core.EntityID
	Wolves  []core.EntityID
}

// TestWorldBuilder строитель для создания тестовых миров (применяет Builder Pattern)
// Устраняет дублирование создания миров в 80+ тестах
type TestWorldBuilder struct {
	worldSize   float32
	seed        int64
	rabbits     []AnimalSpec
	wolves      []AnimalSpec
	withSystems bool
}

// NewTestWorld создает новый строитель тестового мира с дефолтными параметрами
func NewTestWorld() *TestWorldBuilder {
	return &TestWorldBuilder{
		worldSize:   MediumWorldSize, // 640x640 - стандартный размер для большинства тестов
		seed:        DefaultTestSeed, // 42 - детерминированный seed
		withSystems: true,            // По умолчанию создаем системы
	}
}

// WithSize устанавливает размер мира
func (b *TestWorldBuilder) WithSize(size float32) *TestWorldBuilder {
	b.worldSize = size
	return b
}

// WithSmallSize устанавливает маленький размер мира (для unit тестов)
func (b *TestWorldBuilder) WithSmallSize() *TestWorldBuilder {
	b.worldSize = SmallWorldSize
	return b
}

// WithLargeSize устанавливает большой размер мира (для E2E тестов)
func (b *TestWorldBuilder) WithLargeSize() *TestWorldBuilder {
	b.worldSize = LargeWorldSize
	return b
}

// WithSeed устанавливает seed для детерминированности
func (b *TestWorldBuilder) WithSeed(seed int64) *TestWorldBuilder {
	b.seed = seed
	return b
}

// WithoutSystems отключает создание систем (для unit тестов)
func (b *TestWorldBuilder) WithoutSystems() *TestWorldBuilder {
	b.withSystems = false
	return b
}

// AddRabbit добавляет зайца с заданными параметрами
func (b *TestWorldBuilder) AddRabbit(x, y, satiation float32, health int16) *TestWorldBuilder {
	b.rabbits = append(b.rabbits, AnimalSpec{
		X:         x,
		Y:         y,
		Satiation: satiation,
		Health:    health,
	})
	return b
}

// AddHungryRabbit добавляет голодного зайца в стандартной позиции
func (b *TestWorldBuilder) AddHungryRabbit() *TestWorldBuilder {
	return b.AddRabbit(RabbitStartX, RabbitStartY, HungryPercentage, RabbitMaxHealth)
}

// AddSatedRabbit добавляет сытого зайца в стандартной позиции
func (b *TestWorldBuilder) AddSatedRabbit() *TestWorldBuilder {
	return b.AddRabbit(RabbitStartX, RabbitStartY, SatedPercentage, RabbitMaxHealth)
}

// AddDamagedRabbit добавляет поврежденного зайца (для тестов боя)
func (b *TestWorldBuilder) AddDamagedRabbit() *TestWorldBuilder {
	return b.AddRabbit(RabbitStartX, RabbitStartY, SatedPercentage, RabbitMaxHealth/2)
}

// AddWolf добавляет волка с заданными параметрами
func (b *TestWorldBuilder) AddWolf(x, y, hunger float32) *TestWorldBuilder {
	b.wolves = append(b.wolves, AnimalSpec{
		X:         x,
		Y:         y,
		Satiation: hunger,
		Health:    WolfMaxHealth,
	})
	return b
}

// AddHungryWolf добавляет голодного волка рядом с зайцем
func (b *TestWorldBuilder) AddHungryWolf() *TestWorldBuilder {
	return b.AddWolf(WolfStartX, WolfStartY, VeryHungryPercentage)
}

// AddSatedWolf добавляет сытого волка (не будет атаковать)
func (b *TestWorldBuilder) AddSatedWolf() *TestWorldBuilder {
	return b.AddWolf(WolfStartX, WolfStartY, SatedPercentage)
}

// AddWolfNearRabbit добавляет волка на заданном расстоянии от зайца
func (b *TestWorldBuilder) AddWolfNearRabbit(distance, hunger float32) *TestWorldBuilder {
	if len(b.rabbits) == 0 {
		// Если зайцев нет, добавляем стандартного зайца
		b.AddHungryRabbit()
	}

	// Берем позицию последнего добавленного зайца
	lastRabbit := b.rabbits[len(b.rabbits)-1]

	// Размещаем волка на заданном расстоянии справа от зайца
	return b.AddWolf(lastRabbit.X+distance, lastRabbit.Y, hunger)
}

// Build создает мир, системы и сущности на основе настроек
func (b *TestWorldBuilder) Build() (*core.World, *core.SystemManager, TestEntities) {
	// Создаем мир
	world := core.NewWorld(b.worldSize, b.worldSize, b.seed)

	// Создаем системы если нужно
	var systemManager *core.SystemManager
	if b.withSystems {
		systemManager = CreateTestSystemManager(b.worldSize)
	}

	// Создаем сущности
	entities := TestEntities{}

	// Создаем зайцев
	for _, spec := range b.rabbits {
		rabbit := simulation.CreateAnimal(world, core.TypeRabbit, spec.X, spec.Y)
		world.SetSatiation(rabbit, core.Satiation{Value: spec.Satiation})
		world.SetHealth(rabbit, core.Health{Current: spec.Health, Max: RabbitMaxHealth})
		entities.Rabbits = append(entities.Rabbits, rabbit)
	}

	// Создаем волков
	for _, spec := range b.wolves {
		wolf := simulation.CreateAnimal(world, core.TypeWolf, spec.X, spec.Y)
		world.SetSatiation(wolf, core.Satiation{Value: spec.Satiation})
		world.SetHealth(wolf, core.Health{Current: spec.Health, Max: WolfMaxHealth})
		entities.Wolves = append(entities.Wolves, wolf)
	}

	return world, systemManager, entities
}

// BuildWithAnimations создает мир с системами и анимациями (новый метод для тестов боя)
func (b *TestWorldBuilder) BuildWithAnimations() (*core.World, *TestSystemBundle, TestEntities) {
	// Создаем мир
	world := core.NewWorld(b.worldSize, b.worldSize, b.seed)

	// Создаем системы с анимациями если нужно
	var systemBundle *TestSystemBundle
	if b.withSystems {
		systemBundle = CreateTestSystemBundle(b.worldSize)
	}

	// Создаем сущности
	entities := TestEntities{}

	// Создаем зайцев
	for _, spec := range b.rabbits {
		rabbit := simulation.CreateAnimal(world, core.TypeRabbit, spec.X, spec.Y)
		world.SetSatiation(rabbit, core.Satiation{Value: spec.Satiation})
		world.SetHealth(rabbit, core.Health{Current: spec.Health, Max: RabbitMaxHealth})
		entities.Rabbits = append(entities.Rabbits, rabbit)
	}

	// Создаем волков
	for _, spec := range b.wolves {
		wolf := simulation.CreateAnimal(world, core.TypeWolf, spec.X, spec.Y)
		world.SetSatiation(wolf, core.Satiation{Value: spec.Satiation})
		world.SetHealth(wolf, core.Health{Current: spec.Health, Max: WolfMaxHealth})
		entities.Wolves = append(entities.Wolves, wolf)
	}

	return world, systemBundle, entities
}

// BuildWorldOnly создает только мир без систем и сущностей (для unit тестов)
func (b *TestWorldBuilder) BuildWorldOnly() *core.World {
	return core.NewWorld(b.worldSize, b.worldSize, b.seed)
}
