package simulation

import "github.com/aiseeq/savanna/internal/core"

// WolfConfigFactory создаёт конфигурацию волка (Factory Pattern)
// Соблюдает принципы SRP и OCP
type WolfConfigFactory struct{}

// NewWolfConfigFactory создаёт новый factory для волков
func NewWolfConfigFactory() *WolfConfigFactory {
	return &WolfConfigFactory{}
}

// CreateConfig создаёт конфигурацию волка
// Все параметры выводятся от базового радиуса через явные множители из game_balance.go
func (f *WolfConfigFactory) CreateConfig() core.AnimalConfig {
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
