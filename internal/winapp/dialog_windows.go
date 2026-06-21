//go:build windows

package winapp

import (
	"time"
	"unsafe"
)

func ConfirmYesNo(title, message string) bool {
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(message))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		uintptr(mbYesNo|mbIconQuestion|mbDefaultNo|mbSetForeground),
	)
	return ret == idYes
}

func WaitForProcessExit(pid int, timeout time.Duration) bool {
	if pid <= 0 {
		return true
	}
	handle, _, _ := procOpenProcess.Call(processSynchronize, 0, uintptr(uint32(pid)))
	if handle == 0 {
		return true
	}
	defer procCloseHandle.Call(handle)

	milliseconds := uintptr(^uint32(0))
	if timeout >= 0 {
		milliseconds = uintptr(timeout / time.Millisecond)
	}
	ret, _, _ := procWaitForSingleObject.Call(handle, milliseconds)
	return ret == waitObject0
}
