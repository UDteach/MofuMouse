//go:build windows

package winapp

import (
	"image"
	"image/color"
	"testing"
)

func TestTrayIconImageUsesCursorCompanionSprite(t *testing.T) {
	src, err := embeddedSpriteImage("mofumouse_cursor_companion_icon.png")
	if err != nil {
		t.Fatalf("read cursor companion icon: %v", err)
	}
	expected := resizeArea(src, 16)
	got := TrayIconImage(true)

	if !sameImage(got, expected) {
		t.Fatal("tray icon does not match the cursor companion asset")
	}
}

func TestInactiveTrayIconIsGrayscale(t *testing.T) {
	got := TrayIconImage(false)
	bounds := got.Bounds()
	if bounds.Dx() != 16 || bounds.Dy() != 16 {
		t.Fatalf("tray icon bounds = %v, want 16x16", bounds)
	}

	opaquePixels := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := got.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			opaquePixels++
			if r != g || g != b {
				t.Fatalf("inactive icon pixel at %d,%d is not grayscale", x, y)
			}
		}
	}
	if opaquePixels == 0 {
		t.Fatal("inactive tray icon has no visible pixels")
	}
}

func sameImage(a, b image.Image) bool {
	if !a.Bounds().Eq(b.Bounds()) {
		return false
	}
	bounds := a.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.NRGBAModel.Convert(a.At(x, y)) != color.NRGBAModel.Convert(b.At(x, y)) {
				return false
			}
		}
	}
	return true
}
