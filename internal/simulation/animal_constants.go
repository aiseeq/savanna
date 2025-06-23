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
	// УДАЛЕНЫ: RabbitRadius, WolfRadius - не используются
	RabbitSpeed = RabbitBaseSpeed // DEPRECATED: используйте RabbitBaseSpeed
	WolfSpeed   = WolfBaseSpeed   // DEPRECATED: используйте WolfBaseSpeed
)

// Производные константы (вычисляемые от базовых) - ПОКА НЕ ИСПОЛЬЗУЮТСЯ
// УДАЛЕНЫ: WolfAttackRange, VisionRangeRabbit, VisionRangeWolf

// Пороги поведения (алиасы от game_balance.go)
const (
	RabbitHungryThreshold = RabbitHungerThreshold // DEPRECATED: используйте RabbitHungerThreshold
)

// === МИГРАЦИОННОЕ РУКОВОДСТВО ===
//
// УДАЛЕННЫЕ КОНСТАНТЫ (замените в коде):
// Вместо:  simulation.RabbitRadius    → simulation.RabbitBaseRadius
// Вместо:  simulation.WolfRadius      → simulation.WolfBaseRadius
// Вместо:  simulation.WolfAttackRange → simulation.WolfBaseRadius * simulation.WolfAttackRangeMultiplier
// Вместо:  simulation.VisionRangeRabbit → simulation.RabbitBaseRadius * simulation.RabbitVisionMultiplier
// Вместо:  simulation.VisionRangeWolf   → simulation.WolfBaseRadius * simulation.WolfVisionMultiplier
//
// ОСТАВШИЕСЯ DEPRECATED КОНСТАНТЫ (требуют миграции):
// Вместо:  simulation.RabbitSpeed → simulation.RabbitBaseSpeed
// Вместо:  simulation.WolfSpeed   → simulation.WolfBaseSpeed
// Вместо:  simulation.RabbitHungryThreshold → simulation.RabbitHungerThreshold
//
// Все основные константы централизованы в game_balance.go для единого управления балансом.
