package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DiariesTab manages diary comments and signing.
type DiariesTab struct {
	controller Controller
	container  *fyne.Container
}

// NewDiariesTab creates a new diaries tab.
func NewDiariesTab(c Controller) *DiariesTab {
	d := &DiariesTab{controller: c}
	d.container = container.NewStack(
		widget.NewLabelWithStyle("Дневник — выберите класс и четверть", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
	)
	return d
}

// Container returns the tab's root container.
func (d *DiariesTab) Container() fyne.CanvasObject {
	return d.container
}
