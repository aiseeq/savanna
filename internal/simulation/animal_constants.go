package simulation

// animal_constants.go - Алиасы для обратной совместимости (DRY рефакторинг)
//
// ПРИНЦИП DRY: Все реальные константы теперь в game_balance.go
// ПРИНЦИП OCP: Существующий код продолжает работать без изменений
//
// РЕКОМЕНДАЦИЯ: В новом коде используйте константы напрямую из game_balance.go

// === АЛИАСЫ ДЛЯ ОБРАТНОЙ СОВМЕСТИМОСТИ ===

// Размеры и скорости животных (алиасы от game_balance.go)
const (
	RabbitRadius = RabbitBaseRadius // DEPRECATED: используйте RabbitBaseRadius
	RabbitSpeed  = RabbitBaseSpeed  // DEPRECATED: используйте RabbitBaseSpeed
	WolfRadius   = WolfBaseRadius   // DEPRECATED: используйте WolfBaseRadius
	WolfSpeed    = WolfBaseSpeed    // DEPRECATED: используйте WolfBaseSpeed
)

// Производные константы (вычисляемые от базовых)
const (
	WolfAttackRange   = WolfBaseRadius * WolfAttackRangeMultiplier // DEPRECATED
	VisionRangeRabbit = RabbitBaseRadius * RabbitVisionMultiplier  // DEPRECATED
	VisionRangeWolf   = WolfBaseRadius * WolfVisionMultiplier      // DEPRECATED
)

// Пороги поведения (алиасы от game_balance.go)
const (
	RabbitHungryThreshold = RabbitHungerThreshold // DEPRECATED: используйте RabbitHungerThreshold
)

// === МИГРАЦИОННОЕ РУКОВОДСТВО ===
//
// Вместо:  simulation.RabbitRadius
// Теперь:  simulation.RabbitBaseRadius
//
// Вместо:  simulation.WolfAttackRange
// Теперь:  simulation.WolfBaseRadius * simulation.WolfAttackRangeMultiplier
//
// Вместо:  simulation.VisionRangeRabbit
// Теперь:  simulation.RabbitBaseRadius * simulation.RabbitVisionMultiplier
//
// Все основные константы централизованы в game_balance.go для единого управления балансом.
