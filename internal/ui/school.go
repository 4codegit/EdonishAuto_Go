package ui

import (
        "fmt"

        "fyne.io/fyne/v2"
        "fyne.io/fyne/v2/canvas"
        "fyne.io/fyne/v2/container"
        "fyne.io/fyne/v2/layout"
        "fyne.io/fyne/v2/widget"

        "github.com/4codegit/edonish-auto/internal/api"
)

// SchoolPage holds the school selection screen UI components.
type SchoolPage struct {
        app         *App
        schools     []api.School
        schoolList  *widget.List
        statusLabel *widget.Label
}

// NewSchoolPage creates a new school selection page.
func NewSchoolPage(app *App) *SchoolPage {
        return &SchoolPage{app: app}
}

// SetSchools populates the school list and returns the root container.
func (p *SchoolPage) SetSchools(schools []api.School) fyne.CanvasObject {
        p.schools = schools

        p.statusLabel = widget.NewLabel(fmt.Sprintf("Найдено школ: %d", len(schools)))

        icon := canvas.NewImageFromResource(nil)
        icon.FillMode = canvas.ImageFillContain
        icon.SetMinSize(fyne.NewSize(48, 48))

        title := widget.NewLabelWithStyle("Выберите школу", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
        subtitle := widget.NewLabel("У вас есть доступ к нескольким школам")
        subtitle.Alignment = fyne.TextAlignCenter

        p.schoolList = widget.NewList(
                func() int {
                        return len(p.schools)
                },
                func() fyne.CanvasObject {
                        return container.NewVBox(
                                widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
                                widget.NewLabel(""),
                        )
                },
                func(id widget.ListItemID, item fyne.CanvasObject) {
                        school := p.schools[id]
                        box := item.(*container.Scroll).Content.(*fyne.Container)
                        nameLabel := box.Objects[0].(*widget.Label)
                        detailLabel := box.Objects[1].(*widget.Label)

                        roleDisplay := school.Role
                        if roleDisplay == "classroom-teacher" {
                                roleDisplay = "Классный руководитель"
                        } else if roleDisplay == "teacher" {
                                roleDisplay = "Учитель"
                        } else if roleDisplay == "school_admin" {
                                roleDisplay = "Администратор"
                        } else if roleDisplay == "director" {
                                roleDisplay = "Директор"
                        }

                        nameLabel.SetText(school.Name)
                        detailLabel.SetText(fmt.Sprintf("Роль: %s | ID: %d", roleDisplay, school.ID))
                },
        )

        p.schoolList.OnSelected = func(id widget.ListItemID) {
                school := p.schools[id]
                p.app.apiClient.SetSchool(school.ID)
                p.app.LogMessage(fmt.Sprintf("Выбрана школа: %s (ID: %d)", school.Name, school.ID), "info")

                // Save selected school to session
                p.app.SaveSessionSchool(school.ID)

                p.app.showDashboard(p.app.apiClient.UserInfo)
        }

        schoolCard := widget.NewCard("", "", p.schoolList)
        schoolCard.MinSize()

        content := container.NewVBox(
                layout.NewSpacer(),
                container.NewCenter(
                        container.NewVBox(
                                container.NewCenter(icon),
                                container.NewCenter(title),
                                container.NewCenter(subtitle),
                                p.statusLabel,
                                widget.NewSeparator(),
                        ),
                ),
                schoolCard,
                layout.NewSpacer(),
        )

        scroll := container.NewVScroll(content)
        scroll.SetMinSize(fyne.NewSize(500, 400))

        return scroll
}
