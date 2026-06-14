package ui

import (
        "fmt"

        "fyne.io/fyne/v2"
        "fyne.io/fyne/v2/canvas"
        "fyne.io/fyne/v2/container"
        "fyne.io/fyne/v2/dialog"
        "fyne.io/fyne/v2/theme"
        "fyne.io/fyne/v2/widget"

        "github.com/4codegit/edonish-auto/client"
)

// DiariesTab manages the Diaries (Дневник) tab with signatures and diligence marks.
type DiariesTab struct {
        controller Controller
        container  *fyne.Container

        // Filters
        classSel *widget.Select

        // State
        journalOpts   *client.JournalOptions
        selectedGroup *client.JournalGroup
        diaries       []client.DiaryEntry

        // UI
        diariesList *widget.List
        statusLabel *widget.Label
}

// NewDiariesTab creates a new DiariesTab.
func NewDiariesTab(c Controller) *DiariesTab {
        dt := &DiariesTab{
                controller:  c,
                statusLabel: widget.NewLabel("Выберите класс для загрузки дневников"),
        }
        dt.buildUI()
        go dt.loadJournalOptions()
        return dt
}

// Container returns the root container for this tab.
func (dt *DiariesTab) Container() fyne.CanvasObject {
        return dt.container
}

// buildUI creates the full UI layout for the diaries tab.
func (dt *DiariesTab) buildUI() {
        // Filter row
        dt.classSel = widget.NewSelect([]string{}, dt.onClassSelected)
        dt.classSel.PlaceHolder = "Выберите класс..."

        filterRow := container.NewHBox(
                widget.NewLabel("Класс:"),
                dt.classSel,
        )

        // Batch actions
        batchSignBtn := widget.NewButton("Подписать все (одна комбинация)", dt.onBatchSignAll)
        batchSignBtn.Importance = widget.HighImportance

        refreshBtn := widget.NewButtonWithIcon("Обновить", theme.ViewRefreshIcon(), func() {
                if dt.selectedGroup != nil {
                        go dt.loadDiaries()
                }
        })

        actionRow := container.NewHBox(
                batchSignBtn,
                refreshBtn,
        )

        // Placeholder for diaries list
        placeholder := widget.NewLabelWithStyle(
                "Выберите класс для загрузки дневников",
                fyne.TextAlignCenter,
                fyne.TextStyle{Italic: true},
        )

        dt.container = container.NewBorder(
                container.NewVBox(filterRow, actionRow, widget.NewSeparator()),
                dt.statusLabel,
                nil,
                nil,
                placeholder,
        )
}

// loadJournalOptions loads class list from API.
func (dt *DiariesTab) loadJournalOptions() {
        dt.statusLabel.SetText("Загрузка списка классов...")
        opts, err := dt.controller.GetClient().GetJournalOptions()
        if err != nil {
                fyne.Do(func() {
                        dt.statusLabel.SetText(fmt.Sprintf("Ошибка загрузки настроек журнала: %v", err))
                })
                return
        }

        dt.journalOpts = opts

        classNames := make([]string, len(opts.Groups))
        for i, g := range opts.Groups {
                classNames[i] = fmt.Sprintf("%d %s", g.Number, g.Name)
        }

        fyne.Do(func() {
                dt.classSel.Options = classNames
                dt.classSel.Refresh()
                dt.statusLabel.SetText("Выберите класс")
                if len(classNames) > 0 {
                        dt.classSel.SetSelectedIndex(0)
                }
        })
}

// onClassSelected is called when a class is selected from the dropdown.
func (dt *DiariesTab) onClassSelected(selected string) {
        if dt.journalOpts == nil {
                return
        }

        var group *client.JournalGroup
        for i, g := range dt.journalOpts.Groups {
                gName := fmt.Sprintf("%d %s", g.Number, g.Name)
                if gName == selected {
                        group = &dt.journalOpts.Groups[i]
                        break
                }
        }

        if group == nil {
                return
        }

        dt.selectedGroup = group
        go dt.loadDiaries()
}

// loadDiaries calls GetDiaries API and rebuilds the list.
func (dt *DiariesTab) loadDiaries() {
        if dt.selectedGroup == nil {
                return
        }

        fyne.Do(func() {
                dt.statusLabel.SetText("Загрузка дневников...")
        })

        diaries, err := dt.controller.GetClient().GetDiaries(dt.selectedGroup.ID)

        fyne.Do(func() {
                if err != nil {
                        dt.statusLabel.SetText(fmt.Sprintf("Ошибка загрузки дневников: %v", err))
                        return
                }

                dt.diaries = diaries
                dt.rebuildDiariesList()

                dt.statusLabel.SetText(fmt.Sprintf("Загружено дневников: %d", len(diaries)))
        })
}

