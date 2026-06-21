//go:build windows

package winapp

import "syscall"

func EnableDPIAwareness() {
	if ret, _, _ := procSetProcessDPIAwarenessContext.Call(dpiAwarenessContextPerMonitorAwareV2); ret != 0 {
		return
	}
	if ret, _, err := procSetProcessDPIAwareness.Call(uintptr(processPerMonitorDPIAware)); ret == 0 && err == syscall.Errno(0) {
		return
	}
	procSetProcessDPIAware.Call()
}
