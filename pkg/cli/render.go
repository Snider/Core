package cli

import (
	"fmt"
	"strings"
)

// RenderStyle controls how layouts are rendered.
type RenderStyle int

// Render style constants for layout output.
const (
	// RenderFlat uses no borders or decorations.
	RenderFlat RenderStyle = iota
	// RenderSimple uses --- separators between sections.
	RenderSimple
	// RenderBoxed uses Unicode box drawing characters.
	RenderBoxed
)

var currentRenderStyle = RenderFlat

// UseRenderFlat sets the render style to flat (no borders).
func UseRenderFlat() { currentRenderStyle = RenderFlat }

// UseRenderSimple sets the render style to simple (--- separators).
func UseRenderSimple() { currentRenderStyle = RenderSimple }

// UseRenderBoxed sets the render style to boxed (Unicode box drawing).
func UseRenderBoxed() { currentRenderStyle = RenderBoxed }

// Render outputs the layout to terminal.
func (c *Composite) Render() {
	fmt.Print(c.String())
}

// String returns the rendered layout.
func (c *Composite) String() string {
	var sb strings.Builder
	c.renderTo(&sb, 0)
	return sb.String()
}

func (c *Composite) renderTo(sb *strings.Builder, depth int) {
	order := []Region{RegionHeader, RegionLeft, RegionContent, RegionRight, RegionFooter}

	var active []Region
	for _, r := range order {
		if slot, ok := c.regions[r]; ok {
			if len(slot.blocks) > 0 || slot.child != nil {
				active = append(active, r)
			}
		}
	}

	for i, r := range active {
		slot := c.regions[r]
		if i > 0 && currentRenderStyle != RenderFlat {
			c.renderSeparator(sb, depth)
		}
		c.renderSlot(sb, slot, depth)
	}
}

func (c *Composite) renderSeparator(sb *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)
	switch currentRenderStyle {
	case RenderBoxed:
		sb.WriteString(indent + "├" + strings.Repeat("─", 40) + "┤\n")
	case RenderSimple:
		sb.WriteString(indent + strings.Repeat("─", 40) + "\n")
	}
}

func (c *Composite) renderSlot(sb *strings.Builder, slot *Slot, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, block := range slot.blocks {
		for _, line := range strings.Split(block.Render(), "\n") {
			if line != "" {
				sb.WriteString(indent + line + "\n")
			}
		}
	}
	if slot.child != nil {
		slot.child.renderTo(sb, depth+1)
	}
}
