package core

// Файл interface_check.go проверяет что World реализует упрощённые интерфейсы
// Компилятор проверит соответствие на этапе сборки

// Статические проверки интерфейсов (проверяются на этапе компиляции)
var (
	// ПРИНЦИПЫ SOLID: проверка соответствия специализированных интерфейсов
	_ ECSAccess            = (*World)(nil)
	_ SimulationAccess     = (*World)(nil)
	_ MovementSystemAccess = (*World)(nil)
	_ BehaviorSystemAccess = (*World)(nil)
	_ CombatSystemAccess   = (*World)(nil)

	// ISP УЛУЧШЕНИЯ: узкоспециализированные интерфейсы для конкретных систем
	_ SatiationSystemAccess              = (*World)(nil)
	_ GrassSearchSystemAccess            = (*World)(nil)
	_ StarvationDamageSystemAccess       = (*World)(nil)
	_ SatiationSpeedModifierSystemAccess = (*World)(nil)
)