// rebuildDiariesList builds the list showing each diary entry with status.
func (dt *DiariesTab) rebuildDiariesList() {
        if len(dt.diaries) == 0 {
                dt.container.Objects = []fyne.CanvasObject{
                        container.NewBorder(
                                container.NewVBox(
                                        container.NewHBox(widget.NewLabel("Класс:"), dt.classSel),
                                        container.NewHBox(
                                                widget.NewButton("Подписать все (одна комбинация)", dt.onBatchSignAll),
                                                widget.NewButtonWithIcon("Обновить", theme.ViewRefreshIcon(), func() {
                                                        if dt.selectedGroup != nil {
                                                                go dt.loadDiaries()
                                                        }
                                                }),
                                        ),
                                        widget.NewSeparator(),
                                ),
                                dt.statusLabel,
                                nil,
                                nil,
                                widget.NewLabelWithStyle("Нет дневников для этого класса", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
                        ),
                }
                dt.container.Refresh()
                return
        }

        dt.diariesList = widget.NewList(
                func() int {
                        return len(dt.diaries)
                },
                func() fyne.CanvasObject {
                        // Group & Subject
                        groupText := widget.NewLabel("")
                        groupText.TextStyle = fyne.TextStyle{Bold: true}

                        subjectText := widget.NewLabel("")

                        quarterText := widget.NewLabel("")

                        // Diligence mark
                        diligenceText := widget.NewLabel("")
                        diligenceText.TextStyle = fyne.TextStyle{Bold: true}

                        // Signature statuses
                        parentText := widget.NewLabel("")

                        managerText := widget.NewLabel("")

                        // Layout: left side (group, subject, quarter), center (diligence), right (signatures)
                        leftBox := container.NewVBox(groupText, subjectText, quarterText)
                        centerBox := container.NewVBox(diligenceText)
                        rightBox := container.NewVBox(parentText, managerText)

                        row := container.NewBorder(nil, nil, leftBox, rightBox, centerBox)
                        return container.NewPadded(row)
                },
                func(id widget.ListItemID, cell fyne.CanvasObject) {
                        if id < 0 || id >= len(dt.diaries) {
                                return
                        }
                        entry := dt.diaries[id]

                        pad := cell.(*fyne.Container)
                        border := pad.Objects[0].(*fyne.Container)

                        leftBox := border.Objects[0].(*fyne.Container)
                        centerBox := border.Objects[1].(*fyne.Container)
                        rightBox := border.Objects[2].(*fyne.Container)

                        groupText := leftBox.Objects[0].(*widget.Label)
                        subjectText := leftBox.Objects[1].(*widget.Label)
                        quarterText := leftBox.Objects[2].(*widget.Label)

                        diligenceText := centerBox.Objects[0].(*widget.Label)

                        parentText := rightBox.Objects[0].(*widget.Label)
                        managerText := rightBox.Objects[1].(*widget.Label)

                        // Group & Subject
                        groupText.SetText(entry.GroupName)
                        subjectText.SetText(entry.SubjectName)
                        quarterText.SetText(entry.QuarterName)

                        // Diligence mark
                        if entry.DiligenceMark != "" {
                                diligenceText.SetText(fmt.Sprintf("Прилежание: %s", entry.DiligenceMark))
                        } else {
                                diligenceText.SetText("Прилежание: не указано")
                        }

                        // Parent signature status
                        if entry.ParentSigned {
                                parentText.SetText("✓ Подписано")
                        } else {
                                parentText.SetText("✗ Не подписано")
                        }

                        // Manager signature status
                        if entry.ManagerSigned {
                                managerText.SetText("✓ Подписано")
                        } else {
                                managerText.SetText("✗ Не подписано")
                        }
                },
        )

        dt.diariesList.OnSelected = func(id widget.ListItemID) {
                dt.diariesList.Unselect(id)
                dt.showDiaryDialog(id)
        }

        dt.container.Objects = []fyne.CanvasObject{
                container.NewBorder(
                        container.NewVBox(
                                container.NewHBox(widget.NewLabel("Класс:"), dt.classSel),
                                container.NewHBox(
                                        widget.NewButton("Подписать все (одна комбинация)", dt.onBatchSignAll),
                                        widget.NewButtonWithIcon("Обновить", theme.ViewRefreshIcon(), func() {
                                                if dt.selectedGroup != nil {
                                                        go dt.loadDiaries()
                                                }
                                        }),
                                ),
                                widget.NewSeparator(),
                        ),
                        dt.statusLabel,
                        nil,
                        nil,
                        dt.diariesList,
                ),
        }
        dt.container.Refresh()
}

// showDiaryDialog shows a dialog for a single diary entry with signature and diligence actions.
func (dt *DiariesTab) showDiaryDialog(idx int) {
        if idx < 0 || idx >= len(dt.diaries) {
                return
        }

        entry := dt.diaries[idx]

        // Info labels
        groupLabel := widget.NewLabel(fmt.Sprintf("Класс: %s", entry.GroupName))
        subjectLabel := widget.NewLabel(fmt.Sprintf("Предмет: %s", entry.SubjectName))
        quarterLabel := widget.NewLabel(fmt.Sprintf("Четверть: %s", entry.QuarterName))

        // Current diligence display
        var diligenceDisplay string
        if entry.DiligenceMark != "" {
                diligenceDisplay = entry.DiligenceMark
        } else {
                diligenceDisplay = "не указано"
        }
        currentDiligenceLabel := widget.NewLabel(fmt.Sprintf("Текущее прилежание: %s", diligenceDisplay))

        // Diligence selector
        diligenceSel := widget.NewSelect(DiligenceMarks, nil)
        diligenceSel.PlaceHolder = "Выберите прилежание..."
        if entry.DiligenceMark != "" {
                diligenceSel.SetSelected(entry.DiligenceMark)
        }

        // Signature status display
        parentStatus, parentColor := FormatSignedStatus(entry.ParentSigned)
        parentStatusText := canvas.NewText(fmt.Sprintf("Родители: %s", parentStatus), parentColor)
        parentStatusText.TextSize = 12

        managerStatus, managerColor := FormatSignedStatus(entry.ManagerSigned)
        managerStatusText := canvas.NewText(fmt.Sprintf("Руководитель: %s", managerStatus), managerColor)
        managerStatusText.TextSize = 12

        var dlg dialog.Dialog

        // Parent signature button
        parentBtn := widget.NewButton("Подпись родителей", func() {
                go dt.signDiary(entry.DiaryID, "parent", idx)
                dlg.Hide()
        })
        if entry.ParentSigned {
                parentBtn.Disable()
                parentBtn.SetText("✓ Уже подписано (родители)")
        }

        // Manager signature button
        managerBtn := widget.NewButton("Подпись руководителя", func() {
                go dt.signDiary(entry.DiaryID, "manager", idx)
                dlg.Hide()
        })
        if entry.ManagerSigned {
                managerBtn.Disable()
                managerBtn.SetText("✓ Уже подписано (руководитель)")
        }

        // Set diligence button
        diligenceBtn := widget.NewButton("Установить прилежание", func() {
                if diligenceSel.Selected == "" {
                        dialog.ShowInformation("Внимание", "Выберите оценку прилежания", dt.controller.GetWindow())
                        return
                }
                go dt.setDiligence(entry.DiaryID, diligenceSel.Selected, idx)
                dlg.Hide()
        })

        content := container.NewVBox(
                groupLabel,
                subjectLabel,
                quarterLabel,
                widget.NewSeparator(),
                currentDiligenceLabel,
                container.NewHBox(widget.NewLabel("Новое прилежание:"), diligenceSel),
                diligenceBtn,
                widget.NewSeparator(),
                parentStatusText,
                parentBtn,
                managerStatusText,
                managerBtn,
        )

        dialogTitle := fmt.Sprintf("Дневник: %s — %s", entry.GroupName, entry.SubjectName)
        dlg = dialog.NewCustom(dialogTitle, "Закрыть", content, dt.controller.GetWindow())
        dlg.Show()
}

// signDiary signs a diary entry with the specified sign type.
func (dt *DiariesTab) signDiary(diaryID int, signType string, idx int) {
        fyne.Do(func() {
                dt.statusLabel.SetText(fmt.Sprintf("Подписание дневника (%s)...", signType))
        })

        err := dt.controller.GetClient().SignDiary(diaryID, signType)

        fyne.Do(func() {
                if err != nil {
                        dialog.ShowError(fmt.Errorf("Ошибка подписания дневника: %v", err), dt.controller.GetWindow())
                        dt.statusLabel.SetText("Ошибка подписания дневника")
                } else {
                        signLabel := "родителем"
                        if signType == "manager" {
                                signLabel = "руководителем"
                        }
                        dt.statusLabel.SetText(fmt.Sprintf("Дневник подписан %s", signLabel))
                        go dt.loadDiaries()
                }
        })
}

// setDiligence sets the diligence mark for a diary entry.
func (dt *DiariesTab) setDiligence(diaryID int, diligenceMark string, idx int) {
        fyne.Do(func() {
                dt.statusLabel.SetText(fmt.Sprintf("Установка прилежания: %s...", diligenceMark))
        })

        err := dt.controller.GetClient().SetDiaryDiligence(diaryID, diligenceMark)

        fyne.Do(func() {
                if err != nil {
                        dialog.ShowError(fmt.Errorf("Ошибка установки прилежания: %v", err), dt.controller.GetWindow())
                        dt.statusLabel.SetText("Ошибка установки прилежания")
                } else {
                        dt.statusLabel.SetText(fmt.Sprintf("Прилежание установлено: %s", diligenceMark))
                        go dt.loadDiaries()
                }
        })
}

// onBatchSignAll signs all unsigned diaries with a random diligence combination.
// It picks one diligence mark for ALL diaries, sets it, then signs parent AND manager.
func (dt *DiariesTab) onBatchSignAll() {
        if len(dt.diaries) == 0 {
                dialog.ShowInformation("Внимание", "Нет дневников для подписания", dt.controller.GetWindow())
                return
        }

        // Filter unsigned diaries
        var unsigned []client.DiaryEntry
        for _, d := range dt.diaries {
                if !d.ParentSigned || !d.ManagerSigned || d.DiligenceMark == "" {
                        unsigned = append(unsigned, d)
                }
        }

        if len(unsigned) == 0 {
                dialog.ShowInformation("Готово", "Все дневники уже подписаны", dt.controller.GetWindow())
                return
        }

        // Pick one random diligence combination
        chosenDiligence := RandomDiligenceCombo()

        confirmMsg := fmt.Sprintf(
                "Будет установлено прилежание «%s» для %d дневников(я)\nи подписаны все неподписанные записи.\n\nПродолжить?",
                chosenDiligence, len(unsigned),
        )

        dialog.ShowConfirm("Подписать все", confirmMsg, func(ok bool) {
                if !ok {
                        return
                }
                go dt.executeBatchSign(unsigned, chosenDiligence)
        }, dt.controller.GetWindow())
}

// executeBatchSign performs the batch signing operation.
func (dt *DiariesTab) executeBatchSign(unsigned []client.DiaryEntry, diligence string) {
        total := len(unsigned)
        apiClient := dt.controller.GetClient()

        for i, entry := range unsigned {
                progress := fmt.Sprintf("Обработка %d из %d: %s — %s", i+1, total, entry.GroupName, entry.SubjectName)
                fyne.Do(func() {
                        dt.statusLabel.SetText(progress)
                })

                // Set diligence if not set
                if entry.DiligenceMark == "" {
                        if err := apiClient.SetDiaryDiligence(entry.DiaryID, diligence); err != nil {
                                fyne.Do(func() {
                                        dt.statusLabel.SetText(fmt.Sprintf("Ошибка установки прилежания (дневник %d): %v", entry.DiaryID, err))
                                })
                                continue
                        }
                }

                // Sign parent if not signed
                if !entry.ParentSigned {
                        if err := apiClient.SignDiary(entry.DiaryID, "parent"); err != nil {
                                fyne.Do(func() {
                                        dt.statusLabel.SetText(fmt.Sprintf("Ошибка подписания родителем (дневник %d): %v", entry.DiaryID, err))
                                })
                                continue
                        }
                }

                // Sign manager if not signed
                if !entry.ManagerSigned {
                        if err := apiClient.SignDiary(entry.DiaryID, "manager"); err != nil {
                                fyne.Do(func() {
                                        dt.statusLabel.SetText(fmt.Sprintf("Ошибка подписания руководителем (дневник %d): %v", entry.DiaryID, err))
                                })
                                continue
                        }
                }
        }

        fyne.Do(func() {
                dt.statusLabel.SetText(fmt.Sprintf("Готово! Обработано %d дневников (прилежание: %s)", total, diligence))
                go dt.loadDiaries()
        })
}

// Refresh updates the tab with new data from the dashboard context.
// It receives students, group, subject, and quarter from the dashboard
// and triggers a reload of diaries if the group has changed.
func (dt *DiariesTab) Refresh(students []client.Student, group *client.JournalGroup, subject *client.Subject, quarter *client.Quarter) {
        // Update group if provided and different from current
        if group != nil && (dt.selectedGroup == nil || dt.selectedGroup.ID != group.ID) {
                dt.selectedGroup = group
                go dt.loadDiaries()
        }
}
