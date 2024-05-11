package main

import (
	"wallpaper-changer-app/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	MyApp := app.New()
	w := MyApp.NewWindow("Wallpaper Changer")

	// Create UI
	ui.SetupUI(w)

	w.ShowAndRun()
}
