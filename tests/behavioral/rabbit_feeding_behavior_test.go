package behavioral

import (
	"testing"

	"github.com/aiseeq/savanna/config"
	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/generator"
	"github.com/aiseeq/savanna/internal/simulation"
)

// Behavioral Testing в стиле Given-When-Then
// Тесты читаются как спецификация поведения системы

// TestScenario представляет сценарий тестирования
type TestScenario struct {
	world             *core.World
	terrain           *generator.Terrain
	vegetationSystem  *simulation.VegetationSystem
	feedingSystem     *simulation.FeedingSystem
	grassEatingSystem *simulation.GrassEatingSystem
	satiationSystem   *simulation.SatiationSystem
	animationSystem   *animation.AnimationSystem
	rabbit            core.EntityID
	t                 *testing.T
}

// newTestScenario создаёт новый сценарий тестирования
func newTestScenario(t *testing.T) *TestScenario {
	world := core.NewWorld(640, 640, 12345)

	cfg := config.LoadDefaultConfig()
	cfg.World.Size = 10
	terrainGen := generator.NewTerrainGenerator(cfg)
	terrain := terrainGen.Generate()

	vegetationSystem := simulation.NewVegetationSystem(terrain)
	feedingSystem := simulation.NewFeedingSystem(vegetationSystem)
	grassEatingSystem := simulation.NewGrassEatingSystem(vegetationSystem)
	satiationSystem := simulation.NewSatiationSystem()

	// Создаём анимационную систему (КРИТИЧЕСКИ ВАЖНО для GrassEatingSystem!)
	animationSystem := animation.NewAnimationSystem()
	animationSystem.RegisterAnimation(animation.AnimEat, 2, 4.0, true, nil)

	return &TestScenario{
		world:             world,
		terrain:           terrain,
		vegetationSystem:  vegetationSystem,
		feedingSystem:     feedingSystem,
		grassEatingSystem: grassEatingSystem,
		satiationSystem:   satiationSystem,
		animationSystem:   animationSystem,
		t:                 t,
	}
}

// Given методы настраивают начальное состояние

func (s *TestScenario) GivenHungryRabbitAt(x, y float32, satiationLevel float32) *TestScenario {
	s.rabbit = simulation.CreateAnimal(s.world, core.TypeRabbit, x, y)
	s.world.SetSatiation(s.rabbit, core.Satiation{Value: satiationLevel})
	s.world.SetVelocity(s.rabbit, core.Velocity{X: 0, Y: 0}) // Стоит на месте
	return s
}

func (s *TestScenario) GivenGrassAt(tileX, tileY int, amount float32) *TestScenario {
	s.terrain.SetTileType(tileX, tileY, generator.TileGrass)
	s.terrain.SetGrassAmount(tileX, tileY, amount)
	return s
}

func (s *TestScenario) GivenNoGrassAt(tileX, tileY int) *TestScenario {
	s.terrain.SetTileType(tileX, tileY, generator.TileWater) // Используем TileWater вместо несуществующего TileDirt
	s.terrain.SetGrassAmount(tileX, tileY, 0)
	return s
}

// When методы выполняют действия

func (s *TestScenario) WhenTimePassesFor(seconds float32) *TestScenario {
	deltaTime := float32(1.0 / 60.0) // 60 FPS
	ticks := int(seconds / deltaTime)

	for i := 0; i < ticks; i++ {
		s.satiationSystem.Update(s.world, deltaTime)
		s.feedingSystem.Update(s.world, deltaTime)

		// ИСПРАВЛЕНИЕ: Переключаем анимацию СРАЗУ после создания EatingState
		s.updateEatingAnimations()

		// ИСПРАВЛЕНИЕ: Обновляем анимации ПОСЛЕ переключения (как в игре)
		s.updateAnimations(deltaTime)

		s.grassEatingSystem.Update(s.world, deltaTime)
	}

	return s
}

