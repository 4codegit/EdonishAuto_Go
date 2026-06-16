package ui

import (
        "fmt"

        "fyne.io/fyne/v2"
        "fyne.io/fyne/v2/container"
        "fyne.io/fyne/v2/widget"
)

// TopicsTab manages topics and homework.
type TopicsTab struct {
        controller Controller
        container  *fyne.Container
}

// NewTopicsTab creates a new topics tab.
func NewTopicsTab(c Controller) *TopicsTab {
        t := &TopicsTab{controller: c}
        t.container = container.NewStack(
                widget.NewLabelWithStyle("Темы и ДЗ — выберите класс, предмет и четверть", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
        )
        return t
}

// Container returns the tab's root container.
func (t *TopicsTab) Container() fyne.CanvasObject {
        return t.container
}

// LoadData loads topics data for the given filters.
func (t *TopicsTab) LoadData(groupID, subjectID, quarterID string) {
        go func() {
                topics, err := t.controller.GetClient().GetTopics(groupID, subjectID, quarterID)
                fyne.Do(func() {
                        if err != nil {
                                t.container.Objects = []fyne.CanvasObject{
                                        widget.NewLabel(fmt.Sprintf("Ошибка: %v", err)),
                                }
                                t.container.Refresh()
                                return
                        }
                        if len(topics) == 0 {
                                t.container.Objects = []fyne.CanvasObject{
                                        widget.NewLabel("Нет данных о темах"),
                                }
                        } else {
                                t.container.Objects = []fyne.CanvasObject{
                                        widget.NewLabel(fmt.Sprintf("Загружено %d тем", len(topics))),
                                }
                        }
                        t.container.Refresh()
                })
        }()
}
