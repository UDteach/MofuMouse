//go:build windows && amd64

package winapp

import (
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	uia "github.com/openstandia/w32uiautomation"

	"mofumouse/internal/core"
)

const (
	sFalse         = uintptr(1)
	uiaScanTimeout = 1500 * time.Millisecond
)

var uiaScanBusy atomic.Bool

func uiaForegroundAssistTargets(foreground uintptr) []AssistTarget {
	if !uiaScanBusy.CompareAndSwap(false, true) {
		return nil
	}

	results := make(chan []AssistTarget, 1)
	go func() {
		defer uiaScanBusy.Store(false)
		results <- uiaForegroundAssistTargetsSync(foreground)
	}()

	select {
	case targets := <-results:
		return targets
	case <-time.After(uiaScanTimeout):
		return nil
	}
}

func uiaForegroundAssistTargetsSync(foreground uintptr) (targets []AssistTarget) {
	defer func() {
		if recover() != nil {
			targets = nil
		}
	}()

	uninit, ok := initializeUIAThread()
	if !ok {
		return nil
	}
	defer uninit()

	auto, err := uia.NewUIAutomation()
	if err != nil {
		return nil
	}
	defer auto.Release()

	root, err := uiaElementFromHandle(auto, foreground)
	if err != nil || root == nil {
		return nil
	}
	defer root.Release()

	conditionValue := uia.NewVariantInt(int64(uia.UIA_ButtonControlTypeId))
	condition, err := auto.CreatePropertyCondition(uia.UIA_ControlTypePropertyId, conditionValue)
	if err != nil || condition == nil {
		return nil
	}
	defer condition.Release()

	elements, err := root.FindAll(uia.TreeScope_Descendants, condition)
	if err != nil || elements == nil {
		return nil
	}
	defer elements.Release()

	length, err := elements.Get_Length()
	if err != nil || length <= 0 {
		return nil
	}

	targets = make([]AssistTarget, 0, length)
	for index := int32(0); index < length; index++ {
		element, err := elements.GetElement(index)
		if err != nil || element == nil {
			continue
		}
		if target, ok := uiaElementAssistTarget(element, foreground); ok {
			targets = append(targets, target)
		}
		element.Release()
	}
	sortAssistTargets(targets)
	return targets
}

func initializeUIAThread() (func(), bool) {
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err == nil {
		return ole.CoUninitialize, true
	}
	if oleErr, ok := err.(*ole.OleError); ok && oleErr.Code() == sFalse {
		return ole.CoUninitialize, true
	}
	return func() {}, false
}

func uiaElementAssistTarget(element *uia.IUIAutomationElement, fallbackHWND uintptr) (AssistTarget, bool) {
	if !uiaElementBoolProperty(element, uia.UIA_IsEnabledPropertyId, true) {
		return AssistTarget{}, false
	}
	if uiaElementBoolProperty(element, uia.UIA_IsOffscreenPropertyId, false) {
		return AssistTarget{}, false
	}

	label, err := element.Get_CurrentName()
	if err != nil {
		label = ""
	}
	label = strings.TrimSpace(label)
	if label == "" {
		if id, err := element.Get_CurrentAutomationId(); err == nil {
			label = strings.TrimSpace(id)
		}
	}
	if label == "" {
		return AssistTarget{}, false
	}

	uiaRect, err := element.Get_CurrentBoundingRectangle()
	if err != nil {
		return AssistTarget{}, false
	}
	targetRect := rectFromUIA(uiaRect)
	if targetRect.Right <= targetRect.Left || targetRect.Bottom <= targetRect.Top {
		return AssistTarget{}, false
	}

	targetPoint, ok := uiaClickablePoint(element)
	if !ok {
		targetPoint = point{
			X: targetRect.Left + (targetRect.Right-targetRect.Left)/2,
			Y: targetRect.Top + (targetRect.Bottom-targetRect.Top)/2,
		}
	}

	hwnd := fallbackHWND
	if native, err := element.Get_CurrentNativeWindowHandle(); err == nil && native != 0 {
		hwnd = uintptr(native)
	}

	return AssistTarget{
		HWND:  hwnd,
		Label: label,
		Class: core.ClassifyAssistTarget(label),
		Rect:  targetRect,
		Point: targetPoint,
	}, true
}

func uiaElementBoolProperty(element *uia.IUIAutomationElement, property uia.PROPERTYID, fallback bool) bool {
	value, err := element.Get_CurrentPropertyValue(property)
	if err != nil {
		return fallback
	}
	defer value.Clear()
	if boolValue, ok := value.Value().(bool); ok {
		return boolValue
	}
	return fallback
}

func uiaElementFromHandle(auto *uia.IUIAutomation, hwnd uintptr) (*uia.IUIAutomationElement, error) {
	var element *uia.IUIAutomationElement
	hr, _, _ := syscall.Syscall(
		auto.VTable().ElementFromHandle,
		3,
		uintptr(unsafe.Pointer(auto)),
		hwnd,
		uintptr(unsafe.Pointer(&element)),
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return element, nil
}

func uiaClickablePoint(element *uia.IUIAutomationElement) (point, bool) {
	var clickable point
	var gotClickable int32
	hr, _, _ := syscall.Syscall(
		element.VTable().GetClickablePoint,
		3,
		uintptr(unsafe.Pointer(element)),
		uintptr(unsafe.Pointer(&clickable)),
		uintptr(unsafe.Pointer(&gotClickable)),
	)
	if hr != 0 || gotClickable == 0 {
		return point{}, false
	}
	return clickable, true
}

func rectFromUIA(source uia.RECT) rect {
	return rect{
		Left:   int32(source.Left),
		Top:    int32(source.Top),
		Right:  int32(source.Right),
		Bottom: int32(source.Bottom),
	}
}
