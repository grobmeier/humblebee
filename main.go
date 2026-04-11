package main

import (
	"os"

	"github.com/grobmeier/humblebee/internal/guiapp"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	app := guiapp.New()

	err := wails.Run(&options.App{
		Title:  "HumbleBee",
		Width:  1100,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: os.DirFS("frontend/dist"),
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.Startup,
		Bind: []any{
			app,
		},
	})
	if err != nil {
		panic(err)
	}
}