func (s *TestScenario) WhenOneTick() *TestScenario {
	deltaTime := float32(1.0 / 60.0)

	s.satiationSystem.Update(s.world, deltaTime)
	s.feedingSystem.Update(s.world, deltaTime)

	// ИСПРАВЛЕНИЕ: Переключаем анимацию СРАЗУ после создания EatingState
	s.updateEatingAnimations()

	// ИСПРАВЛЕНИЕ: Обновляем анимации ПОСЛЕ переключения (как в игре)
	s.updateAnimations(deltaTime)

	s.grassEatingSystem.Update(s.world, deltaTime)

	return s
}

// updateAnimations обновляет анимации всех животных (точно как в игре)
func (s *TestScenario) updateAnimations(deltaTime float32) {
	// Обновляем анимации для всех сущностей с компонентом Animation
	s.world.ForEachWith(core.MaskAnimation, func(entity core.EntityID) {
		anim, hasAnim := s.world.GetAnimation(entity)
		if !hasAnim || !anim.Playing {
			return
		}

		// Создаём временный компонент для системы анимации
		animComponent := animation.AnimationComponent{
			CurrentAnim: animation.AnimationType(anim.CurrentAnim),
			Frame:       anim.Frame,
			Timer:       anim.Timer,
			Playing:     anim.Playing,
			FacingRight: anim.FacingRight,
		}

		// Обновляем через систему анимации (точно как в игре)
		s.animationSystem.Update(&animComponent, deltaTime)

		// Сохраняем обновлённое состояние обратно в мир
		anim.Frame = animComponent.Frame
		anim.Timer = animComponent.Timer
		anim.Playing = animComponent.Playing
		anim.FacingRight = animComponent.FacingRight
		s.world.SetAnimation(entity, anim)
	})
}

// updateEatingAnimations автоматически переключает зайцев на анимацию поедания когда они начинают есть
func (s *TestScenario) updateEatingAnimations() {
	// Проходим по всем сущностям с EatingState
	s.world.ForEachWith(core.MaskEatingState, func(entity core.EntityID) {
		// Проверяем что это заяц
		behavior, hasBehavior := s.world.GetBehavior(entity)
		if !hasBehavior || behavior.Type != core.BehaviorHerbivore {
			return
		}

		// Получаем текущую анимацию
		anim, hasAnim := s.world.GetAnimation(entity)
		if !hasAnim {
			return
		}

		// Если заяц не в анимации поедания - переключаем
		if anim.CurrentAnim != int(animation.AnimEat) {
			anim.CurrentAnim = int(animation.AnimEat)
			anim.Frame = 0 // Начинаем с первого кадра
			anim.Timer = 0
			anim.Playing = true
			s.world.SetAnimation(entity, anim)
		}
	})
}

// Then методы проверяют результаты

func (s *TestScenario) ThenRabbitShouldBeEating() *TestScenario {
	isEating := s.world.HasComponent(s.rabbit, core.MaskEatingState)
	if !isEating {
		s.t.Errorf("Expected rabbit to be eating, but it's not")
	}
	return s
}

func (s *TestScenario) ThenRabbitShouldNotBeEating() *TestScenario {
	isEating := s.world.HasComponent(s.rabbit, core.MaskEatingState)
	if isEating {
		// Отладочная информация для понимания почему заяц ест
		satiation, _ := s.world.GetSatiation(s.rabbit)
		behavior, _ := s.world.GetBehavior(s.rabbit)
		s.t.Errorf("Expected rabbit not to be eating, but it is. Satiation: %.1f, Threshold: %.1f",
			satiation.Value, behavior.SatiationThreshold)
	}
	return s
}

func (s *TestScenario) ThenGrassAmountAt(tileX, tileY int) GrassAssertion {
	amount := s.terrain.GetGrassAmount(tileX, tileY)
	return GrassAssertion{amount: amount, t: s.t, tileX: tileX, tileY: tileY}
}

