package plt

import (
	"code.google.com/p/plotinum/vecgfx"
	"image/color"
	"fmt"
	"math"
)

const (
	DefaultFont = "Times-Roman"
)

type Axis struct{
	// Min and Max are the minimum and maximum data
	// coordinates on this axis.
	Min, Max float64

	// Label is the axis label
	Label string

	// LabelStyle is the text style of the label on the axis.
	LabelStyle TextStyle

	// AxisStyle is the style of the axis's line.
	AxisStyle LineStyle

	// Padding between the axis line and the data in inches.
	Padding float64

	// Ticks are the tick marks on the axis.
	Ticks TickMarks
}

// MakeAxis returns a default axis.
func MakeAxis() Axis {
	labelFont, err := MakeFont(DefaultFont, 12)
	if err != nil {
		panic(err)
	}
	return Axis{
		Min: math.Inf(1),
		Max: math.Inf(-1),
		Label: "",
		LabelStyle: TextStyle{
			Color: Black,
			Font: labelFont,
		},
		AxisStyle: LineStyle{
			Color: Black,
			Width: 1.0/64.0,
		},
		Padding: 1.0/8.0,
		Ticks: MakeTickMarks(),
	}
}

// X transfroms the data point x to the drawing coordinate
// for the given drawing area.
func (a *Axis) X(da *DrawArea, x float64) float64 {
	p := (x - a.Min) / (a.Max - a.Min)
	return da.Min.X + p*(da.Max().X - da.Min.X)
}

// Y transforms the data point y to the drawing coordinate
// for the given drawing area.
func (a *Axis) Y(da *DrawArea, y float64) float64 {
	p := (y - a.Min) / (a.Max - a.Min)
	return da.Min.Y + p*(da.Max().Y - da.Min.Y)
}

// height returns the height of the axis in inches
//  if it is drawn as a horizontal axis.
func (a *Axis) height() (h float64) {
	if a.Label != "" {
		h += a.LabelStyle.Font.Extents().Height/vecgfx.PtInch
	}
	marks := a.Ticks.Marks(a.Min, a.Max)
	if len(marks) > 0 {
		h += a.Ticks.Length + a.Ticks.labelHeight(marks)
	}
	h += a.AxisStyle.Width/2
	h += a.Padding
	return
}

// drawHoriz draws the axis onto the given area.
func (a *Axis) drawHoriz(da *DrawArea) {
	y := da.Min.Y
	if a.Label != "" {
		da.SetTextStyle(a.LabelStyle)
		y += -(a.LabelStyle.Font.Extents().Descent/vecgfx.PtInch * da.DPI())
		da.Text(da.Center().X, y, -0.5, 0, a.Label)
		y += a.LabelStyle.Font.Extents().Ascent/vecgfx.PtInch * da.DPI()
	}
	marks := a.Ticks.Marks(a.Min, a.Max)
	if len(marks) > 0 {
		da.SetLineStyle(a.Ticks.MarkStyle)
		da.SetTextStyle(a.Ticks.LabelStyle)
		for _, t := range marks {
			if t.minor() {
				continue
			}
			da.Text(a.X(da, t.Value), y, -0.5, 0, t.Label)
		}
		y += a.Ticks.labelHeight(marks) * da.DPI()

		len := a.Ticks.Length*da.DPI()
		for _, t := range marks {
			x := a.X(da, t.Value)
			da.Line([]Point{{x, y + t.lengthOffset(len)}, {x, y + len}})
		}
		y += len
	}
	da.SetLineStyle(a.AxisStyle)
	da.Line([]Point{{da.Min.X, y}, {da.Max().X, y}})
}

// width returns the width of the axis in inches
//  if it is drawn as a vertically axis.
func (a *Axis) width() (w float64) {
	if a.Label != "" {
		w += a.LabelStyle.Font.Extents().Ascent/vecgfx.PtInch
	}
	marks := a.Ticks.Marks(a.Min, a.Max)
	if len(marks) > 0 {
		if lwidth := a.Ticks.labelWidth(marks); lwidth > 0 {
			w += lwidth
			// Add a space after tick labels to separate
			// them from the tick marks
			w += a.Ticks.LabelStyle.Font.Width(" ")/vecgfx.PtInch
		}
		w += a.Ticks.Length
	}
	w += a.AxisStyle.Width/2
	w += a.Padding
	return
}

