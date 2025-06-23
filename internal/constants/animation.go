package constants

// AnimationType тип анимации - перемещено сюда для избежания зависимостей от GUI
type AnimationType uint8

const (
	AnimIdle AnimationType = iota
	AnimWalk
	AnimRun
	AnimDeathDying
	AnimDeathDecay
	AnimEat
	AnimAttack
)

// String возвращает название анимации
func (at AnimationType) String() string {
	switch at {
	case AnimIdle:
		return "Idle"
	case AnimWalk:
		return "Walk"
	case AnimRun:
		return "Run"
	case AnimDeathDying:
		return "DeathDying"
	case AnimDeathDecay:
		return "DeathDecay"
	case AnimEat:
		return "Eat"
	case AnimAttack:
		return "Attack"
	default:
		return "Unknown"
	}
}
