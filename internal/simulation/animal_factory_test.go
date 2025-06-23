package simulation

import (
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
)

func TestCreateAnimal_Rabbit(t *testing.T) {
	world := core.NewWorld(200, 200, 12345)
	entity := CreateAnimal(world, core.TypeRabbit, 100, 100)

	// Проверяем, что сущность создана
	if entity == 0 {
		t.Fatal("Failed to create rabbit entity")
	}

	// Проверяем позицию
	if !world.HasComponent(entity, core.MaskPosition) {
		t.Fatal("Rabbit should have Position component")
	}
	pos, ok := world.GetPosition(entity)
	if !ok {
		t.Fatal("Failed to get position component")
	}
	if pos.X != 100 || pos.Y != 100 {
		t.Errorf("Expected position (100, 100), got (%f, %f)", pos.X, pos.Y)
	}

	// Проверяем здоровье
	if !world.HasComponent(entity, core.MaskHealth) {
		t.Fatal("Rabbit should have Health component")
	}
	health, ok := world.GetHealth(entity)
	if !ok {
		t.Fatal("Failed to get health component")
	}
	if health.Max != RabbitMaxHealth {
		t.Errorf("Expected max health %d, got %d", RabbitMaxHealth, health.Max)
	}
	if health.Current != RabbitMaxHealth {
		t.Errorf("Expected current health %d, got %d", RabbitMaxHealth, health.Current)
	}

	// Проверяем голод
	if !world.HasComponent(entity, core.MaskHunger) {
		t.Fatal("Rabbit should have Hunger component")
	}
	hunger, ok := world.GetHunger(entity)
	if !ok {
		t.Fatal("Failed to get hunger component")
	}
	if hunger.Value != RabbitInitialHunger {
		t.Errorf("Expected initial hunger %f, got %f", RabbitInitialHunger, hunger.Value)
	}

	// Проверяем размеры (должны быть в пикселях после конвертации)
	if !world.HasComponent(entity, core.MaskSize) {
		t.Fatal("Rabbit should have Size component")
	}
	size, ok := world.GetSize(entity)
	if !ok {
		t.Fatal("Failed to get size component")
	}
	expectedRadius := float32(RabbitBaseRadius * constants.TileSizePixels)
	if size.Radius != expectedRadius {
		t.Errorf("Expected radius %f pixels, got %f", expectedRadius, size.Radius)
	}
	// Заяц не должен иметь радиус атаки
	if size.AttackRange != 0 {
		t.Errorf("Expected attack range 0, got %f", size.AttackRange)
	}

	// Проверяем скорость
	if !world.HasComponent(entity, core.MaskSpeed) {
		t.Fatal("Rabbit should have Speed component")
	}
	speed, ok := world.GetSpeed(entity)
	if !ok {
		t.Fatal("Failed to get speed component")
	}
	if speed.Base != RabbitBaseSpeed {
		t.Errorf("Expected base speed %f, got %f", RabbitBaseSpeed, speed.Base)
	}

	// Проверяем AnimalConfig
	if !world.HasComponent(entity, core.MaskAnimalConfig) {
		t.Fatal("Rabbit should have AnimalConfig component")
	}
	config, ok := world.GetAnimalConfig(entity)
	if !ok {
		t.Fatal("Failed to get animal config component")
	}
	if config.BaseRadius != RabbitBaseRadius {
		t.Errorf("Expected base radius %f, got %f", RabbitBaseRadius, config.BaseRadius)
	}
	if config.VisionRange != RabbitBaseRadius*RabbitVisionMultiplier {
		t.Errorf("Expected vision range %f, got %f", RabbitBaseRadius*RabbitVisionMultiplier, config.VisionRange)
	}
	if config.HungerThreshold != RabbitHungerThreshold {
		t.Errorf("Expected hunger threshold %f, got %f", RabbitHungerThreshold, config.HungerThreshold)
	}

	// Проверяем поведение
	if !world.HasComponent(entity, core.MaskBehavior) {
		t.Fatal("Rabbit should have Behavior component")
	}
	behavior, ok := world.GetBehavior(entity)
	if !ok {
		t.Fatal("Failed to get behavior component")
	}
	if behavior.Type != core.BehaviorHerbivore {
		t.Errorf("Expected behavior type %v, got %v", core.BehaviorHerbivore, behavior.Type)
	}

	// Проверяем тип животного
	if !world.HasComponent(entity, core.MaskAnimalType) {
		t.Fatal("Rabbit should have AnimalType component")
	}
	animalType, ok := world.GetAnimalType(entity)
	if !ok {
		t.Fatal("Failed to get animal type component")
	}
	if animalType != core.TypeRabbit {
		t.Errorf("Expected animal type %v, got %v", core.TypeRabbit, animalType)
	}
}

