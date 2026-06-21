//go:build windows

package winapp

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"mofumouse/internal/core"
)

const (
	companionSideLeft  int32 = -1
	companionSideRight int32 = 1
	cursorGap          int32 = 6
	spritePadding      int32 = 6
	movingSpriteScale        = 1.15
	movingSpriteHold         = 450 * time.Millisecond
	assistScanInterval       = 350 * time.Millisecond
	assistTargetGrace        = 8 * time.Second
)

var runtimeSpriteTiers = []int32{32, 48, 64, 96}

type Overlay struct {
	state *core.State
	hwnd  atomic.Uintptr

	sprites map[string]map[int32]*dibSprite

	companionSide    int32
	lastCursor       point
	lastCursorValid  bool
	lastMove         time.Time
	lastTick         time.Time
	lastAssistScan   time.Time
	lastAssistSeen   time.Time
	assistForeground uintptr
	assistTarget     AssistTarget
	assistValid      bool
	renderX          float64
	renderY          float64
	renderValid      bool
}

func NewOverlay(state *core.State) *Overlay {
	return &Overlay{state: state, sprites: loadOverlaySprites(), companionSide: companionSideLeft}
}

func loadOverlaySprites() map[string]map[int32]*dibSprite {
	poses := []string{
		"idle",
		"front",
		"walk",
		"walk2",
		"run",
		"return",
		"guard",
		"patpat",
		"sleepy",
		"sniff",
		"groom",
		"nibble",
		"dig",
		"roll",
	}
	coats := []core.CoatColor{
		core.CoatAgouti,
		core.CoatGray,
		core.CoatDark,
		core.CoatCream,
		core.CoatWhite,
		core.CoatPied,
	}
	sprites := make(map[string]map[int32]*dibSprite, len(poses)*len(coats))
	for _, pose := range poses {
		for _, coat := range coats {
			key := spriteKey(pose, coat)
			for _, tier := range runtimeSpriteTiers {
				name := fmt.Sprintf("%d/degu_%s_%s.png", tier, pose, coat)
				if sprite, err := loadSprite(name); err == nil {
					if sprites[key] == nil {
						sprites[key] = map[int32]*dibSprite{}
					}
					sprites[key][tier] = sprite
				}
			}
			name := fmt.Sprintf("degu_%s_%s.png", pose, coat)
			if sprite, err := loadSprite(name); err == nil {
				if sprites[key] == nil {
					sprites[key] = map[int32]*dibSprite{}
				}
				if sprites[key][32] == nil {
					sprites[key][32] = sprite
				}
			}
		}
	}
	return sprites
}

func (o *Overlay) Run(ctx context.Context) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	className := utf16Ptr("MofuMouseOverlayWindow")
	instance := getModuleHandle()
	wndProc := makeCallback(o.windowProc)

	wc := wndClassEx{
		Size:      uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:   wndProc,
		Instance:  instance,
		Cursor:    loadCursor(idcArrow),
		ClassName: className,
	}
	if ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 {
		return fmt.Errorf("RegisterClassExW: %w", err)
	}

	width, height := o.windowSize(time.Now())
	exStyle := uintptr(wsExLayered | wsExTransparent | wsExTopmost | wsExNoActivate | wsExToolWindow)
	hwnd, _, err := procCreateWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("MofuMouse"))),
		uintptr(wsPopup),
		0, 0, uintptr(width), uintptr(height),
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW: %w", err)
	}
	o.hwnd.Store(hwnd)

	procSetLayeredWindowAttrs.Call(hwnd, transparentColor, 0, lwaColorKey)
	procSetTimer.Call(hwnd, 1, 33, 0)
	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)

	go func() {
		<-ctx.Done()
		if h := o.hwnd.Load(); h != 0 {
			procPostMessageW.Call(h, wmAppStop, 0, 0)
		}
	}()

	var m msg
	for {
		ret, _, callErr := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			return fmt.Errorf("GetMessageW: %w", callErr)
		}
		if ret == 0 {
			return nil
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (o *Overlay) windowProc(hwnd uintptr, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case wmTimer:
		o.followCursor(hwnd)
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0
	case wmPaint:
		o.paint(hwnd)
		return 0
	case wmClose, wmAppStop:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		o.hwnd.Store(0)
		procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
		return ret
	}
}

