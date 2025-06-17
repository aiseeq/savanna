package core

// Файл interface_check.go проверяет что World реализует все специализированные интерфейсы
// Компилятор проверит соответствие на этапе сборки

// Статические проверки интерфейсов (проверяются на этапе компиляции)
var (
	// Базовые интерфейсы
	_ PositionAccess   = (*World)(nil)
	_ MovementAccess   = (*World)(nil)
	_ HealthAccess     = (*World)(nil)
	_ HungerAccess     = (*World)(nil)
	_ SizeAccess       = (*World)(nil)
	_ AnimationAccess  = (*World)(nil)
	_ BehaviorAccess   = (*World)(nil)
	_ AnimalTypeAccess = (*World)(nil)
	_ ECSCore          = (*World)(nil)
	_ SpatialQueries   = (*World)(nil)
	_ RandomAccess     = (*World)(nil)
	_ EntityManagement = (*World)(nil)

	// Состояния
	_ CombatStateAccess = (*World)(nil)
	_ EatingStateAccess = (*World)(nil)
	_ CorpseAccess      = (*World)(nil)
	_ ComponentRemoval  = (*World)(nil)

	// Композитные интерфейсы для систем
	_ MovementSystemAccess = (*World)(nil)
	_ FeedingSystemAccess  = (*World)(nil)
	_ BehaviorSystemAccess = (*World)(nil)
	_ AttackSystemAccess   = (*World)(nil)
	_ EatingSystemAccess   = (*World)(nil)
	_ DamageSystemAccess   = (*World)(nil)
	_ CorpseSystemAccess   = (*World)(nil)
	_ WorldAccess          = (*World)(nil)
)
