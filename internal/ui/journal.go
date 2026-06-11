package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/4codegit/edonish-auto/internal/engine"
)

// JournalPage holds the journal viewer UI components.
type JournalPage struct {
	app *App

	classSelect   *widget.Select
	subjectSelect *widget.Select
	quarterSelect *widget.Select
	journalEntry  *widget.Entry
}

// NewJournalPage creates a new journal page.
func NewJournalPage(app *App) *JournalPage {
	return &JournalPage{app: app}
}

// Build creates the journal view and returns the root container.
func (p *JournalPage) Build() fyne.CanvasObject {
	p.classSelect = widget.NewSelect([]string{"Выберите..."}, func(s string) {
		p.onClassChange(s)
	})
	p.classSelect.PlaceHolder = "Класс"

	p.subjectSelect = widget.NewSelect([]string{"Выберите..."}, func(s string) {})
	p.subjectSelect.PlaceHolder = "Предмет"

	p.quarterSelect = widget.NewSelect([]string{"Выберите..."}, func(s string) {})
	p.quarterSelect.PlaceHolder = "Четверть"

	loadBtn := widget.NewButton("📥 Загрузить", func() {
		p.loadJournal()
	})
	loadBtn.Importance = widget.HighImportance

	p.journalEntry = widget.NewMultiLineEntry()
	p.journalEntry.SetPlaceHolder("Выберите класс, предмет и четверть для просмотра журнала")
	p.journalEntry.Wrapping = fyne.TextWrapWord
	p.journalEntry.SetMinRowsVisible(20)

	toolbar := widget.NewCard("📖 Просмотр журнала", "", container.NewVBox(
		container.NewHBox(
			p.classSelect,
			p.subjectSelect,
			p.quarterSelect,
			loadBtn,
		),
	))

	journalCard := widget.NewCard("", "", p.journalEntry)

	content := container.NewVBox(
		toolbar,
		journalCard,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(900, 600))

	return scroll
}

// UpdateDropdowns populates dropdowns with loaded data.
func (p *JournalPage) UpdateDropdowns() {
	classOpts := []string{"Все классы"}
	for _, g := range p.app.groupsData {
		name, _ := g["name"].(string)
		classOpts = append(classOpts, name)
	}
	p.classSelect.Options = classOpts
	p.classSelect.SetSelectedIndex(0)
	p.classSelect.Refresh()

	subjectOpts := []string{"Все предметы"}
	for _, s := range p.app.teacherSubjects {
		name, _ := s["subjectName"].(string)
		subjectOpts = append(subjectOpts, name)
	}
	p.subjectSelect.Options = subjectOpts
	p.subjectSelect.SetSelectedIndex(0)
	p.subjectSelect.Refresh()

	quarterOpts := []string{"Все четверти"}
	for _, q := range p.app.quartersData {
		name, _ := q["name"].(string)
		quarterOpts = append(quarterOpts, name)
	}
	p.quarterSelect.Options = quarterOpts
	p.quarterSelect.SetSelectedIndex(0)
	p.quarterSelect.Refresh()
}

