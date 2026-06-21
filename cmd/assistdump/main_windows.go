//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"mofumouse/internal/core"
	"mofumouse/internal/winapp"
)

func main() {
	configPath := flag.String("config", "", "settings file path")
	safeOnly := flag.Bool("safe", false, "print only safe foreground assist targets")
	delayMS := flag.Int("delay-ms", 500, "delay before scanning")
	flag.Parse()

	var rules []core.AssistRule
	if *configPath != "" {
		settings, err := core.LoadSettingsFile(*configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		rules = core.NewState(settings).TargetRules()
	}

	if *delayMS > 0 {
		time.Sleep(time.Duration(*delayMS) * time.Millisecond)
	}

	targets := winapp.ForegroundAssistTargetsWithRules(rules)
	if *safeOnly {
		targets = winapp.SafeForegroundAssistTargetsWithRules(rules)
	}
	for index, target := range targets {
		fmt.Printf(
			"%d\tlabel=%q\tpriority=%d\tdangerous=%v\tapp=%q\twindow=%q\tpoint=%d,%d\n",
			index,
			target.Label,
			target.Class.Priority,
			target.Class.Dangerous,
			target.App,
			target.Window,
			target.Point.X,
			target.Point.Y,
		)
	}
}
