package simulation

import "github.com/aiseeq/savanna/internal/core"

// RabbitConfigFactory создаёт конфигурацию зайца (Factory Pattern)
// Соблюдает принципы SRP и OCP
type RabbitConfigFactory struct{}

// NewRabbitConfigFactory создаёт новый factory для зайцев
func NewRabbitConfigFactory() *RabbitConfigFactory {
	return &RabbitConfigFactory{}
}

// CreateConfig создаёт конфигурацию зайца
// Все параметры выводятся от базового радиуса через явные множители из game_balance.go
func (f *RabbitConfigFactory) CreateConfig() core.AnimalConfig {
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
