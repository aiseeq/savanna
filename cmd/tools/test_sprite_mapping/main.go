package main

import (
	"fmt"
)

// Копируем enum из animation/system.go
type AnimationType uint8

const (
	AnimIdle AnimationType = iota
	AnimWalk
	AnimRun
	AnimDeathDying
	AnimDeathDecay
	AnimEat
	AnimAttack
)

func main() {
	fmt.Println("🎬 Тестирование соответствия анимаций и спрайтов")
	fmt.Println("==============================================")

	// Анимации которые загружаются в sprite_renderer.go
	loadedAnimations := map[AnimationType]string{
		AnimIdle:       "idle",
		AnimWalk:       "walk",
		AnimRun:        "run",
		AnimAttack:     "attack",
		AnimEat:        "eat",
		AnimDeathDying: "dead",
	}

	fmt.Println("\n📋 Соответствие enum -> спрайты:")
	for animType, spriteName := range loadedAnimations {
		fmt.Printf("  AnimationType %d (%s) -> \"%s\" спрайты\n",
			int(animType), getAnimName(animType), spriteName)
	}

	// Проверяем что AnimIdle = 0 (что устанавливается при создании животного)
	fmt.Printf("\n🔍 Начальная анимация: AnimIdle = %d\n", int(AnimIdle))

	// Проверяем все значения enum
	fmt.Println("\n📊 Полный список AnimationType:")
	allAnims := []AnimationType{AnimIdle, AnimWalk, AnimRun, AnimDeathDying, AnimDeathDecay, AnimEat, AnimAttack}
	for _, anim := range allAnims {
		spriteName, loaded := loadedAnimations[anim]
		status := "❌ НЕ ЗАГРУЖЕН"
		if loaded {
			status = "✅ загружен как \"" + spriteName + "\""
		}
		fmt.Printf("  %d: %s - %s\n", int(anim), getAnimName(anim), status)
	}

	// Ищем потенциальные проблемы
	fmt.Println("\n⚠️  Потенциальные проблемы:")

	if _, hasDecay := loadedAnimations[AnimDeathDecay]; !hasDecay {
		fmt.Printf("  - AnimDeathDecay (%d) не загружается, но может использоваться\n", int(AnimDeathDecay))
	}

	fmt.Println("\n🎯 Вывод:")
	fmt.Println("  Если животное создается с AnimIdle (0), то должны загружаться \"idle\" спрайты")
	fmt.Println("  Проверьте что в assets есть hare_idle_1.png и wolf_idle_1.png")
}

func getAnimName(anim AnimationType) string {
	switch anim {
	case AnimIdle:
		return "AnimIdle"
	case AnimWalk:
		return "AnimWalk"
	case AnimRun:
		return "AnimRun"
	case AnimDeathDying:
		return "AnimDeathDying"
	case AnimDeathDecay:
		return "AnimDeathDecay"
	case AnimEat:
		return "AnimEat"
	case AnimAttack:
		return "AnimAttack"
	default:
		return "Unknown"
	}
}
