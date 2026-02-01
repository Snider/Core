package cli

import "fmt"

// CheckBuilder provides fluent API for check results.
type CheckBuilder struct {
	name     string
	status   string
	style    *AnsiStyle
	icon     string
	duration string
}

// Check starts building a check result line.
//
//	cli.Check("audit").Pass()
//	cli.Check("fmt").Fail().Duration("2.3s")
//	cli.Check("test").Skip()
func Check(name string) *CheckBuilder {
	return &CheckBuilder{name: name}
}

// Pass marks the check as passed.
func (c *CheckBuilder) Pass() *CheckBuilder {
	c.status = "passed"
	c.style = SuccessStyle
	c.icon = Glyph(":check:")
	return c
}

// Fail marks the check as failed.
func (c *CheckBuilder) Fail() *CheckBuilder {
	c.status = "failed"
	c.style = ErrorStyle
	c.icon = Glyph(":cross:")
	return c
}

// Skip marks the check as skipped.
func (c *CheckBuilder) Skip() *CheckBuilder {
	c.status = "skipped"
	c.style = DimStyle
	c.icon = "-"
	return c
}

// Warn marks the check as warning.
func (c *CheckBuilder) Warn() *CheckBuilder {
	c.status = "warning"
	c.style = WarningStyle
	c.icon = Glyph(":warn:")
	return c
}

// Duration adds duration to the check result.
func (c *CheckBuilder) Duration(d string) *CheckBuilder {
	c.duration = d
	return c
}

// Message adds a custom message instead of status.
func (c *CheckBuilder) Message(msg string) *CheckBuilder {
	c.status = msg
	return c
}

// String returns the formatted check line.
func (c *CheckBuilder) String() string {
	icon := c.icon
	if c.style != nil {
		icon = c.style.Render(c.icon)
	}

	status := c.status
	if c.style != nil && c.status != "" {
		status = c.style.Render(c.status)
	}

	if c.duration != "" {
		return fmt.Sprintf("  %s %-20s %-10s %s", icon, c.name, status, DimStyle.Render(c.duration))
	}
	if status != "" {
		return fmt.Sprintf("  %s %s %s", icon, c.name, status)
	}
	return fmt.Sprintf("  %s %s", icon, c.name)
}

// Print outputs the check result.
func (c *CheckBuilder) Print() {
	fmt.Println(c.String())
}