// drawVert draws the axis onto the given area.
func (a *Axis) drawVert(da *DrawArea) {
	x := da.Min.X
	if a.Label != "" {
		x += a.LabelStyle.Font.Extents().Ascent/vecgfx.PtInch * da.DPI()
		da.SetTextStyle(a.LabelStyle)
		da.Push()
		da.Rotate(math.Pi/2)
		da.Text(da.Center().Y, -x, -0.5, 0, a.Label)
		da.Pop()
		x += -a.LabelStyle.Font.Extents().Descent/vecgfx.PtInch * da.DPI()
	}
	marks := a.Ticks.Marks(a.Min, a.Max)
	if len(marks) > 0 {
		da.SetLineStyle(a.Ticks.MarkStyle)
		da.SetTextStyle(a.Ticks.LabelStyle)
		if lwidth := a.Ticks.labelWidth(marks); lwidth > 0 {
			x += lwidth * da.DPI()
			x += a.Ticks.LabelStyle.Font.Width(" ")/vecgfx.PtInch * da.DPI()
		}
		for _, t := range marks {
			if t.minor() {
				continue
			}
			da.Text(x, a.Y(da, t.Value), -1, -0.5, t.Label + " ")
		}
		len := a.Ticks.Length*da.DPI()
		for _, t := range marks {
			y := a.Y(da, t.Value)
			da.Line([]Point{{x + t.lengthOffset(len), y}, {x + len, y}})
		}
		x += len
	}
	da.SetLineStyle(a.AxisStyle)
	da.Line([]Point{{x, da.Min.Y}, {x, da.Max().Y}})
}

// TickMarks specifies the style and location of the tick marks
// on an axis.
type TickMarks struct {
	// LabelStyle is the TextStyle on the tick labels.
	LabelStyle TextStyle

	// MarkStyle is the LineStyle of the tick mark lines.
	MarkStyle LineStyle

	// Length is the length of a major tick mark in inches.
	// Minor tick marks are half of the length of major
	// tick marks.
	Length float64

	// TickMarker locates the tick marks given the
	// minimum and maximum values.
	TickMarker
}

// A TickMarker returns a slice of ticks between a given
// range of values. 
type TickMarker interface{
	// Marks returns a slice of ticks for the given range.
	Marks(min, max float64) []Tick
}

// A Tick is a single tick mark
type Tick struct {
	Value float64
	Label string
}

// minor returns true if this is a minor tick mark.
func (t Tick) minor() bool {
	return t.Label == ""
}

// lengthOffset returns an offset that should be added to the
// tick mark's line to accout for its length.  I.e., the start of
// the line for a minor tick mark must be shifted by half of
// the length.
func (t Tick) lengthOffset(len float64) float64 {
	if t.minor() {
		return len/2
	}
	return 0
}

// MakeTickMarks returns a TickMarks using the default style
// and TickMarker.
func MakeTickMarks() TickMarks {
	labelFont, err := MakeFont(DefaultFont, 10)
	if err != nil {
		panic(err)
	}
	return TickMarks{
		LabelStyle: TextStyle{
			Color: color.RGBA{A: 255},
			Font: labelFont,
		},
		MarkStyle: LineStyle{
			Color: color.RGBA{A:255},
			Width: 1.0/64.0,
		},
		Length: 1.0/10.0,
		TickMarker: DefaultTicks(struct{}{}),
	}
}
// labelHeight returns the label height in inches.
func (tick TickMarks) labelHeight(ticks []Tick) float64 {
	for _, t := range ticks {
		if t.minor() {
			continue
		}
		font := tick.LabelStyle.Font
		return font.Extents().Ascent/vecgfx.PtInch
	}
	return 0
}

// labelWidth returns the label width in inches.
func (tick TickMarks) labelWidth(ticks []Tick) float64 {
	maxWidth := 0.0
	for _, t := range ticks {
		if t.minor() {
			continue
		}
		w := tick.LabelStyle.Font.Width(t.Label)
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth/vecgfx.PtInch
}

// A DefalutTicks returns a default set of tick marks within
// the given range.
type DefaultTicks struct{}

// Marks implements the TickMarker Marks method.
func (_ DefaultTicks) Marks(min, max float64) []Tick {
	return []Tick{
			{ Value: min, Label: fmt.Sprintf("%g", min) },
			{ Value: min + (max-min)/4, },
			{ Value: min + (max-min)/2, Label: fmt.Sprintf("%g", min + (max-min)/2) },
			{ Value: min + 3*(max-min)/4, },
			{ Value: max, Label: fmt.Sprintf("%g", max) },
	}
}

// A ConstTicks always returns the same set of tick marks.
type ConstantTicks []Tick

// Marks implements the TickMarker Marks method.
func (tks ConstantTicks) Marks(min, max float64) []Tick {
	return tks
}