package simulation

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

func TestBehaviorSystem_UsesAnimalConfig(t *testing.T) {
	// Создаем мир и систему поведения
	world := core.NewWorld(100, 100, 12345)

	// Создаем простой terrain для VegetationSystem
	terrain := &generator.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([][]generator.TileType, 10),
		Grass:  make([][]float32, 10),
	}
	// Заполняем terrain травой
	for y := 0; y < 10; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 10)
		terrain.Grass[y] = make([]float32, 10)
		for x := 0; x < 10; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	behaviorSystem := NewAnimalBehaviorSystem(vegetationSystem)

	// Создаем кастомное травоядное с большим радиусом зрения
	entity := world.CreateEntity()

	// Добавляем все необходимые компоненты
	world.AddPosition(entity, core.Position{X: 50, Y: 50})
	world.AddVelocity(entity, core.Velocity{X: 0, Y: 0})
	world.AddHealth(entity, core.Health{Current: 100, Max: 100})
	world.AddSatiation(entity, core.Satiation{Value: 50.0}) // Голодное животное
	world.AddAnimalType(entity, core.TypeRabbit)

	// Создаем кастомную конфигурацию с большим радиусом зрения
	customConfig := core.AnimalConfig{
		BaseRadius:         0.5,
		MaxHealth:          100,
		BaseSpeed:          1.0,
		CollisionRadius:    0.5,
		AttackRange:        0.0,  // Травоядное
		VisionRange:        10.0, // БОЛЬШОЙ радиус зрения - 10 тайлов
		SatiationThreshold: 80.0, // Ест при голоде < 80%
		FleeThreshold:      2.0,
		SearchSpeed:        0.8,
		WanderingSpeed:     0.7,
		ContentSpeed:       0.3,
		MinDirectionTime:   1.0,
		MaxDirectionTime:   4.0,
		AttackDamage:       0,
		AttackCooldown:     0.0,
		HitChance:          0.0,
	}
	world.AddAnimalConfig(entity, customConfig)

	// Добавляем размер (конвертируем тайлы в пиксели)
	world.AddSize(entity, core.Size{
		Radius:      customConfig.CollisionRadius * 32, // 32 пикселя/тайл
		AttackRange: 0,
	})

	// Добавляем скорость
	world.AddSpeed(entity, core.Speed{
		Current: customConfig.BaseSpeed,
		Base:    customConfig.BaseSpeed,
	})

	// Добавляем поведение
	world.AddBehavior(entity, core.Behavior{
		Type:               core.BehaviorHerbivore,
		DirectionTimer:     0,
		SatiationThreshold: customConfig.SatiationThreshold,
		FleeThreshold:      customConfig.FleeThreshold,
		SearchSpeed:        customConfig.SearchSpeed,
		WanderingSpeed:     customConfig.WanderingSpeed,
		ContentSpeed:       customConfig.ContentSpeed,
		VisionRange:        customConfig.VisionRange,
		MinDirectionTime:   customConfig.MinDirectionTime,
		MaxDirectionTime:   customConfig.MaxDirectionTime,
	})

	// Запоминаем исходную скорость
	initialVel, _ := world.GetVelocity(entity)

	// Обновляем поведение
	deltaTime := float32(1.0 / 60.0)
	behaviorSystem.Update(world, deltaTime)

	// Проверяем что скорость изменилась (животное начало двигаться)
	finalVel, _ := world.GetVelocity(entity)

	// Животное должно начать двигаться (искать траву)
	if finalVel.X == initialVel.X && finalVel.Y == initialVel.Y {
		t.Error("Hungry animal should start moving to search for food")
	}

	// Проверяем что AnimalConfig используется корректно
	config, hasConfig := world.GetAnimalConfig(entity)
	if !hasConfig {
		t.Fatal("Entity should have AnimalConfig component")
	}

	if config.VisionRange != 10.0 {
		t.Errorf("Expected vision range 10.0, got %f", config.VisionRange)
	}

	if config.SatiationThreshold != 80.0 {
		t.Errorf("Expected hunger threshold 80.0, got %f", config.SatiationThreshold)
	}

	t.Logf("SUCCESS: Custom animal with vision range %f is using AnimalConfig correctly",
		config.VisionRange)
}