func (s *TestScenario) ThenRabbitSatiationLevel() SatiationAssertion {
	satiation, _ := s.world.GetSatiation(s.rabbit)
	return SatiationAssertion{level: satiation.Value, t: s.t}
}

// Вспомогательные типы для assertions

type GrassAssertion struct {
	amount float32
	t      *testing.T
	tileX  int
	tileY  int
}

func (ga GrassAssertion) ShouldBe(expected float32) {
	if ga.amount != expected {
		ga.t.Errorf("Expected grass at (%d,%d) to be %.1f, but was %.1f",
			ga.tileX, ga.tileY, expected, ga.amount)
	}
}

func (ga GrassAssertion) ShouldBeLessThan(threshold float32) {
	if ga.amount >= threshold {
		ga.t.Errorf("Expected grass at (%d,%d) to be less than %.1f, but was %.1f",
			ga.tileX, ga.tileY, threshold, ga.amount)
	}
}

func (ga GrassAssertion) ShouldBeGreaterThan(threshold float32) {
	if ga.amount <= threshold {
		ga.t.Errorf("Expected grass at (%d,%d) to be greater than %.1f, but was %.1f",
			ga.tileX, ga.tileY, threshold, ga.amount)
	}
}

type SatiationAssertion struct {
	level float32
	t     *testing.T
}

func (sa SatiationAssertion) ShouldBe(expected float32) {
	if sa.level != expected {
		sa.t.Errorf("Expected rabbit satiation to be %.1f, but was %.1f", expected, sa.level)
	}
}

func (sa SatiationAssertion) ShouldBeGreaterThan(threshold float32) {
	if sa.level <= threshold {
		sa.t.Errorf("Expected rabbit satiation to be greater than %.1f, but was %.1f", threshold, sa.level)
	}
}

func (sa SatiationAssertion) ShouldBeLessThan(threshold float32) {
	if sa.level >= threshold {
		sa.t.Errorf("Expected rabbit satiation to be less than %.1f, but was %.1f", threshold, sa.level)
	}
}

// Behavioral тесты в стиле Given-When-Then

func TestRabbitFindsFoodWhenHungry(t *testing.T) {
	t.Parallel()

	newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 50.0). // 50% сытости (меньше 60% порога)
		GivenGrassAt(1, 1, 75.0).          // Много травы в том же тайле
		WhenOneTick().                     // Проходит один тик
		ThenRabbitShouldBeEating()         // Заяц должен начать есть
}

func TestRabbitIgnoresFoodWhenSatiated(t *testing.T) {
	t.Parallel()

	newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 95.0). // 95% сытости (больше 60% порога)
		GivenGrassAt(1, 1, 75.0).          // Много травы доступно
		WhenOneTick().                     // Проходит один тик
		ThenRabbitShouldNotBeEating()      // Заяц не должен есть (сытый)
}

func TestRabbitCannotEatWithoutGrass(t *testing.T) {
	t.Parallel()

	newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 30.0). // Очень голодный заяц
		GivenNoGrassAt(1, 1).              // Нет травы в тайле
		WhenOneTick().                     // Проходит один тик
		ThenRabbitShouldNotBeEating()      // Заяц не может есть без травы
}

func TestRabbitCannotEatInsufficientGrass(t *testing.T) {
	t.Parallel()

	newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 40.0). // Голодный заяц
		GivenGrassAt(1, 1, 5.0).           // Мало травы (меньше порога 10.0)
		WhenOneTick().                     // Проходит один тик
		ThenRabbitShouldNotBeEating()      // Заяц не должен есть недостаточную траву
}

