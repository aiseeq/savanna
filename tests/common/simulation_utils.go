package common

import (
	"testing"
	"time"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// RunSimulation запускает симуляцию на заданное количество тиков
// Устраняет дублирование кода запуска симуляции в тестах
func RunSimulation(world *core.World, systemManager *core.SystemManager, ticks int) {
	for i := 0; i < ticks; i++ {
		systemManager.Update(world, StandardDeltaTime)
	}
}

// RunSimulationWithAnimations запускает симуляцию с анимациями (для тестов боя)
// ИСПРАВЛЕНИЕ: Анимации обновляются ПЕРЕД системами, как в GUI режиме
func RunSimulationWithAnimations(world *core.World, systemBundle *TestSystemBundle, ticks int) {
	if systemBundle == nil {
		return
	}

	for i := 0; i < ticks; i++ {
		// 1. СНАЧАЛА обновляем анимации (как в GUI режиме)
		if systemBundle.AnimationAdapter != nil {
			systemBundle.AnimationAdapter.Update(world, StandardDeltaTime)
		}

		// 2. ПОТОМ обновляем все игровые системы
		systemBundle.SystemManager.Update(world, StandardDeltaTime)
	}
}

// RunSimulationForDuration запускает симуляцию на заданное время
func RunSimulationForDuration(world *core.World, systemManager *core.SystemManager, duration time.Duration) {
	ticks := int(duration.Seconds() * StandardTPS)
	RunSimulation(world, systemManager, ticks)
}

// RunSimulationUntilCondition запускает симуляцию до выполнения условия или таймаута
func RunSimulationUntilCondition(
	world *core.World,
	systemManager *core.SystemManager,
	condition func() bool,
	maxTicks int,
) int {
	for i := 0; i < maxTicks; i++ {
		if condition() {
			return i
		}
		systemManager.Update(world, StandardDeltaTime)
	}
	return maxTicks
}

// SimulationStep выполняет один шаг симуляции с логированием (для отладки)
func SimulationStep(world *core.World, systemManager *core.SystemManager, tick int, logInfo func(tick int)) {
	if logInfo != nil {
		logInfo(tick)
	}
	systemManager.Update(world, StandardDeltaTime)
}

// WaitForEntityState ждет пока сущность достигнет определенного состояния
func WaitForEntityState(world *core.World, systemManager *core.SystemManager, entity core.EntityID,
	checkState func(core.EntityID) bool, maxTicks int) bool {

	for i := 0; i < maxTicks; i++ {
		if checkState(entity) {
			return true
		}
		systemManager.Update(world, StandardDeltaTime)
	}
	return false
}

// AssertEntityHealth проверяет здоровье сущности
func AssertEntityHealth(t *testing.T, world *core.World, entity core.EntityID, expectedHealth int16, message string) {
	t.Helper()

	health, hasHealth := world.GetHealth(entity)
	if !hasHealth {
		t.Fatalf("%s: у сущности %d нет компонента Health", message, entity)
	}

	if health.Current != expectedHealth {
		t.Errorf("%s: ожидалось здоровье %d, получено %d", message, expectedHealth, health.Current)
	}
}

// HungerTestParams параметры для проверки голода
type HungerTestParams struct {
	Entity         core.EntityID
	ExpectedHunger float32
	Tolerance      float32
	Message        string
}

// AssertEntityHunger проверяет голод сущности
func AssertEntityHunger(t *testing.T, world *core.World, params HungerTestParams) {
	t.Helper()

	hunger, hasHunger := world.GetSatiation(params.Entity)
	if !hasHunger {
		t.Fatalf("%s: у сущности %d нет компонента Hunger", params.Message, params.Entity)
	}

	diff := hunger.Value - params.ExpectedHunger
	if diff < 0 {
		diff = -diff
	}

	if diff > params.Tolerance {
		t.Errorf("%s: ожидался голод %.1f±%.1f, получен %.1f",
			params.Message, params.ExpectedHunger, params.Tolerance, hunger.Value)
	}
}

// AssertEntityDead проверяет что сущность мертва
func AssertEntityDead(t *testing.T, world *core.World, entity core.EntityID, message string) {
	t.Helper()

	if world.IsAlive(entity) {
		t.Errorf("%s: сущность %d должна быть мертва, но жива", message, entity)
	}
}

// AssertEntityAlive проверяет что сущность жива
func AssertEntityAlive(t *testing.T, world *core.World, entity core.EntityID, message string) {
	t.Helper()

	if !world.IsAlive(entity) {
		t.Errorf("%s: сущность %d должна быть жива, но мертва", message, entity)
	}
}

// AssertEntityHasComponent проверяет наличие компонента у сущности
func AssertEntityHasComponent(
	t *testing.T,
	world *core.World,
	entity core.EntityID,
	component core.ComponentMask,
	message string,
) {
	t.Helper()

	if !world.HasComponent(entity, component) {
		t.Errorf("%s: у сущности %d нет компонента %d", message, entity, component)
	}
}

// AssertEntityLacksComponent проверяет отсутствие компонента у сущности
func AssertEntityLacksComponent(
	t *testing.T,
	world *core.World,
	entity core.EntityID,
	component core.ComponentMask,
	message string,
) {
	t.Helper()

	if world.HasComponent(entity, component) {
		t.Errorf("%s: у сущности %d есть компонент %d, но не должно быть", message, entity, component)
	}
}

// AssertRabbitDamaged проверяет что заяц получил урон (стандартная проверка для тестов боя)
func AssertRabbitDamaged(t *testing.T, world *core.World, rabbit core.EntityID) {
	t.Helper()

	health, hasHealth := world.GetHealth(rabbit)
	if !hasHealth {
		t.Fatal("У зайца нет компонента Health")
	}

	if health.Current >= RabbitMaxHealth {
		t.Errorf("Заяц не получил урон: здоровье %d/%d", health.Current, RabbitMaxHealth)
	}
}

// AssertWolfFed проверяет что волк поел (голод увеличился)
func AssertWolfFed(t *testing.T, world *core.World, wolf core.EntityID, initialHunger float32) {
	t.Helper()

	hunger, hasHunger := world.GetSatiation(wolf)
	if !hasHunger {
		t.Fatal("У волка нет компонента Hunger")
	}

	if hunger.Value <= initialHunger {
		t.Errorf("Волк не поел: голод был %.1f, стал %.1f", initialHunger, hunger.Value)
	}
}

// GetEntityDistance возвращает расстояние между двумя сущностями
func GetEntityDistance(world *core.World, entity1, entity2 core.EntityID) float32 {
	pos1, has1 := world.GetPosition(entity1)
	pos2, has2 := world.GetPosition(entity2)

	if !has1 || !has2 {
		return -1 // Ошибка
	}

	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return dx*dx + dy*dy // Квадрат расстояния для быстрого сравнения
}

// IsWolfAttacking проверяет атакует ли волк (унификация логики из разных тестов)
func IsWolfAttacking(world *core.World, wolf core.EntityID) bool {
	// Проверяем голод волка
	hunger, hasHunger := world.GetSatiation(wolf)
	if !hasHunger || hunger.Value > WolfAttackSatiationThreshold {
		return false
	}

	// Проверяем есть ли состояние атаки
	_, hasAttack := world.GetAttackState(wolf)
	return hasAttack
}

// IsRabbitEating проверяет ест ли заяц траву
func IsRabbitEating(world *core.World, rabbit core.EntityID) bool {
	eatingState, hasEating := world.GetEatingState(rabbit)
	if !hasEating {
		return false
	}

	// Target = 0 означает поедание травы (не трупа)
	return eatingState.Target == simulation.GrassEatingTarget
}

// LogEntityState выводит подробную информацию о состоянии сущности (для отладки)
func LogEntityState(t *testing.T, world *core.World, entity core.EntityID, entityName string) {
	t.Helper()

	if !world.IsAlive(entity) {
		t.Logf("%s %d: МЕРТВ", entityName, entity)
		return
	}

	// Позиция
	if pos, has := world.GetPosition(entity); has {
		t.Logf("%s %d: позиция (%.1f, %.1f)", entityName, entity, pos.X, pos.Y)
	}

	// Здоровье
	if health, has := world.GetHealth(entity); has {
		t.Logf("%s %d: здоровье %d/%d", entityName, entity, health.Current, health.Max)
	}

	// Голод
	if hunger, has := world.GetSatiation(entity); has {
		t.Logf("%s %d: голод %.1f%%", entityName, entity, hunger.Value)
	}

	// Состояния
	if _, has := world.GetAttackState(entity); has {
		t.Logf("%s %d: АТАКУЕТ", entityName, entity)
	}

	if eatingState, has := world.GetEatingState(entity); has {
		if eatingState.Target == 0 {
			t.Logf("%s %d: ЕСТ ТРАВУ", entityName, entity)
		} else {
			t.Logf("%s %d: ЕСТ ТРУП %d", entityName, entity, eatingState.Target)
		}
	}
}

// LogSimulationSummary выводит сводку по симуляции
func LogSimulationSummary(t *testing.T, world *core.World, entities TestEntities, ticks int) {
	t.Helper()

	t.Logf("=== СВОДКА СИМУЛЯЦИИ ПОСЛЕ %d ТИКОВ ===", ticks)

	aliveRabbits := 0
	aliveWolves := 0

	for _, rabbit := range entities.Rabbits {
		if world.IsAlive(rabbit) {
			aliveRabbits++
		}
	}

	for _, wolf := range entities.Wolves {
		if world.IsAlive(wolf) {
			aliveWolves++
		}
	}

	t.Logf("Живых зайцев: %d/%d", aliveRabbits, len(entities.Rabbits))
	t.Logf("Живых волков: %d/%d", aliveWolves, len(entities.Wolves))

	// Количество трупов
	corpseCount := world.CountEntitiesWith(core.MaskCorpse)
	t.Logf("Трупов: %d", corpseCount)
}
