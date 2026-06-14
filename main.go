package main

import (
	"os"
	"runtime"
)

func main() {
	// For Windows, force software rendering to avoid driver OpenGL crashes
	if runtime.GOOS == "windows" && os.Getenv("FYNE_RENDER") == "" {
		os.Setenv("FYNE_RENDER", "software")
	}

	controller := NewAppController()
	controller.Run()
}
