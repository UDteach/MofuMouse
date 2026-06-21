//go:build windows && !amd64

package winapp

func uiaForegroundAssistTargets(foreground uintptr) []AssistTarget {
	return nil
}
