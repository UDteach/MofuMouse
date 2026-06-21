package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"mofumouse/internal/core"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	configFlag := flag.String("config", "", "settings file path")
	flag.Parse()

	configPath := *configFlag
	if configPath == "" {
		var err error
		configPath, err = core.DefaultConfigPath()
		if err != nil {
			configPath = "settings.json"
		}
	}

	app := NewControlCenterApp(configPath)
	err := wails.Run(&options.App{
		Title:     "MofuMouse Control Center",
		Width:     1060,
		Height:    780,
		MinWidth:  900,
		MinHeight: 660,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 250, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "MofuMouse Control Center:", err)
		os.Exit(1)
	}
}
