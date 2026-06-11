package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/4codegit/edonish-auto/internal/api"
)

// SchoolPage holds the school selection screen UI components.
type SchoolPage struct {
	app     *App
	schools []api.School
}

// NewSchoolPage creates a new school selection page.
func NewSchoolPage(app *App) *SchoolPage {
	return &SchoolPage{app: app}
}

// SetSchools builds the school selection view and returns the root container.
// Uses Button widgets instead of widget.List to avoid type assertion panics.
func (p *SchoolPage) SetSchools(schools []api.School) fyne.CanvasObject {
	p.schools = schools

	icon := canvas.NewImageFromResource(theme.ComputerIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(64, 64))

	title := widget.NewLabelWithStyle("Выберите школу", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel(fmt.Sprintf("Найдено школ: %d. Нажмите для выбора.", len(schools)))
	subtitle.Alignment = fyne.TextAlignCenter

	// Build a card with a button for each school
	var schoolButtons []fyne.CanvasObject
	for i := range schools {
		school := schools[i] // capture for closure

		roleDisplay := school.Role
		switch roleDisplay {
		case "classroom-teacher":
			roleDisplay = "Классный руководитель"
		case "teacher":
			roleDisplay = "Учитель"
		case "school_admin":
			roleDisplay = "Администратор школы"
		case "director":
			roleDisplay = "Директор"
		case "headteacher":
			roleDisplay = "Завуч"
		}

		schoolNameLabel := widget.NewLabelWithStyle(school.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		schoolDetailLabel := widget.NewLabel(fmt.Sprintf("Роль: %s  |  ID: %d", roleDisplay, school.ID))
		schoolDetailLabel.TextStyle = fyne.TextStyle{Italic: true}

		selectBtn := widget.NewButton("Выбрать", func() {
			p.app.apiClient.SetSchool(school.ID)
			p.app.LogMessage(fmt.Sprintf("Выбрана школа: %s (ID: %d)", school.Name, school.ID), "info")
			p.app.SaveSessionSchool(school.ID)
			p.app.showDashboard(p.app.apiClient.UserInfo)
		})
		selectBtn.Importance = widget.HighImportance

		card := widget.NewCard("", "", container.NewVBox(
			schoolNameLabel,
			schoolDetailLabel,
			selectBtn,
		))

		schoolButtons = append(schoolButtons, card)
	}

	schoolList := container.NewVBox(schoolButtons...)

	scroll := container.NewVScroll(schoolList)
	scroll.SetMinSize(fyne.NewSize(500, 300))

	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(icon),
		container.NewCenter(title),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
		scroll,
		layout.NewSpacer(),
	)

	return content
}
