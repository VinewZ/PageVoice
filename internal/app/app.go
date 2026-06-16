package app

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func Run(assets embed.FS) {
	app := application.New(application.Options{
		Name:        "PageVoice",
		Description: "Give voice to your texts.",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Main",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
