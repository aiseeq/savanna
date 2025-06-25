package simulation

import "github.com/aiseeq/savanna/internal/core"

// CreateAnimalConfig создаёт конфигурацию животного по типу
// Устраняет нарушения SOLID: заменяет захардкоженные константы на компонентную архитектуру
func CreateAnimalConfig(animalType core.AnimalType) core.AnimalConfig {
	switch animalType {
	case core.TypeRabbit:
		return createRabbitConfig()
	case core.TypeWolf:
		return createWolfConfig()
	default:
		// Базовая конфигурация для неизвестных типов
		return core.AnimalConfig{
			BaseRadius:         DefaultAnimalRadius,
			MaxHealth:          DefaultAnimalHealth,
			BaseSpeed:          DefaultAnimalSpeed,
			CollisionRadius:    DefaultAnimalRadius * CollisionRadiusMultiplier,
			AttackRange:        PacifistAttackDamage, // Не атакует
			VisionRange:        DefaultAnimalRadius * DefaultVisionMultiplier,
			SatiationThreshold: DefaultSatiationThreshold,
			FleeThreshold:      DefaultAnimalRadius * RabbitFleeDistanceMultiplier, // Используем множитель зайца как базовый
			SearchSpeed:        SearchSpeedMultiplier,
			WanderingSpeed:     WanderingSpeedMultiplier,
			ContentSpeed:       ContentSpeedMultiplier,
			MinDirectionTime:   DefaultMinDirectionTime,
			MaxDirectionTime:   DefaultMaxDirectionTime,
			AttackDamage:       PacifistAttackDamage,
			AttackCooldown:     PacifistAttackCooldown,
			HitChance:          PacifistHitChance,
		}
	}
}

// createRabbitConfig создаёт конфигурацию зайца
// Все параметры выводятся от базового радиуса через явные множители из game_balance.go
func createRabbitConfig() core.AnimalConfig {
	return core.AnimalConfig{
		// Базовые параметры
		BaseRadius: RabbitBaseRadius,
		MaxHealth:  RabbitMaxHealth,
		BaseSpeed:  RabbitBaseSpeed,

		// Размеры (выводятся от базового радиуса)
		CollisionRadius: RabbitBaseRadius * CollisionRadiusMultiplier,
		AttackRange:     PacifistAttackDamage, // Заяц не атакует
		VisionRange:     RabbitBaseRadius * RabbitVisionMultiplier,

		// Поведение травоядного
		SatiationThreshold: RabbitSatiationThreshold,
		FleeThreshold:      RabbitBaseRadius * RabbitFleeDistanceMultiplier,

		// Скорости в разных состояниях
		SearchSpeed:    SearchSpeedMultiplier,
		WanderingSpeed: WanderingSpeedMultiplier,
		ContentSpeed:   ContentSpeedMultiplier,

		// Таймеры поведения
		MinDirectionTime: RabbitMinDirectionTime,
		MaxDirectionTime: RabbitMaxDirectionTime,

		// Боевые характеристики (заяц мирный)
		AttackDamage:   PacifistAttackDamage,
		AttackCooldown: PacifistAttackCooldown,
		HitChance:      PacifistHitChance,
	}
}

// createWolfConfig создаёт конфигурацию волка
// Все параметры выводятся от базового радиуса через явные множители из game_balance.go
func createWolfConfig() core.AnimalConfig {
	return core.AnimalConfig{
		// Базовые параметры
		BaseRadius: WolfBaseRadius,
		MaxHealth:  WolfMaxHealth,
		BaseSpeed:  WolfBaseSpeed,

		// Размеры (выводятся от базового радиуса)
		CollisionRadius: WolfBaseRadius * CollisionRadiusMultiplier,
		AttackRange:     WolfBaseRadius * WolfAttackRangeMultiplier,
		VisionRange:     WolfBaseRadius * WolfVisionMultiplier,

		// Поведение хищника
		SatiationThreshold: WolfSatiationThreshold,
		FleeThreshold:      PacifistAttackDamage, // Волк не убегает (используем 0.0)

		// Скорости в разных состояниях
		SearchSpeed:    HuntingSpeedMultiplier, // Полная скорость при охоте
		WanderingSpeed: WanderingSpeedMultiplier,
		ContentSpeed:   ContentSpeedMultiplier,

		// Таймеры поведения
		MinDirectionTime: WolfMinDirectionTime,
		MaxDirectionTime: WolfMaxDirectionTime,

		// Боевые характеристики
		AttackDamage:   WolfAttackDamageDefault,
		AttackCooldown: WolfAttackCooldown,
		HitChance:      WolfHitChance,
	}
}
