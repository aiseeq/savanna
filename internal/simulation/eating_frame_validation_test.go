package simulation

import (
	"testing"

	"github.com/aiseeq/savanna/internal/constants"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
)

func TestEatingFrameValidation_DiscreteNutrition(t *testing.T) {
	// Создаем мир и системы
	world := core.NewWorld(10, 10, 12345)

	// Создаем terrain с травой
	terrain := &generator.Terrain{
		Width:  10,
		Height: 10,
		Size:   10, // ИСПРАВЛЕНИЕ: нужно установить Size для GetSize()
		Tiles:  make([][]generator.TileType, 10),
		Grass:  make([][]float32, 10),
	}
	for y := 0; y < 10; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 10)
		terrain.Grass[y] = make([]float32, 10)
		for x := 0; x < 10; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	_ = NewGrassSearchSystem(vegetationSystem)
	grassEatingSystem := NewGrassEatingSystem(vegetationSystem)

	// Создаем зайца
	rabbit := CreateAnimal(world, core.TypeRabbit, 5, 5)

	// Устанавливаем голод чтобы заяц захотел есть
	world.SetSatiation(rabbit, core.Satiation{Value: 50.0}) // Голодный

	// Добавляем состояние поедания травы
	eatingState := core.EatingState{
		Target:          GrassEatingTarget,
		TargetType:      core.EatingTargetGrass,
		EatingProgress:  0.0,
		NutritionGained: 0.0,
	}
	world.AddEatingState(rabbit, eatingState)

	// Устанавливаем анимацию поедания
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       0, // Начинаем с кадра 0
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	initialHunger, _ := world.GetSatiation(rabbit)
	t.Logf("Initial hunger: %.1f", initialHunger.Value)

	// Тест кадра 0 - НЕ должен давать питательность
	deltaTime := float32(1.0 / 60.0)
	grassEatingSystem.Update(world, deltaTime)

	hungerAfterFrame0, _ := world.GetSatiation(rabbit)
	if hungerAfterFrame0.Value != initialHunger.Value {
		t.Errorf("Frame 0 should NOT give nutrition. Expected hunger %.1f, got %.1f",
			initialHunger.Value, hungerAfterFrame0.Value)
	}
	t.Logf("Frame 0: hunger = %.1f (correct, no nutrition given)", hungerAfterFrame0.Value)

	// Переключаем на кадр 1 - ДОЛЖЕН дать питательность
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       1, // Переходим на кадр 1
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	grassEatingSystem.Update(world, deltaTime)

	hungerAfterFrame1, _ := world.GetSatiation(rabbit)
	expectedNutrition := GrassPerEatingTick * GrassNutritionValue // 1.0 * 2.0 = 2.0
	expectedHunger := initialHunger.Value + float32(expectedNutrition)

	if hungerAfterFrame1.Value != expectedHunger {
		t.Errorf("Frame 0→1 transition should give nutrition. Expected hunger %.1f, got %.1f",
			expectedHunger, hungerAfterFrame1.Value)
	}
	t.Logf("Frame 0→1: hunger = %.1f (correct, +%.1f nutrition)", hungerAfterFrame1.Value, expectedNutrition)

	// Тест кадра 1 (повторно) - НЕ должен давать дополнительную питательность
	grassEatingSystem.Update(world, deltaTime)

	hungerAfterFrame1Again, _ := world.GetSatiation(rabbit)
	if hungerAfterFrame1Again.Value != hungerAfterFrame1.Value {
		t.Errorf("Staying on frame 1 should NOT give additional nutrition. Expected hunger %.1f, got %.1f",
			hungerAfterFrame1.Value, hungerAfterFrame1Again.Value)
	}
	t.Logf("Frame 1 (repeat): hunger = %.1f (correct, no additional nutrition)", hungerAfterFrame1Again.Value)

	// Переключаем обратно на кадр 0 - НЕ должен давать питательность
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       0, // Возвращаемся на кадр 0
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	grassEatingSystem.Update(world, deltaTime)

	hungerAfterBackToFrame0, _ := world.GetSatiation(rabbit)
	if hungerAfterBackToFrame0.Value != hungerAfterFrame1Again.Value {
		t.Errorf("Frame 1→0 transition should NOT give nutrition. Expected hunger %.1f, got %.1f",
			hungerAfterFrame1Again.Value, hungerAfterBackToFrame0.Value)
	}
	t.Logf("Frame 1→0: hunger = %.1f (correct, only 0→1 gives nutrition)", hungerAfterBackToFrame0.Value)

	// Еще один цикл 0→1 для подтверждения
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       1, // Снова переходим на кадр 1
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	grassEatingSystem.Update(world, deltaTime)

	hungerAfterSecondCycle, _ := world.GetSatiation(rabbit)
	expectedHungerSecondCycle := hungerAfterBackToFrame0.Value + float32(expectedNutrition)

	if hungerAfterSecondCycle.Value != expectedHungerSecondCycle {
		t.Errorf("Second 0→1 transition should give nutrition again. Expected hunger %.1f, got %.1f",
			expectedHungerSecondCycle, hungerAfterSecondCycle.Value)
	}
	t.Logf("Second 0→1: hunger = %.1f (correct, +%.1f nutrition again)", hungerAfterSecondCycle.Value, expectedNutrition)

	t.Logf("SUCCESS: Discrete eating works correctly - nutrition only on 0→1 frame transitions")
}

