package simulation

import (
	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/physics"
)

// AttackSystem отвечает ТОЛЬКО за атаки и нанесение урона (устраняет нарушение SRP)
type AttackSystem struct {
	attackCooldowns map[core.EntityID]float32 // Кулдауны атак
}

// NewAttackSystem создаёт новую систему атак
func NewAttackSystem() *AttackSystem {
	return &AttackSystem{
		attackCooldowns: make(map[core.EntityID]float32),
	}
}

// Update обновляет систему атак
func (as *AttackSystem) Update(world *core.World, deltaTime float32) {
	// Обновляем кулдауны атак
	as.updateAttackCooldowns(deltaTime)

	// Обрабатываем атаки хищников
	as.handlePredatorAttacks(world, deltaTime)

	// Очищаем кулдауны мертвых животных
	as.cleanupCooldowns(world)
}

// updateAttackCooldowns обновляет кулдауны атак
func (as *AttackSystem) updateAttackCooldowns(deltaTime float32) {
	for entityID, cooldown := range as.attackCooldowns {
		cooldown -= deltaTime
		if cooldown <= 0 {
			delete(as.attackCooldowns, entityID)
		} else {
			as.attackCooldowns[entityID] = cooldown
		}
	}
}

// handlePredatorAttacks обрабатывает атаки хищников (устраняет нарушение OCP)
// Работает с любыми хищниками через компонент Behavior, а не захардкоженные типы
func (as *AttackSystem) handlePredatorAttacks(world *core.World, deltaTime float32) {
	// Обрабатываем ВСЕХ животных с поведением хищника БЕЗ AttackState
	world.ForEachWith(core.MaskBehavior|core.MaskPosition|core.MaskSize, func(predator core.EntityID) {
		as.tryStartAttack(world, predator)
	})

	// Обрабатываем хищников С AttackState - обновляем фазы атаки
	world.ForEachWith(core.MaskAttackState, func(predator core.EntityID) {
		as.updateAttackState(world, predator, deltaTime)
	})
}

// findAttackTarget ищет цель для атаки (универсальная логика для любых хищников)
// Устраняет нарушение OCP - теперь работает с любыми типами через компонент Behavior
func (as *AttackSystem) findAttackTarget(world *core.World, attacker core.EntityID) core.EntityID {
	attackerPos, hasPos := world.GetPosition(attacker)
	if !hasPos {
		return 0
	}

	// Получаем поведение атакующего
	behavior, hasBehavior := world.GetBehavior(attacker)
	if !hasBehavior {
		return 0
	}

	// Получаем размеры атакующего
	attackerSize, hasAttackerSize := world.GetSize(attacker)
	if !hasAttackerSize || attackerSize.AttackRange <= 0 {
		return 0 // Не хищник или не может атаковать
	}

	// Используем дальность видения из поведения (универсально!)
	searchRadius := behavior.VisionRange

	// ПОИСК ЛЮБЫХ ТРАВОЯДНЫХ (устраняет захардкоженность TypeRabbit)
	// Ищем ближайшее животное с поведением травоядного
	var closestTarget core.EntityID
	var closestDistanceSquared float32 = searchRadius*searchRadius + 1 // За пределами радиуса

	world.ForEachWith(core.MaskBehavior|core.MaskPosition|core.MaskSize, func(candidate core.EntityID) {
		if !as.isValidHerbivoreTarget(world, attacker, candidate) {
			return
		}

		candidatePos, _ := world.GetPosition(candidate)
		// Вычисляем квадрат расстояния
		dx := attackerPos.X - candidatePos.X
		dy := attackerPos.Y - candidatePos.Y
		distanceSquared := dx*dx + dy*dy

		if distanceSquared < closestDistanceSquared {
			closestDistanceSquared = distanceSquared
			closestTarget = candidate
		}
	})

	if closestTarget == 0 {
		return 0
	}

	// Проверяем что цель в радиусе атаки
	targetPos, hasTargetPos := world.GetPosition(closestTarget)
	if !hasTargetPos {
		return 0
	}

	targetSize, hasTargetSize := world.GetSize(closestTarget)
	if !hasTargetSize {
		return 0
	}

	// Проверяем возможность атаки
	if as.isWithinAttackRange(attackerPos, targetPos, attackerSize.AttackRange, targetSize.Radius) {
		return closestTarget
	}

	return 0
}