func (o *Overlay) followCursor(hwnd uintptr) {
	if !o.state.AssistantEnabled() {
		procShowWindow.Call(hwnd, swHide)
		return
	}
	procShowWindow.Call(hwnd, swShow)
	pos, ok := CursorPosition()
	if !ok {
		return
	}

	now := time.Now()
	if !o.lastCursorValid {
		o.lastCursor = pos
		o.lastCursorValid = true
	} else if pos.X != o.lastCursor.X || pos.Y != o.lastCursor.Y {
		o.updateCompanionSide(pos.X - o.lastCursor.X)
		o.lastCursor = pos
		o.lastMove = now
	}
	width, height := o.windowSize(now)
	targetX, targetY := o.targetWindowPosition(pos, width, height, o.spriteDrawSize(now))
	o.updateRenderPosition(now, targetX, targetY)
	x := int32(math.Round(o.renderX))
	y := int32(math.Round(o.renderY))
	procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 1)
}

func (o *Overlay) targetWindowPosition(pos point, width, height, spriteSize int32) (float64, float64) {
	x := o.sidePosition(pos, o.companionSide, spriteSize) + int32(o.state.CompanionOffsetX)
	y := pos.Y - spritePadding - spriteSize/2 - int32(o.state.CompanionOffsetY)

	screenX, screenY, screenW, screenH := virtualScreen()
	screenRight := screenX + screenW
	screenBottom := screenY + screenH
	if x < screenX && o.companionSide == companionSideLeft {
		o.companionSide = companionSideRight
		x = o.sidePosition(pos, companionSideRight, spriteSize) + int32(o.state.CompanionOffsetX)
	}
	if x+width > screenRight && o.companionSide == companionSideRight {
		o.companionSide = companionSideLeft
		x = o.sidePosition(pos, companionSideLeft, spriteSize) + int32(o.state.CompanionOffsetX)
	}
	if x < screenX {
		x = screenX
	}
	if y+height > screenBottom {
		y = screenBottom - height
	}
	if y < screenY {
		y = screenY
	}
	return float64(x), float64(y)
}

func (o *Overlay) updateCompanionSide(dx int32) {
	if abs32(dx) < 3 {
		return
	}
	if dx > 0 {
		o.companionSide = companionSideLeft
		return
	}
	o.companionSide = companionSideRight
}

func (o *Overlay) refreshAssistTarget(now time.Time, cursor point) {
	if !o.state.TargetAssistEnabled() {
		o.assistValid = false
		o.assistForeground = 0
		debugAssist("disabled")
		return
	}
	if !o.lastAssistScan.IsZero() && now.Sub(o.lastAssistScan) < assistScanInterval {
		return
	}
	o.lastAssistScan = now

	foreground, _, _ := procGetForegroundWindow.Call()
	ctx := foregroundAssistContext(foreground)
	targets := SafeForegroundAssistTargetsWithRules(o.state.TargetRules())
	if len(targets) == 0 {
		if o.keepAssistTargetOnMiss(now, foreground) {
			debugAssist("no safe targets; keeping previous target label=%q rules=%d app=%q window=%q hwnd=%d", o.assistTarget.Label, len(o.state.TargetRules()), ctx.App, ctx.Window, foreground)
			return
		}
		o.assistValid = false
		o.assistForeground = 0
		debugAssist("no safe targets rules=%d app=%q window=%q hwnd=%d", len(o.state.TargetRules()), ctx.App, ctx.Window, foreground)
		return
	}

	o.assistTarget = targets[0]
	o.assistValid = true
	o.lastAssistSeen = now
	o.assistForeground = foreground
	debugAssist("target label=%q point=%d,%d rules=%d app=%q window=%q hwnd=%d", o.assistTarget.Label, o.assistTarget.Point.X, o.assistTarget.Point.Y, len(o.state.TargetRules()), ctx.App, ctx.Window, foreground)
	if o.assistTarget.Point.X >= cursor.X {
		o.companionSide = companionSideLeft
		return
	}
	o.companionSide = companionSideRight
}

func (o *Overlay) keepAssistTargetOnMiss(now time.Time, foreground uintptr) bool {
	if !o.assistValid || o.assistForeground == 0 || foreground != o.assistForeground || o.lastAssistSeen.IsZero() {
		return false
	}
	return now.Sub(o.lastAssistSeen) <= assistTargetGrace
}

