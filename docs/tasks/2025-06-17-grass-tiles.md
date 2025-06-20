# ГЕНЕРАЦИЯ ТРАВЯНЫХ ТАЙЛОВ ДЛЯ САВАННЫ

## КОНЦЕПЦИЯ СИСТЕМЫ

Создаём **4 состояния травы** в высоком разрешении (1024x1024px) с последующим уменьшением до 32x32px. Динамическая система цветов покажет водность участков через оттенки в игре.

### СОСТОЯНИЯ ТРАВЫ:
- **Пустая земля** (0%) - голая почва саванны
- **Молодые ростки** (25%) - первые всходы после дождя  
- **Средняя трава** (50%) - активно растущая растительность
- **Густая трава** (100%) - полностью развитая саванна

### АНТИ-ПОВТОР ТАКТИКИ:
- ✅ **Зеркалирование** - горизонтальный/вертикальный флип
- ✅ **Цветовые вариации** - тёплые/холодные оттенки
- ✅ **Поворот на 90°** - 4 ориентации каждого тайла
- 🆕 **Микросдвиги травинок** - слегка разные позиции элементов
- 🆕 **Вариации плотности** - разное количество травинок в кластерах

---

## ТЕХНИЧЕСКИЕ ТРЕБОВАНИЯ

**РАЗМЕР И ФОРМАТ:**
* Исходник: 1024x1024 пикселей (для максимальной детализации)
* Финальный размер: 32x32 пикселя после уменьшения
* Прозрачный фон (PNG с альфа-каналом)
* Изометрический вид (45° наклон, вид сверху)

**ФОРМА ТАЙЛА:**
* Ромбовидная основа занимает нижнюю половину изображения (512 пикселей высоты)
* Верхняя половина остаётся прозрачной для роста травы
* Базовая почва: плодородная земля с молодой растительностью
* Трава растёт В ВЕРХНЕЙ ЗОНЕ, не меняя положение ромба почвы

**VISUAL STYLE:**
* Ручная отрисовка, НЕ пиксель-арт
* Яркие, насыщенные цвета африканской саванны
* Чёткие контуры для хорошего масштабирования
* Детализация должна читаться даже в 32x32

