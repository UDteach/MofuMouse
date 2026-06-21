//go:build windows

package winapp

import (
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shcore   = syscall.NewLazyDLL("shcore.dll")

	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procDestroyWindow                 = user32.NewProc("DestroyWindow")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procEnumChildWindows              = user32.NewProc("EnumChildWindows")
	procGetClassNameW                 = user32.NewProc("GetClassNameW")
	procGetCursorPos                  = user32.NewProc("GetCursorPos")
	procGetForegroundWindow           = user32.NewProc("GetForegroundWindow")
	procGetLastInputInfo              = user32.NewProc("GetLastInputInfo")
	procGetMessageW                   = user32.NewProc("GetMessageW")
	procGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
	procGetWindowRect                 = user32.NewProc("GetWindowRect")
	procGetWindowTextLengthW          = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW                = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcID         = user32.NewProc("GetWindowThreadProcessId")
	procCloseHandle                   = kernel32.NewProc("CloseHandle")
	procGetModuleHandleW              = kernel32.NewProc("GetModuleHandleW")
	procGetTickCount                  = kernel32.NewProc("GetTickCount")
	procInvalidateRect                = user32.NewProc("InvalidateRect")
	procLoadCursorW                   = user32.NewProc("LoadCursorW")
	procMessageBoxW                   = user32.NewProc("MessageBoxW")
	procMoveWindow                    = user32.NewProc("MoveWindow")
	procKeybdEvent                    = user32.NewProc("keybd_event")
	procMouseEvent                    = user32.NewProc("mouse_event")
	procPostQuitMessage               = user32.NewProc("PostQuitMessage")
	procPostMessageW                  = user32.NewProc("PostMessageW")
	procSendMessageW                  = user32.NewProc("SendMessageW")
	procSetCursorPos                  = user32.NewProc("SetCursorPos")
	procSetLayeredWindowAttrs         = user32.NewProc("SetLayeredWindowAttributes")
	procSetProcessDPIAware            = user32.NewProc("SetProcessDPIAware")
	procSetProcessDPIAwareness        = shcore.NewProc("SetProcessDpiAwareness")
	procSetProcessDPIAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procSetStretchBltMode             = gdi32.NewProc("SetStretchBltMode")
	procSetWindowTextW                = user32.NewProc("SetWindowTextW")
	procSetThreadExecution            = kernel32.NewProc("SetThreadExecutionState")
	procSetTimer                      = user32.NewProc("SetTimer")
	procShowWindow                    = user32.NewProc("ShowWindow")
	procIsWindowEnabled               = user32.NewProc("IsWindowEnabled")
	procIsWindowVisible               = user32.NewProc("IsWindowVisible")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procUpdateWindow                  = user32.NewProc("UpdateWindow")
	procBeginPaint                    = user32.NewProc("BeginPaint")
	procEndPaint                      = user32.NewProc("EndPaint")
	procCreateSolidBrush              = gdi32.NewProc("CreateSolidBrush")
	procCreateFontW                   = gdi32.NewProc("CreateFontW")
	procDeleteObject                  = gdi32.NewProc("DeleteObject")
	procEllipse                       = gdi32.NewProc("Ellipse")
	procFillRect                      = user32.NewProc("FillRect")
	procMoveToEx                      = gdi32.NewProc("MoveToEx")
	procOpenProcess                   = kernel32.NewProc("OpenProcess")
	procLineTo                        = gdi32.NewProc("LineTo")
	procRectangle                     = gdi32.NewProc("Rectangle")
	procSelectObject                  = gdi32.NewProc("SelectObject")
	procSetBkMode                     = gdi32.NewProc("SetBkMode")
	procSetTextColor                  = gdi32.NewProc("SetTextColor")
	procStretchDIBits                 = gdi32.NewProc("StretchDIBits")
	procTextOutW                      = gdi32.NewProc("TextOutW")
	procQueryFullProcessImage         = kernel32.NewProc("QueryFullProcessImageNameW")
	procWaitForSingleObject           = kernel32.NewProc("WaitForSingleObject")
)

const (
	cwUseDefault = 0x80000000

	wsPopup            = 0x80000000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsTabStop          = 0x00010000
	wsBorder           = 0x00800000
	wsOverlappedWindow = 0x00CF0000

	wsExLayered     = 0x00080000
	wsExTransparent = 0x00000020
	wsExTopmost     = 0x00000008
	wsExNoActivate  = 0x08000000
	wsExToolWindow  = 0x00000080

	lwaColorKey = 0x00000001

	swShow    = 5
	swHide    = 0
	wmCommand = 0x0111
	wmPaint   = 0x000F
	wmSetFont = 0x0030
	wmTimer   = 0x0113
	wmClose   = 0x0010
	wmDestroy = 0x0002
	wmAppStop = 0x8001

	bsPushButton    = 0x00000000
	bsAutoCheckBox  = 0x00000003
	esAutoHScroll   = 0x00000080
	cbsDropDownList = 0x00000003
	cbsHasStrings   = 0x00000200

	bmGetCheck = 0x00F0
	bmSetCheck = 0x00F1
	bstChecked = 1

	cbAddString = 0x0143
	cbGetCurSel = 0x0147
	cbSetCurSel = 0x014E

	idcArrow = 32512

	transparentColor    = 0x00FF00FF
	colorOnColorStretch = 3

	processQueryLimitedInformation = 0x1000
	processSynchronize             = 0x00100000

	waitObject0 = 0x00000000
	waitTimeout = 0x00000102

	mbYesNo         = 0x00000004
	mbIconQuestion  = 0x00000020
	mbDefaultNo     = 0x00000100
	mbSetForeground = 0x00010000
	idYes           = 6

	dpiAwarenessContextPerMonitorAwareV2 = ^uintptr(3)
	processPerMonitorDPIAware            = 2

	smXVirtualScreen  = 76
	smYVirtualScreen  = 77
	smCXVirtualScreen = 78
	smCYVirtualScreen = 79
)

func utf16Ptr(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func getModuleHandle() uintptr {
	ret, _, _ := procGetModuleHandleW.Call(0)
	return ret
}

func boolToInt(v bool) uintptr {
	if v {
		return 1
	}
	return 0
}

func makeCallback(fn any) uintptr {
	return syscall.NewCallback(fn)
}

func errnoIfZero(ret uintptr, err error) error {
	if ret != 0 {
		return nil
	}
	if err != syscall.Errno(0) {
		return err
	}
	return syscall.EINVAL
}

func stringToUTF16Len(s string) (*uint16, uintptr) {
	chars := syscall.StringToUTF16(s)
	return &chars[0], uintptr(len(chars) - 1)
}

func ptr(v any) uintptr {
	return uintptr(unsafe.Pointer(&v))
}
