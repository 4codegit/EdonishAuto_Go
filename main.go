// eDonish Auto — Modern desktop application (Go + Fyne UI)
// Automated grade management for edonish.tj
package main

import (
	"os"

	"github.com/4codegit/edonish-auto/internal/ui"
)

func main() {
	// Force software rendering on Windows if no OpenGL driver available.
	// This prevents crashes on systems without proper GPU/OpenGL drivers.
	// The FYNE_RENDER=software environment variable tells Fyne to use
	// CPU-based rendering instead of OpenGL, which is slower but compatible.
	if os.Getenv("FYNE_RENDER") == "" {
		os.Setenv("FYNE_RENDER", "software")
	}

	app := ui.NewApp()
	app.Run()
}
