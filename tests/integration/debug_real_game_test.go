package integration

import (
	"fmt"
	"testing"

	"github.com/aiseeq/savanna/internal/adapters"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
)

// TestDebugRealGame диагностический тест для поиска проблемы в реальной игре
func TestDebugRealGame(t *testing.T) {
	t.Parallel()
	world := core.NewWorld(1600, 1600, 42)

	// Создаём точно такие же системы как в main.go
	systemManager := core.NewSystemManager()
	combatSystem := simulation.NewCombatSystem()
	animalBehaviorSystem := simulation.NewAnimalBehaviorSystem(nil)
	movementSystem := simulation.NewMovementSystem(1600, 1600)

	// Важно: инициализируем анимационные компоненты для животных
	// чтобы боевая система могла проверять анимации

	// Добавляем в том же порядке что в main.go
	systemManager.AddSystem(&adapters.BehaviorSystemAdapter{System: animalBehaviorSystem})
	systemManager.AddSystem(&adapters.MovementSystemAdapter{System: movementSystem})
	systemManager.AddSystem(combatSystem)

	// Создаём животных
	rabbit := simulation.CreateRabbit(world, 800, 800)
	wolf := simulation.CreateWolf(world, 810, 800) // на расстоянии 10 единиц

	// Проверяем начальное расстояние
	wolfPos, _ := world.GetPosition(wolf)
	rabbitPos, _ := world.GetPosition(rabbit)
	initialDistance := (wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)
	t.Logf("Начальное расстояние между волком и зайцем: %.1f (требуется <= 144 для атаки)", initialDistance)

	// Делаем волка очень голодным
	world.SetHunger(wolf, core.Hunger{Value: 10.0})

	t.Logf("=== ДИАГНОСТИКА РЕАЛЬНОЙ ИГРЫ ===")

	initialHealth, _ := world.GetHealth(rabbit)
	t.Logf("Начальное здоровье зайца: %d", initialHealth.Current)

	deltaTime := float32(1.0 / 60.0)

	// Отслеживаем компоненты детально
	for i := 0; i < 300; i++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime)

		// Детальная диагностика каждые 10 тиков
		if i%10 == 0 {
			health, _ := world.GetHealth(rabbit)
			hunger, _ := world.GetHunger(wolf)

			// Проверяем компоненты
			hasDamageFlash := world.HasComponent(rabbit, core.MaskDamageFlash)
			hasCorpse := world.HasComponent(rabbit, core.MaskCorpse)
			hasEatingState := world.HasComponent(wolf, core.MaskEatingState)

			var damageFlashInfo string
			if hasDamageFlash {
				flash, _ := world.GetDamageFlash(rabbit)
				damageFlashInfo = fmt.Sprintf("ЕСТЬ (%.2f сек)", flash.Timer)
			} else {
				damageFlashInfo = "НЕТ"
			}

			var corpseInfo string
			if hasCorpse {
				corpse, _ := world.GetCorpse(rabbit)
				corpseInfo = fmt.Sprintf("ЕСТЬ (%.1f)", corpse.NutritionalValue)
			} else {
				corpseInfo = "НЕТ"
			}

			var eatingInfo string
			if hasEatingState {
				eating, _ := world.GetEatingState(wolf)
				eatingInfo = fmt.Sprintf("ЕСТЬ (цель: %d)", eating.Target)
			} else {
				eatingInfo = "НЕТ"
			}

			t.Logf("Тик %3d: HP %2d, голод %.0f%%, DamageFlash %s, Corpse %s, Eating %s",
				i, health.Current, hunger.Value, damageFlashInfo, corpseInfo, eatingInfo)
		}

		// Проверяем изменения здоровья
		currentHealth, _ := world.GetHealth(rabbit)
		if currentHealth.Current != initialHealth.Current {
			t.Logf("🩸 УРОН на тике %d: %d -> %d", i, initialHealth.Current, currentHealth.Current)

			// КРИТИЧЕСКИЙ МОМЕНТ: проверяем DamageFlash сразу после урона
			if world.HasComponent(rabbit, core.MaskDamageFlash) {
				flash, _ := world.GetDamageFlash(rabbit)
				t.Logf("✅ DamageFlash ЕСТЬ сразу после урона: таймер %.3f", flash.Timer)
			} else {
				t.Logf("❌ DamageFlash НЕТ сразу после урона!")
			}

			initialHealth = currentHealth
		}

		// Проверяем создание трупа
		if !world.HasComponent(rabbit, core.MaskCorpse) && currentHealth.Current == 0 {
			// Заяц умер но труп не создался
			if i > 0 { // не в первый тик
				t.Logf("⚠️ ПРОБЛЕМА: заяц умер но труп не создался на тике %d", i)
			}
		}

		// Если труп создался, проверяем начало поедания
		if world.HasComponent(rabbit, core.MaskCorpse) && !world.HasComponent(wolf, core.MaskEatingState) {
			// Есть труп но нет поедания
			wolfPos, _ := world.GetPosition(wolf)
			rabbitPos, _ := world.GetPosition(rabbit)
			distance := (wolfPos.X-rabbitPos.X)*(wolfPos.X-rabbitPos.X) + (wolfPos.Y-rabbitPos.Y)*(wolfPos.Y-rabbitPos.Y)

			if distance <= 15.0*15.0 { // В радиусе поедания
				t.Logf("⚠️ ПРОБЛЕМА: есть труп рядом (дист %.1f) но поедание не началось на тике %d", distance, i)
			}
		}

		// Если заяц полностью исчез - успех
		if !world.IsAlive(rabbit) {
			t.Logf("🎉 УСПЕХ: заяц полностью съеден на тике %d", i)
			return
		}
	}

	t.Logf("Тест завершён без полного поедания")
}