func TestEatingFrameValidation_IntegrationWithSystems(t *testing.T) {
	// Интеграционный тест проверяющий что вся система работает вместе
	world := core.NewWorld(10, 10, 12345)

	// Создаем terrain с травой
	terrain := &generator.Terrain{
		Width:  10,
		Height: 10,
		Size:   10, // ИСПРАВЛЕНИЕ: нужно установить Size для GetSize()
		Tiles:  make([][]generator.TileType, 10),
		Grass:  make([][]float32, 10),
	}
	for y := 0; y < 10; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 10)
		terrain.Grass[y] = make([]float32, 10)
		for x := 0; x < 10; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	grassSearchSystem := NewGrassSearchSystem(vegetationSystem)
	grassEatingSystem := NewGrassEatingSystem(vegetationSystem)

	// Создаем зайца
	rabbit := CreateAnimal(world, core.TypeRabbit, 5, 5)

	// Устанавливаем голод чтобы GrassSearchSystem создал EatingState
	world.SetSatiation(rabbit, core.Satiation{Value: 50.0}) // Голодный

	initialHunger, _ := world.GetSatiation(rabbit)

	// Запускаем GrassSearchSystem - должен создать EatingState
	grassSearchSystem.Update(world, 1.0/60.0)

	// Проверяем что создалось состояние поедания
	if !world.HasComponent(rabbit, core.MaskEatingState) {
		t.Fatal("GrassSearchSystem should create EatingState for hungry rabbit on grass")
	}

	eatingState, _ := world.GetEatingState(rabbit)
	if eatingState.TargetType != core.EatingTargetGrass {
		t.Errorf("EatingState should be for grass, got %d", eatingState.TargetType)
	}

	// Устанавливаем анимацию поедания на кадр 0
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Запускаем GrassEatingSystem на кадре 0 - не должен дать питательность
	grassEatingSystem.Update(world, 1.0/60.0)

	hungerAfterFrame0, _ := world.GetSatiation(rabbit)
	if hungerAfterFrame0.Value != initialHunger.Value {
		t.Errorf("Frame 0 should not give nutrition in integration test")
	}

	// Переключаем на кадр 1 и запускаем GrassEatingSystem
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       1,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	grassEatingSystem.Update(world, 1.0/60.0)

	hungerAfterFrame1, _ := world.GetSatiation(rabbit)
	expectedIncrease := GrassPerEatingTick * GrassNutritionValue

	if hungerAfterFrame1.Value <= initialHunger.Value {
		t.Errorf("Frame 0→1 should increase hunger from %.1f to %.1f, got %.1f",
			initialHunger.Value, initialHunger.Value+float32(expectedIncrease), hungerAfterFrame1.Value)
	}

	t.Logf("SUCCESS: Integration test shows discrete eating works with frame transitions")
	t.Logf("Hunger: %.1f → %.1f (+%.1f)", initialHunger.Value, hungerAfterFrame1.Value,
		hungerAfterFrame1.Value-initialHunger.Value)
}

func TestEatingFrameValidation_SystemOrder(t *testing.T) {
	// Тест проверяет что системы выполняются в правильном порядке
	world := core.NewWorld(10, 10, 12345)

	terrain := &generator.Terrain{
		Width:  10,
		Height: 10,
		Size:   10, // ИСПРАВЛЕНИЕ: нужно установить Size для GetSize()
		Tiles:  make([][]generator.TileType, 10),
		Grass:  make([][]float32, 10),
	}
	for y := 0; y < 10; y++ {
		terrain.Tiles[y] = make([]generator.TileType, 10)
		terrain.Grass[y] = make([]float32, 10)
		for x := 0; x < 10; x++ {
			terrain.Tiles[y][x] = generator.TileGrass
			terrain.Grass[y][x] = 100.0
		}
	}

	vegetationSystem := NewVegetationSystem(terrain)
	grassSearchSystem := NewGrassSearchSystem(vegetationSystem)
	grassEatingSystem := NewGrassEatingSystem(vegetationSystem)

	// Создаем голодного зайца
	rabbit := CreateAnimal(world, core.TypeRabbit, 5, 5)
	world.SetSatiation(rabbit, core.Satiation{Value: 50.0})

	// Шаг 1: GrassSearchSystem создает EatingState
	grassSearchSystem.Update(world, 1.0/60.0)

	if !world.HasComponent(rabbit, core.MaskEatingState) {
		t.Fatal("Step 1: GrassSearchSystem should create EatingState")
	}

	// Шаг 2: GrassEatingSystem обрабатывает дискретное питание
	// Сначала устанавливаем анимацию на кадр 0
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       0,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	// Вызываем GrassEatingSystem на кадре 0 - не должен дать питательность
	grassEatingSystem.Update(world, 1.0/60.0)

	// Теперь переключаем на кадр 1 - должен дать питательность при переходе 0→1
	world.SetAnimation(rabbit, core.Animation{
		CurrentAnim: int(constants.AnimEat),
		Frame:       1,
		Timer:       0,
		Playing:     true,
		FacingRight: true,
	})

	initialHunger, _ := world.GetSatiation(rabbit)
	grassEatingSystem.Update(world, 1.0/60.0)
	finalHunger, _ := world.GetSatiation(rabbit)

	if finalHunger.Value <= initialHunger.Value {
		t.Errorf("Step 2: GrassEatingSystem should increase hunger from %.1f to %.1f",
			initialHunger.Value, finalHunger.Value)
	}

	t.Logf("SUCCESS: System order test passed")
	t.Logf("GrassSearchSystem → EatingState created")
	t.Logf("GrassEatingSystem → Nutrition given (%.1f → %.1f)",
		initialHunger.Value, finalHunger.Value)
}
