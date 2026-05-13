package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Цветовая палитра интерфейса
var (
	colorPanelBg   = color.RGBA{30, 30, 40, 255}
	colorSliderBg  = color.RGBA{55, 55, 70, 255}
	colorSliderFg  = color.RGBA{100, 180, 255, 255}
	colorBtn       = color.RGBA{55, 55, 70, 255}
	colorBtnHover  = color.RGBA{80, 80, 100, 255}
	colorBtnActive = color.RGBA{100, 180, 255, 255}
	colorText      = color.RGBA{220, 220, 230, 255}
	colorTextDim   = color.RGBA{140, 140, 160, 255}
	colorBorder    = color.RGBA{70, 70, 90, 255}
	colorSimBg     = color.RGBA{15, 15, 25, 255}
	colorGrid      = color.RGBA{25, 25, 38, 255}
)

// --------- Slider ---------

// Slider — горизонтальный ползунок с целочисленным диапазоном.
type Slider struct {
	X, Y, W, H int
	Min, Max   int
	Value      int
	Label      string
	dragging   bool
}

type NumericUpDown struct {
	X, Y, W  int
	Min, Max int
	Value    int
	btnUp    *Button
	btnDown  *Button
	inputBuf string
	focused  bool
}

// NewSlider создаёт ползунок.
func NewSlider(x, y, w int, label string, min, max, val int) *Slider {
	return &Slider{X: x, Y: y, W: w, H: 20, Min: min, Max: max, Value: val, Label: label}
}

func NewNumericUpDown(x, y, w, min, max, val int) *NumericUpDown {
	arrowW := 18
	arrowH := 14
	return &NumericUpDown{
		X: x, Y: y, W: w,
		Min: min, Max: max, Value: val,
		inputBuf: fmt.Sprintf("%d", val),
		btnUp:    NewButton(x+w-arrowW, y, arrowW, arrowH, "+"),
		btnDown:  NewButton(x+w-arrowW, y+arrowH, arrowW, arrowH, "-"),
	}
}

// Update обрабатывает ввод мыши для ползунка.
func (sl *Slider) Update() bool {
	mx, my := ebiten.CursorPosition()
	inBounds := mx >= sl.X && mx <= sl.X+sl.W && my >= sl.Y-4 && my <= sl.Y+sl.H+4

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && inBounds {
		sl.dragging = true
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && sl.dragging {
		t := float64(mx-sl.X) / float64(sl.W)
		t = math.Max(0, math.Min(1, t))
		newVal := sl.Min + int(math.Round(t*float64(sl.Max-sl.Min)))
		if newVal != sl.Value {
			sl.Value = newVal
			return true
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		sl.dragging = false
	}
	return false
}

// Draw рисует ползунок на экране.
func (sl *Slider) Draw(screen *ebiten.Image) {
	// Трек ползунка
	vector.DrawFilledRect(screen, float32(sl.X), float32(sl.Y)+6, float32(sl.W), 8, colorSliderBg, false)

	// Заполненная часть
	t := float64(sl.Value-sl.Min) / float64(sl.Max-sl.Min)
	filled := int(t * float64(sl.W))
	if filled > 0 {
		vector.DrawFilledRect(screen, float32(sl.X), float32(sl.Y)+6, float32(filled), 8, colorSliderFg, false)
	}

	// Ручка
	thumbX := float32(sl.X + filled)
	vector.DrawFilledCircle(screen, thumbX, float32(sl.Y)+10, 8, colorSliderFg, true)
	vector.StrokeCircle(screen, thumbX, float32(sl.Y)+10, 8, 1.5, colorText, true)
}

func (n *NumericUpDown) Update() bool {
	arrowW := 18
	changed := false

	// Кнопки стрелок
	n.btnUp.X, n.btnUp.Y = n.X+n.W-arrowW, n.Y
	n.btnDown.X, n.btnDown.Y = n.X+n.W-arrowW, n.Y+14

	if n.btnUp.Update() && n.Value < n.Max {
		n.Value++
		n.inputBuf = fmt.Sprintf("%d", n.Value)
		changed = true
	}
	if n.btnDown.Update() && n.Value > n.Min {
		n.Value--
		n.inputBuf = fmt.Sprintf("%d", n.Value)
		changed = true
	}

	// Клик по полю ввода — фокус
	mx, my := ebiten.CursorPosition()
	fieldW := n.W - arrowW
	inField := mx >= n.X && mx <= n.X+fieldW && my >= n.Y && my <= n.Y+28
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		n.focused = inField
	}

	// Ввод цифр с клавиатуры
	if n.focused {
		// Цифры 0–9
		for k := ebiten.Key0; k <= ebiten.Key9; k++ {
			if inpututil.IsKeyJustPressed(k) {
				digit := fmt.Sprintf("%d", k-ebiten.Key0)
				if n.inputBuf == "0" {
					n.inputBuf = digit
				} else if len(n.inputBuf) < 3 {
					n.inputBuf += digit
				}
			}
		}
		// Backspace
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(n.inputBuf) > 0 {
			n.inputBuf = n.inputBuf[:len(n.inputBuf)-1]
			if n.inputBuf == "" {
				n.inputBuf = "0"
			}
		}
		// Enter — применить
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			var v int
			fmt.Sscanf(n.inputBuf, "%d", &v)
			if v < n.Min {
				v = n.Min
			}
			if v > n.Max {
				v = n.Max
			}
			n.Value = v
			n.inputBuf = fmt.Sprintf("%d", v)
			n.focused = false
			changed = true
		}
	}

	return changed
}

