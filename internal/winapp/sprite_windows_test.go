//go:build windows

package winapp

import "testing"

func TestCompositeOverMagentaDropsFaintAlphaPixels(t *testing.T) {
	r, g, b := compositeOverMagenta(10, 20, 30, spriteAlphaCutoff-1)
	if r != 255 || g != 0 || b != 255 {
		t.Fatalf("faint alpha pixel = %d,%d,%d, want transparent magenta", r, g, b)
	}
}

func TestCompositeOverMagentaKeepsOpaquePixels(t *testing.T) {
	r, g, b := compositeOverMagenta(10, 20, 30, spriteAlphaCutoff)
	if r != 10 || g != 20 || b != 30 {
		t.Fatalf("opaque-enough pixel = %d,%d,%d, want original color", r, g, b)
	}
}
