package animation

import (
	"math"

	"github.com/aiseeq/savanna/internal/core"
)

// AnimationResolver разрешает типы анимаций для животных
type AnimationResolver struct{}

// NewAnimationResolver создаёт новый резолвер анимаций
func NewAnimationResolver() *AnimationResolver {
	return &AnimationResolver{}
}

// ResolveAnimalAnimationType определяет тип анимации для животного (устраняет дублирование логики)
func (ar *AnimationResolver) ResolveAnimalAnimationType(
	world *core.World,
	entity core.EntityID,
	animalType core.AnimalType,
) AnimationType {
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

	// Вычисляем квадрат скорости
	speedSquared := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speedSquared < SpeedThresholds.Idle {
		return AnimIdle
	} else if speedSquared < SpeedThresholds.WolfWalk {
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

	// Вычисляем квадрат скорости
	speedSquared := velocity.X*velocity.X + velocity.Y*velocity.Y

	if speedSquared < SpeedThresholds.Idle {
		return AnimIdle
	} else if speedSquared < SpeedThresholds.RabbitWalk {
		return AnimWalk
	} else {
		return AnimRun
	}
}

// isWolfAttacking проверяет атакует ли волк
func (ar *AnimationResolver) isWolfAttacking(world *core.World, wolf core.EntityID) bool {
	satiation, hasSatiation := world.GetSatiation(wolf)
	if !hasSatiation || satiation.Value > AttackThresholds.WolfSatiation {
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

	// Вычисляем расстояние
	dx := pos.X - rabbitPos.X
	dy := pos.Y - rabbitPos.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	return distance <= AttackThresholds.WolfAttackDistance
}

// Константы порогов скорости для анимаций (устраняет магические числа)
const (
	// УМЕНЬШЕНИЕ В 20 РАЗ: соответствует уменьшенным скоростям животных
	// Базовая скорость зайца теперь 0.15, волка 0.2 тайла/сек
	// С учетом множителей поведения (0.3-1.0) реальная скорость 0.05-0.2 тайла/сек
	IdleSpeedThreshold       = 0.005 // Порог покоя (было 0.1, теперь 0.005)
	WolfWalkSpeedThreshold   = 0.2   // Порог ходьбы волка (было 4.0, теперь 0.2)
	RabbitWalkSpeedThreshold = 0.11  // Порог ходьбы зайца (было 2.25, теперь 0.11)

	// Пороги для логики атак
	WolfSatiationAttackThreshold = 60.0 // Волк атакует если сытость < 60%
	WolfSearchRadius             = 15.0 // Радиус поиска добычи для атаки
	WolfAttackDistance           = 12.0 // Дистанция атаки волка
)

// SpeedThresholds константы порогов скорости для анимаций
var SpeedThresholds = struct {
	Idle       float32
	WolfWalk   float32
	RabbitWalk float32
}{
	Idle:       IdleSpeedThreshold,       // 0.005 (уменьшено в 20 раз)
	WolfWalk:   WolfWalkSpeedThreshold,   // 0.2 (уменьшено в 20 раз)
	RabbitWalk: RabbitWalkSpeedThreshold, // 0.11 (уменьшено в ~20 раз)
}

// AttackThresholds константы для логики атак
var AttackThresholds = struct {
	WolfSatiation      float32
	WolfSearchRadius   float32
	WolfAttackDistance float32
}{
	WolfSatiation:      WolfSatiationAttackThreshold,
	WolfSearchRadius:   WolfSearchRadius,
	WolfAttackDistance: WolfAttackDistance,
}