func debugAssist(format string, args ...any) {
	path := os.Getenv("MOFUMOUSE_ASSIST_LOG")
	if path == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintf(file, time.Now().Format(time.RFC3339Nano)+" "+format, args...)
	fmt.Fprintln(file)
}

func (o *Overlay) sidePosition(pos point, side, spriteSize int32) int32 {
	if side == companionSideRight {
		return pos.X + cursorGap - spritePadding
	}
	return pos.X - cursorGap - spritePadding - spriteSize
}

func abs32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}

func (o *Overlay) updateRenderPosition(now time.Time, targetX, targetY float64) {
	if !o.renderValid {
		o.renderX = targetX
		o.renderY = targetY
		o.renderValid = true
		o.lastTick = now
		return
	}

	dx := targetX - o.renderX
	dy := targetY - o.renderY
	dist := math.Hypot(dx, dy)
	if dist > 180 {
		o.renderX = targetX
		o.renderY = targetY
		o.lastTick = now
		return
	}

	dt := now.Sub(o.lastTick)
	if dt <= 0 || dt > 120*time.Millisecond {
		dt = 33 * time.Millisecond
	}
	o.lastTick = now

	// 30 Hz timer: quick catch-up keeps the companion beside the cursor without snapping.
	alpha := 1 - math.Pow(0.55, float64(dt)/(33*float64(time.Millisecond)))
	o.renderX += dx * alpha
	o.renderY += dy * alpha
	if math.Hypot(targetX-o.renderX, targetY-o.renderY) < 0.45 {
		o.renderX = targetX
		o.renderY = targetY
	}
}

func (o *Overlay) paint(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	now := time.Now()
	width, height := o.windowSize(now)
	bg := createBrush(transparentColor)
	defer deleteObject(bg)
	r := rect{Left: 0, Top: 0, Right: width, Bottom: height}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), bg)

	spriteSize := o.spriteDrawSize(now)
	sprite, action := o.currentSprite(now, spriteSize)
	x := int32(6)
	y := int32(6 + o.bob(now, action))
	if sprite != nil {
		drawSprite(hdc, sprite, x, y, spriteSize, spriteSize, o.companionSide == companionSideRight)
	} else {
		o.paintPlaceholder(hdc, x, y, spriteSize)
	}

	if o.state.ShowPetName() {
		o.paintName(hdc, spriteSize)
	}
}

func (o *Overlay) currentSprite(now time.Time, targetSize int32) (*dibSprite, string) {
	action := o.state.CurrentIdleAction(now)
	if action == "return" {
		return o.sequenceSprite(now, []string{"return", "walk", "return"}, 150, targetSize), "return"
	}
	if !o.lastMove.IsZero() && now.Sub(o.lastMove) < movingSpriteHold {
		return o.sequenceSprite(now, []string{"walk", "walk2", "run", "walk2"}, 95, targetSize), "move"
	}
	if action != "" {
		switch action {
		case "nibble", "groom", "patpat", "sniff", "sleepy", "dig", "roll", "guard":
			return o.sequenceSprite(now, []string{action, "idle", action}, 260, targetSize), action
		}
	}

	if o.state.Jiggle1px() {
		return o.sequenceSprite(now, []string{"guard", "idle", "guard"}, 480, targetSize), "guard"
	}

	return o.spriteForPose("idle", targetSize), "idle"
}

func (o *Overlay) sequenceSprite(now time.Time, poses []string, frameMS int64, targetSize int32) *dibSprite {
	if len(poses) == 0 {
		return o.spriteForPose("idle", targetSize)
	}
	index := int((now.UnixMilli() / frameMS) % int64(len(poses)))
	for i := range poses {
		pose := poses[(index+i)%len(poses)]
		if sprite := o.spriteForPose(pose, targetSize); sprite != nil {
			return sprite
		}
	}
	return o.spriteForPose("idle", targetSize)
}

func (o *Overlay) spriteForPose(pose string, targetSize int32) *dibSprite {
	coat := o.state.CoatColor()
	if sprite := o.bestSprite(spriteKey(pose, coat), targetSize); sprite != nil {
		return sprite
	}
	if coat != core.CoatAgouti {
		if sprite := o.bestSprite(spriteKey("idle", coat), targetSize); sprite != nil {
			return sprite
		}
	}
	if sprite := o.bestSprite(spriteKey(pose, core.CoatAgouti), targetSize); sprite != nil {
		return sprite
	}
	return o.bestSprite(spriteKey("idle", core.CoatAgouti), targetSize)
}

