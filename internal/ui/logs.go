package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// LogsPage holds the logs viewer UI components.
type LogsPage struct {
	app      *App
	logEntry *widget.Entry
}

// NewLogsPage creates a new logs page.
func NewLogsPage(app *App) *LogsPage {
	return &LogsPage{app: app}
}

// Build creates the logs view and returns the root container.
func (p *LogsPage) Build() fyne.CanvasObject {
	p.logEntry = widget.NewMultiLineEntry()
	p.logEntry.SetPlaceHolder("Логи появятся здесь...")
	p.logEntry.Wrapping = fyne.TextWrapWord
	p.logEntry.SetMinRowsVisible(25)

	clearBtn := widget.NewButton("Очистить", func() {
		p.app.ClearLogs()
	})

	copyBtn := widget.NewButton("Копировать всё", func() {
		windows := p.app.fyneApp.Driver().AllWindows()
		if len(windows) > 0 {
			windows[0].Clipboard().SetContent(p.app.GetLogText())
			p.app.LogMessage("Логи скопированы в буфер обмена", "info")
		}
	})

	toolbar := container.NewHBox(
		widget.NewLabelWithStyle("Логи", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		copyBtn,
		clearBtn,
	)

	logCard := widget.NewCard("", "", p.logEntry)

	content := container.NewBorder(toolbar, nil, nil, nil, logCard)
	return content
}

// AppendLog appends a log entry to the display.
func (p *LogsPage) AppendLog(entry string) {
	if p.logEntry == nil {
		return
	}
	fyne.Do(func() {
		current := p.logEntry.Text
		if current != "" {
			current += "\n"
		}
		current += entry
		p.logEntry.SetText(current)
	})
}

// Clear clears the log display.
func (p *LogsPage) Clear() {
	if p.logEntry != nil {
		p.logEntry.SetText("")
	}
}
