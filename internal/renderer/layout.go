package renderer

import (
	"math"
	"strconv"
)

// BoundingBox represents the dimensions and anchor points of a rendered element
type BoundingBox struct {
	X, Y          float64
	Width, Height float64

	// Anchor points for connecting lines
	AnchorLeft  float64 // X coordinate of left connection point
	AnchorRight float64 // X coordinate of right connection point
	AnchorY     float64 // Y coordinate of horizontal connection line (centerline)
}

// NewBoundingBox creates a bounding box with default anchors
func NewBoundingBox(x, y, width, height float64) BoundingBox {
	return BoundingBox{
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		AnchorLeft:  x,
		AnchorRight: x + width,
		AnchorY:     y + height/2,
	}
}

// X2 returns the right edge coordinate
func (b BoundingBox) X2() float64 {
	return b.X + b.Width
}

// Y2 returns the bottom edge coordinate
func (b BoundingBox) Y2() float64 {
	return b.Y + b.Height
}

// CenterX returns the horizontal center
func (b BoundingBox) CenterX() float64 {
	return b.X + b.Width/2
}

// CenterY returns the vertical center
func (b BoundingBox) CenterY() float64 {
	return b.Y + b.Height/2
}

// Translate returns a new bounding box shifted by dx, dy
func (b BoundingBox) Translate(dx, dy float64) BoundingBox {
	return BoundingBox{
		X:           b.X + dx,
		Y:           b.Y + dy,
		Width:       b.Width,
		Height:      b.Height,
		AnchorLeft:  b.AnchorLeft + dx,
		AnchorRight: b.AnchorRight + dx,
		AnchorY:     b.AnchorY + dy,
	}
}

// RenderedNode represents a node that has been rendered
type RenderedNode struct {
	Element SVGElement
	BBox    BoundingBox
}

// SpaceHorizontally arranges items horizontally with padding
// Returns the total bounding box and transforms each item's position
func SpaceHorizontally(items []RenderedNode, padding float64) ([]RenderedNode, BoundingBox) {
	if len(items) == 0 {
		return items, BoundingBox{}
	}

	// Find the maximum anchor Y across all items (for vertical alignment)
	maxAnchorY := 0.0
	for _, item := range items {
		if item.BBox.AnchorY > maxAnchorY {
			maxAnchorY = item.BBox.AnchorY
		}
	}

	// Position items horizontally and align vertically
	result := make([]RenderedNode, len(items))
	x := 0.0
	minY := math.MaxFloat64
	maxY := 0.0

	for i, item := range items {
		// Calculate vertical offset to align anchors
		dy := maxAnchorY - item.BBox.AnchorY
		newBBox := item.BBox.Translate(x-item.BBox.X, dy)

		result[i] = RenderedNode{
			Element: wrapWithTransform(item.Element, x-item.BBox.X, dy),
			BBox:    newBBox,
		}

		if newBBox.Y < minY {
			minY = newBBox.Y
		}
		if newBBox.Y2() > maxY {
			maxY = newBBox.Y2()
		}

		x = newBBox.X2() + padding
	}

	// Calculate total bounding box
	totalWidth := x - padding // Remove last padding
	if len(result) > 0 {
		totalWidth = result[len(result)-1].BBox.X2()
	}

	totalBBox := BoundingBox{
		X:           0,
		Y:           minY,
		Width:       totalWidth,
		Height:      maxY - minY,
		AnchorLeft:  result[0].BBox.AnchorLeft,
		AnchorRight: result[len(result)-1].BBox.AnchorRight,
		AnchorY:     maxAnchorY,
	}

	return result, totalBBox
}

