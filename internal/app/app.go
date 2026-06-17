package app

import (
	"embed"
	"log"

	"github.com/vinewz/PageVoice/internal/services/textupload"
	"github.com/vinewz/PageVoice/internal/tts/piper"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func Run(assets embed.FS) {
	piperDir, err := piper.EnsureExtracted()
	if err != nil {
		log.Fatalf("piper setup: %v", err)
	}
	_ = piperDir

	textSvc := textupload.New()

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

	app.RegisterService(application.NewService(textSvc))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Main",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	})

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
