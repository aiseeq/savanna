package animation

import "github.com/aiseeq/savanna/internal/core"

// AnimationResolver разрешает типы анимаций для животных (устраняет дублирование между GUI и headless)
type AnimationResolver struct{}

// NewAnimationResolver создаёт новый резолвер анимаций
func NewAnimationResolver() *AnimationResolver {
	return &AnimationResolver{}
}

// ResolveAnimalAnimationType определяет тип анимации для животного (устраняет дублирование логики)
func (ar *AnimationResolver) ResolveAnimalAnimationType(world *core.World, entity core.EntityID, animalType core.AnimalType) AnimationType {
	switch animalType {
	case core.TypeWolf:
		return ar.resolveWolfAnimationType(world, entity)
	case core.TypeRabbit:
		return ar.resolveRabbitAnimationType(world, entity)
	default:
		return AnimIdle
	}
}

// resolveWolfAnimationType определяет анимацию волка по приоритетам
func (ar *AnimationResolver) resolveWolfAnimationType(world *core.World, entity core.EntityID) AnimationType {
	// ПРИОРИТЕТ 1: Если волк ест
	if world.HasComponent(entity, core.MaskEatingState) {
		return AnimEat
	}

	// ПРИОРИТЕТ 2: Если волк атакует
	if ar.isWolfAttacking(world, entity) {
		return AnimAttack
	}

	// ПРИОРИТЕТ 3: Движение
	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speed < SpeedThresholds.Idle {
		return AnimIdle
	} else if speed < SpeedThresholds.WolfWalk {
		return AnimWalk
	} else {
		return AnimRun
	}
}

// resolveRabbitAnimationType определяет анимацию зайца по приоритетам
func (ar *AnimationResolver) resolveRabbitAnimationType(world *core.World, entity core.EntityID) AnimationType {
	// ПРИОРИТЕТ 1: Труп
	if world.HasComponent(entity, core.MaskCorpse) {
		return AnimDeathDying
	}

	// ПРИОРИТЕТ 2: Если заяц ест
	if world.HasComponent(entity, core.MaskEatingState) {
		return AnimEat
	}

	// ПРИОРИТЕТ 3: Движение
	velocity, hasVel := world.GetVelocity(entity)
	if !hasVel {
		return AnimIdle
	}

	speed := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speed < SpeedThresholds.Idle {
		return AnimIdle
	} else if speed < SpeedThresholds.RabbitWalk {
		return AnimWalk
	} else {
		return AnimRun
	}
}

// isWolfAttacking проверяет атакует ли волк (общая логика из GUI и headless)
func (ar *AnimationResolver) isWolfAttacking(world *core.World, wolf core.EntityID) bool {
	hunger, hasHunger := world.GetHunger(wolf)
	if !hasHunger || hunger.Value > AttackThresholds.WolfHunger {
		return false
	}

	pos, hasPos := world.GetPosition(wolf)
	if !hasPos {
		return false
	}

	nearestRabbit, foundRabbit := world.FindNearestByType(pos.X, pos.Y, AttackThresholds.WolfSearchRadius, core.TypeRabbit)
	if !foundRabbit {
		return false
	}

	if world.HasComponent(nearestRabbit, core.MaskCorpse) {
		return false
	}

	rabbitPos, hasRabbitPos := world.GetPosition(nearestRabbit)
	if !hasRabbitPos {
		return false
	}

	distance := (pos.X-rabbitPos.X)*(pos.X-rabbitPos.X) + (pos.Y-rabbitPos.Y)*(pos.Y-rabbitPos.Y)
	return distance <= AttackThresholds.WolfAttackDistance*AttackThresholds.WolfAttackDistance
}

// Константы порогов скорости для анимаций (устраняет магические числа)
const (
	// Пороги скорости для переключения анимаций (квадрат скорости для оптимизации)
	IdleSpeedThreshold       = 0.1   // Порог покоя (очень медленное движение)
	WolfWalkSpeedThreshold   = 400.0 // Порог ходьбы волка (20*20 = 400)
	RabbitWalkSpeedThreshold = 300.0 // Порог ходьбы зайца (около 17*17 = 289)

	// Пороги для логики атак
	WolfHungerAttackThreshold = 60.0 // Волк атакует если голод < 60%
	WolfSearchRadius          = 15.0 // Радиус поиска добычи для атаки
	WolfAttackDistance        = 12.0 // Дистанция атаки волка
)

// SpeedThresholds константы порогов скорости для анимаций
var SpeedThresholds = struct {
	Idle       float32
	WolfWalk   float32
	RabbitWalk float32
}{
	Idle:       IdleSpeedThreshold,
	WolfWalk:   WolfWalkSpeedThreshold,
	RabbitWalk: RabbitWalkSpeedThreshold,
}

// AttackThresholds константы для логики атак
var AttackThresholds = struct {
	WolfHunger         float32
	WolfSearchRadius   float32
	WolfAttackDistance float32
}{
	WolfHunger:         WolfHungerAttackThreshold,
	WolfSearchRadius:   WolfSearchRadius,
	WolfAttackDistance: WolfAttackDistance,
}