func (o *Overlay) bestSprite(key string, targetSize int32) *dibSprite {
	tiers := o.sprites[key]
	if len(tiers) == 0 {
		return nil
	}
	var largestTier int32
	var largest *dibSprite
	for _, tier := range runtimeSpriteTiers {
		sprite := tiers[tier]
		if sprite == nil {
			continue
		}
		if tier >= targetSize {
			return sprite
		}
		if tier > largestTier {
			largestTier = tier
			largest = sprite
		}
	}
	return largest
}

func spriteKey(pose string, coat core.CoatColor) string {
	return pose + ":" + string(coat)
}

func (o *Overlay) bob(now time.Time, action string) int32 {
	switch action {
	case "idle":
		if (now.UnixMilli()/700)%2 == 0 {
			return 1
		}
	case "move":
		if (now.UnixMilli()/140)%2 == 0 {
			return -1
		}
	case "nibble", "groom", "patpat", "sniff":
		if (now.UnixMilli()/220)%2 == 0 {
			return 1
		}
	case "guard":
		if (now.UnixMilli()/500)%2 == 0 {
			return 1
		}
	}
	return 0
}

func (o *Overlay) windowSize(now time.Time) (int32, int32) {
	size := o.spriteDrawSize(now)
	width := size + 12
	height := size + 12
	if o.state.ShowPetName() {
		width = max32(width, 92)
		height += 22
	}
	return width, height
}

func (o *Overlay) spriteDrawSize(now time.Time) int32 {
	size := int32(o.state.SpriteSizePx())
	if !o.lastMove.IsZero() && now.Sub(o.lastMove) < movingSpriteHold {
		return int32(math.Round(float64(size) * movingSpriteScale))
	}
	return size
}

func (o *Overlay) paintName(hdc uintptr, spriteSize int32) {
	font := createFont(14, "Yu Gothic UI")
	var oldFont uintptr
	if font != 0 {
		oldFont, _, _ = procSelectObject.Call(hdc, font)
		defer deleteObject(font)
	}
	labelBrush := createBrush(0x00EDE8D6)
	defer deleteObject(labelBrush)
	nameRect := rect{Left: 2, Top: spriteSize + 8, Right: 90, Bottom: spriteSize + 28}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&nameRect)), labelBrush)
	procSetBkMode.Call(hdc, 1)
	procSetTextColor.Call(hdc, 0x00000000)
	name, nameLen := stringToUTF16Len(o.state.Name())
	procTextOutW.Call(hdc, 5, uintptr(spriteSize+10), uintptr(unsafe.Pointer(name)), nameLen)
	if oldFont != 0 {
		procSelectObject.Call(hdc, oldFont)
	}
}

func (o *Overlay) paintPlaceholder(hdc uintptr, x, y, size int32) {
	body := createBrush(0x006C8C9C)
	defer deleteObject(body)
	eye := createBrush(0x00000000)
	defer deleteObject(eye)

	old, _, _ := procSelectObject.Call(hdc, body)
	procEllipse.Call(hdc, uintptr(x), uintptr(y+size/4), uintptr(x+size), uintptr(y+size))
	procSelectObject.Call(hdc, eye)
	procEllipse.Call(hdc, uintptr(x+size*2/3), uintptr(y+size/2), uintptr(x+size*2/3+3), uintptr(y+size/2+3))
	procSelectObject.Call(hdc, old)
}

func loadCursor(id uintptr) uintptr {
	ret, _, _ := procLoadCursorW.Call(0, id)
	return ret
}

func createBrush(color uintptr) uintptr {
	ret, _, _ := procCreateSolidBrush.Call(color)
	return ret
}

func createFont(height int32, face string) uintptr {
	ret, _, _ := procCreateFontW.Call(
		uintptr(int(height)),
		0, 0, 0,
		400,
		0, 0, 0,
		128,
		0, 0, 0, 0,
		uintptr(unsafe.Pointer(utf16Ptr(face))),
	)
	return ret
}

func deleteObject(obj uintptr) {
	if obj != 0 {
		procDeleteObject.Call(obj)
	}
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
