package integration

import (
	"fmt"
	"math"
	"testing"

	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/tests/common"
)

// TestCornerClusteringBehavior проводит детальный анализ кластеризации зайцев в углах
// при наличии волков в центре мира для выявления причин такого поведения
func TestCornerClusteringBehavior(t *testing.T) {
	t.Logf("=== НАЧАЛО АНАЛИЗА КЛАСТЕРИЗАЦИИ ЗАЙЦЕВ В УГЛАХ ===")

	// Создаем мир среднего размера для хорошей видимости кластеризации
	worldSize := float32(common.MediumWorldSize) // 640x640
	world, systemManager, entities := common.NewTestWorld().
		WithSize(worldSize).
		WithSeed(42). // Детерминированный seed для воспроизводимости
		// Создаем зайцев в кольце вокруг центра
		AddRabbit(320, 220, common.HungryPercentage, common.RabbitMaxHealth). // North
		AddRabbit(420, 320, common.HungryPercentage, common.RabbitMaxHealth). // East
		AddRabbit(320, 420, common.HungryPercentage, common.RabbitMaxHealth). // South
		AddRabbit(220, 320, common.HungryPercentage, common.RabbitMaxHealth). // West
		AddRabbit(270, 270, common.HungryPercentage, common.RabbitMaxHealth). // NW
		AddRabbit(370, 270, common.HungryPercentage, common.RabbitMaxHealth). // NE
		AddRabbit(370, 370, common.HungryPercentage, common.RabbitMaxHealth). // SE
		AddRabbit(270, 370, common.HungryPercentage, common.RabbitMaxHealth). // SW
		// Добавляем волков в центр мира
		AddWolf(320, 320, common.VeryHungryPercentage). // Центральный волк
		AddWolf(310, 310, common.VeryHungryPercentage). // Второй волк рядом
		Build()

	t.Logf("Создан мир %dx%.0f с %d зайцами вокруг центра и %d волками в центре",
		int(worldSize), worldSize, len(entities.Rabbits), len(entities.Wolves))

	// Логируем начальные позиции
	t.Logf("\n--- НАЧАЛЬНЫЕ ПОЗИЦИИ ---")
	for i, rabbit := range entities.Rabbits {
		pos, _ := world.GetPosition(rabbit)
		t.Logf("Заяц %d: (%.1f, %.1f)", i+1, pos.X, pos.Y)
	}
	for i, wolf := range entities.Wolves {
		pos, _ := world.GetPosition(wolf)
		t.Logf("Волк %d: (%.1f, %.1f)", i+1, pos.X, pos.Y)
	}

	// Структуры для отслеживания поведения
	type RabbitSnapshot struct {
		Frame           int
		Position        core.Position
		Velocity        core.Velocity
		FleeingFromWolf bool
		DistanceToWolf  float32
		DistanceToEdge  float32
		InCorner        bool
		BehaviorType    core.BehaviorType
	}

	rabbitHistories := make([][]RabbitSnapshot, len(entities.Rabbits))
	for i := range rabbitHistories {
		rabbitHistories[i] = make([]RabbitSnapshot, 0, 1200) // 20 секунд * 60 FPS
	}

	// Функция для определения близости к углу
	isInCorner := func(pos core.Position, worldSize float32, cornerThreshold float32) bool {
		margin := cornerThreshold
		// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для сравнения
		nearLeft := pos.X < margin
		nearRight := pos.X > worldSize-margin
		nearTop := pos.Y < margin
		nearBottom := pos.Y > worldSize-margin

		return (nearLeft || nearRight) && (nearTop || nearBottom)
	}

	// Функция для расчета расстояния до ближайшего края
	distanceToEdge := func(pos core.Position, worldSize float32) float32 {
		// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для вычислений
		distToLeft := pos.X
		distToRight := worldSize - pos.X
		distToTop := pos.Y
		distToBottom := worldSize - pos.Y

		minDist := distToLeft
		if distToRight < minDist {
			minDist = distToRight
		}
		if distToTop < minDist {
			minDist = distToTop
		}
		if distToBottom < minDist {
			minDist = distToBottom
		}

		return minDist
	}

	// Функция для нахождения ближайшего волка
	findNearestWolf := func(rabbitPos core.Position) (float32, bool) {
		minDist := float32(math.Inf(1))
		found := false

		for _, wolf := range entities.Wolves {
			if !world.IsAlive(wolf) {
				continue
			}
			wolfPos, _ := world.GetPosition(wolf)
			dx := rabbitPos.X - wolfPos.X
			dy := rabbitPos.Y - wolfPos.Y
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

			if dist < minDist {
				minDist = dist
				found = true
			}
		}

		return minDist, found
	}

	// Симулируем 20 секунд (1200 тиков)
	const simulationTicks = 1200
	const logInterval = 120 // Логируем каждые 2 секунды

	t.Logf("\n--- НАЧАЛО СИМУЛЯЦИИ НА %d ТИКОВ (%d СЕКУНД) ---", simulationTicks, simulationTicks/60)

	for tick := 0; tick < simulationTicks; tick++ {
		systemManager.Update(world, common.StandardDeltaTime)

		// Собираем данные о каждом зайце
		for i, rabbit := range entities.Rabbits {
			if !world.IsAlive(rabbit) {
				continue // Заяц мертв, пропускаем
			}

			pos, _ := world.GetPosition(rabbit)
			vel, _ := world.GetVelocity(rabbit)
			behavior, _ := world.GetBehavior(rabbit)

			distToWolf, foundWolf := findNearestWolf(pos)
			edgeDist := distanceToEdge(pos, worldSize)
			inCorner := isInCorner(pos, worldSize, 50.0)        // 50 пикселей от угла
			fleeingFromWolf := foundWolf && distToWolf <= 100.0 // В пределах видимости зайца

			snapshot := RabbitSnapshot{
				Frame:           tick,
				Position:        pos,
				Velocity:        vel,
				FleeingFromWolf: fleeingFromWolf,
				DistanceToWolf:  distToWolf,
				DistanceToEdge:  edgeDist,
				InCorner:        inCorner,
				BehaviorType:    behavior.Type,
			}

			rabbitHistories[i] = append(rabbitHistories[i], snapshot)
		}

		// Логируем состояние каждые 2 секунды
		if tick%logInterval == 0 {
			t.Logf("\n--- СОСТОЯНИЕ НА ТИКЕ %d (%.1f сек) ---", tick, float32(tick)/60.0)

			corneredRabbits := 0
			fleeingRabbits := 0

			for i, rabbit := range entities.Rabbits {
				if !world.IsAlive(rabbit) {
					t.Logf("Заяц %d: МЕРТВ", i+1)
					continue
				}

				pos, _ := world.GetPosition(rabbit)
				vel, _ := world.GetVelocity(rabbit)

				distToWolf, foundWolf := findNearestWolf(pos)
				edgeDist := distanceToEdge(pos, worldSize)
				inCorner := isInCorner(pos, worldSize, 50.0)
				fleeingFromWolf := foundWolf && distToWolf <= 100.0

				if inCorner {
					corneredRabbits++
				}
				if fleeingFromWolf {
					fleeingRabbits++
				}

				status := ""
				if fleeingFromWolf {
					status += "УБЕГАЕТ "
				}
				if inCorner {
					status += "В_УГЛУ "
				}
				if edgeDist < 20 {
					status += "У_КРАЯ "
				}

				t.Logf("Заяц %d: pos=(%.1f,%.1f) vel=(%.1f,%.1f) distWolf=%.1f distEdge=%.1f %s",
					i+1, pos.X, pos.Y, vel.X, vel.Y, distToWolf, edgeDist, status)
			}

			t.Logf("СТАТИСТИКА: %d зайцев в углах, %d убегают от волков", corneredRabbits, fleeingRabbits)
		}
	}

	t.Logf("\n--- АНАЛИЗ РЕЗУЛЬТАТОВ ---")

	// Анализируем каждого зайца
	for i, history := range rabbitHistories {
		if len(history) == 0 {
			t.Logf("Заяц %d: Нет данных (возможно умер рано)", i+1)
			continue
		}

		startPos := history[0].Position
		endPos := history[len(history)-1].Position

		// Подсчитываем время в различных состояниях
		timeInCorner := 0
		timeFleeing := 0
		timeNearEdge := 0
		minDistToWolf := float32(math.Inf(1))
		maxDistToEdge := float32(0)

		for _, snapshot := range history {
			if snapshot.InCorner {
				timeInCorner++
			}
			if snapshot.FleeingFromWolf {
				timeFleeing++
			}
			if snapshot.DistanceToEdge < 30 {
				timeNearEdge++
			}
			if snapshot.DistanceToWolf < minDistToWolf {
				minDistToWolf = snapshot.DistanceToWolf
			}
			if snapshot.DistanceToEdge > maxDistToEdge {
				maxDistToEdge = snapshot.DistanceToEdge
			}
		}

		// Вычисляем проценты
		totalTime := len(history)
		cornerPercent := float32(timeInCorner) / float32(totalTime) * 100
		fleeingPercent := float32(timeFleeing) / float32(totalTime) * 100
		edgePercent := float32(timeNearEdge) / float32(totalTime) * 100

		// Анализируем траекторию движения
		totalDistance := float32(0)
		for j := 1; j < len(history); j++ {
			prev := history[j-1].Position
			curr := history[j].Position
			dx := curr.X - prev.X
			dy := curr.Y - prev.Y
			totalDistance += float32(math.Sqrt(float64(dx*dx + dy*dy)))
		}

		// Проверяем достиг ли заяц угла
		reachedCorner := isInCorner(endPos, worldSize, 50.0)

		t.Logf("\nЗАЯЦ %d ИТОГОВЫЙ АНАЛИЗ:", i+1)
		t.Logf("  Начальная позиция: (%.1f, %.1f)", startPos.X, startPos.Y)
		t.Logf("  Конечная позиция:  (%.1f, %.1f)", endPos.X, endPos.Y)
		t.Logf("  Достиг угла: %v", reachedCorner)
		t.Logf("  Время в углу: %.1f%% (%d/%d тиков)", cornerPercent, timeInCorner, totalTime)
		t.Logf("  Время убегания: %.1f%% (%d/%d тиков)", fleeingPercent, timeFleeing, totalTime)
		t.Logf("  Время у края: %.1f%% (%d/%d тиков)", edgePercent, timeNearEdge, totalTime)
		t.Logf("  Мин. расстояние до волка: %.1f", minDistToWolf)
		t.Logf("  Макс. расстояние до края: %.1f", maxDistToEdge)
		t.Logf("  Общее расстояние движения: %.1f пикселей", totalDistance)

		// Детальный анализ кластеризации
		if cornerPercent > 50 {
			t.Logf("  ⚠️  ОБНАРУЖЕНА КЛАСТЕРИЗАЦИЯ! Заяц провел >50%% времени в углу")

			// Анализируем как заяц попал в угол
			firstCornerTime := -1
			for j, snapshot := range history {
				if snapshot.InCorner {
					firstCornerTime = j
					break
				}
			}

			if firstCornerTime >= 0 {
				t.Logf("  📍 Первый раз попал в угол на тике %d (%.1f сек)", firstCornerTime, float32(firstCornerTime)/60.0)

				// Анализируем что происходило перед попаданием в угол
				if firstCornerTime > 10 {
					for k := firstCornerTime - 10; k < firstCornerTime; k++ {
						if k >= 0 && k < len(history) {
							snap := history[k]
							t.Logf("    Тик %d: pos=(%.1f,%.1f) fleeing=%v distWolf=%.1f",
								k, snap.Position.X, snap.Position.Y, snap.FleeingFromWolf, snap.DistanceToWolf)
						}
					}
				}
			}
		}
	}

	// Общая статистика по кластеризации
	t.Logf("\n--- ОБЩАЯ СТАТИСТИКА КЛАСТЕРИЗАЦИИ ---")
	clusteredRabbits := 0
	aliveRabbits := 0

	for i, rabbit := range entities.Rabbits {
		if !world.IsAlive(rabbit) {
			continue
		}
		aliveRabbits++

		pos, _ := world.GetPosition(rabbit)
		if isInCorner(pos, worldSize, 50.0) {
			clusteredRabbits++

			// Находим ближайший угол
			corners := []string{"Верхний-левый", "Верхний-правый", "Нижний-левый", "Нижний-правый"}
			cornerPositions := [][2]float32{
				{0, 0}, {worldSize, 0}, {0, worldSize}, {worldSize, worldSize},
			}

			minDist := float32(math.Inf(1))
			nearestCorner := ""

			for j, cornerPos := range cornerPositions {
				// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для вычислений
				dx := pos.X - cornerPos[0]
				dy := pos.Y - cornerPos[1]
				dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
				if dist < minDist {
					minDist = dist
					nearestCorner = corners[j]
				}
			}

			t.Logf("Заяц %d в углу %s (расстояние %.1f)", i+1, nearestCorner, minDist)
		}
	}

	clusteringRate := float32(clusteredRabbits) / float32(aliveRabbits) * 100
	t.Logf("КЛАСТЕРИЗАЦИЯ: %d из %d зайцев в углах (%.1f%%)", clusteredRabbits, aliveRabbits, clusteringRate)

	// Анализ системы отражения от границ
	t.Logf("\n--- АНАЛИЗ СИСТЕМЫ ОТРАЖЕНИЯ ГРАНИЦ ---")
	boundaryReflections := 0
	for i, history := range rabbitHistories {
		if len(history) < 2 {
			continue
		}

		for j := 1; j < len(history); j++ {
			prev := history[j-1]
			curr := history[j]

			// Проверяем резкое изменение направления скорости у границы
			if prev.DistanceToEdge < 10 && curr.DistanceToEdge < 10 {
				// Проверяем отражение скорости
				velMagnitudePrev := float32(math.Sqrt(float64(prev.Velocity.X*prev.Velocity.X + prev.Velocity.Y*prev.Velocity.Y)))
				velMagnitudeCurr := float32(math.Sqrt(float64(curr.Velocity.X*curr.Velocity.X + curr.Velocity.Y*curr.Velocity.Y)))

				if velMagnitudePrev > 5 && velMagnitudeCurr > 5 {
					// Вычисляем угол между векторами скорости
					dotProduct := prev.Velocity.X*curr.Velocity.X + prev.Velocity.Y*curr.Velocity.Y
					// Вычисляем косинус угла
					cosAngle := dotProduct / (velMagnitudePrev * velMagnitudeCurr)

					// Если угол близок к 180 градусов (отражение)
					if cosAngle < -0.5 {
						boundaryReflections++
						t.Logf("Заяц %d: отражение от границы на тике %d, pos=(%.1f,%.1f), vel=(%.1f,%.1f)->(%.1f,%.1f)",
							i+1, j, curr.Position.X, curr.Position.Y,
							prev.Velocity.X, prev.Velocity.Y, curr.Velocity.X, curr.Velocity.Y)
					}
				}
			}
		}
	}

	t.Logf("Всего обнаружено отражений от границ: %d", boundaryReflections)

	// Выводы о причинах кластеризации
	t.Logf("\n--- ВЫВОДЫ О КЛАСТЕРИЗАЦИИ ---")
	if clusteringRate > 50 {
		t.Logf("🔴 ПРОБЛЕМА: Высокий уровень кластеризации (%.1f%%)!", clusteringRate)
		t.Logf("Возможные причины:")
		t.Logf("1. Зайцы убегают от волков к краям мира")
		t.Logf("2. Система отражения от границ не позволяет им вернуться к центру")
		t.Logf("3. Углы становятся 'ловушками' для зайцев")
		t.Logf("4. Недостаточная мотивация возвращаться к центру (поиск травы)")
	} else if clusteringRate > 25 {
		t.Logf("🟡 ВНИМАНИЕ: Умеренная кластеризация (%.1f%%)", clusteringRate)
		t.Logf("Поведение частично соответствует ожиданиям")
	} else {
		t.Logf("🟢 НОРМА: Низкий уровень кластеризации (%.1f%%)", clusteringRate)
		t.Logf("Зайцы распределены равномерно")
	}

	t.Logf("\n=== АНАЛИЗ КЛАСТЕРИЗАЦИИ ЗАВЕРШЕН ===")
}