// isValidHerbivoreTarget проверяет что цель подходит для атаки (снижает сложность)
func (as *AttackSystem) isValidHerbivoreTarget(world *core.World, attacker, candidate core.EntityID) bool {
	if candidate == attacker || world.HasComponent(candidate, core.MaskCorpse) {
		return false
	}

	candidateBehavior, hasBehavior := world.GetBehavior(candidate)
	return hasBehavior && candidateBehavior.Type == core.BehaviorHerbivore
}

// tryStartAttack пытается начать атаку для хищника (упрощена через вспомогательные методы)
func (as *AttackSystem) tryStartAttack(world *core.World, predator core.EntityID) {
	// Валидация базовых условий для атаки
	if !as.canStartAttack(world, predator) {
		return
	}

	// Поиск цели и начало атаки
	target := as.findAttackTarget(world, predator)
	if target != 0 {
		as.startAttackState(world, predator, target)
	}
}

// canStartAttack проверяет может ли хищник начать атаку (извлечено из tryStartAttack)
func (as *AttackSystem) canStartAttack(world *core.World, predator core.EntityID) bool {
	// Проверяем что это хищник
	behavior, ok := world.GetBehavior(predator)
	if !ok || behavior.Type != core.BehaviorPredator {
		return false
	}

	// Проверяем возможность атаковать
	size, hasSize := world.GetSize(predator)
	if !hasSize || size.AttackRange <= 0 {
		return false
	}

	// Проверяем что не атакует уже
	if world.HasComponent(predator, core.MaskAttackState) {
		return false
	}

	// ИСПРАВЛЕНИЕ БАГА: Проверяем что не ест уже (волк должен доесть текущий труп)
	if world.HasComponent(predator, core.MaskEatingState) {
		return false // Волк занят поеданием - не может атаковать
	}

	// Проверяем кулдаун
	if !as.canAttack(predator) {
		return false
	}

	// Проверяем голод
	hunger, hasHunger := world.GetSatiation(predator)
	isHungry := hasHunger && hunger.Value < behavior.SatiationThreshold

	return isHungry
}

// startAttackState создаёт состояние атаки и анимацию (снижает сложность)
func (as *AttackSystem) startAttackState(world *core.World, predator, target core.EntityID) {
	attackState := core.AttackState{
		Target:     target,
		Phase:      core.AttackPhaseWindup,
		PhaseTimer: 0.0,
		TotalTimer: 0.0,
		HasStruck:  false,
	}
	world.AddAttackState(predator, attackState)

	// АВТОМАТИЧЕСКИ устанавливаем анимацию атаки (если есть компонент Animation)
	if world.HasComponent(predator, core.MaskAnimation) {
		world.SetAnimation(predator, core.Animation{
			CurrentAnim: int(constants.AnimAttack),
			Frame:       AttackFrameWindup,
			Timer:       0,
			Playing:     true,
			FacingRight: true, // TODO: определять направление по положению цели
		})
	}
}

// updateAttackState обновляет состояние атаки хищника
func (as *AttackSystem) updateAttackState(world *core.World, predator core.EntityID, deltaTime float32) {
	attackState, hasAttack := world.GetAttackState(predator)
	if !hasAttack {
		return
	}

	// Обновляем таймеры
	attackState.PhaseTimer += deltaTime
	attackState.TotalTimer += deltaTime

	// Получаем анимацию для синхронизации (если есть)
	anim, hasAnim := world.GetAnimation(predator)

	// ИСПРАВЛЕНИЕ: Используем АНИМАЦИЮ как основной механизм нанесения урона
	// Урон наносится на кадре 1 анимации, а не по таймеру
	if hasAnim && as.updateAttackStateWithAnimation(world, predator, &attackState, anim) {
		return // Состояние удалено
	}

	// Резервный механизм по таймерам (если нет анимации)
	if !hasAnim && as.updateAttackStateWithTimer(world, predator, &attackState) {
		return // Состояние удалено
	}

	// ВАЖНО: Сохраняем обновленное состояние в конце, ТОЛЬКО если состояние ещё существует
	if world.HasComponent(predator, core.MaskAttackState) {
		world.SetAttackState(predator, attackState)
	}
}

