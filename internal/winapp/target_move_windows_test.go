//go:build windows

package winapp

import "testing"

func TestPointInsideRect(t *testing.T) {
	tests := []struct {
		name string
		p    point
		want bool
	}{
		{name: "inside", p: point{X: 10, Y: 20}, want: true},
		{name: "left edge inside", p: point{X: 0, Y: 20}, want: true},
		{name: "right edge outside", p: point{X: 100, Y: 20}, want: false},
		{name: "bottom edge outside", p: point{X: 10, Y: 100}, want: false},
		{name: "negative y origin inside", p: point{X: 5, Y: -5}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pointInsideRect(tt.p, 0, -10, 100, 110)
			if got != tt.want {
				t.Fatalf("pointInsideRect = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointsNear(t *testing.T) {
	if !pointsNear(point{X: 100, Y: 100}, point{X: 103, Y: 96}, 4) {
		t.Fatal("points should be near")
	}
	if pointsNear(point{X: 100, Y: 100}, point{X: 105, Y: 100}, 4) {
		t.Fatal("points should not be near")
	}
}

func TestNudgePointInsideRect(t *testing.T) {
	tests := []struct {
		name string
		p    point
		rect [4]int32
		want point
		ok   bool
	}{
		{name: "nudge right inside", p: point{X: 10, Y: 20}, rect: [4]int32{0, 0, 100, 100}, want: point{X: 11, Y: 20}, ok: true},
		{name: "nudge left at right edge", p: point{X: 99, Y: 20}, rect: [4]int32{0, 0, 100, 100}, want: point{X: 98, Y: 20}, ok: true},
		{name: "negative monitor origin", p: point{X: -10, Y: 20}, rect: [4]int32{-100, 0, 100, 100}, want: point{X: -9, Y: 20}, ok: true},
		{name: "outside virtual screen", p: point{X: -101, Y: 20}, rect: [4]int32{-100, 0, 100, 100}, ok: false},
		{name: "too narrow", p: point{X: 0, Y: 0}, rect: [4]int32{0, 0, 1, 1}, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := nudgePointInsideRect(tt.p, tt.rect[0], tt.rect[1], tt.rect[2], tt.rect[3])
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Fatalf("point = %+v, want %+v", got, tt.want)
			}
		})
	}
}
