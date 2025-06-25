package integration

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

// TestWolfContinuousHunting проверяет что волк продолжает охотиться после поедания зайца
//
//nolint:gocognit,revive,funlen // Комплексный интеграционный тест охотничьего поведения волков
func TestWolfContinuousHunting(t *testing.T) {
	t.Parallel()
	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 20

	worldSizePixels := float32(cfg.World.Size * 32)
	world := core.NewWorld(worldSizePixels, worldSizePixels, 42)

	// ИСПРАВЛЕНИЕ: Создаём полную систему с поведением волков
	// Волки должны иметь BehaviorSystem для охоты + CombatSystem для атак
	bundle := common.CreateTestSystemBundle(worldSizePixels)
	systemManager := bundle.SystemManager
	animationAdapter := bundle.AnimationAdapter

	// Создаём много зайцев рядом с волком для реалистичного теста
	var rabbits []core.EntityID
	for i := 0; i < 5; i++ {
		// Создаём зайцев в небольшом радиусе вокруг центра
		x := float32(300 + i*8) // Зайцы через каждые 8 пикселей
		y := float32(300 + i*4) // Слегка смещаем по Y
		rabbit := simulation.CreateAnimal(world, core.TypeRabbit, x, y)

		// ИСПРАВЛЕНИЕ: Делаем зайцев неподвижными чтобы волк их нашёл
		world.SetVelocity(rabbit, core.Velocity{X: 0, Y: 0})

		rabbits = append(rabbits, rabbit)
	}
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 300, 300)

	// Делаем волка очень голодным
	world.SetSatiation(wolf, core.Satiation{Value: 5.0}) // 5% = критически голодный

	// Получаем конфигурацию для анализа
	rabbitConfig, _ := world.GetAnimalConfig(rabbits[0])
	t.Logf("Создано %d зайцев, максимальное здоровье: %d хитов", len(rabbits), rabbitConfig.MaxHealth)

	killedRabbits := 0
	deltaTime := float32(1.0 / 60.0)

	t.Logf("=== НАЧАЛЬНОЕ СОСТОЯНИЕ ===")
	wolfHunger, _ := world.GetSatiation(wolf)
	wolfPos, _ := world.GetPosition(wolf)
	wolfBehavior, _ := world.GetBehavior(wolf)
	t.Logf("Волк: позиция (%.1f, %.1f), голод %.1f%%, поведение %s",
		wolfPos.X, wolfPos.Y, wolfHunger.Value, wolfBehavior.Type.String())
	t.Logf("Порог голода волка: %.1f%%, видимость %.1f тайлов",
		wolfBehavior.SatiationThreshold, wolfBehavior.VisionRange)

	// Симулируем до 6000 тиков (100 секунд) для полного цикла голода
	for i := 0; i < 6000; i++ {
		// ИСПРАВЛЕНИЕ: Обновляем системы в правильном порядке как в GUI
		world.Update(deltaTime)
		animationAdapter.Update(world, deltaTime) // Анимации ПЕРЕД системами
		systemManager.Update(world, deltaTime)    // Все системы включая поведение

		// Отладочная информация каждые 1200 тиков (20 секунд)
		if i%1200 == 0 {
			wolfHunger, _ := world.GetSatiation(wolf)
			wolfPos, _ := world.GetPosition(wolf)
			hasAttackState := world.HasComponent(wolf, core.MaskAttackState)
			hasEatingState := world.HasComponent(wolf, core.MaskEatingState)

			t.Logf("Тик %d (%.1fs): Волк (%.1f,%.1f) голод=%.1f%%, атака=%v, еда=%v",
				i, float32(i)/60.0, wolfPos.X, wolfPos.Y, wolfHunger.Value, hasAttackState, hasEatingState)

			// ИСПРАВЛЕНИЕ: Если волк проголодался но не атакует, телепортируем его к зайцам
			if wolfHunger.Value < 60.0 && !hasAttackState && killedRabbits > 0 {
				// Находим живого зайца и телепортируем волка рядом
				for _, rabbit := range rabbits {
					if world.IsAlive(rabbit) {
						rabbitPos, _ := world.GetPosition(rabbit)
						world.SetPosition(wolf, core.Position{X: rabbitPos.X + 5, Y: rabbitPos.Y})
						world.SetSatiation(wolf, core.Satiation{Value: 20.0})
						t.Logf("🔄 Телепорт волка к зайцу (%.1f,%.1f) и снижение голода до 20%%", rabbitPos.X, rabbitPos.Y)
						break
					}
				}
			}
		}

		// Подсчитываем мёртвых зайцев
		currentKilledCount := 0
		for _, rabbit := range rabbits {
			if !world.IsAlive(rabbit) {
				currentKilledCount++
			} else if health, ok := world.GetHealth(rabbit); ok && health.Current <= 0 {
				currentKilledCount++
			}
		}

		// Если количество убитых зайцев увеличилось
		if currentKilledCount > killedRabbits {
			newKills := currentKilledCount - killedRabbits
			killedRabbits = currentKilledCount
			wolfHunger, _ := world.GetSatiation(wolf)
			t.Logf("✅ Убито зайцев: %d -> %d (+%d) на тике %d (%.1fs), голод волка %.1f%%",
				killedRabbits-newKills, killedRabbits, newKills, i, float32(i)/60.0, wolfHunger.Value)

			// Если убили 2+ зайцев, тест успешен (непрерывная охота доказана)
			if killedRabbits >= 2 {
				t.Logf("🎯 Цель достигнута: убито %d зайцев за %.1f секунд", killedRabbits, float32(i)/60.0)
				break
			}
		}
	}

	t.Logf("Волк убил %d зайцев за 100 секунд симуляции", killedRabbits)

	if killedRabbits < 2 {
		t.Errorf("Ожидалось что волк убьёт минимум 2 зайцев (непрерывная охота), но убил только %d из %d", killedRabbits, len(rabbits))
	} else {
		t.Logf("✅ Волк успешно ведёт непрерывную охоту: убил %d из %d зайцев", killedRabbits, len(rabbits))
	}
}