**ЦВЕТОВАЯ СХЕМА (базовая):**
* Почва: тёплый красновато-коричневый (#CD853F, #D2691E)
* Молодая трава: светло-зелёный (#90EE90, #32CD32)
* Зрелая трава: золотисто-зелёный (#9ACD32, #6B8E23)
* Сухая трава: жёлто-коричневый (#DAA520, #B8860B)

---

## СТРАТЕГИЯ ГЕНЕРАЦИИ

**ПОРЯДОК СОЗДАНИЯ:**
1. 🌿 **НАЧИНАЕМ с густой травы (100%)** - эталонный тайл с максимальной растительностью
2. 🌾 Уменьшаем до средней травы (50%) - убираем часть травы, **ромб почвы остаётся на месте**
3. 🌱 Ещё больше уменьшаем до молодых ростков (25%) - **ромб почвы НЕ ДВИГАЕТСЯ**
4. 🟫 Полностью убираем траву до голой земли (0%) - **ромб почвы в том же положении**

**КЛЮЧЕВОЙ ПРИНЦИП:** Форма и положение ромбовидной почвы остаётся константой во всех 4 состояниях. Меняется только количество и высота травы в верхней зоне.

---

## ПРОМПТЫ ДЛЯ ГЕНЕРАЦИИ

### 🌿 ПРОМПТ 1 — Густая трава (100%) [НАЧИНАТЬ С ЭТОГО!]

**Create an isometric grass tile showing dense, fully-grown savanna grass coverage.**

**VISUAL STYLE:**
* Hand-drawn illustration style (NOT pixel art)
* Vibrant, saturated African savanna colors
* Clean, crisp edges for good scaling  
* 1024x1024 square image
* Transparent background
* 45-degree isometric perspective (top-down view)

**TILE STRUCTURE:**
* Diamond-shaped soil base occupies BOTTOM HALF of image (512px height)
* Grass grows in UPPER ZONE above the soil diamond
* Soil position MUST remain constant for all variations

**TERRAIN FEATURES:**
* Dense grass coverage growing upward from soil diamond
* 20+ grass clumps filling the upper growing zone
* Mature grass: golden-brown savanna color (#DAA520, #B8860B)
* Grass height: tall, reaching into upper 50% of image
* Large 5-7 blade clusters
* Varied grass orientations for natural look
* Rich, mature savanna appearance
* Small gaps between major clumps

**LIGHTING:**
* Bright African sun from upper-left
* Complex shadow patterns through grass
* Warm, golden color dominance

---

### 🌾 ПРОМПТ 3 — Средняя трава (50%) [ДЕЛАТЬ ВТОРЫМ]

**Create an isometric grass tile showing medium savanna grass coverage with growing vegetation.**

**VISUAL STYLE:**
* Hand-drawn illustration style (NOT pixel art)
* Vibrant, saturated African savanna colors  
* Clean, crisp edges for good scaling
* 1024x1024 square image
* Transparent background
* 45-degree isometric perspective (top-down view)

**TILE STRUCTURE:**
* Diamond-shaped soil base occupies BOTTOM HALF of image (512px height)
* Medium grass grows in UPPER ZONE above the soil diamond
* Soil position EXACTLY MATCHES the dense grass tile variation

**TERRAIN FEATURES:**
* Same reddish-brown soil base as dense grass version
* 12-15 grass clumps of varying sizes in upper zone
* Medium grass: golden-green (#9ACD32, #6B8E23)
* Grass height: medium, extending into upper zone
* Mix of 3-5 blade clusters
* Some gaps in upper growing area
* Transitioning from bright green to golden
* SAME diamond soil shape and position as dense version

**LIGHTING:**
* Bright African sun from upper-left
* Clear grass shadows on soil
* Balanced warm and cool color harmony

---

### 🌱 ПРОМПТ 2 — Молодые ростки (25%) [ДЕЛАТЬ ТРЕТЬИМ]

**Create an isometric grass tile showing early savanna grass growth with sparse young shoots.**

**VISUAL STYLE:**
* Hand-drawn illustration style (NOT pixel art) 
* Vibrant, saturated African savanna colors
* Clean, crisp edges for good scaling
* 1024x1024 square image
* Transparent background
* 45-degree isometric perspective (top-down view)

**TILE STRUCTURE:**
* Diamond-shaped soil base occupies BOTTOM HALF of image (512px height)
* Young grass grows in UPPER ZONE above the soil diamond
* Soil position EXACTLY MATCHES the dense grass tile variation

**TERRAIN FEATURES:**
* Same reddish-brown soil base as dense grass version
* 5-8 small grass shoots scattered in upper growing zone
* Young grass: bright light green (#90EE90, #32CD32)
* Shoots are thin, 2-3 blades each
* Height: very short, barely reaching into upper zone
* Most upper area still empty/transparent
* Fresh, vibrant green color
* SAME diamond soil shape and position as other variants

**LIGHTING:**
* Bright African sun from upper-left
* Grass casts tiny shadows
* Mix of warm soil and cool green tones

---

### 🟫 ПРОМПТ 4 — Пустая земля (0%) [ДЕЛАТЬ ПОСЛЕДНИМ]

**Create an isometric grass tile showing barren savanna soil with no vegetation.**

**VISUAL STYLE:**
* Hand-drawn illustration style (NOT pixel art)
* Vibrant, saturated African savanna colors
* Clean, crisp edges for good scaling
* 1024x1024 square image
* Transparent background
* 45-degree isometric perspective (top-down view)

**TILE STRUCTURE:**
* Diamond-shaped soil base occupies BOTTOM HALF of image (512px height)
* Upper half remains transparent (for future grass growth)
* Soil position EXACTLY MATCHES the previous grass tile variations

**TERRAIN FEATURES:**
* Reddish-brown savanna soil (#CD853F, #D2691E)
* Subtle cracks and weathering in the earth
* Small pebbles and stones scattered sparsely
* Gentle shadows to show surface texture
* Maybe 1-2 tiny dried twigs or leaves on soil surface
* Completely barren - no green vegetation
* SAME diamond soil shape and position as grass variants

**LIGHTING:**
* Bright African sun from upper-left
* Strong shadows for depth
* Warm, golden undertones

---

## ПОСТОБРАБОТКА В ИГРЕ

### ДИНАМИЧЕСКИЕ ЦВЕТОВЫЕ МОДИФИКАЦИИ:

**Водность участков:**
* 🔵 **Высокая влажность** - более синий оттенок (+Blue channel)
* 🟡 **Средняя влажность** - базовые цвета (без изменений)
* 🔴 **Засуха** - более красный/жёлтый (+Red channel, -Green)

**Эффекты в коде:**
```glsl
// Псевдокод шейдера
color.rgb *= mix(
    vec3(1.2, 0.8, 0.7), // Засуха (красноватый)
    vec3(0.8, 1.0, 1.2), // Влажность (синеватый)  
    waterLevel
);
```
