//go:build windows

package winapp

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"unsafe"

	"mofumouse/internal/core"
)

type SettingsWindowHooks struct {
	Save         func() error
	Refresh      func()
	OpenConfig   func()
	CheckUpdates func()
}

type settingsWindow struct {
	state *core.State
	hooks SettingsWindowHooks

	hwnd        uintptr
	callback    uintptr
	font        uintptr
	pendingSize int

	nameEdit    uintptr
	sizeLabel   uintptr
	coatCombo   uintptr
	assistant   uintptr
	launch      uintptr
	quickKeys   uintptr
	quickScroll uintptr
	jiggle      uintptr
	idleActions uintptr
	edgeWarp    uintptr
	showName    uintptr
	status      uintptr
}

type coatOption struct {
	label string
	coat  core.CoatColor
}

var (
	settingsWindowsMu sync.Mutex
	settingsWindows   = map[uintptr]*settingsWindow{}
	coatOptions       = []coatOption{
		{label: "ノーマル（茶色）", coat: core.CoatAgouti},
		{label: "グレー", coat: core.CoatGray},
		{label: "ダークブラウン", coat: core.CoatDark},
		{label: "クリーム", coat: core.CoatCream},
		{label: "ホワイト", coat: core.CoatWhite},
		{label: "ぶち模様", coat: core.CoatPied},
	}
)

const (
	settingsIDName         = 1001
	settingsIDCoat         = 1002
	settingsIDAssistant    = 1010
	settingsIDLaunch       = 1011
	settingsIDQuickKeys    = 1012
	settingsIDQuickScroll  = 1023
	settingsIDJiggle       = 1017
	settingsIDIdle         = 1018
	settingsIDEdgeWarp     = 1019
	settingsIDShowName     = 1022
	settingsIDSizeDown     = 1020
	settingsIDSizeUp       = 1021
	settingsIDApply        = 1030
	settingsIDOpenConfig   = 1031
	settingsIDCheckUpdates = 1032
	settingsIDClose        = 1033
)

func RunSettingsWindow(ctx context.Context, state *core.State, hooks SettingsWindowHooks) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	w := &settingsWindow{
		state:       state,
		hooks:       hooks,
		pendingSize: state.SpriteSizePx(),
	}
	if err := w.create(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		if w.hwnd != 0 {
			procPostMessageW.Call(w.hwnd, wmAppStop, 0, 0)
		}
	}()

	var m msg
	for {
		ret, _, callErr := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			return callErr
		}
		if ret == 0 {
			return nil
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (w *settingsWindow) create() error {
	className := utf16Ptr("MofuMouseSettingsWindow")
	instance := getModuleHandle()
	w.callback = makeCallback(settingsWindowProc)

	wc := wndClassEx{
		Size:       uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:    w.callback,
		Instance:   instance,
		Cursor:     loadCursor(idcArrow),
		ClassName:  className,
		Background: uintptr(6),
	}
	if ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 && err != syscall.Errno(1410) {
		return fmt.Errorf("RegisterClassExW settings: %w", err)
	}

	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("MofuMouse の設定"))),
		uintptr(wsOverlappedWindow|wsVisible),
		180, 24, 560, 980,
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW settings: %w", err)
	}
	w.hwnd = hwnd
	w.font = createFont(16, "Yu Gothic UI")

	settingsWindowsMu.Lock()
	settingsWindows[hwnd] = w
	settingsWindowsMu.Unlock()

	w.createControls()
	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)
	return nil
}

