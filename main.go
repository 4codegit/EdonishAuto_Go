// eDonish Auto — Modern desktop application (Go + Fyne UI)
// Automated grade management for edonish.tj
package main

import (
	"os"
	"runtime"

	"fyne.io/fyne/v2/app"
)

// main - точка входа в приложение
func main() {
	// On Windows, force software rendering to avoid OpenGL driver crashes.
	if runtime.GOOS == "windows" && os.Getenv("FYNE_RENDER") == "" {
		os.Setenv("FYNE_RENDER", "software")
	}

	// Создаём приложение Fyne
	a := app.New()

	// Создаём контроллер приложения
	ctrl := NewAppController(a)

	// Запускаем приложение
	ctrl.Run()
}
