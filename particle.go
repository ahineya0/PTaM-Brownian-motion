package main

import (
	"math"
	"math/rand"
)

// Particle представляет одну частицу в симуляции броуновского движения.
type Particle struct {
	X, Y   float64 // позиция центра
	VX, VY float64 // вектор скорости
	R      float64 // радиус (для визуализации и обнаружения столкновений)
	Hue    float64 // цвет (HSL, оттенок 0–360)
}

// NewParticle создаёт частицу со случайным направлением скорости.
func NewParticle(x, y, speed, radius float64) *Particle {
	angle := rand.Float64() * 2 * math.Pi
	return &Particle{
		X:   x,
		Y:   y,
		VX:  math.Cos(angle) * speed,
		VY:  math.Sin(angle) * speed,
		R:   radius,
		Hue: rand.Float64() * 360,
	}
}

// Move перемещает частицу на один шаг симуляции.
func (p *Particle) Move() {
	p.X += p.VX
	p.Y += p.VY
}

// Speed возвращает модуль скорости частицы.
func (p *Particle) Speed() float64 {
	return math.Hypot(p.VX, p.VY)
}

// ReflectWalls обрабатывает зеркальное отражение от границ прямоугольной области.
// Закон отражения: нормальная к стене компонента скорости меняет знак,
// тангенциальная остаётся без изменений. Модуль скорости не меняется.
func (p *Particle) ReflectWalls(width, height float64) {
	if p.X-p.R < 0 {
		p.X = p.R
		p.VX = math.Abs(p.VX)
	}
	if p.X+p.R > width {
		p.X = width - p.R
		p.VX = -math.Abs(p.VX)
	}
	if p.Y-p.R < 0 {
		p.Y = p.R
		p.VY = math.Abs(p.VY)
	}
	if p.Y+p.R > height {
		p.Y = height - p.R
		p.VY = -math.Abs(p.VY)
	}
}

// ResolveCollision обрабатывает абсолютно упругое столкновение двух частиц
// равной массы методом нормальных и тангенциальных компонент скоростей
// вдоль линии центров (см. Приложение А ТЗ и раздел 1.3 анализа предметной области).
// Возвращает true, если столкновение произошло.
func ResolveCollision(a, b *Particle) bool {
	dx := b.X - a.X
	dy := b.Y - a.Y
	dist := math.Hypot(dx, dy)
	minDist := a.R + b.R

	if dist >= minDist || dist < 1e-9 {
		return false
	}

	// Единичный вектор нормали вдоль линии центров
	nx := dx / dist
	ny := dy / dist

	// Раздвигаем частицы, чтобы устранить перекрытие
	overlap := (minDist - dist) / 2.0
	a.X -= nx * overlap
	a.Y -= ny * overlap
	b.X += nx * overlap
	b.Y += ny * overlap

	// Нормальные проекции скоростей: vn = (v · n)
	v1n := a.VX*nx + a.VY*ny
	v2n := b.VX*nx + b.VY*ny

	// Тангенциальные проекции (ось, перпендикулярная нормали): vt = (v · τ)
	// τ = (-ny, nx)
	v1t := -a.VX*ny + a.VY*nx
	v2t := -b.VX*ny + b.VY*nx

	// Для равных масс нормальные компоненты обмениваются (центральный удар),
	// тангенциальные остаются неизменными (трения нет).
	u1n := v2n
	u2n := v1n
	u1t := v1t
	u2t := v2t

	// Собираем результирующие векторы скоростей: u = un·n + ut·τ
	a.VX = u1n*nx - u1t*ny
	a.VY = u1n*ny + u1t*nx
	b.VX = u2n*nx - u2t*ny
	b.VY = u2n*ny + u2t*nx

	return true
}
