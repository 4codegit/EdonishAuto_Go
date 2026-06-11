package ui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/4codegit/edonish-auto/internal/engine"
)

// ─── Data models ──────────────────────────────────────────────

// journalData holds the structured data for the journal table view.
type journalData struct {
	groupName   string
	subjectName string
	quarterName string
	dates       []dateCol    // column headers (dates)
	students    []studentRow // rows (one per student)
}

// dateCol represents a single date column in the journal.
type dateCol struct {
	dateID   string
	dateStr  string // full date, e.g. "2025-01-15"
	shortStr string // short date, e.g. "01-15"
}

// studentRow represents a single student row with marks.
type studentRow struct {
	studentID   int
	name        string
	marks       map[string]string // dateID -> display text
	markValues  map[string]int    // dateID -> numeric mark value
	avg         float64
	min         int
	max         int
	gradeCount  int
	missing     int
}

// ─── JournalPage ──────────────────────────────────────────────

// JournalPage holds the journal viewer UI components.
type JournalPage struct {
	app *App

	// Filters
	classSelect   *widget.Select
	subjectSelect *widget.Select
	quarterSelect *widget.Select

	// Status
	statusLabel *widget.Label

	// Table view
	journalTable *widget.Table
	journalData  []journalData

	// Student detail popup
	studentDetail *widget.Entry
	detailCard    *widget.Card
	selectedStudent string

	// Layout references
	tableScroll   *container.Scroll
	detailPanel   *fyne.Container
	splitLayout   *fyne.Container
}

// NewJournalPage creates a new journal page.
func NewJournalPage(app *App) *JournalPage {
	return &JournalPage{app: app}
}

// Build creates the journal view and returns the root container.
func (p *JournalPage) Build() fyne.CanvasObject {
	// ── Filter bar ────────────────────────────────────────
	p.classSelect = widget.NewSelect([]string{}, func(s string) {
		p.onClassChange(s)
	})
	p.classSelect.PlaceHolder = "Класс"

	p.subjectSelect = widget.NewSelect([]string{}, func(s string) {
		p.onSubjectChange(s)
	})
	p.subjectSelect.PlaceHolder = "Предмет"

	p.quarterSelect = widget.NewSelect([]string{}, func(s string) {
		p.onQuarterChange(s)
	})
	p.quarterSelect.PlaceHolder = "Четверть"

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		p.loadJournal()
	})

	// ── Status label ──────────────────────────────────────
	p.statusLabel = widget.NewLabelWithStyle("Выберите класс и предмет для загрузки журнала",
		fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// ── Journal Table ─────────────────────────────────────
	p.journalTable = widget.NewTable(
		func() (int, int) { return p.tableRowCount(), p.tableColCount() },
		func() fyne.CanvasObject {
			return widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			p.tableCellUpdate(id, cell.(*widget.Label))
		},
	)
	p.journalTable.SetColumnWidth(0, 35)  // #
	p.journalTable.SetColumnWidth(1, 180) // ФИО
	p.journalTable.OnSelected = func(id widget.TableCellID) {
		p.onCellSelected(id)
	}

	p.tableScroll = container.NewScroll(p.journalTable)
	p.tableScroll.SetMinSize(fyne.NewSize(800, 400))

	// ── Student detail panel (right side, hidden by default) ──
	p.studentDetail = widget.NewMultiLineEntry()
	p.studentDetail.Wrapping = fyne.TextWrapWord
	p.studentDetail.TextStyle = fyne.TextStyle{Monospace: true}
	p.studentDetail.SetMinRowsVisible(15)

	closeDetailBtn := widget.NewButton("Закрыть", func() {
		p.hideDetail()
	})

	p.detailCard = widget.NewCard("Анализ ученика", "", container.NewVBox(
		p.studentDetail,
		closeDetailBtn,
	))

	p.detailPanel = container.NewVBox()
	p.detailPanel.Hide()

	// ── Main layout ───────────────────────────────────────
	filterRow := container.NewBorder(nil, nil, nil, refreshBtn,
		container.NewGridWithColumns(3,
			p.classSelect,
			p.subjectSelect,
			p.quarterSelect,
		),
	)

	header := widget.NewCard("", "", container.NewVBox(
		filterRow,
		p.statusLabel,
	))

	tableArea := container.NewHSplit(
		p.tableScroll,
		p.detailPanel,
	)
	tableArea.SetOffset(0.75)

	content := container.NewBorder(header, nil, nil, nil, tableArea)

	return content
}

// ─── Filter change handlers (auto-load) ──────────────────────