func (w *settingsWindow) createControls() {
	w.staticText("MofuMouse の設定", 18, 16, 480, 24)
	w.staticText("デグーの名前", 24, 56, 110, 22)
	w.nameEdit = w.child("EDIT", w.state.Name(), wsBorder|esAutoHScroll|wsTabStop, 150, 52, 260, 26, settingsIDName)
	w.hint("トレイの状態表示に使います。名前を画面に出すかは下で選べます。", 150, 82)

	w.staticText("デグーの大きさ", 24, 120, 120, 22)
	w.sizeLabel = w.staticText("", 150, 120, 82, 22)
	w.button("小さく", 248, 116, 80, 28, settingsIDSizeDown)
	w.button("大きく", 338, 116, 80, 28, settingsIDSizeUp)
	w.hint("カーソル横に出るデグーの表示サイズです。見づらい時だけ大きくします。", 150, 150)
	w.updateSizeLabel()

	w.staticText("毛色", 24, 188, 110, 22)
	w.coatCombo = w.child("COMBOBOX", "", cbsDropDownList|cbsHasStrings|wsTabStop, 150, 184, 220, 160, settingsIDCoat)
	for _, option := range coatOptions {
		procSendMessageW.Call(w.coatCombo, cbAddString, 0, uintptr(unsafe.Pointer(utf16Ptr(option.label))))
	}
	procSendMessageW.Call(w.coatCombo, cbSetCurSel, uintptr(coatIndex(w.state.CoatColor())), 0)
	w.hint("見た目だけの設定です。あとから何度でも変えられます。", 150, 214)

	w.staticText("動きと補助機能", 24, 250, 480, 22)
	w.assistant = w.checkbox("カーソル横にデグーを表示する", 24, 284, w.state.AssistantEnabled(), settingsIDAssistant)
	w.hint("マウスカーソルのすぐ横で、移動方向に合わせて左右へ回ります。", 48, 308)
	w.jiggle = w.checkbox("スクリーンセーバー防止の1px移動", 24, 336, w.state.Jiggle1px(), settingsIDJiggle)
	w.hint("しばらく触っていない時だけ、マウスを1pxだけ戻して動かします。", 48, 360)
	w.idleActions = w.checkbox("止まっている時にしぐさを出す", 24, 388, w.state.IdleActionsEnabled(), settingsIDIdle)
	w.hint("かじる、毛づくろい、ぺちぺちなどの待機アクションを出します。", 48, 412)
	w.edgeWarp = w.checkbox("画面端で反対側へ移動する", 24, 440, w.state.EdgeWarpEnabled(), settingsIDEdgeWarp)
	w.hint("カーソルが画面の端に触れた時、反対側へ移動します。慣れてから使う設定です。", 48, 464)
	w.showName = w.checkbox("デグーの名前を画面に出す", 24, 492, w.state.ShowPetName(), settingsIDShowName)
	w.hint("普段はオフ推奨です。オンにするとデグーの下に名前を表示します。", 48, 516)
	w.launch = w.checkbox("Windows起動時にMofuMouseを始める", 24, 544, w.state.LaunchAtLogin(), settingsIDLaunch)
	w.hint("サインインしたら自動でMofuMouseを起動します。いつでもオフにできます。", 48, 568)
	w.quickKeys = w.checkbox("クイックキーを使う", 24, 596, w.state.QuickKeysEnabled(), settingsIDQuickKeys)
	w.hint("検索、コピー、全選択、新しいタブだけをトレイから実行できます。", 48, 620)
	w.quickScroll = w.checkbox("クイックスクロールを使う", 24, 648, w.state.QuickScrollEnabled(), settingsIDQuickScroll)
	w.hint("トレイから、今のウィンドウを少しだけ上下にスクロールできます。", 48, 672)
	w.button("保存して反映", 24, 780, 110, 32, settingsIDApply)
	w.button("設定ファイルを開く", 146, 780, 130, 32, settingsIDOpenConfig)
	w.button("アップデート確認", 288, 780, 124, 32, settingsIDCheckUpdates)
	w.button("閉じる", 424, 780, 70, 32, settingsIDClose)
	w.status = w.staticText("変更したら「保存して反映」を押してください。", 24, 824, 480, 22)
}

func (w *settingsWindow) child(className, text string, style uintptr, x, y, width, height int32, id uintptr) uintptr {
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(wsChild|wsVisible)|style,
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		w.hwnd, id, getModuleHandle(), 0,
	)
	if w.font != 0 && hwnd != 0 {
		procSendMessageW.Call(hwnd, wmSetFont, w.font, 1)
	}
	return hwnd
}

func (w *settingsWindow) staticText(text string, x, y, width, height int32) uintptr {
	return w.child("STATIC", text, 0, x, y, width, height, 0)
}

func (w *settingsWindow) hint(text string, x, y int32) uintptr {
	return w.staticText(text, x, y, 470, 22)
}