// TestBoundaryReflectionMechanics тестирует механику отражения от границ изолированно
func TestBoundaryReflectionMechanics(t *testing.T) {
	t.Logf("=== ТЕСТ МЕХАНИКИ ОТРАЖЕНИЯ ОТ ГРАНИЦ ===")

	worldSize := float32(200) // Маленький мир для быстрого достижения границ
	world, systemManager, entities := common.NewTestWorld().
		WithSize(worldSize).
		WithSeed(12345).
		// Размещаем зайцев очень близко к каждой границе
		AddRabbit(10, 100, common.SatedPercentage, common.RabbitMaxHealth).  // Левая граница
		AddRabbit(190, 100, common.SatedPercentage, common.RabbitMaxHealth). // Правая граница
		AddRabbit(100, 10, common.SatedPercentage, common.RabbitMaxHealth).  // Верхняя граница
		AddRabbit(100, 190, common.SatedPercentage, common.RabbitMaxHealth). // Нижняя граница
		Build()

	t.Logf("Создан тестовый мир %dx%.0f с зайцами у каждой границы", int(worldSize), worldSize)

	// Задаем зайцам скорости направленные к границам
	world.SetVelocity(entities.Rabbits[0], core.Velocity{X: -20, Y: 0}) // К левой границе
	world.SetVelocity(entities.Rabbits[1], core.Velocity{X: 20, Y: 0})  // К правой границе
	world.SetVelocity(entities.Rabbits[2], core.Velocity{X: 0, Y: -20}) // К верхней границе
	world.SetVelocity(entities.Rabbits[3], core.Velocity{X: 0, Y: 20})  // К нижней границе

	directions := []string{"Левая", "Правая", "Верхняя", "Нижняя"}

	// Логируем начальное состояние
	t.Logf("\n--- НАЧАЛЬНОЕ СОСТОЯНИЕ ---")
	for i, rabbit := range entities.Rabbits {
		pos, _ := world.GetPosition(rabbit)
		vel, _ := world.GetVelocity(rabbit)
		t.Logf("Заяц %s границы: pos=(%.1f,%.1f) vel=(%.1f,%.1f)",
			directions[i], pos.X, pos.Y, vel.X, vel.Y)
	}

	// Симулируем до первого отражения или 300 тиков (5 секунд)
	for tick := 0; tick < 300; tick++ {
		systemManager.Update(world, common.StandardDeltaTime)

		// Проверяем каждого зайца на отражение
		for i, rabbit := range entities.Rabbits {
			pos, _ := world.GetPosition(rabbit)
			vel, _ := world.GetVelocity(rabbit)

			// Проверяем не вышел ли за границы
			// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32 для сравнения
			if pos.X < 0 || pos.X > worldSize || pos.Y < 0 || pos.Y > worldSize {
				t.Errorf("ОШИБКА: Заяц %s вышел за границы! pos=(%.1f,%.1f)", directions[i], pos.X, pos.Y)
			}

			// Логируем когда заяц достигает границы
			margin := float32(10)
			atBoundary := false
			boundaryType := ""

			switch i {
			case 0: // Левая граница
				// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
				if pos.X <= margin {
					atBoundary = true
					boundaryType = "левой"
				}
			case 1: // Правая граница
				// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
				if pos.X >= worldSize-margin {
					atBoundary = true
					boundaryType = "правой"
				}
			case 2: // Верхняя граница
				// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
				if pos.Y <= margin {
					atBoundary = true
					boundaryType = "верхней"
				}
			case 3: // Нижняя граница
				// ТИПОБЕЗОПАСНОСТЬ: конвертируем physics.Pixels в float32
				if pos.Y >= worldSize-margin {
					atBoundary = true
					boundaryType = "нижней"
				}
			}

			if atBoundary {
				t.Logf("Тик %d: Заяц %s достиг %s границы: pos=(%.1f,%.1f) vel=(%.1f,%.1f)",
					tick, directions[i], boundaryType, pos.X, pos.Y, vel.X, vel.Y)
			}
		}

		// Логируем состояние каждые 60 тиков
		if tick%60 == 0 {
			t.Logf("\n--- СОСТОЯНИЕ НА ТИКЕ %d ---", tick)
			for i, rabbit := range entities.Rabbits {
				pos, _ := world.GetPosition(rabbit)
				vel, _ := world.GetVelocity(rabbit)
				t.Logf("Заяц %s: pos=(%.1f,%.1f) vel=(%.1f,%.1f)",
					directions[i], pos.X, pos.Y, vel.X, vel.Y)
			}
		}
	}

	// Финальный анализ
	t.Logf("\n--- ФИНАЛЬНОЕ СОСТОЯНИЕ ---")
	for i, rabbit := range entities.Rabbits {
		pos, _ := world.GetPosition(rabbit)
		vel, _ := world.GetVelocity(rabbit)
		t.Logf("Заяц %s: pos=(%.1f,%.1f) vel=(%.1f,%.1f)",
			directions[i], pos.X, pos.Y, vel.X, vel.Y)

		// Анализируем отражение скорости
		expectedDirection := ""
		actualDirection := ""

		switch i {
		case 0: // Левая граница - скорость должна стать положительной по X
			expectedDirection = "вправо (X > 0)"
			if vel.X > 0 {
				actualDirection = "вправо ✓"
			} else {
				actualDirection = fmt.Sprintf("влево X=%.1f ✗", vel.X)
			}
		case 1: // Правая граница - скорость должна стать отрицательной по X
			expectedDirection = "влево (X < 0)"
			if vel.X < 0 {
				actualDirection = "влево ✓"
			} else {
				actualDirection = fmt.Sprintf("вправо X=%.1f ✗", vel.X)
			}
		case 2: // Верхняя граница - скорость должна стать положительной по Y
			expectedDirection = "вниз (Y > 0)"
			if vel.Y > 0 {
				actualDirection = "вниз ✓"
			} else {
				actualDirection = fmt.Sprintf("вверх Y=%.1f ✗", vel.Y)
			}
		case 3: // Нижняя граница - скорость должна стать отрицательной по Y
			expectedDirection = "вверх (Y < 0)"
			if vel.Y < 0 {
				actualDirection = "вверх ✓"
			} else {
				actualDirection = fmt.Sprintf("вниз Y=%.1f ✗", vel.Y)
			}
		}

		t.Logf("  Ожидаемое направление: %s", expectedDirection)
		t.Logf("  Фактическое направление: %s", actualDirection)
	}

	t.Logf("\n=== ТЕСТ ОТРАЖЕНИЯ ЗАВЕРШЕН ===")
}