func (n *NumericUpDown) Draw(screen *ebiten.Image) {
	arrowW := 18

	// Поле ввода
	fieldW := n.W - arrowW
	fieldBg := colorSliderBg
	if n.focused {
		fieldBg = color.RGBA{60, 60, 80, 255}
	}
	vector.DrawFilledRect(screen, float32(n.X), float32(n.Y), float32(fieldW), 28, fieldBg, false)
	vector.StrokeRect(screen, float32(n.X), float32(n.Y), float32(fieldW), 28, 1, colorBorder, false)

	// Текст в поле
	label := n.inputBuf
	if n.focused {
		label += "|" // курсор
	}
	ebitenutil.DebugPrintAt(screen, label, n.X+6, n.Y+8)

	// Кнопки стрелок
	n.btnUp.Draw(screen)
	n.btnDown.Draw(screen)
}

// --------- Button ---------

// Button — кликабельная кнопка.
type Button struct {
	X, Y, W, H int
	Label      string
	hovered    bool
	pressed    bool
}

// NewButton создаёт кнопку.
func NewButton(x, y, w, h int, label string) *Button {
	return &Button{X: x, Y: y, W: w, H: h, Label: label}
}

// Update возвращает true при клике на кнопку.
func (b *Button) Update() bool {
	mx, my := ebiten.CursorPosition()
	b.hovered = mx >= b.X && mx <= b.X+b.W && my >= b.Y && my <= b.Y+b.H
	if b.hovered && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		b.pressed = true
		return true
	}
	b.pressed = false
	return false
}

// Draw рисует кнопку на экране.
func (b *Button) Draw(screen *ebiten.Image) {
	bg := colorBtn
	if b.pressed {
		bg = colorBtnActive
	} else if b.hovered {
		bg = colorBtnHover
	}
	vector.DrawFilledRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), bg, false)
	vector.StrokeRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), 1, colorBorder, false)
	// Текст кнопки
	tw := len(b.Label) * 6
	ebitenutil.DebugPrintAt(screen, b.Label, b.X+(b.W-tw)/2, b.Y+(b.H-12)/2)
}

// --------- Helpers ---------

// drawRect рисует контур прямоугольника.
func drawRect(screen *ebiten.Image, x, y, w, h int, c color.RGBA) {
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, c, false)
}

// drawLabel рисует текстовую метку на панели управления.
func drawLabel(screen *ebiten.Image, text string, x, y int, c color.RGBA) {
	// ebitenutil.DebugPrintAt не поддерживает цвет напрямую — используем отдельный Image
	tmp := ebiten.NewImage(len(text)*6+4, 16)
	ebitenutil.DebugPrint(tmp, text)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	// Простая перекраска через ColorScale (приближение)
	_ = c
	screen.DrawImage(tmp, op)
}
