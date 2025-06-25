package integration

import (
	"testing"

	"github.com/aiseeq/savanna/internal/animation"
	"github.com/aiseeq/savanna/internal/core"
	"github.com/aiseeq/savanna/internal/simulation"
	"github.com/aiseeq/savanna/tests/common"
)

// TestWolfSwitchingTargetsBug - TDD тест для бага переключения волка между зайцами
//
// БАГ: Волк убивает первого зайца, но не доедает его, а переключается на второго
// ОЖИДАНИЕ: Волк должен полностью съесть первого зайца перед атакой второго
//
//nolint:gocognit,revive,funlen // TDD тест для воспроизведения конкретного бага
func TestWolfSwitchingTargetsBug(t *testing.T) {
	t.Parallel()

	// Создаём маленький мир 3x3 тайла как в описании бага
	worldSize := float32(3 * 32) // 96x96 пикселей
	world := core.NewWorld(worldSize, worldSize, 42)

	// ИСПРАВЛЕНИЕ: Используем централизованный системный менеджер для правильного порядка систем
	systemManager := common.CreateTestSystemManager(worldSize)

	// Создаём анимационные системы для полноты симуляции
	wolfAnimSystem := animation.NewAnimationSystem()
	rabbitAnimSystem := animation.NewAnimationSystem()

	// Регистрируем анимации
	wolfAnimSystem.RegisterAnimation(animation.AnimAttack, 2, 6.0, false, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimEat, 2, 2.0, true, nil)
	wolfAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)

	rabbitAnimSystem.RegisterAnimation(animation.AnimIdle, 2, 1.0, true, nil)
	rabbitAnimSystem.RegisterAnimation(animation.AnimDeathDying, 2, 3.0, false, nil)

	animManager := animation.NewAnimationManager(wolfAnimSystem, rabbitAnimSystem)

	// Создаём сценарий: 1 волк + 2 зайца близко друг к другу
	rabbit1 := simulation.CreateAnimal(world, core.TypeRabbit, 48, 48) // Центр мира
	rabbit2 := simulation.CreateAnimal(world, core.TypeRabbit, 49, 49) // Рядом с первым
	wolf := simulation.CreateAnimal(world, core.TypeWolf, 47, 48)      // Близко к обоим

	// Делаем волка очень голодным
	world.SetSatiation(wolf, core.Satiation{Value: 20.0}) // 20% - очень голодный

	t.Logf("=== НАЧАЛЬНОЕ СОСТОЯНИЕ ===")
	t.Logf("Волк на (47, 48), голод 20%%")
	t.Logf("Заяц1 на (48, 48)")
	t.Logf("Заяц2 на (49, 49)")

	deltaTime := float32(1.0 / 60.0)
	firstTargetKilled := false
	firstTargetEntity := core.EntityID(0)

	// Симулируем до 10 секунд (600 тиков)
	for i := 0; i < 600; i++ {
		world.Update(deltaTime)
		systemManager.Update(world, deltaTime) // Используем централизованный менеджер
		animManager.UpdateAllAnimations(world, deltaTime)

		// Проверяем кто умер первым
		if !firstTargetKilled {
			if !world.IsAlive(rabbit1) || world.HasComponent(rabbit1, core.MaskCorpse) {
				firstTargetKilled = true
				firstTargetEntity = rabbit1
				t.Logf("Тик %d: Заяц1 убит, стал трупом", i)
			} else if !world.IsAlive(rabbit2) || world.HasComponent(rabbit2, core.MaskCorpse) {
				firstTargetKilled = true
				firstTargetEntity = rabbit2
				t.Logf("Тик %d: Заяц2 убит, стал трупом", i)
			}
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: после убийства первого зайца
		if firstTargetKilled {
			// Волк должен есть первого убитого зайца
			if world.HasComponent(wolf, core.MaskEatingState) {
				eatingState, _ := world.GetEatingState(wolf)
				if eatingState.Target == firstTargetEntity {
					// Хорошо - волк ест первого зайца
					if i%60 == 0 {
						t.Logf("Тик %d: Волк правильно ест первого зайца (entity %d)", i, firstTargetEntity)
					}
				} else {
					// БАГ! Волк переключился на другую цель
					t.Errorf("БАГ ОБНАРУЖЕН на тике %d: Волк переключился с первого убитого зайца (entity %d) "+
						"на другую цель (entity %d)", i, firstTargetEntity, eatingState.Target)
					t.Errorf("Волк должен полностью съесть первого зайца перед атакой второго!")
					return
				}
			} else if world.HasComponent(wolf, core.MaskAttackState) {
				// БАГ! Волк атакует второго зайца не доев первого
				attackState, _ := world.GetAttackState(wolf)
				if world.HasComponent(firstTargetEntity, core.MaskCorpse) {
					t.Errorf("БАГ ОБНАРУЖЕН на тике %d: Волк атакует второго зайца (entity %d) не доев первого трупа (entity %d)",
						i, attackState.Target, firstTargetEntity)

					// Проверяем что первый труп всё ещё существует
					if corpse, hasCorpse := world.GetCorpse(firstTargetEntity); hasCorpse {
						t.Errorf("Первый труп всё ещё имеет питательность %.1f - волк должен его доесть!", corpse.NutritionalValue)
					}
					return
				}
			}

			// Проверяем полное исчезновение первого зайца (полностью съеден)
			// Заяц считается съеденным если либо полностью исчез либо труп исчез
			firstTargetGone := !world.IsAlive(firstTargetEntity) && !world.HasComponent(firstTargetEntity, core.MaskCorpse)

			if firstTargetGone {
				t.Logf("✅ Тик %d: Первый заяц полностью съеден (исчез), волк может атаковать второго", i)

				// Проверяем что голод волка восстановился
				hunger, _ := world.GetSatiation(wolf)
				if hunger.Value > 20.0 {
					t.Logf("✅ Голод волка восстановился: %.1f%%", hunger.Value)
				}

				// Разрешаем волку атаковать второго зайца
				break
			}
		}

		// Логируем прогресс каждую секунду
		if i%60 == 0 {
			rabbit1Alive := world.IsAlive(rabbit1)
			rabbit2Alive := world.IsAlive(rabbit2)
			hunger, _ := world.GetSatiation(wolf)

			t.Logf("Секунда %d: Заяц1=%v, Заяц2=%v, голод волка=%.1f%%",
				i/60+1, rabbit1Alive, rabbit2Alive, hunger.Value)
		}
	}

	// Финальная проверка - важно НЕ что оба зайца убиты, а что волк НЕ переключался между ними
	rabbit1Alive := world.IsAlive(rabbit1)
	rabbit2Alive := world.IsAlive(rabbit2)
	rabbit1IsCorpse := world.HasComponent(rabbit1, core.MaskCorpse)
	rabbit2IsCorpse := world.HasComponent(rabbit2, core.MaskCorpse)
	finalHunger, _ := world.GetSatiation(wolf)

	t.Logf("=== ФИНАЛЬНОЕ СОСТОЯНИЕ ===")
	t.Logf("Заяц1: жив=%v, труп=%v", rabbit1Alive, rabbit1IsCorpse)
	t.Logf("Заяц2: жив=%v, труп=%v", rabbit2Alive, rabbit2IsCorpse)
	t.Logf("Голод волка: %.1f%% (начальный: 20%%)", finalHunger.Value)

	// ГЛАВНАЯ ПРОВЕРКА: Первый заяц должен быть мёртв (убит или съеден)
	if rabbit1Alive && !rabbit1IsCorpse {
		t.Error("Первый заяц остался живым - волк должен был его убить")
	}

	// ПРОВЕРКА УСПЕХА: Если первый заяц убит, а второй цел - значит волк не переключался
	if (rabbit1IsCorpse || !rabbit1Alive) && rabbit2Alive && !rabbit2IsCorpse {
		t.Logf("✅ УСПЕХ: Волк убил первого зайца и НЕ переключился на второго")
		t.Logf("✅ БАГ ИСПРАВЛЕН: Волк правильно завершает поедание перед новой атакой")
	}

	// Проверка роста голода
	if finalHunger.Value <= 20.0 {
		t.Error("Голод волка не увеличился - возможно он не ел")
	} else {
		t.Logf("✅ Голод волка увеличился с 20%% до %.1f%%", finalHunger.Value)
	}
}