func (p *JournalPage) onClassChange(selected string) {
	p.updateSubjectsForClass(selected)
	// Auto-load if subject is also selected
	if p.subjectSelect.Selected != "" && p.subjectSelect.Selected != "Все предметы" {
		p.loadJournal()
	}
}

func (p *JournalPage) onSubjectChange(selected string) {
	if selected != "" && selected != "Все предметы" {
		p.loadJournal()
	}
}

func (p *JournalPage) onQuarterChange(selected string) {
	if p.subjectSelect.Selected != "" && p.subjectSelect.Selected != "Все предметы" {
		p.loadJournal()
	}
}

// ─── Dropdowns ────────────────────────────────────────────────

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

// updateSubjectsForClass filters subject dropdown for the selected class.
func (p *JournalPage) updateSubjectsForClass(selected string) {
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

// ─── Table helpers ────────────────────────────────────────────

func (p *JournalPage) tableRowCount() int {
	rows := 0
	for _, jd := range p.journalData {
		rows += 1 // header row (group/subject/quarter title)
		rows += 1 // date header row
		rows += len(jd.students)
		rows += 1 // spacer
	}
	if rows == 0 {
		rows = 1
	}
	return rows
}

func (p *JournalPage) tableColCount() int {
	maxCols := 4 // #, ФИО, Ср, (Мин/Макс combined)
	for _, jd := range p.journalData {
		cols := 2 + len(jd.dates) + 2 // #, ФИО, dates..., Ср, Диап
		if cols > maxCols {
			maxCols = cols
		}
	}
	return maxCols
}

func (p *JournalPage) tableCellUpdate(id widget.TableCellID, label *widget.Label) {
	if len(p.journalData) == 0 {
		if id.Row == 0 && id.Col == 1 {
			label.SetText("Выберите класс и предмет")
		} else {
			label.SetText("")
		}
		return
	}

	rowIdx := 0
	for _, jd := range p.journalData {
		// Title row
		if id.Row == rowIdx {
			p.titleCell(id.Col, jd, label)
			return
		}
		rowIdx++

		// Date header row
		if id.Row == rowIdx {
			p.dateHeaderCell(id.Col, jd, label)
			return
		}
		rowIdx++

		// Student rows
		if id.Row < rowIdx+len(jd.students) {
			si := id.Row - rowIdx
			p.studentCell(id.Col, jd, jd.students[si], label)
			return
		}
		rowIdx += len(jd.students)

		// Spacer row
		if id.Row == rowIdx {
			label.SetText("")
			return
		}
		rowIdx++
	}

	label.SetText("")
}

func (p *JournalPage) titleCell(col int, jd journalData, label *widget.Label) {
	label.TextStyle = fyne.TextStyle{Bold: true}
	if col == 1 {
		label.SetText(fmt.Sprintf("%s — %s (%s)", jd.groupName, jd.subjectName, jd.quarterName))
	} else {
		label.SetText("")
	}
}

func (p *JournalPage) dateHeaderCell(col int, jd journalData, label *widget.Label) {
	label.TextStyle = fyne.TextStyle{Bold: true}
	totalCols := 2 + len(jd.dates) + 2
	if col >= totalCols {
		label.SetText("")
		return
	}
	switch {
	case col == 0:
		label.SetText("№")
	case col == 1:
		label.SetText("ФИО ученика")
	case col >= 2 && col < 2+len(jd.dates):
		label.SetText(jd.dates[col-2].shortStr)
	case col == 2+len(jd.dates):
		label.SetText("Ср")
	case col == 2+len(jd.dates)+1:
		label.SetText("Диап")
	}
}

func (p *JournalPage) studentCell(col int, jd journalData, sr studentRow, label *widget.Label) {
	totalCols := 2 + len(jd.dates) + 2
	if col >= totalCols {
		label.SetText("")
		return
	}

	label.TextStyle = fyne.TextStyle{}

	switch {
	case col == 0:
		for i, s := range jd.students {
			if s.studentID == sr.studentID {
				label.SetText(fmt.Sprintf("%d", i+1))
				return
			}
		}
		label.SetText("")
	case col == 1:
		label.SetText(sr.name)
		label.Alignment = fyne.TextAlignLeading
	case col >= 2 && col < 2+len(jd.dates):
		dateID := jd.dates[col-2].dateID
		if display, ok := sr.marks[dateID]; ok {
			label.SetText(display)
			if val, okv := sr.markValues[dateID]; okv {
				if val >= 9 {
					label.TextStyle = fyne.TextStyle{Bold: true}
				} else if val <= 3 {
					label.TextStyle = fyne.TextStyle{Italic: true}
				}
			}
		} else {
			label.SetText("")
		}
	case col == 2+len(jd.dates):
		if sr.gradeCount > 0 {
			label.SetText(fmt.Sprintf("%.1f", sr.avg))
		} else {
			label.SetText("—")
		}
	case col == 2+len(jd.dates)+1:
		if sr.gradeCount > 0 {
			label.SetText(fmt.Sprintf("%d-%d", sr.min, sr.max))
		} else {
			label.SetText("—")
		}
	}
}

// ─── Cell click → show student detail ────────────────────────

func (p *JournalPage) onCellSelected(id widget.TableCellID) {
	if len(p.journalData) == 0 {
		return
	}

	// Find which student was clicked
	rowIdx := 0
	for _, jd := range p.journalData {
		rowIdx += 2 // skip title + date header

		if id.Row >= rowIdx && id.Row < rowIdx+len(jd.students) {
			si := id.Row - rowIdx
			sr := jd.students[si]
			p.showStudentDetail(sr, jd)
			return
		}
		rowIdx += len(jd.students) + 1 // students + spacer
	}
}

func (p *JournalPage) showStudentDetail(sr studentRow, jd journalData) {
	p.selectedStudent = sr.name
	p.detailCard.SetTitle(sr.name)

	var lines []string
	lines = append(lines, fmt.Sprintf("Ученик: %s", sr.name))
	lines = append(lines, fmt.Sprintf("Класс: %s  |  Предмет: %s  |  %s", jd.groupName, jd.subjectName, jd.quarterName))
	lines = append(lines, strings.Repeat("─", 40))

	if sr.gradeCount > 0 {
		lines = append(lines, fmt.Sprintf("Средний балл: %.1f", sr.avg))
		lines = append(lines, fmt.Sprintf("Минимум: %d  |  Максимум: %d", sr.min, sr.max))
		lines = append(lines, fmt.Sprintf("Разброс: %d  |  Оценок: %d  |  Пропусков: %d", sr.max-sr.min, sr.gradeCount, sr.missing))
		lines = append(lines, "")
		lines = append(lines, "Визуальный разброс:")
		lines = append(lines, makeVisualSpread(sr.min, sr.max, sr.avg, 10))
		lines = append(lines, "")

		// Distribution
		grades := make([]int, 0, sr.gradeCount)
		for _, v := range sr.markValues {
			if v > 0 {
				grades = append(grades, v)
			}
		}
		if len(grades) > 0 {
			lines = append(lines, "Распределение оценок:")
			lines = append(lines, makeDistribution(grades))
		}
	} else {
		lines = append(lines, "Нет оценок")
	}

	p.studentDetail.SetText(strings.Join(lines, "\n"))
	p.detailPanel.Objects = []fyne.CanvasObject{p.detailCard}
	p.detailPanel.Show()
}

func (p *JournalPage) hideDetail() {
	p.detailPanel.Hide()
	p.selectedStudent = ""
}

// ─── Load journal data ────────────────────────────────────────

func (p *JournalPage) loadJournal() {
	classSelected := p.classSelect.Selected
	subjectSelected := p.subjectSelect.Selected
	quarterSelected := p.quarterSelect.Selected

	if subjectSelected == "" || subjectSelected == "Все предметы" {
		p.statusLabel.SetText("Выберите предмет")
		return
	}

	p.statusLabel.SetText("Загрузка журнала...")
	p.app.LogMessage(fmt.Sprintf("Загрузка журнала: %s / %s / %s", classSelected, subjectSelected, quarterSelected), "info")
	p.hideDetail()

	go func() {
		groups := p.getSelectedGroups(classSelected)
		if len(groups) == 0 {
			fyne.Do(func() {
				p.statusLabel.SetText("Не выбран класс")
			})
			return
		}

		var allData []journalData

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

					datesData, err := p.app.apiClient.GetJournalDates(groupID, subjectID, qpropID)
					if err != nil {
						continue
					}
					days := engine.ExtractDays(datesData)
					if len(days) == 0 {
						continue
					}

					studentsData, err := p.app.apiClient.GetJournalStudents(groupID, subjectID, qpropID)
					if err != nil {
						continue
					}
					students := engine.ExtractStudents(studentsData)
					if len(students) == 0 {
						continue
					}

					var dateCols []dateCol
					for _, day := range days {
						dateID := mapStr(day, "assignmentDateId")
						dateStr := mapStr(day, "assignmentDate")
						shortStr := dateStr
						if len(dateStr) >= 10 {
							shortStr = dateStr[5:10]
						}
						dateCols = append(dateCols, dateCol{
							dateID:   dateID,
							dateStr:  dateStr,
							shortStr: shortStr,
						})
					}

					var studentRows []studentRow
					for _, student := range students {
						studentID := mapInt(student, "studentId")
						studentName := fmt.Sprintf("%s %s", mapStr(student, "lastName"), mapStr(student, "firstName"))

						existingMarks := engine.ExtractExistingMarks(student)
						markDetails := extractMarkDetails(student)

						sr := studentRow{
							studentID:  studentID,
							name:       studentName,
							marks:      make(map[string]string),
							markValues: make(map[string]int),
						}

						var grades []int
						for _, dc := range dateCols {
							if _, has := existingMarks[dc.dateID]; has {
								if mi, ok := markDetails[dc.dateID]; ok {
									display := engine.ParseGradeDisplay(mi.shortName, mi.markValue)
									sr.marks[dc.dateID] = display
									sr.markValues[dc.dateID] = mi.markValue
									if mi.markValue > 0 {
										grades = append(grades, mi.markValue)
									}
								} else {
									sr.marks[dc.dateID] = "+"
								}
							} else {
								sr.missing++
							}
						}

						sr.gradeCount = len(grades)
						if len(grades) > 0 {
							sr.min = grades[0]
							sr.max = grades[0]
							sum := 0
							for _, g := range grades {
								sum += g
								if g < sr.min {
									sr.min = g
								}
								if g > sr.max {
									sr.max = g
								}
							}
							sr.avg = float64(sum) / float64(len(grades))
						}

						studentRows = append(studentRows, sr)
					}

					allData = append(allData, journalData{
						groupName:   groupName,
						subjectName: subjectName,
						quarterName: quarterName,
						dates:       dateCols,
						students:    studentRows,
					})
				}
			}
		}

		fyne.Do(func() {
			p.journalData = allData
			p.journalTable.Refresh()

			// Set date column widths
			for _, jd := range allData {
				for i := range jd.dates {
					p.journalTable.SetColumnWidth(2+i, 45)
				}
				cols := len(jd.dates)
				p.journalTable.SetColumnWidth(2+cols, 50)   // Ср
				p.journalTable.SetColumnWidth(2+cols+1, 55) // Диап
			}

			if len(allData) == 0 {
				p.statusLabel.SetText("Нет данных для отображения")
			} else {
				totalStudents := 0
				for _, jd := range allData {
					totalStudents += len(jd.students)
				}
				p.statusLabel.SetText(fmt.Sprintf("Загружено: %d класс/предмет/четверть, %d учеников",
					len(allData), totalStudents))
			}
		})
	}()
}

