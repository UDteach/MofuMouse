//go:build windows

package winapp

import (
	"context"
	"time"
	"unsafe"

	"mofumouse/internal/core"
)

func RunJiggler(ctx context.Context, state *core.State) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var last point
	var hasLast bool

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !state.Jiggle1px() {
				hasLast = false
				continue
			}
			pos, ok := CursorPosition()
			if !ok {
				continue
			}
			if !hasLast {
				last = pos
				hasLast = true
				continue
			}
			if pos != last {
				last = pos
				continue
			}
			if IdleDuration() < time.Duration(state.JiggleIdleSec)*time.Second {
				continue
			}
			nudgeCursor(pos)
			last = pos
		}
	}
}

func CursorPosition() (point, bool) {
	var p point
	ret, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	return p, ret != 0
}

func IdleDuration() time.Duration {
	var info lastInputInfo
	info.Size = uint32(unsafe.Sizeof(info))
	ret, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return 0
	}
	now, _, _ := procGetTickCount.Call()
	elapsed := uint32(now) - info.Time
	return time.Duration(elapsed) * time.Millisecond
}

func nudgeCursor(p point) {
	x, y, w, h := virtualScreen()
	next, ok := nudgePointInsideRect(p, x, y, w, h)
	if !ok {
		return
	}
	setCursorPosition(next)
	time.Sleep(80 * time.Millisecond)
	setCursorPosition(p)
}

func nudgePointInsideRect(p point, x, y, w, h int32) (point, bool) {
	if !pointInsideRect(p, x, y, w, h) || w < 2 || h < 1 {
		return point{}, false
	}
	next := p
	if p.X+1 < x+w {
		next.X = p.X + 1
		return next, true
	}
	if p.X-1 >= x {
		next.X = p.X - 1
		return next, true
	}
	return point{}, false
}
