package main

import (
	"embed"
	"fmt"

	"github.com/vinewz/PageVoice/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	fmt.Println("Lesgo")
	app.Run(assets)
}