// SpaceVertically arranges items vertically with padding, centered horizontally
func SpaceVertically(items []RenderedNode, padding float64) ([]RenderedNode, BoundingBox) {
	if len(items) == 0 {
		return items, BoundingBox{}
	}

	// Find the maximum width for horizontal centering
	maxWidth := 0.0
	for _, item := range items {
		if item.BBox.Width > maxWidth {
			maxWidth = item.BBox.Width
		}
	}

	// Position items vertically
	result := make([]RenderedNode, len(items))
	y := 0.0

	for i, item := range items {
		// Center horizontally
		dx := (maxWidth - item.BBox.Width) / 2
		newBBox := item.BBox.Translate(dx-item.BBox.X, y-item.BBox.Y)

		result[i] = RenderedNode{
			Element: wrapWithTransform(item.Element, dx-item.BBox.X, y-item.BBox.Y),
			BBox:    newBBox,
		}

		y = newBBox.Y2() + padding
	}

	totalHeight := y - padding
	if len(result) > 0 {
		totalHeight = result[len(result)-1].BBox.Y2()
	}

	totalBBox := BoundingBox{
		X:           0,
		Y:           0,
		Width:       maxWidth,
		Height:      totalHeight,
		AnchorLeft:  0,
		AnchorRight: maxWidth,
		AnchorY:     totalHeight / 2,
	}

	return result, totalBBox
}

// wrapWithTransform wraps an element in a group with a transform
func wrapWithTransform(elem SVGElement, dx, dy float64) SVGElement {
	if dx == 0 && dy == 0 {
		return elem
	}
	return &Group{
		Transform: "translate(" + fmtFloat(dx) + "," + fmtFloat(dy) + ")",
		Children:  []SVGElement{elem},
	}
}

// MeasureText estimates the width of text given the configuration
func MeasureText(text string, cfg *Config) float64 {
	return float64(len(text)) * cfg.CharWidth
}

// PathBuilder helps construct SVG path data
type PathBuilder struct {
	commands []string
}

// NewPathBuilder creates a new path builder
func NewPathBuilder() *PathBuilder {
	return &PathBuilder{}
}

// MoveTo adds a move command (M)
func (pb *PathBuilder) MoveTo(x, y float64) *PathBuilder {
	pb.commands = append(pb.commands, "M "+fmtFloat(x)+" "+fmtFloat(y))
	return pb
}

// LineTo adds a line command (L)
func (pb *PathBuilder) LineTo(x, y float64) *PathBuilder {
	pb.commands = append(pb.commands, "L "+fmtFloat(x)+" "+fmtFloat(y))
	return pb
}

// HorizontalTo adds a horizontal line command (H)
func (pb *PathBuilder) HorizontalTo(x float64) *PathBuilder {
	pb.commands = append(pb.commands, "H "+fmtFloat(x))
	return pb
}

// VerticalTo adds a vertical line command (V)
func (pb *PathBuilder) VerticalTo(y float64) *PathBuilder {
	pb.commands = append(pb.commands, "V "+fmtFloat(y))
	return pb
}

// QuadraticTo adds a quadratic bezier curve (Q)
func (pb *PathBuilder) QuadraticTo(cx, cy, x, y float64) *PathBuilder {
	pb.commands = append(pb.commands, "Q "+fmtFloat(cx)+" "+fmtFloat(cy)+" "+fmtFloat(x)+" "+fmtFloat(y))
	return pb
}

// CubicTo adds a cubic bezier curve (C)
func (pb *PathBuilder) CubicTo(c1x, c1y, c2x, c2y, x, y float64) *PathBuilder {
	pb.commands = append(pb.commands, "C "+fmtFloat(c1x)+" "+fmtFloat(c1y)+" "+fmtFloat(c2x)+" "+fmtFloat(c2y)+" "+fmtFloat(x)+" "+fmtFloat(y))
	return pb
}

// ArcTo adds an arc command (A)
func (pb *PathBuilder) ArcTo(rx, ry, rotation float64, largeArc, sweep bool, x, y float64) *PathBuilder {
	la, sw := 0, 0
	if largeArc {
		la = 1
	}
	if sweep {
		sw = 1
	}
	pb.commands = append(pb.commands, "A "+fmtFloat(rx)+" "+fmtFloat(ry)+" "+fmtFloat(rotation)+" "+strconv.Itoa(la)+" "+strconv.Itoa(sw)+" "+fmtFloat(x)+" "+fmtFloat(y))
	return pb
}

// String returns the complete path data
func (pb *PathBuilder) String() string {
	return joinPath(pb.commands)
}

func joinPath(commands []string) string {
	result := ""
	for i, cmd := range commands {
		if i > 0 {
			result += " "
		}
		result += cmd
	}
	return result
}