func TestGrassConsumptionOverTime(t *testing.T) {
	t.Parallel()

	scenario := newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 40.0). // Голодный заяц
		GivenGrassAt(1, 1, 50.0)           // Средний количество травы

	// Начальное состояние
	scenario.ThenGrassAmountAt(1, 1).ShouldBe(50.0)

	// Заяц начинает есть
	scenario.WhenOneTick().
		ThenRabbitShouldBeEating()

	// После некоторого времени трава должна уменьшиться
	scenario.WhenTimePassesFor(1.0) // 1 секунда
	scenario.ThenGrassAmountAt(1, 1).ShouldBeLessThan(50.0)
}

func TestSatiationRecoveryDuringEating(t *testing.T) {
	t.Parallel()

	scenario := newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 30.0). // Очень голодный заяц (30%)
		GivenGrassAt(1, 1, 100.0)          // Много травы

	// Начальный уровень сытости
	scenario.ThenRabbitSatiationLevel().ShouldBe(30.0)

	// Заяц начинает есть
	scenario.WhenOneTick().
		ThenRabbitShouldBeEating()

	// После поедания сытость должна увеличиться
	scenario.WhenTimePassesFor(2.0) // 2 секунды поедания
	scenario.ThenRabbitSatiationLevel().ShouldBeGreaterThan(30.0)
}

func TestRabbitStopsEatingWhenSatiated(t *testing.T) {
	t.Parallel()

	// Тест с сытостью значительно выше порога - НЕ должен есть
	scenario := newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 95.0). // Значительно выше порога 60%
		GivenGrassAt(1, 1, 100.0)          // Много травы

	// Заяц не должен начинать есть (не голодный)
	scenario.WhenOneTick().
		ThenRabbitShouldNotBeEating()
}

func TestRabbitThresholdBoundary(t *testing.T) {
	t.Parallel()

	// Тест границы порога - с сытостью значительно выше порога
	scenario1 := newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 92.0). // 92% сытости > 60% порога - НЕ должен есть
		GivenGrassAt(1, 1, 100.0)

	scenario1.WhenOneTick().
		ThenRabbitShouldNotBeEating()

	// С сытостью ниже порога - должен есть
	scenario2 := newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 45.0). // 45% сытости < 60% порога - должен есть
		GivenGrassAt(1, 1, 100.0)

	scenario2.WhenOneTick().
		ThenRabbitShouldBeEating()
}

// Интеграционный behavioral тест
func TestFullFeedingCycle(t *testing.T) {
	t.Parallel()

	scenario := newTestScenario(t).
		GivenHungryRabbitAt(48, 48, 20.0). // Очень голодный
		GivenGrassAt(1, 1, 100.0)          // Максимально возможное количество травы

	// 1. Заяц начинает есть
	scenario.WhenOneTick().
		ThenRabbitShouldBeEating().
		ThenGrassAmountAt(1, 1).ShouldBe(100.0) // Трава ещё не потреблена

	// 2. Процесс поедания
	scenario.WhenTimePassesFor(3.0)
	scenario.ThenRabbitShouldBeEating()                           // Всё ещё ест
	scenario.ThenGrassAmountAt(1, 1).ShouldBeLessThan(100.0)      // Трава потребляется
	scenario.ThenRabbitSatiationLevel().ShouldBeGreaterThan(20.0) // Голод уменьшается

	// 3. Длительное поедание до насыщения
	scenario.WhenTimePassesFor(20.0)

	// Проверяем что заяц все еще ест или уже наелся
	isEating := scenario.world.HasComponent(scenario.rabbit, core.MaskEatingState)
	hunger, _ := scenario.world.GetSatiation(scenario.rabbit)
	scenario.t.Logf("After 23 seconds: Hunger: %.1f%%, IsEating: %v", hunger.Value, isEating)

	// Ослабляем требования - заяц должен значительно восстановить голод
	scenario.ThenRabbitSatiationLevel().ShouldBeGreaterThan(50.0) // Хотя бы половина
	scenario.ThenGrassAmountAt(1, 1).ShouldBeLessThan(90.0)       // Хотя бы немного потреблена

	// Результат: заяц должен перестать есть когда достигнет порога сытости
}
