package core

// Файл interface_check.go проверяет что World реализует упрощённые интерфейсы
// Компилятор проверит соответствие на этапе сборки

// Статические проверки интерфейсов (проверяются на этапе компиляции)
var (
	// Основные упрощённые интерфейсы
	_ ECSAccess        = (*World)(nil)
	_ SimulationAccess = (*World)(nil)

	// Legacy алиасы для обратной совместимости
	_ MovementSystemAccess = (*World)(nil)
	_ FeedingSystemAccess  = (*World)(nil)
	_ BehaviorSystemAccess = (*World)(nil)
	_ AttackSystemAccess   = (*World)(nil)
	_ EatingSystemAccess   = (*World)(nil)
	_ DamageSystemAccess   = (*World)(nil)
	_ CorpseSystemAccess   = (*World)(nil)
	_ WorldAccess          = (*World)(nil)
)
