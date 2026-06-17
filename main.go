package main

import (
	"embed"

	"github.com/vinewz/PageVoice/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app.Run(assets)
}