func (w *settingsWindow) button(text string, x, y, width, height int32, id uintptr) uintptr {
	return w.child("BUTTON", text, bsPushButton|wsTabStop, x, y, width, height, id)
}

func (w *settingsWindow) checkbox(text string, x, y int32, checked bool, id uintptr) uintptr {
	hwnd := w.child("BUTTON", text, bsAutoCheckBox|wsTabStop, x, y, 360, 26, id)
	if checked {
		procSendMessageW.Call(hwnd, bmSetCheck, bstChecked, 0)
	}
	return hwnd
}

func settingsWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	settingsWindowsMu.Lock()
	w := settingsWindows[hwnd]
	settingsWindowsMu.Unlock()

	switch message {
	case wmCommand:
		if w != nil {
			w.handleCommand(wParam)
			return 0
		}
	case wmClose, wmAppStop:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		if w != nil && w.font != 0 {
			deleteObject(w.font)
		}
		settingsWindowsMu.Lock()
		delete(settingsWindows, hwnd)
		settingsWindowsMu.Unlock()
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return ret
}

func (w *settingsWindow) handleCommand(wParam uintptr) {
	id := wParam & 0xffff
	switch id {
	case settingsIDSizeDown:
		w.pendingSize -= 8
		if w.pendingSize < 16 {
			w.pendingSize = 16
		}
		w.updateSizeLabel()
	case settingsIDSizeUp:
		w.pendingSize += 8
		if w.pendingSize > 96 {
			w.pendingSize = 96
		}
		w.updateSizeLabel()
	case settingsIDApply:
		w.apply()
	case settingsIDOpenConfig:
		if w.hooks.OpenConfig != nil {
			w.hooks.OpenConfig()
		}
	case settingsIDCheckUpdates:
		if w.hooks.CheckUpdates != nil {
			w.hooks.CheckUpdates()
		}
		w.setStatus("アップデート確認を開始しました。")
	case settingsIDClose:
		procDestroyWindow.Call(w.hwnd)
	}
}

func (w *settingsWindow) apply() {
	w.state.SetName(windowText(w.nameEdit))
	w.state.SetSpriteSizePx(w.pendingSize)
	w.state.SetCoatColor(coatOptions[comboIndex(w.coatCombo)].coat)
	w.state.SetAssistantEnabled(checked(w.assistant))
	w.state.SetLaunchAtLogin(checked(w.launch))
	w.state.SetQuickKeysEnabled(checked(w.quickKeys))
	w.state.SetQuickScrollEnabled(checked(w.quickScroll))
	w.state.SetJiggle1px(checked(w.jiggle))
	w.state.SetIdleActionsEnabled(checked(w.idleActions))
	w.state.SetEdgeWarpEnabled(checked(w.edgeWarp))
	w.state.SetShowPetName(checked(w.showName))

	if w.hooks.Save != nil {
		if err := w.hooks.Save(); err != nil {
			w.setStatus("保存に失敗: " + err.Error())
			return
		}
	}
	if w.hooks.Refresh != nil {
		w.hooks.Refresh()
	}
	w.setStatus("設定を保存しました。")
}

func (w *settingsWindow) updateSizeLabel() {
	procSetWindowTextW.Call(w.sizeLabel, uintptr(unsafe.Pointer(utf16Ptr(strconv.Itoa(w.pendingSize)+" px"))))
}

func (w *settingsWindow) setStatus(text string) {
	procSetWindowTextW.Call(w.status, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func windowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	buf := make([]uint16, int(length)+1)
	if len(buf) == 0 {
		return ""
	}
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func checked(hwnd uintptr) bool {
	ret, _, _ := procSendMessageW.Call(hwnd, bmGetCheck, 0, 0)
	return ret == bstChecked
}

func comboIndex(hwnd uintptr) int {
	index, _, _ := procSendMessageW.Call(hwnd, cbGetCurSel, 0, 0)
	if int(index) < 0 || int(index) >= len(coatOptions) {
		return 0
	}
	return int(index)
}

func coatIndex(coat core.CoatColor) int {
	for i, option := range coatOptions {
		if option.coat == coat {
			return i
		}
	}
	return 0
}