// updateAttackStateWithAnimation обновляет атаку на основе анимации (возвращает true если удалено)
func (as *AttackSystem) updateAttackStateWithAnimation(
	world *core.World, predator core.EntityID, attackState *core.AttackState, anim core.Animation,
) bool {
	// ИСПРАВЛЕНИЕ: Проверяем анимацию атаки ИЛИ любую анимацию которая завершилась (Playing=false)
	// Это позволяет работать с тестами которые меняют анимацию на другую после завершения атаки
	isAttackAnim := anim.CurrentAnim == int(constants.AnimAttack)
	animFinished := !anim.Playing

	// ГЛАВНАЯ ЛОГИКА: Урон наносится на кадре 1 анимации (только для анимации атаки)
	if isAttackAnim && anim.Frame == AttackFrameStrike && !attackState.HasStruck {
		// Кадр 1 (Strike) - наносим урон
		as.executeStrike(world, predator, attackState.Target)
		attackState.HasStruck = true
		attackState.Phase = core.AttackPhaseStrike
	} else if isAttackAnim && anim.Frame == AttackFrameWindup {
		// Кадр 0 (Windup) - подготовка к удару
		attackState.Phase = core.AttackPhaseWindup
		attackState.HasStruck = false
	}

	// ИСПРАВЛЕНИЕ: Завершаем атаку когда анимация завершилась (Playing = false) ИЛИ анимация сменилась
	if animFinished || (!isAttackAnim && attackState.HasStruck) {
		// Анимация атаки завершена или сменилась после нанесения урона - устанавливаем кулдаун и удаляем состояние
		as.setAttackCooldown(world, predator)
		world.RemoveAttackState(predator)
		return true // Состояние удалено
	}

	// ИСПРАВЛЕНИЕ: Дополнительная проверка по кадрам - анимация атаки имеет только 2 кадра (0,1)
	// Если мы уже нанесли урон на кадре 1 и анимация проиграла достаточно времени, завершаем атаку
	if attackState.HasStruck && attackState.Phase == core.AttackPhaseStrike {
		// Получаем данные анимации для расчета времени
		const AttackFrameCount = 2 // Анимация атаки имеет 2 кадра (из loader.go)

		// Время полной анимации = количество кадров / FPS
		fullAnimationDuration := float32(AttackFrameCount) / AttackAnimationFPS

		// Если прошло достаточно времени для завершения полной анимации, завершаем атаку
		if attackState.TotalTimer >= fullAnimationDuration {
			as.setAttackCooldown(world, predator)
			world.RemoveAttackState(predator)
			return true // Состояние удалено
		}
	}

	// Резервный механизм завершения по таймеру (если анимация зависла)
	if attackState.TotalTimer >= (AttackWindupDuration + AttackStrikeDuration) {
		// Атака завершена - устанавливаем кулдаун и удаляем состояние
		as.setAttackCooldown(world, predator)
		world.RemoveAttackState(predator)
		return true // Состояние удалено
	}

	return false
}

// syncAttackWithAnimation синхронизирует атаку с анимацией (не управляет основной логикой)
func (as *AttackSystem) syncAttackWithAnimation(
	world *core.World, predator core.EntityID, attackState *core.AttackState, anim core.Animation,
) {
	// Синхронизируем кадр анимации с фазой атаки
	switch attackState.Phase {
	case core.AttackPhaseWindup:
		// Убеждаемся что анимация на правильном кадре
		if anim.CurrentAnim == int(constants.AnimAttack) && anim.Frame != AttackFrameWindup {
			anim.Frame = AttackFrameWindup
			world.SetAnimation(predator, anim)
		}

	case core.AttackPhaseStrike:
		// Убеждаемся что анимация на правильном кадре
		if anim.CurrentAnim == int(constants.AnimAttack) && anim.Frame != AttackFrameStrike {
			anim.Frame = AttackFrameStrike
			world.SetAnimation(predator, anim)
		}
	}
}

