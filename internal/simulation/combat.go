package simulation

import (
	"github.com/aiseeq/savanna/internal/core"
)

// Константы боевой системы
const (
	// WOLF_ATTACK_RANGE импортируется из animal.go

	// ПАРАМЕТРЫ АТАК
	AttackCooldownSeconds = 0.2 // Кулдаун между атаками (секунды)
	AttackHitChance       = 0.8 // Шанс попадания атаки (80%)
	WolfAttackDamage      = 25  // Урон от атаки волка

	// ПАРАМЕТРЫ ТРУПОВ
	CorpseNutritionalValue = 200.0 // Питательность трупа зайца (увеличено для достижения порога 80%)
	CorpseDecayTime        = 60.0  // Время разложения трупа (секунды)

	// ПАРАМЕТРЫ ПОЕДАНИЯ (перенесены в game_balance.go)
	// EatingRange, CorpseNutritionPerTick, NutritionToHungerRatio теперь в game_balance.go

	// ПАРАМЕТРЫ УРОНА
	DamageFlashDuration = 0.16 // Длительность эффекта мигания при уроне (сек) - быстрое угасание в 5 раз

	// КЛЮЧЕВЫЕ КАДРЫ АНИМАЦИИ АТАКИ (устраняет магические числа)
	AttackFrameWindup = 0 // Кадр 0: замах перед атакой
	AttackFrameStrike = 1 // Кадр 1: удар и нанесение урона
)

// CombatSystem координирует работу всех боевых систем (устраняет нарушение SRP)
// Применяет паттерн Facade для упрощения использования множества специализированных систем
type CombatSystem struct {
	attackSystem *AttackSystem // Система атак
	eatingSystem *EatingSystem // Система поедания
	corpseSystem *CorpseSystem // Система трупов
	damageSystem *DamageSystem // Система эффектов урона
}

// NewCombatSystem создаёт новую объединенную систему боя
func NewCombatSystem() *CombatSystem {
	return &CombatSystem{
		attackSystem: NewAttackSystem(),
		eatingSystem: NewEatingSystem(),
		corpseSystem: NewCorpseSystem(),
		damageSystem: NewDamageSystem(),
	}
}

// Update обновляет все боевые подсистемы (паттерн Facade)
func (cs *CombatSystem) Update(world *core.World, deltaTime float32) {
	// Порядок важен: сначала атаки, потом эффекты, потом поедание
	cs.attackSystem.Update(world, deltaTime)
	cs.damageSystem.Update(world, deltaTime)
	cs.corpseSystem.Update(world, deltaTime)
	cs.eatingSystem.Update(world, deltaTime)
}
