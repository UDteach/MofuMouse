//go:build windows

package winapp

import (
	"context"
	"time"

	"mofumouse/internal/core"
)

var idleActionIDs = []string{"nibble", "groom", "patpat", "sniff", "sleepy", "dig", "roll"}

func RunIdleActions(ctx context.Context, state *core.State) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	nextIndex := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !state.IdleActionsEnabled() {
				continue
			}
			if state.CurrentIdleAction(time.Now()) != "" {
				continue
			}
			if IdleDuration() < time.Duration(state.IdleActionSec)*time.Second {
				continue
			}
			action := idleActionIDs[nextIndex%len(idleActionIDs)]
			nextIndex++
			state.SetIdleAction(action, 7*time.Second)
		}
	}
}
