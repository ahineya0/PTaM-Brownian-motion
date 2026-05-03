package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// hslToRGB конвертирует HSL (h: 0–360, s,l: 0–1) в RGBA.
func hslToRGB(h, s, l float64) color.RGBA {
	h = math.Mod(h, 360)
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

// darken делает цвет темнее для обводки.
func darken(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: 255,
	}
}

// DrawSimField заливает область симуляции фоном.
func DrawSimField(screen *ebiten.Image, offsetX, offsetY, w, h int) {
	vector.DrawFilledRect(screen,
		float32(offsetX), float32(offsetY),
		float32(w), float32(h),
		colorSimBg, false)

	// Лёгкая сетка для наглядности
	gridStep := 40
	gridColor := colorGrid
	for x := offsetX; x < offsetX+w; x += gridStep {
		vector.StrokeLine(screen,
			float32(x), float32(offsetY),
			float32(x), float32(offsetY+h),
			1, gridColor, false)
	}
	for y := offsetY; y < offsetY+h; y += gridStep {
		vector.StrokeLine(screen,
			float32(offsetX), float32(y),
			float32(offsetX+w), float32(y),
			1, gridColor, false)
	}

	// Рамка области
	vector.StrokeRect(screen,
		float32(offsetX), float32(offsetY),
		float32(w), float32(h),
		2, colorBorder, false)
}

// DrawParticles рисует все частицы симуляции.
// offsetX, offsetY — смещение области симуляции на экране.
func DrawParticles(screen *ebiten.Image, sim *Simulation, offsetX, offsetY int) {
	for _, p := range sim.Particles {
		sx := float32(p.X) + float32(offsetX)
		sy := float32(p.Y) + float32(offsetY)
		r := float32(p.R)

		fill := hslToRGB(p.Hue, 0.75, 0.55)
		stroke := darken(fill, 0.6)

		vector.DrawFilledCircle(screen, sx, sy, r, fill, true)
		vector.StrokeCircle(screen, sx, sy, r, 1.2, stroke, true)

		// Небольшая «бликовая» точка для объёма
		vector.DrawFilledCircle(screen, sx-r*0.28, sy-r*0.28, r*0.22,
			color.RGBA{255, 255, 255, 60}, true)
	}
}

// DrawVelocityVectors рисует векторы скоростей поверх частиц (режим отладки).
// TODO (прототип): в финальной версии добавить переключатель в UI.
func DrawVelocityVectors(screen *ebiten.Image, sim *Simulation, offsetX, offsetY int, scale float64) {
	arrowColor := color.RGBA{255, 200, 50, 180}
	for _, p := range sim.Particles {
		sx := float32(p.X) + float32(offsetX)
		sy := float32(p.Y) + float32(offsetY)
		ex := sx + float32(p.VX*scale)
		ey := sy + float32(p.VY*scale)
		vector.StrokeLine(screen, sx, sy, ex, ey, 1.5, arrowColor, false)
	}
}
