package cli

import "testing"

func TestParseVariant(t *testing.T) {
	c, err := ParseVariant("H[LC]F")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if _, ok := c.regions[RegionHeader]; !ok {
		t.Error("Expected Header region")
	}
	if _, ok := c.regions[RegionFooter]; !ok {
		t.Error("Expected Footer region")
	}
	
	hSlot := c.regions[RegionHeader]
	if hSlot.child == nil {
		t.Error("Header should have child layout")
	} else {
		if _, ok := hSlot.child.regions[RegionLeft]; !ok {
			t.Error("Child should have Left region")
		}
	}
}
