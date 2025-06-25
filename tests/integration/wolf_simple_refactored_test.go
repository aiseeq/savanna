package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/tests/common"
)

// TestWolfEatsRabbitRefactored - рефакторинговая версия TestWolfEatsRabbit
// Демонстрирует использование новой общей тестовой инфраструктуры
// Сокращение кода: с 80+ строк до 20 строк
func TestWolfEatsRabbitRefactored(t *testing.T) {
	t.Parallel()
	// Проверяем что боевая система работает с рефакторенной инфраструктурой

	// ДО: 15+ строк дублированного кода создания мира и систем
	// ПОСЛЕ: 4 строки с Builder Pattern + анимации для тестов боя
	world, systemBundle, entities := common.NewTestWorld().
		WithSize(common.MediumWorldSize).
		AddRabbit(300, 300, common.SatedPercentage, common.RabbitMaxHealth).
		AddWolfNearRabbit(common.CloseDistance, common.VeryHungryPercentage).
		BuildWithAnimations()

	rabbit := entities.Rabbits[0]
	wolf := entities.Wolves[0]

	// ДО: 10+ строк проверок начального состояния
	// ПОСЛЕ: 2 строки с готовыми утилитами
	common.LogEntityState(t, world, rabbit, "Заяц")
	common.LogEntityState(t, world, wolf, "Волк")

	// ДО: 20+ строк симуляции с ручными циклами
	// ПОСЛЕ: 1 строка С АНИМАЦИЯМИ (исправление для тестов боя)
	common.RunSimulationWithAnimations(world, systemBundle, common.FiveSecondTicks)

	// ДО: 15+ строк проверок результата
	// ПОСЛЕ: 3 строки с готовыми assert'ами

	// Заяц должен быть убит (проверяем создание трупа)
	corpseCount := world.CountEntitiesWith(core.MaskCorpse)
	if corpseCount == 0 {
		t.Error("Заяц должен умереть и стать трупом, но трупов нет")
	} else {
		t.Logf("Заяц успешно убит: создано %d трупов", corpseCount)
	}

	common.AssertWolfFed(t, world, wolf, common.VeryHungryPercentage)
	common.LogSimulationSummary(t, world, entities, common.FiveSecondTicks)
}

// TestWolfIgnoresSatedRabbit - новый тест стал проще писать
func TestWolfIgnoresSatedRabbit(t *testing.T) {
	t.Parallel()

	// Сытый волк не должен атаковать
	world, systems, entities := common.NewTestWorld().
		AddSatedRabbit().
		AddSatedWolf().
		Build()

	rabbit := entities.Rabbits[0]

	common.RunSimulation(world, systems, common.TenSecondTicks)

	// Заяц должен остаться невредимым
	common.AssertEntityHealth(t, world, rabbit, common.RabbitMaxHealth, "Сытый волк не должен атаковать")
	common.AssertEntityAlive(t, world, rabbit, "Заяц должен быть жив")
}

// TestSingleWolfKillRabbit - упрощённый тест одного волка убивающего зайца
func TestSingleWolfKillRabbit(t *testing.T) {
	t.Parallel()
	// Проверяем что боевая система работает с рефакторенной инфраструктурой

	// Один голодный волк рядом с зайцем
	world, systemBundle, entities := common.NewTestWorld().
		WithSmallSize().
		AddRabbit(100, 100, common.SatedPercentage, common.RabbitMaxHealth).
		AddWolf(102, 100, common.VeryHungryPercentage). // Очень близко
		BuildWithAnimations()

	rabbit := entities.Rabbits[0]
	wolf := entities.Wolves[0]

	// Проверяем начальное состояние
	t.Logf("Начальное состояние:")
	common.LogEntityState(t, world, rabbit, "Заяц")
	common.LogEntityState(t, world, wolf, "Волк")

	common.RunSimulationWithAnimations(world, systemBundle, common.FiveSecondTicks)

	// Заяц должен умереть (проверяем создание трупа)
	corpseCount := world.CountEntitiesWith(core.MaskCorpse)
	if corpseCount == 0 {
		t.Error("Заяц должен умереть и стать трупом, но трупов нет")
	} else {
		t.Logf("Заяц успешно убит: создано %d трупов", corpseCount)
	}

	// Волк должен поесть
	hunger, hasHunger := world.GetSatiation(wolf)
	if hasHunger && hunger.Value > common.VeryHungryPercentage {
		t.Logf("Волк поел: голод вырос с %.1f до %.1f", common.VeryHungryPercentage, hunger.Value)
	} else {
		t.Errorf("Волк не поел: голод остался %.1f", hunger.Value)
	}
}
