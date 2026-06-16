package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FinalGradesTab manages quarter/semester/year marks.
type FinalGradesTab struct {
	controller Controller
	container  *fyne.Container
}

// NewFinalGradesTab creates a new final grades tab.
func NewFinalGradesTab(c Controller) *FinalGradesTab {
	f := &FinalGradesTab{controller: c}
	f.container = container.NewStack(
		widget.NewLabelWithStyle("Итоговые оценки — выберите класс, предмет и четверть", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
	)
	return f
}

// Container returns the tab's root container.
func (f *FinalGradesTab) Container() fyne.CanvasObject {
	return f.container
}
