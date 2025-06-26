package simulation

import "github.com/aiseeq/savanna/internal/core"

// DefaultConfigFactory создаёт базовую конфигурацию для неизвестных типов животных (Factory Pattern)
// Соблюдает принципы SRP и OCP
type DefaultConfigFactory struct{}

// NewDefaultConfigFactory создаёт новый factory для базовой конфигурации
func NewDefaultConfigFactory() *DefaultConfigFactory {
	return &DefaultConfigFactory{}
}

// CreateConfig создаёт базовую конфигурацию для неизвестных типов
func (f *DefaultConfigFactory) CreateConfig() core.AnimalConfig {
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