// ─── Helper functions ─────────────────────────────────────────

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

// makeVisualSpread creates a visual text-based spread indicator.
func makeVisualSpread(min, max int, avg float64, scale int) string {
	if max < min || scale <= 0 {
		return "[---]"
	}
	barWidth := scale * 2
	bar := make([]rune, barWidth)
	for i := range bar {
		bar[i] = '.'
	}
	for i := 0; i < barWidth; i++ {
		pos := float64(i) / float64(barWidth-1) * float64(max)
		if pos >= float64(min) && pos <= float64(max) {
			bar[i] = '='
		}
	}
	avgPos := int((avg - 1) / float64(max) * float64(barWidth-1))
	if avgPos < 0 {
		avgPos = 0
	}
	if avgPos >= barWidth {
		avgPos = barWidth - 1
	}
	bar[avgPos] = 'o'
	minPos := int(float64(min-1) / float64(max) * float64(barWidth-1))
	maxPos := int(float64(max-1) / float64(max) * float64(barWidth-1))
	if minPos >= 0 && minPos < barWidth {
		bar[minPos] = '['
	}
	if maxPos >= 0 && maxPos < barWidth {
		bar[maxPos] = ']'
	}
	return fmt.Sprintf("%d %s %d", min, string(bar), max)
}

// makeDistribution creates a text-based histogram of grade distribution.
func makeDistribution(grades []int) string {
	if len(grades) == 0 {
		return ""
	}
	counts := make(map[int]int)
	for _, g := range grades {
		counts[g]++
	}
	maxCount := 0
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}
	var gradeValues []int
	for g := range counts {
		gradeValues = append(gradeValues, g)
	}
	sort.Ints(gradeValues)

	var parts []string
	for _, g := range gradeValues {
		c := counts[g]
		barLen := 0
		if maxCount > 0 {
			barLen = int(float64(c) / float64(maxCount) * 8)
		}
		bar := strings.Repeat("#", barLen)
		parts = append(parts, fmt.Sprintf("%d:%s%d", g, bar, c))
	}
	return strings.Join(parts, " ")
}

// ─── Data selection helpers ───────────────────────────────────

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