func TestCreateAnimal_Wolf(t *testing.T) {
	world := core.NewWorld(300, 300, 12345)
	entity := CreateAnimal(world, core.TypeWolf, 200, 200)

	// Проверяем, что сущность создана
	if entity == 0 {
		t.Fatal("Failed to create wolf entity")
	}

	// Проверяем позицию
	if !world.HasComponent(entity, core.MaskPosition) {
		t.Fatal("Wolf should have Position component")
	}
	pos, ok := world.GetPosition(entity)
	if !ok {
		t.Fatal("Failed to get position component")
	}
	if pos.X != 200 || pos.Y != 200 {
		t.Errorf("Expected position (200, 200), got (%f, %f)", pos.X, pos.Y)
	}

	// Проверяем здоровье
	if !world.HasComponent(entity, core.MaskHealth) {
		t.Fatal("Wolf should have Health component")
	}
	health, ok := world.GetHealth(entity)
	if !ok {
		t.Fatal("Failed to get health component")
	}
	if health.Max != WolfMaxHealth {
		t.Errorf("Expected max health %d, got %d", WolfMaxHealth, health.Max)
	}

	// Проверяем голод
	if !world.HasComponent(entity, core.MaskHunger) {
		t.Fatal("Wolf should have Hunger component")
	}
	hunger, ok := world.GetHunger(entity)
	if !ok {
		t.Fatal("Failed to get hunger component")
	}
	if hunger.Value != WolfInitialHunger {
		t.Errorf("Expected initial hunger %f, got %f", WolfInitialHunger, hunger.Value)
	}

	// Проверяем размеры (должны быть в пикселях после конвертации)
	if !world.HasComponent(entity, core.MaskSize) {
		t.Fatal("Wolf should have Size component")
	}
	size, ok := world.GetSize(entity)
	if !ok {
		t.Fatal("Failed to get size component")
	}
	expectedRadius := float32(WolfBaseRadius * constants.TileSizePixels)
	if size.Radius != expectedRadius {
		t.Errorf("Expected radius %f pixels, got %f", expectedRadius, size.Radius)
	}
	expectedAttackRange := float32(WolfBaseRadius * WolfAttackRangeMultiplier * constants.TileSizePixels)
	if size.AttackRange != expectedAttackRange {
		t.Errorf("Expected attack range %f pixels, got %f", expectedAttackRange, size.AttackRange)
	}

	// Проверяем скорость
	if !world.HasComponent(entity, core.MaskSpeed) {
		t.Fatal("Wolf should have Speed component")
	}
	speed, ok := world.GetSpeed(entity)
	if !ok {
		t.Fatal("Failed to get speed component")
	}
	if speed.Base != WolfBaseSpeed {
		t.Errorf("Expected base speed %f, got %f", WolfBaseSpeed, speed.Base)
	}

	// Проверяем AnimalConfig
	if !world.HasComponent(entity, core.MaskAnimalConfig) {
		t.Fatal("Wolf should have AnimalConfig component")
	}
	config, ok := world.GetAnimalConfig(entity)
	if !ok {
		t.Fatal("Failed to get animal config component")
	}
	if config.BaseRadius != WolfBaseRadius {
		t.Errorf("Expected base radius %f, got %f", WolfBaseRadius, config.BaseRadius)
	}
	if config.VisionRange != WolfBaseRadius*WolfVisionMultiplier {
		t.Errorf("Expected vision range %f, got %f", WolfBaseRadius*WolfVisionMultiplier, config.VisionRange)
	}
	if config.HungerThreshold != WolfHungerThreshold {
		t.Errorf("Expected hunger threshold %f, got %f", WolfHungerThreshold, config.HungerThreshold)
	}
	if config.AttackDamage != WolfAttackDamageDefault {
		t.Errorf("Expected attack damage %d, got %d", WolfAttackDamageDefault, config.AttackDamage)
	}

	// Проверяем поведение
	if !world.HasComponent(entity, core.MaskBehavior) {
		t.Fatal("Wolf should have Behavior component")
	}
	behavior, ok := world.GetBehavior(entity)
	if !ok {
		t.Fatal("Failed to get behavior component")
	}
	if behavior.Type != core.BehaviorPredator {
		t.Errorf("Expected behavior type %v, got %v", core.BehaviorPredator, behavior.Type)
	}

	// Проверяем тип животного
	if !world.HasComponent(entity, core.MaskAnimalType) {
		t.Fatal("Wolf should have AnimalType component")
	}
	animalType, ok := world.GetAnimalType(entity)
	if !ok {
		t.Fatal("Failed to get animal type component")
	}
	if animalType != core.TypeWolf {
		t.Errorf("Expected animal type %v, got %v", core.TypeWolf, animalType)
	}
}

func TestCreateAnimal_AllParametersFromGameBalance(t *testing.T) {
	world := core.NewWorld(400, 400, 12345)

	// Тест для зайца - все параметры должны быть из game_balance.go
	rabbit := CreateAnimal(world, core.TypeRabbit, 0, 0)
	rabbitConfig, ok := world.GetAnimalConfig(rabbit)
	if !ok {
		t.Fatal("Failed to get rabbit config")
	}

	// Проверяем, что все ключевые параметры взяты из констант баланса
	if rabbitConfig.BaseSpeed != RabbitBaseSpeed {
		t.Errorf("Rabbit speed should be %f from game_balance.go, got %f", RabbitBaseSpeed, rabbitConfig.BaseSpeed)
	}
	if rabbitConfig.VisionRange != RabbitBaseRadius*RabbitVisionMultiplier {
		t.Errorf("Rabbit vision should be calculated from game_balance.go constants")
	}

	// Тест для волка
	wolf := CreateAnimal(world, core.TypeWolf, 0, 0)
	wolfConfig, ok := world.GetAnimalConfig(wolf)
	if !ok {
		t.Fatal("Failed to get wolf config")
	}

	if wolfConfig.BaseSpeed != WolfBaseSpeed {
		t.Errorf("Wolf speed should be %f from game_balance.go, got %f", WolfBaseSpeed, wolfConfig.BaseSpeed)
	}
	if wolfConfig.AttackRange != WolfBaseRadius*WolfAttackRangeMultiplier {
		t.Errorf("Wolf attack range should be calculated from game_balance.go constants")
	}
}