// updateAttackStateWithTimer обновляет атаку по таймерам (возвращает true если удалено)
func (as *AttackSystem) updateAttackStateWithTimer(
	world *core.World, predator core.EntityID, attackState *core.AttackState,
) bool {
	switch attackState.Phase {
	case core.AttackPhaseWindup:
		if attackState.PhaseTimer >= AttackWindupDuration {
			// Время перейти к удару
			attackState.Phase = core.AttackPhaseStrike
			attackState.PhaseTimer = 0.0
		}

	case core.AttackPhaseStrike:
		// Наносим урон сразу при входе в фазу Strike
		if !attackState.HasStruck {
			as.executeStrike(world, predator, attackState.Target)
			attackState.HasStruck = true
		}

		// Завершаем атаку по таймеру
		if attackState.PhaseTimer >= AttackStrikeDuration {
			// Атака завершена - устанавливаем кулдаун и удаляем состояние
			as.setAttackCooldown(world, predator)
			world.RemoveAttackState(predator)
			return true // Состояние удалено
		}
	}
	return false
}

// executeStrike выполняет удар - наносит урон цели (упрощена через вспомогательные методы)
func (as *AttackSystem) executeStrike(world *core.World, attacker, target core.EntityID) {
	// Валидация цели и позиций
	if !as.validateStrikeTarget(world, target) {
		return
	}

	// Валидация дальности атаки
	if !as.validateStrikeRange(world, attacker, target) {
		return
	}

	// Выполнение атаки с проверкой попадания
	as.performStrikeAttempt(world, attacker, target)
}

// validateStrikeTarget проверяет что цель валидна для атаки (извлечено из executeStrike)
func (as *AttackSystem) validateStrikeTarget(world *core.World, target core.EntityID) bool {
	return world.IsAlive(target) && !world.HasComponent(target, core.MaskCorpse)
}

// validateStrikeRange проверяет что цель в радиусе атаки (извлечено из executeStrike)
func (as *AttackSystem) validateStrikeRange(world *core.World, attacker, target core.EntityID) bool {
	attackerPos, hasAttackerPos := world.GetPosition(attacker)
	targetPos, hasTargetPos := world.GetPosition(target)
	if !hasAttackerPos || !hasTargetPos {
		return false
	}

	attackerSize, hasAttackerSize := world.GetSize(attacker)
	targetSize, hasTargetSize := world.GetSize(target)
	if !hasAttackerSize || !hasTargetSize {
		return false
	}

	return as.isWithinAttackRange(attackerPos, targetPos, attackerSize.AttackRange, targetSize.Radius)
}

// performStrikeAttempt выполняет попытку атаки с проверкой попадания (извлечено из executeStrike)
func (as *AttackSystem) performStrikeAttempt(world *core.World, attacker, target core.EntityID) {
	config, hasConfig := world.GetAnimalConfig(attacker)
	if !hasConfig {
		return // Нет конфигурации - не можем атаковать
	}

	// Проверяем шанс попадания
	rng := world.GetRNG()
	if rng.Float32() < config.HitChance {
		as.dealDamageToTarget(world, attacker, target, config.AttackDamage)
	}
}