func TestBehaviorSystem_AnimalConfigVsHardcodedValues(t *testing.T) {
	// Сравниваем поведение животных созданных через фабрику с кастомными настройками
	world := core.NewWorld(100, 100, 12345)

	terrain := &generator.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([][]generator.TileType, 10),
		Grass:  make([][]float32, 10),
	}
	for y := 0; y < 10; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 10)
		terrain.Grass[y] = make([]float32, 10)
		for x := 0; x < 10; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	behaviorSystem := NewAnimalBehaviorSystem(vegetationSystem)

	// Создаем стандартного зайца через фабрику
	standardRabbit := CreateAnimal(world, core.TypeRabbit, 25, 25)

	// Проверяем что у него есть AnimalConfig
	standardConfig, hasStandardConfig := world.GetAnimalConfig(standardRabbit)
	if !hasStandardConfig {
		t.Fatal("Standard rabbit should have AnimalConfig component")
	}

	// Создаем волка через фабрику
	standardWolf := CreateAnimal(world, core.TypeWolf, 75, 75)

	// Проверяем что у него есть AnimalConfig
	wolfConfig, hasWolfConfig := world.GetAnimalConfig(standardWolf)
	if !hasWolfConfig {
		t.Fatal("Standard wolf should have AnimalConfig component")
	}

	// Проверяем что значения из AnimalConfig соответствуют game_balance.go
	if standardConfig.BaseSpeed != RabbitBaseSpeed {
		t.Errorf("Rabbit base speed from AnimalConfig (%f) should match game_balance.go (%f)",
			standardConfig.BaseSpeed, RabbitBaseSpeed)
	}

	if standardConfig.VisionRange != RabbitBaseRadius*RabbitVisionMultiplier {
		t.Errorf("Rabbit vision range from AnimalConfig (%f) should be calculated from game_balance.go (%f)",
			standardConfig.VisionRange, RabbitBaseRadius*RabbitVisionMultiplier)
	}

	if wolfConfig.BaseSpeed != WolfBaseSpeed {
		t.Errorf("Wolf base speed from AnimalConfig (%f) should match game_balance.go (%f)",
			wolfConfig.BaseSpeed, WolfBaseSpeed)
	}

	if wolfConfig.VisionRange != WolfBaseRadius*WolfVisionMultiplier {
		t.Errorf("Wolf vision range from AnimalConfig (%f) should be calculated from game_balance.go (%f)",
			wolfConfig.VisionRange, WolfBaseRadius*WolfVisionMultiplier)
	}

	// Тестируем поведение
	deltaTime := float32(1.0 / 60.0)
	behaviorSystem.Update(world, deltaTime)

	t.Logf("SUCCESS: Standard animals created through factory use correct AnimalConfig values")
	t.Logf("Rabbit: speed=%f, vision=%f", standardConfig.BaseSpeed, standardConfig.VisionRange)
	t.Logf("Wolf: speed=%f, vision=%f", wolfConfig.BaseSpeed, wolfConfig.VisionRange)
}

func TestBehaviorSystem_NoHardcodedAnimalTypes(t *testing.T) {
	// Тест проверяет что система поведения НЕ использует hardcoded проверки типа животного
	world := core.NewWorld(100, 100, 12345)

	terrain := &generator.Terrain{
		Width:  10,
		Height: 10,
		Tiles:  make([][]generator.TileType, 10),
		Grass:  make([][]float32, 10),
	}
	for y := 0; y < 10; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 10)
		terrain.Grass[y] = make([]float32, 10)
		for x := 0; x < 10; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	behaviorSystem := NewAnimalBehaviorSystem(vegetationSystem)

	// Создаем "неизвестное" животное с поведением травоядного
	unknownHerbivore := world.CreateEntity()

	// Добавляем все компоненты но с неизвестным типом животного
	world.AddPosition(unknownHerbivore, core.Position{X: 50, Y: 50})
	world.AddVelocity(unknownHerbivore, core.Velocity{X: 0, Y: 0})
	world.AddHealth(unknownHerbivore, core.Health{Current: 100, Max: 100})
	world.AddSatiation(unknownHerbivore, core.Satiation{Value: 50.0})
	world.AddAnimalType(unknownHerbivore, core.TypeNone) // Неизвестный тип!

	// Но с конфигурацией травоядного
	herbivoreConfig := core.AnimalConfig{
		BaseRadius:         0.3,
		MaxHealth:          100,
		BaseSpeed:          1.5,
		CollisionRadius:    0.3,
		AttackRange:        0.0, // НЕ атакует = травоядное
		VisionRange:        2.4,
		SatiationThreshold: 85.0,
		FleeThreshold:      1.2,
		SearchSpeed:        0.8,
		WanderingSpeed:     0.7,
		ContentSpeed:       0.3,
		MinDirectionTime:   1.0,
		MaxDirectionTime:   4.0,
		AttackDamage:       0,
		AttackCooldown:     0.0,
		HitChance:          0.0,
	}
	world.AddAnimalConfig(unknownHerbivore, herbivoreConfig)
	world.AddSize(unknownHerbivore, core.Size{Radius: 9.6, AttackRange: 0})
	world.AddSpeed(unknownHerbivore, core.Speed{Current: 1.5, Base: 1.5})
	world.AddBehavior(unknownHerbivore, core.Behavior{
		Type:               core.BehaviorHerbivore, // Поведение травоядного
		DirectionTimer:     0,
		SatiationThreshold: herbivoreConfig.SatiationThreshold,
		FleeThreshold:      herbivoreConfig.FleeThreshold,
		SearchSpeed:        herbivoreConfig.SearchSpeed,
		WanderingSpeed:     herbivoreConfig.WanderingSpeed,
		ContentSpeed:       herbivoreConfig.ContentSpeed,
		VisionRange:        herbivoreConfig.VisionRange,
		MinDirectionTime:   herbivoreConfig.MinDirectionTime,
		MaxDirectionTime:   herbivoreConfig.MaxDirectionTime,
	})

	// Система должна работать без ошибок
	deltaTime := float32(1.0 / 60.0)
	behaviorSystem.Update(world, deltaTime)

	// Проверяем что животное ведёт себя как травоядное несмотря на неизвестный тип
	finalVel, _ := world.GetVelocity(unknownHerbivore)

	// Голодное травоядное должно начать двигаться
	if finalVel.X == 0 && finalVel.Y == 0 {
		t.Error("Unknown herbivore should behave like herbivore and start moving when hungry")
	}

	t.Logf("SUCCESS: Unknown animal type with herbivore behavior works correctly")
}
