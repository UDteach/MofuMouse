//go:build windows

package winapp

import "fmt"

const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

func SetKeepAwake(enabled bool) error {
	flags := uintptr(esContinuous)
	if enabled {
		flags |= esSystemRequired | esDisplayRequired
	}
	ret, _, err := procSetThreadExecution.Call(flags)
	if ret == 0 {
		return fmt.Errorf("SetThreadExecutionState: %w", err)
	}
	return nil
}