// dealDamageToTarget наносит урон цели
func (as *AttackSystem) dealDamageToTarget(world *core.World, attacker, target core.EntityID, damage int16) {
	health, hasHealth := world.GetHealth(target)
	if !hasHealth {
		return
	}

	// Наносим урон
	oldHealth := health.Current
	health.Current -= damage
	if health.Current < 0 {
		health.Current = 0
	}

	world.SetHealth(target, health)

	if damage > 0 {
		_ = oldHealth // Избегаем warning о неиспользуемой переменной
	}

	// Добавляем эффект мигания
	world.AddDamageFlash(target, core.DamageFlash{
		Timer:     DamageFlashDuration,
		Duration:  DamageFlashDuration,
		Intensity: MaxDamageFlashIntensity, // Максимальная интенсивность для лучшей видимости
	})

	// Если цель умерла, превращаем её в труп
	if health.Current == 0 {
		// ИСПРАВЛЕНИЕ: Сначала создаём труп, потом создаём EatingState для трупа
		// createCorpse() возвращает ID новой сущности-трупа
		corpseEntity := CreateCorpseAndGetID(world, target)

		// ИСПРАВЛЕНИЕ КРИТИЧЕСКОГО БАГА: Сбрасываем AttackState при убийстве цели
		// Согласно требованию 3.2.3 из docs/tasks/2025-06-22-rework.md
		world.RemoveAttackState(attacker)

		// ИСПРАВЛЕНИЕ БАГА: Автоматически начинаем поедание ТРУПА АТАКУЮЩИМ
		// Это предотвращает переключение на других целей
		if corpseEntity != 0 && !world.HasComponent(attacker, core.MaskEatingState) {
			world.AddEatingState(attacker, core.EatingState{
				Target:          corpseEntity,               // Атакующий ест ТРУП (новая сущность)
				TargetType:      core.EatingTargetAnimal,    // Тип: поедание животного
				EatingProgress:  constants.InitialProgress,  // Начальный прогресс
				NutritionGained: constants.InitialNutrition, // Начальная питательность
			})
		}
	}
}

// canAttack проверяет может ли животное атаковать (нет кулдауна)
func (as *AttackSystem) canAttack(entity core.EntityID) bool {
	_, hasCooldown := as.attackCooldowns[entity]
	return !hasCooldown
}

// setAttackCooldown устанавливает кулдаун атаки из конфигурации животного
func (as *AttackSystem) setAttackCooldown(world *core.World, entity core.EntityID) {
	// Используем кулдаун из конфигурации животного вместо захардкоженной константы
	if config, hasConfig := world.GetAnimalConfig(entity); hasConfig {
		as.attackCooldowns[entity] = config.AttackCooldown
	} else {
		// Fallback на константу из combat.go если нет конфигурации
		as.attackCooldowns[entity] = AttackCooldownSeconds
	}
}

// cleanupCooldowns очищает кулдауны для мертвых животных
func (as *AttackSystem) cleanupCooldowns(world *core.World) {
	for entityID := range as.attackCooldowns {
		if !world.IsAlive(entityID) {
			delete(as.attackCooldowns, entityID)
		}
	}
}

// isWithinAttackRange проверяет находится ли цель в радиусе атаки (ТИПОБЕЗОПАСНО)
func (as *AttackSystem) isWithinAttackRange(
	attackerPos, targetPos core.Position, attackRange, targetRadius float32,
) bool {
	// ЭЛЕГАНТНАЯ МАТЕМАТИКА: работаем напрямую с позициями
	attackerTilePos := physics.PixelPosition{X: physics.NewPixels(attackerPos.X), Y: physics.NewPixels(attackerPos.Y)}.ToTiles()
	targetTilePos := physics.PixelPosition{X: physics.NewPixels(targetPos.X), Y: physics.NewPixels(targetPos.Y)}.ToTiles()

	// Параметры уже в тайлах
	attackRangeTiles := physics.NewTiles(attackRange)
	targetRadiusTiles := physics.NewTiles(targetRadius)

	// Вычисляем квадрат расстояния между центрами (в тайлах)
	dx := attackerTilePos.X.Sub(targetTilePos.X)
	dy := attackerTilePos.Y.Sub(targetTilePos.Y)
	distanceSquared := dx.Float32()*dx.Float32() + dy.Float32()*dy.Float32()

	// Максимальная дистанция атаки: радиус атаки + радиус цели (в тайлах)
	maxAttackDistance := attackRangeTiles.Add(targetRadiusTiles)

	// Сравниваем квадраты для избежания sqrt
	return distanceSquared <= maxAttackDistance.Float32()*maxAttackDistance.Float32()
}