// onClassChange updates subject dropdown when class selection changes.
func (p *JournalPage) onClassChange(selected string) {
	if p.app.journalOptions == nil {
		return
	}

	var subjects []string
	if optionsMap, ok := p.app.journalOptions.(map[string]interface{}); ok {
		if groups, ok := optionsMap["groups"].([]interface{}); ok {
			for _, g := range groups {
				if gm, ok := g.(map[string]interface{}); ok {
					gname := fmt.Sprintf("%s%s", mapStr(gm, "number"), mapStr(gm, "name"))
					if gname == selected || selected == "Все классы" {
						if subs, ok := gm["subjects"].([]interface{}); ok {
							for _, s := range subs {
								if sm, ok := s.(map[string]interface{}); ok {
									name := mapStr(sm, "subjectName")
									if name != "" {
										subjects = append(subjects, name)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := []string{"Все предметы"}
	for _, s := range subjects {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}

	p.subjectSelect.Options = unique
	p.subjectSelect.SetSelectedIndex(0)
	p.subjectSelect.Refresh()
}

// loadJournal loads and displays the journal for selected filters.
func (p *JournalPage) loadJournal() {
	classSelected := p.classSelect.Selected
	subjectSelected := p.subjectSelect.Selected
	quarterSelected := p.quarterSelect.Selected

	p.journalEntry.SetText("Загрузка журнала...")
	p.app.LogMessage(fmt.Sprintf("Загрузка журнала: %s / %s / %s", classSelected, subjectSelected, quarterSelected), "info")

	go func() {
		groups := p.getSelectedGroups(classSelected)
		if len(groups) == 0 {
			p.journalEntry.SetText("Не выбран класс")
			return
		}

		var allLines []string

		for _, group := range groups {
			groupID := mapInt(group, "id")
			groupName := mapStr(group, "name")

			subjects := p.getSubjectsForGroup(group, subjectSelected)
			quarters := p.getSelectedQuarters(quarterSelected)

			for _, subject := range subjects {
				subjectID := mapInt(subject, "subjectId")
				subjectName := mapStr(subject, "subjectName")

				for _, quarter := range quarters {
					qpropID := mapInt(quarter, "qpropId")
					quarterName := mapStr(quarter, "name")

					// Get dates
					datesData, err := p.app.apiClient.GetJournalDates(groupID, subjectID, qpropID)
					if err != nil {
						allLines = append(allLines, fmt.Sprintf("❌ Ошибка дат: %v", err))
						continue
					}

					days := engine.ExtractDays(datesData)
					if len(days) == 0 {
						allLines = append(allLines, fmt.Sprintf("⏭️ Нет дат: %s | %s | %s", groupName, subjectName, quarterName))
						continue
					}

					// Get students
					studentsData, err := p.app.apiClient.GetJournalStudents(groupID, subjectID, qpropID)
					if err != nil {
						allLines = append(allLines, fmt.Sprintf("❌ Ошибка студентов: %v", err))
						continue
					}

					students := engine.ExtractStudents(studentsData)

					// Build journal display
					allLines = append(allLines, "")
					allLines = append(allLines, fmt.Sprintf("═══ %s | %s | %s ═══", groupName, subjectName, quarterName))
					allLines = append(allLines, "")

					// Header row with dates
					header := "Ученик\t\t"
					for _, day := range days {
						dateStr := mapStr(day, "assignmentDate")
						if len(dateStr) >= 10 {
							dateStr = dateStr[5:10] // MM-DD
						}
						header += dateStr + "\t"
					}
					allLines = append(allLines, header)
					allLines = append(allLines, strings.Repeat("─", len(header)+20))

					// Student rows
					for _, student := range students {
						studentName := fmt.Sprintf("%s %s", mapStr(student, "lastName"), mapStr(student, "firstName"))
						row := studentName + "\t"

						existingMarks := engine.ExtractExistingMarks(student)
						markDetails := extractMarkDetails(student)

						for _, day := range days {
							dateID := mapStr(day, "assignmentDateId")
							if _, has := existingMarks[dateID]; has {
								if markInfo, ok := markDetails[dateID]; ok {
									display := engine.ParseGradeDisplay(markInfo.shortName, markInfo.markValue)
									row += display + "\t"
								} else {
									row += "✓\t"
								}
							} else {
								row += "—\t"
							}
						}

						allLines = append(allLines, row)
					}
				}
			}
		}

		if len(allLines) == 0 {
			allLines = append(allLines, "Нет данных для отображения")
		}

		p.journalEntry.SetText(strings.Join(allLines, "\n"))
		p.journalEntry.Refresh()
	}()
}

// markInfo holds grade display information.
type markInfo struct {
	shortName string
	markValue int
}

func extractMarkDetails(student map[string]interface{}) map[string]markInfo {
	result := make(map[string]markInfo)
	if subjectMarks, ok := student["subjectMarks"].([]interface{}); ok {
		for _, m := range subjectMarks {
			if mm, ok := m.(map[string]interface{}); ok {
				dateID := mapStr(mm, "assignmentDateId")
				shortName := mapStr(mm, "shortName")
				markValue := mapInt(mm, "mark")
				result[dateID] = markInfo{shortName: shortName, markValue: markValue}
			}
		}
	}
	return result
}

func (p *JournalPage) getSelectedGroups(selected string) []map[string]interface{} {
	if selected == "Все классы" || selected == "" {
		return p.app.groupsData
	}
	var result []map[string]interface{}
	for _, g := range p.app.groupsData {
		if name, _ := g["name"].(string); name == selected {
			result = append(result, g)
		}
	}
	return result
}

func (p *JournalPage) getSubjectsForGroup(group map[string]interface{}, selected string) []map[string]interface{} {
	var result []map[string]interface{}
	if subjects, ok := group["subjects"].([]interface{}); ok {
		for _, s := range subjects {
			if sm, ok := s.(map[string]interface{}); ok {
				if selected == "Все предметы" || selected == "" {
					result = append(result, sm)
				} else if mapStr(sm, "subjectName") == selected {
					result = append(result, sm)
				}
			}
		}
	}
	return result
}

func (p *JournalPage) getSelectedQuarters(selected string) []map[string]interface{} {
	if selected == "Все четверти" || selected == "" {
		return p.app.quartersData
	}
	var result []map[string]interface{}
	for _, q := range p.app.quartersData {
		if name, _ := q["name"].(string); name == selected {
			result = append(result, q)
		}
	}
	return result
}

// mapInt extracts an int field from a map[string]interface{}.
func mapInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}
