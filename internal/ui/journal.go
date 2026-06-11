package ui

import (
        "fmt"
        "math"
        "sort"
        "strconv"
        "strings"
        "time"

        "fyne.io/fyne/v2"
        "fyne.io/fyne/v2/container"
        "fyne.io/fyne/v2/theme"
        "fyne.io/fyne/v2/widget"

        "github.com/4codegit/edonish-auto/internal/config"
        "github.com/4codegit/edonish-auto/internal/engine"
)

// ─── Data models ──────────────────────────────────────────────

// dateInfo holds date column data.
type dateInfo struct {
        dateID   string
        dateStr  string // full date, e.g. "2025-01-15"
        shortStr string // short date, e.g. "01-15"
        weekday  string // short weekday name
        topic    string
        homeWork string
}

// studentInfo holds student row data.
type studentInfo struct {
        studentID  int
        name       string
        marks      map[string]string // dateID -> display text
        markValues map[string]int    // dateID -> numeric mark value
        markIDs    map[string]string // dateID -> assignmentMarkId

        // Quarter/Semester/Year marks
        quarterMarkVal    string
        quarterMarkID     string
        semesterMarkVal   string
        semesterPropID    int
        yearMarkVal       string
        yearPropID        int

        // Computed stats
        avg        float64
        min        int
        max        int
        gradeCount int
        missing    int
}

// gradeCellData stores data for an editable grade cell.
type gradeCellData struct {
        row        int
        col        int
        studentID  int
        dateID     string
        qpropID    int
        markID     string
        value      string
        origValue  string
        isPastDate bool
        isNA       bool
}

// journalParams stores the current journal API query parameters.
type journalParams struct {
        groupID              int
        subjectID            int
        qpropID              int
        curriculumPropertyID int
        quarterName          string
}

// ─── JournalPage ──────────────────────────────────────────────

// JournalPage holds the interactive journal viewer UI.
type JournalPage struct {
        app *App

        // Filters
        classSelect   *widget.Select
        subjectSelect *widget.Select
        quarterSelect *widget.Select

        // Action buttons
        loadBtn  *widget.Button
        saveBtn  *widget.Button
        clearBtn *widget.Button

        // Status
        statusLabel *widget.Label
        countLabel  *widget.Label

        // Grid data
        dates    []dateInfo
        students []studentInfo
        params   *journalParams

        // Grade cells for editing
        gradeCells map[string]*widget.Entry // "row,col" -> Entry
        gradeData  map[string]*gradeCellData

        // Grid dimensions
        gridRows int
        gridCols int

        // State
        journalLoaded bool

        // Container for the grid
        gridContainer *fyne.Container
        gridScroll    *container.Scroll

        // Topics section
        topicsLabel *widget.Label
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

        p.loadBtn = widget.NewButtonWithIcon("Загрузить", theme.DownloadIcon(), func() {
                p.onLoadJournal()
        })
        p.loadBtn.Importance = widget.HighImportance

        p.saveBtn = widget.NewButtonWithIcon("Сохранить", theme.DocumentSaveIcon(), func() {
                // Reload to save current state (auto-saves on cell edit)
                p.onLoadJournal()
        })
        p.saveBtn.Disable()

        p.clearBtn = widget.NewButtonWithIcon("Очистить все", theme.DeleteIcon(), func() {
                p.onClearAllGrades()
        })
        p.clearBtn.Disable()

        p.statusLabel = widget.NewLabelWithStyle("Выберите класс, предмет и четверть, затем нажмите «Загрузить»",
                fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

        p.countLabel = widget.NewLabel("")

        p.topicsLabel = widget.NewLabel("")

        // Placeholder grid
        p.gradeCells = make(map[string]*widget.Entry)
        p.gradeData = make(map[string]*gradeCellData)

        placeholder := widget.NewLabelWithStyle("Журнал не загружен\n\nВыберите конкретный класс, предмет и четверть\nи нажмите «Загрузить»",
                fyne.TextAlignCenter, fyne.TextStyle{})

        p.gridContainer = container.NewVBox(placeholder)
        p.gridScroll = container.NewVScroll(p.gridContainer)
        p.gridScroll.SetMinSize(fyne.NewSize(900, 450))

        // ── Toolbar ───────────────────────────────────────────
        toolbar := widget.NewCard("Просмотр журнала", "", container.NewVBox(
                container.NewGridWithColumns(5,
                        p.classSelect,
                        p.subjectSelect,
                        p.quarterSelect,
                        p.loadBtn,
                        container.NewHBox(p.saveBtn, p.clearBtn),
                ),
                p.countLabel,
        ))

        content := container.NewVBox(
                toolbar,
                p.statusLabel,
                p.gridScroll,
                p.topicsLabel,
        )

        scroll := container.NewVScroll(content)
        scroll.SetMinSize(fyne.NewSize(900, 600))

        return scroll
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

// ─── Load journal ─────────────────────────────────────────────

func (p *JournalPage) onLoadJournal() {
        className := p.classSelect.Selected
        subjectName := p.subjectSelect.Selected
        quarterName := p.quarterSelect.Selected

        // Require specific class, subject, quarter (like Python version)
        if className == "" || className == "Все классы" {
                p.statusLabel.SetText("Выберите конкретный класс!")
                return
        }
        if subjectName == "" || subjectName == "Все предметы" {
                p.statusLabel.SetText("Выберите предмет!")
                return
        }

        // Look up IDs
        var groupID, subjectID, qpropID, curriculumPropertyID int
        for _, g := range p.app.groupsData {
                if name, _ := g["name"].(string); name == className {
                        groupID = mapInt(g, "id")
                        break
                }
        }
        for _, s := range p.app.teacherSubjects {
                if name, _ := s["subjectName"].(string); name == subjectName {
                        subjectID = mapInt(s, "subjectId")
                        curriculumPropertyID = mapInt(s, "curriculumPropertyId")
                        break
                }
        }
        // Look up quarter ID from journal_options for this group
        if optionsMap, ok := p.app.journalOptions.(map[string]interface{}); ok {
                if groups, ok := optionsMap["groups"].([]interface{}); ok {
                        for _, g := range groups {
                                if gm, ok := g.(map[string]interface{}); ok {
                                        gname := fmt.Sprintf("%s%s", mapStr(gm, "number"), mapStr(gm, "name"))
                                        if gname == className {
                                                if quarters, ok := gm["quarters"].([]interface{}); ok {
                                                        for _, q := range quarters {
                                                                if qm, ok := q.(map[string]interface{}); ok {
                                                                        if mapStr(qm, "name") == quarterName {
                                                                                qpropID = mapInt(qm, "id")
                                                                                break
                                                                        }
                                                                }
                                                        }
                                                }
                                                break
                                        }
                                }
                        }
                }
        }
        // Fallback to quarters_data
        if qpropID == 0 {
                for _, q := range p.app.quartersData {
                        if name, _ := q["name"].(string); name == quarterName {
                                qpropID = mapInt(q, "qpropId")
                                break
                        }
                }
        }

        if groupID == 0 || subjectID == 0 || qpropID == 0 {
                p.statusLabel.SetText("Не удалось определить параметры журнала!")
                return
        }

        p.params = &journalParams{
                groupID:              groupID,
                subjectID:            subjectID,
                qpropID:              qpropID,
                curriculumPropertyID: curriculumPropertyID,
                quarterName:          quarterName,
        }

        p.statusLabel.SetText("Загрузка журнала...")
        p.app.LogMessage(fmt.Sprintf("Загрузка журнала: %s / %s / %s", className, subjectName, quarterName), "info")

        go p.loadJournalData()
}

func (p *JournalPage) loadJournalData() {
        params := p.params
        if params == nil {
                return
        }

        // Get dates
        datesData, err := p.app.apiClient.GetJournalDates(params.groupID, params.subjectID, params.qpropID)
        if err != nil {
                fyne.Do(func() { p.statusLabel.SetText(fmt.Sprintf("Ошибка загрузки дат: %v", err)) })
                return
        }
        days := engine.ExtractDays(datesData)

        // Get students
        studentsData, err := p.app.apiClient.GetJournalStudents(params.groupID, params.subjectID, params.qpropID)
        if err != nil {
                fyne.Do(func() { p.statusLabel.SetText(fmt.Sprintf("Ошибка загрузки студентов: %v", err)) })
                return
        }
        students := engine.ExtractStudents(studentsData)

        if len(days) == 0 || len(students) == 0 {
                fyne.Do(func() { p.statusLabel.SetText("Нет данных для отображения") })
                return
        }

        // Build date columns
        var dates []dateInfo
        for _, day := range days {
                dateID := mapStr(day, "assignmentDateId")
                dateStr := mapStr(day, "assignmentDate")
                shortStr := dateStr
                if len(dateStr) >= 10 {
                        shortStr = dateStr[5:10] // MM-DD
                }
                dates = append(dates, dateInfo{
                        dateID:   dateID,
                        dateStr:  dateStr,
                        shortStr: shortStr,
                        weekday:  mapStr(day, "weekdayShortName"),
                        topic:    mapStr(day, "topic"),
                        homeWork: mapStr(day, "homeWork"),
                })
        }

        // Build student rows
        var studentRows []studentInfo
        totalMarks := 0
        emptyCells := 0

        for _, student := range students {
                studentID := mapInt(student, "studentId")
                studentName := fmt.Sprintf("%s %s", mapStr(student, "lastName"), mapStr(student, "firstName"))

                existingMarks := engine.ExtractExistingMarks(student)
                markDetails := extractMarkDetails(student)

                si := studentInfo{
                        studentID:  studentID,
                        name:       studentName,
                        marks:      make(map[string]string),
                        markValues: make(map[string]int),
                        markIDs:    make(map[string]string),
                }

                var grades []int
                for _, dc := range dates {
                        if _, has := existingMarks[dc.dateID]; has {
                                if mi, ok := markDetails[dc.dateID]; ok {
                                        display := parseGradeDisplay(mi.shortName, mi.markValue)
                                        si.marks[dc.dateID] = display
                                        si.markValues[dc.dateID] = mi.markValue
                                        si.markIDs[dc.dateID] = mi.markID
                                        if mi.markValue > 0 {
                                                grades = append(grades, mi.markValue)
                                        }
                                        totalMarks++
                                } else {
                                        si.marks[dc.dateID] = "+"
                                        totalMarks++
                                }
                        } else {
                                si.missing++
                                emptyCells++
                        }
                }

                // Compute stats
                si.gradeCount = len(grades)
                if len(grades) > 0 {
                        si.min = grades[0]
                        si.max = grades[0]
                        sum := 0
                        for _, g := range grades {
                                sum += g
                                if g < si.min {
                                        si.min = g
                                }
                                if g > si.max {
                                        si.max = g
                                }
                        }
                        si.avg = float64(sum) / float64(len(grades))
                }

                // Extract quarter/semester/year marks
                if qmList, ok := student["quarterMark"].([]interface{}); ok && len(qmList) > 0 {
                        if qm, ok := qmList[0].(map[string]interface{}); ok {
                                si.quarterMarkVal = parseGradeDisplay(mapStr(qm, "shortName"), mapInt(qm, "mark"))
                                si.quarterMarkID = mapStr(qm, "quarterMarkId")
                                if si.quarterMarkID == "" {
                                        si.quarterMarkID = mapStr(qm, "assignmentMarkId")
                                }
                        }
                }
                if smList, ok := student["semesterMark"].([]interface{}); ok && len(smList) > 0 {
                        if sm, ok := smList[0].(map[string]interface{}); ok {
                                si.semesterMarkVal = parseGradeDisplay(mapStr(sm, "shortName"), mapInt(sm, "mark"))
                                si.semesterPropID = mapInt(sm, "semesterPropertyId")
                        }
                }
                if ymList, ok := student["yearMark"].([]interface{}); ok && len(ymList) > 0 {
                        if ym, ok := ymList[0].(map[string]interface{}); ok {
                                si.yearMarkVal = parseGradeDisplay(mapStr(ym, "shortName"), mapInt(ym, "mark"))
                                si.yearPropID = mapInt(ym, "yearPropertyId")
                        }
                }

                studentRows = append(studentRows, si)
        }

        // Update UI on main thread
        fyne.Do(func() {
                p.dates = dates
                p.students = studentRows
                p.gridRows = len(studentRows)
                p.gridCols = len(dates)
                p.journalLoaded = true

                p.saveBtn.Enable()
                p.clearBtn.Enable()

                pct := 0.0
                if totalMarks+emptyCells > 0 {
                        pct = float64(totalMarks) / float64(totalMarks+emptyCells) * 100
                }
                p.countLabel.SetText(fmt.Sprintf("%d учеников | %d дат | Заполнено: %d | Пустых: %d | Заполненность: %.0f%%",
                        len(studentRows), len(dates), totalMarks, emptyCells, pct))

                p.buildGrid()

                p.statusLabel.SetText(fmt.Sprintf("Журнал загружен: %d оценок, %d пустых", totalMarks, emptyCells))
        })
}

// ─── Build grid ───────────────────────────────────────────────

func (p *JournalPage) buildGrid() {
        p.gradeCells = make(map[string]*widget.Entry)
        p.gradeData = make(map[string]*gradeCellData)

        if len(p.students) == 0 || len(p.dates) == 0 {
                p.gridContainer.Objects = []fyne.CanvasObject{
                        widget.NewLabelWithStyle("Нет данных", fyne.TextAlignCenter, fyne.TextStyle{}),
                }
                p.gridContainer.Refresh()
                return
        }

        var rows []fyne.CanvasObject

        // ── Header row ──
        headerCells := []fyne.CanvasObject{
                widget.NewLabelWithStyle("#", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
                widget.NewLabelWithStyle("Ученик", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
                widget.NewLabelWithStyle("🎲", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
        }
        for _, d := range p.dates {
                headerCells = append(headerCells,
                        widget.NewLabelWithStyle(d.shortStr, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
                )
        }
        // Чтв / Смст / Год columns
        headerCells = append(headerCells,
                widget.NewLabelWithStyle("Чтв", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
                widget.NewLabelWithStyle("Смст", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
                widget.NewLabelWithStyle("Год", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
        )
        rows = append(rows, container.NewHBox(headerCells...))

        // ── Student rows ──
        now := time.Now().Format("2006-01-02")

        for rowIdx, si := range p.students {
                rowCells := []fyne.CanvasObject{
                        widget.NewLabel(fmt.Sprintf("%d", rowIdx+1)),
                        widget.NewLabel(si.name),
                }

                // Random grade button for this student
                diceBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
                        p.onRandomGradeForStudent(rowIdx)
                })
                diceBtn.ToolTip = widget.NewToolTip("Рандомная оценка для строки")
                rowCells = append(rowCells, diceBtn)

                // Grade cells for each date
                for colIdx, d := range p.dates {
                        cellKey := fmt.Sprintf("%d,%d", rowIdx, colIdx)
                        value := si.marks[d.dateID]
                        markID := si.markIDs[d.dateID]
                        isPastDate := d.dateStr[:minInt(10, len(d.dateStr))] < now

                        data := &gradeCellData{
                                row:        rowIdx,
                                col:        colIdx,
                                studentID:  si.studentID,
                                dateID:     d.dateID,
                                qpropID:    p.params.qpropID,
                                markID:     markID,
                                value:      value,
                                origValue:  value,
                                isPastDate: isPastDate,
                                isNA:       value == config.ABSENT_SHORT,
                        }
                        p.gradeData[cellKey] = data

                        cell := widget.NewEntry()
                        cell.SetText(value)
                        cell.PlaceHolder = "—"
                        cell.Wrapping = fyne.TextWrapOff
                        cell.TextStyle = fyne.TextStyle{Monospace: true}

                        // Store reference
                        p.gradeCells[cellKey] = cell

                        // On submit: set grade via API
                        r, c := rowIdx, colIdx
                        cell.OnSubmitted = func(text string) {
                                p.onCellSubmit(r, c, text)
                        }

                        rowCells = append(rowCells, cell)
                }

                // Quarter mark cell (clickable label)
                quarterText := si.quarterMarkVal
                if quarterText == "" {
                        quarterText = "—"
                }
                quarterLabel := widget.NewButton(quarterText, func() {
                        p.onSetQuarterMark(rowIdx)
                })
                quarterLabel.Importance = widget.MediumImportance
                rowCells = append(rowCells, quarterLabel)

                // Semester mark cell
                semesterText := si.semesterMarkVal
                if semesterText == "" {
                        semesterText = "—"
                }
                semesterLabel := widget.NewButton(semesterText, func() {
                        p.onSetSemesterMark(rowIdx)
                })
                semesterLabel.Importance = widget.LowImportance
                rowCells = append(rowCells, semesterLabel)

                // Year mark cell
                yearText := si.yearMarkVal
                if yearText == "" {
                        yearText = "—"
                }
                yearLabel := widget.NewButton(yearText, func() {
                        p.onSetYearMark(rowIdx)
                })
                yearLabel.Importance = widget.LowImportance
                rowCells = append(rowCells, yearLabel)

                rows = append(rows, container.NewHBox(rowCells...))
        }

        // ── Stats row ──
        pct := 0.0
        totalMarks, emptyCells := 0, 0
        for _, si := range p.students {
                totalMarks += si.gradeCount
                emptyCells += si.missing
        }
        if totalMarks+emptyCells > 0 {
                pct = float64(totalMarks) / float64(totalMarks+emptyCells) * 100
        }
        statsText := fmt.Sprintf("Заполнено: %d | Пустых: %d | Заполненность: %.0f%%", totalMarks, emptyCells, pct)
        rows = append(rows, widget.NewSeparator())
        rows = append(rows, widget.NewLabelWithStyle(statsText, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}))

        // ── Topics section ──
        rows = append(rows, widget.NewSeparator())
        rows = append(rows, widget.NewLabelWithStyle("Темы уроков", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
        for _, d := range p.dates {
                topicStr := d.topic
                if topicStr == "" {
                        topicStr = "(пусто)"
                }
                hwStr := d.homeWork
                if hwStr == "" {
                        hwStr = "—"
                }
                topicLine := fmt.Sprintf("  %s %s | Тема: %s | ДЗ: %s", d.shortStr, d.weekday, topicStr, hwStr)
                rows = append(rows, widget.NewLabel(topicLine))
        }

        // ── Help text ──
        rows = append(rows, widget.NewSeparator())
        rows = append(rows, widget.NewLabelWithStyle(
                "Enter: поставить оценку | Чтв: ceil(ср.) | Смст: ceil((чтв1+чтв2)/2) | Год: ceil((чтв1+чтв2+чтв3+чтв4)/4) | 🎲: рандом",
                fyne.TextAlignLeading, fyne.TextStyle{Italic: true},
        ))

        p.gridContainer.Objects = rows
        p.gridContainer.Refresh()
}

// ─── Cell edit handlers ───────────────────────────────────────

func (p *JournalPage) onCellSubmit(row, col int, text string) {
        data := p.gradeData[fmt.Sprintf("%d,%d", row, col)]
        if data == nil {
                return
        }

        text = strings.TrimSpace(text)

        // Check for absent input variants
        naVariants := map[string]bool{"н/а": true, "на": true, "n/a": true, "na": true}
        if naVariants[strings.ToLower(text)] {
                p.setCellGrade(row, col, config.ABSENT_MARK)
                return
        }

        // Parse numeric grade
        grade, err := strconv.Atoi(text)
        if err != nil {
                p.app.LogMessage(fmt.Sprintf("Неверная оценка: %s", text), "error")
                // Restore original value
                if cell, ok := p.gradeCells[fmt.Sprintf("%d,%d", row, col)]; ok {
                        cell.SetText(data.value)
                }
                return
        }

        if grade < config.MinGrade || grade > config.MaxGradeAllow {
                p.app.LogMessage(fmt.Sprintf("Оценка должна быть от %d до %d или %s", config.MinGrade, config.MaxGradeAllow, config.ABSENT_DISPLAY), "error")
                if cell, ok := p.gradeCells[fmt.Sprintf("%d,%d", row, col)]; ok {
                        cell.SetText(data.value)
                }
                return
        }

        p.setCellGrade(row, col, grade)
}

func (p *JournalPage) setCellGrade(row, col, grade int) {
        data := p.gradeData[fmt.Sprintf("%d,%d", row, col)]
        if data == nil {
                return
        }

        cell := p.gradeCells[fmt.Sprintf("%d,%d", row, col)]
        if cell == nil {
                return
        }

        if data.isPastDate {
                p.app.LogMessage("Внимание: дата этой ячейки уже прошла. Сервер может заблокировать изменение.", "warning")
        }

        displayVal := config.ABSENT_SHORT
        if grade != config.ABSENT_MARK {
                displayVal = strconv.Itoa(grade)
        }
        p.app.LogMessage(fmt.Sprintf("Установка оценки %s в ячейке (строка %d)", displayVal, row+1), "info")

        go func() {
                // Delete existing mark first
                if data.markID != "" {
                        p.app.LogMessage(fmt.Sprintf("  Удаление старой оценки (ID: %s)", data.markID), "info")
                        _, err := p.app.apiClient.DeleteMark(data.markID)
                        if err != nil {
                                p.app.LogMessage(fmt.Sprintf("  Ошибка удаления (пропускаем): %v", err), "warning")
                        }
                }

                // Create new mark
                p.app.LogMessage(fmt.Sprintf("  Создание оценки: student=%d, date=%s, grade=%d", data.studentID, data.dateID, grade), "info")
                result, err := p.app.apiClient.CreateMark(
                        data.studentID,
                        data.dateID,
                        grade,
                        0, // markTypeID will be set correctly in CreateMark
                        data.qpropID,
                )

                if err != nil {
                        p.app.LogMessage(fmt.Sprintf("  Ошибка: %v", err), "error")
                        fyne.Do(func() {
                                cell.SetText(data.value)
                        })
                        return
                }

                if resultMap, ok := result.(map[string]interface{}); ok {
                        if errMsg, exists := resultMap["error"]; exists && errMsg != nil {
                                p.app.LogMessage(fmt.Sprintf("  Ошибка API: %v", errMsg), "error")
                                fyne.Do(func() {
                                        cell.SetText(data.value)
                                })
                                return
                        }
                        // Success
                        newMarkID := mapStr(resultMap, "assignmentMarkId")
                        data.markID = newMarkID
                        data.value = displayVal
                        data.origValue = displayVal
                        data.isNA = (grade == config.ABSENT_MARK)
                        p.app.LogMessage(fmt.Sprintf("  Успех! Mark ID: %s", newMarkID), "info")

                        fyne.Do(func() {
                                cell.SetText(displayVal)
                        })
                } else {
                        data.value = displayVal
                        data.origValue = displayVal
                        data.isNA = (grade == config.ABSENT_MARK)
                        p.app.LogMessage("  Успех!", "info")
                        fyne.Do(func() {
                                cell.SetText(displayVal)
                        })
                }
        }()
}

// ─── Random grade for student row ─────────────────────────────

func (p *JournalPage) onRandomGradeForStudent(row int) {
        if !p.journalLoaded || len(p.dates) == 0 {
                return
        }

        filled := 0
        for col := 0; col < len(p.dates); col++ {
                data := p.gradeData[fmt.Sprintf("%d,%d", row, col)]
                if data != nil && data.value == "" {
                        // Random grade in min-max range
                        grade := config.MinGrade + int(math.Round(float64(config.MaxGrade-config.MinGrade)*float64(time.Now().Nanosecond()%100)/100))
                        if grade < config.MinGrade {
                                grade = config.MinGrade
                        }
                        if grade > config.MaxGrade {
                                grade = config.MaxGrade
                        }
                        p.setCellGrade(row, col, grade)
                        filled++
                }
        }

        if filled == 0 {
                p.app.LogMessage("Все ячейки уже заполнены", "info")
        } else {
                p.app.LogMessage(fmt.Sprintf("Заполнено %d ячеек рандомом", filled), "info")
        }
}

// ─── Quarter mark ─────────────────────────────────────────────

func (p *JournalPage) onSetQuarterMark(row int) {
        if !p.journalLoaded || p.params == nil {
                return
        }
        if row >= len(p.students) {
                return
        }

        si := p.students[row]
        params := p.params

        p.app.LogMessage(fmt.Sprintf("Расчёт четвертной для %s (строка %d)...", si.name, row+1), "info")

        go func() {
                // Fetch fresh student data from API
                studentsData, err := p.app.apiClient.GetJournalStudents(params.groupID, params.subjectID, params.qpropID)
                if err != nil {
                        p.app.LogMessage(fmt.Sprintf("Ошибка API: %v", err), "error")
                        return
                }

                students := engine.ExtractStudents(studentsData)
                var student map[string]interface{}
                for _, s := range students {
                        if mapInt(s, "studentId") == si.studentID {
                                student = s
                                break
                        }
                }
                if student == nil {
                        p.app.LogMessage("Ученик не найден в ответе API", "error")
                        return
                }

                // Extract grades from fresh API response
                var gradeValues []int
                if subjectMarks, ok := student["subjectMarks"].([]interface{}); ok {
                        for _, m := range subjectMarks {
                                if mm, ok := m.(map[string]interface{}); ok {
                                        sn := mapStr(mm, "shortName")
                                        mv := mapInt(mm, "mark")
                                        display := parseGradeDisplay(sn, mv)
                                        if display != "" && display != config.ABSENT_SHORT {
                                                if v, err := strconv.Atoi(display); err == nil {
                                                        if config.MinGrade <= v && v <= config.MaxGradeAllow {
                                                                gradeValues = append(gradeValues, v)
                                                        }
                                                }
                                        }
                                }
                        }
                }

                if len(gradeValues) == 0 {
                        p.app.LogMessage("У ученика нет оценок для расчёта четвертной", "error")
                        return
                }

                // Calculate ceil(average)
                sum := 0
                for _, v := range gradeValues {
                        sum += v
                }
                avg := float64(sum) / float64(len(gradeValues))
                grade := int(math.Ceil(avg))
                if grade < config.MinGrade {
                        grade = config.MinGrade
                }
                if grade > config.MaxGradeAllow {
                        grade = config.MaxGradeAllow
                }

                p.app.LogMessage(fmt.Sprintf("Четвертная: оценки=%v, ср.=%.2f, ceil=%d", gradeValues, avg, grade), "info")

                // Save quarter mark
                result, err := p.app.apiClient.CreateQuarterMark(
                        si.studentID,
                        params.qpropID,
                        grade,
                        params.subjectID,
                        params.curriculumPropertyID,
                )
                if err != nil {
                        p.app.LogMessage(fmt.Sprintf("Ошибка четвертной: %v", err), "error")
                        return
                }
                if resultMap, ok := result.(map[string]interface{}); ok {
                        if errMsg, exists := resultMap["error"]; exists && errMsg != nil {
                                p.app.LogMessage(fmt.Sprintf("Ошибка API: %v", errMsg), "error")
                                return
                        }
                }
                p.app.LogMessage(fmt.Sprintf("Четвертная оценка %d поставлена (%s)", grade, si.name), "info")

                // Reload journal to show updated quarter mark
                p.loadJournalData()
        }()
}

// ─── Semester mark ────────────────────────────────────────────

func (p *JournalPage) onSetSemesterMark(row int) {
        if !p.journalLoaded || p.params == nil {
                return
        }
        if row >= len(p.students) {
                return
        }

        si := p.students[row]
        params := p.params

        if si.semesterPropID == 0 {
                p.app.LogMessage("Нет semester_property_id — невозможно поставить полугодие", "error")
                return
        }

        p.app.LogMessage(fmt.Sprintf("Расчёт полугодия для %s...", si.name), "info")

        go func() {
                // Fetch all 4 quarters to calculate semester mark
                quarterMarks := make(map[string]int)
                for _, q := range p.app.quartersData {
                        qname := mapStr(q, "name")
                        qpropID := mapInt(q, "qpropId")
                        studentsData, err := p.app.apiClient.GetJournalStudents(params.groupID, params.subjectID, qpropID)
                        if err != nil {
                                continue
                        }
                        students := engine.ExtractStudents(studentsData)
                        for _, s := range students {
                                if mapInt(s, "studentId") == si.studentID {
                                        if qmList, ok := s["quarterMark"].([]interface{}); ok && len(qmList) > 0 {
                                                if qm, ok := qmList[0].(map[string]interface{}); ok {
                                                        sn := mapStr(qm, "shortName")
                                                        mv := mapInt(qm, "mark")
                                                        display := parseGradeDisplay(sn, mv)
                                                        if display != "" && display != config.ABSENT_SHORT {
                                                                if v, err := strconv.Atoi(display); err == nil {
                                                                        if config.MinGrade <= v && v <= config.MaxGradeAllow {
                                                                                quarterMarks[qname] = v
                                                                        }
                                                                }
                                                        }
                                                }
                                        }
                                        break
                                }
                        }
                }

                // Determine which semester based on current quarter
                currentQuarter := params.quarterName
                qNum := extractQuarterNumber(currentQuarter)

                var semesterGrade *int
                if qNum == 1 || qNum == 2 {
                        q1 := findQuarterMark(quarterMarks, 1)
                        q2 := findQuarterMark(quarterMarks, 2)
                        if q1 != nil && q2 != nil {
                                avg := float64(*q1+*q2) / 2
                                g := int(math.Ceil(avg))
                                if g < config.MinGrade {
                                        g = config.MinGrade
                                }
                                if g > config.MaxGradeAllow {
                                        g = config.MaxGradeAllow
                                }
                                semesterGrade = &g
                                p.app.LogMessage(fmt.Sprintf("1-е полугодие: чтв1=%d, чтв2=%d, ср.=%.2f → %d", *q1, *q2, avg, g), "info")
                        } else {
                                p.app.LogMessage("Нельзя поставить 1-е полугодие: нужны обе четверти (1 и 2)", "error")
                                return
                        }
                } else if qNum == 3 || qNum == 4 {
                        q3 := findQuarterMark(quarterMarks, 3)
                        q4 := findQuarterMark(quarterMarks, 4)
                        if q3 != nil && q4 != nil {
                                avg := float64(*q3+*q4) / 2
                                g := int(math.Ceil(avg))
                                if g < config.MinGrade {
                                        g = config.MinGrade
                                }
                                if g > config.MaxGradeAllow {
                                        g = config.MaxGradeAllow
                                }
                                semesterGrade = &g
                                p.app.LogMessage(fmt.Sprintf("2-е полугодие: чтв3=%d, чтв4=%d, ср.=%.2f → %d", *q3, *q4, avg, g), "info")
                        } else {
                                p.app.LogMessage("Нельзя поставить 2-е полугодие: нужны обе четверти (3 и 4)", "error")
                                return
                        }
                } else {
                        p.app.LogMessage(fmt.Sprintf("Не удалось определить номер четверти из '%s'", currentQuarter), "error")
                        return
                }

                if semesterGrade == nil {
                        return
                }

                result, err := p.app.apiClient.CreateSemesterMark(si.studentID, si.semesterPropID, *semesterGrade)
                if err != nil {
                        p.app.LogMessage(fmt.Sprintf("Ошибка полугодия: %v", err), "error")
                        return
                }
                if resultMap, ok := result.(map[string]interface{}); ok {
                        if errMsg, exists := resultMap["error"]; exists && errMsg != nil {
                                p.app.LogMessage(fmt.Sprintf("Ошибка API: %v", errMsg), "error")
                                return
                        }
                }
                p.app.LogMessage(fmt.Sprintf("Полугодие %d поставлено (%s)", *semesterGrade, si.name), "info")
                p.loadJournalData()
        }()
}

// ─── Year mark ────────────────────────────────────────────────

func (p *JournalPage) onSetYearMark(row int) {
        if !p.journalLoaded || p.params == nil {
                return
        }
        if row >= len(p.students) {
                return
        }

        si := p.students[row]
        params := p.params

        if si.yearPropID == 0 {
                p.app.LogMessage("Нет year_property_id — невозможно поставить годовую", "error")
                return
        }

        p.app.LogMessage(fmt.Sprintf("Расчёт годовой для %s...", si.name), "info")

        go func() {
                quarterMarks := make(map[string]int)
                for _, q := range p.app.quartersData {
                        qname := mapStr(q, "name")
                        qpropID := mapInt(q, "qpropId")
                        studentsData, err := p.app.apiClient.GetJournalStudents(params.groupID, params.subjectID, qpropID)
                        if err != nil {
                                continue
                        }
                        students := engine.ExtractStudents(studentsData)
                        for _, s := range students {
                                if mapInt(s, "studentId") == si.studentID {
                                        if qmList, ok := s["quarterMark"].([]interface{}); ok && len(qmList) > 0 {
                                                if qm, ok := qmList[0].(map[string]interface{}); ok {
                                                        sn := mapStr(qm, "shortName")
                                                        mv := mapInt(qm, "mark")
                                                        display := parseGradeDisplay(sn, mv)
                                                        if display != "" && display != config.ABSENT_SHORT {
                                                                if v, err := strconv.Atoi(display); err == nil {
                                                                        if config.MinGrade <= v && v <= config.MaxGradeAllow {
                                                                                quarterMarks[qname] = v
                                                                        }
                                                                }
                                                        }
                                                }
                                        }
                                        break
                                }
                        }
                }

                // Check if quarter 4 exists
                q4 := findQuarterMark(quarterMarks, 4)
                if q4 == nil {
                        p.app.LogMessage("Нельзя поставить годовую: 4-я четверть ещё не поставлена", "error")
                        return
                }

                // Calculate year mark from all available quarters
                var qVals []int
                for num := 1; num <= 4; num++ {
                        if v := findQuarterMark(quarterMarks, num); v != nil {
                                qVals = append(qVals, *v)
                        }
                }

                if len(qVals) == 0 {
                        p.app.LogMessage("Нет четвертных оценок для расчёта годовой", "error")
                        return
                }

                avg := float64(0)
                for _, v := range qVals {
                        avg += float64(v)
                }
                avg /= float64(len(qVals))
                yearGrade := int(math.Ceil(avg))
                if yearGrade < config.MinGrade {
                        yearGrade = config.MinGrade
                }
                if yearGrade > config.MaxGradeAllow {
                        yearGrade = config.MaxGradeAllow
                }

                p.app.LogMessage(fmt.Sprintf("Годовая: четверти=%v, ср.=%.2f → %d", qVals, avg, yearGrade), "info")

                result, err := p.app.apiClient.CreateYearMark(si.studentID, si.yearPropID, yearGrade)
                if err != nil {
                        p.app.LogMessage(fmt.Sprintf("Ошибка годовой: %v", err), "error")
                        return
                }
                if resultMap, ok := result.(map[string]interface{}); ok {
                        if errMsg, exists := resultMap["error"]; exists && errMsg != nil {
                                p.app.LogMessage(fmt.Sprintf("Ошибка API: %v", errMsg), "error")
                                return
                        }
                }
                p.app.LogMessage(fmt.Sprintf("Годовая оценка %d поставлена (%s)", yearGrade, si.name), "info")
                p.loadJournalData()
        }()
}

// ─── Clear all grades ─────────────────────────────────────────

func (p *JournalPage) onClearAllGrades() {
        if !p.journalLoaded {
                return
        }

        p.app.LogMessage("Удаление всех оценок из журнала...", "info")

        go func() {
                deleted := 0
                errors := 0
                for _, si := range p.students {
                        for _, markID := range si.markIDs {
                                if markID != "" {
                                        _, err := p.app.apiClient.DeleteMark(markID)
                                        if err != nil {
                                                errors++
                                                p.app.LogMessage(fmt.Sprintf("Ошибка удаления: %v", err), "error")
                                        } else {
                                                deleted++
                                        }
                                        time.Sleep(200 * time.Millisecond)
                                }
                        }
                }
                p.app.LogMessage(fmt.Sprintf("Удалено %d оценок, ошибок: %d", deleted, errors), "info")
                // Reload journal
                p.loadJournalData()
        }()
}

// ─── Helper functions ─────────────────────────────────────────

// markDetail holds mark shortName, markValue and markID.
type markDetail struct {
        shortName string
        markValue int
        markID    string
}

// extractMarkDetails extracts mark details from a student data map.
func extractMarkDetails(student map[string]interface{}) map[string]markDetail {
        result := make(map[string]markDetail)
        if subjectMarks, ok := student["subjectMarks"].([]interface{}); ok {
                for _, m := range subjectMarks {
                        if mm, ok := m.(map[string]interface{}); ok {
                                dateID := mapStr(mm, "assignmentDateId")
                                if dateID != "" {
                                        result[dateID] = markDetail{
                                                shortName: mapStr(mm, "shortName"),
                                                markValue: mapInt(mm, "mark"),
                                                markID:    mapStr(mm, "assignmentMarkId"),
                                        }
                                }
                        }
                }
        }
        return result
}

// parseGradeDisplay converts API shortName to display text.
// Matches the Python version's _parse_grade_display logic.
func parseGradeDisplay(shortName string, markValue int) string {
        if shortName == "" || strings.TrimSpace(shortName) == "" {
                return ""
        }
        shortName = strings.TrimSpace(shortName)

        // Already absent
        lower := strings.ToLower(shortName)
        if lower == "н/а" || lower == "n/a" || lower == "на" || lower == "na" {
                return config.ABSENT_SHORT
        }

        // Fractional format: "X/Y"
        if strings.Contains(shortName, "/") {
                parts := strings.SplitN(shortName, "/", 2)
                numerator := strings.TrimSpace(parts[0])
                if num, err := strconv.Atoi(numerator); err == nil {
                        // Absent detection: mark_value=0 OR numerator < MIN_GRADE
                        if markValue == config.ABSENT_MARK {
                                return config.ABSENT_SHORT
                        }
                        if num == config.ABSENT_MARK || (num > 0 && num < config.MinGrade) {
                                return config.ABSENT_SHORT
                        }
                        return strconv.Itoa(num)
                }
                return shortName
        }

        return shortName
}

// extractQuarterNumber gets the quarter number from a quarter name.
func extractQuarterNumber(name string) int {
        for _, ch := range name {
                if ch >= '1' && ch <= '4' {
                        return int(ch - '0')
                }
        }
        // Try Roman numerals
        romanMap := map[string]int{"I": 1, "II": 2, "III": 3, "IV": 4}
        for roman, num := range romanMap {
                if strings.Contains(name, roman) {
                        return num
                }
        }
        return 0
}

// findQuarterMark looks up a quarter mark by quarter number.
func findQuarterMark(marks map[string]int, num int) *int {
        prefixes := []string{
                fmt.Sprintf("%d четверть", num),
                fmt.Sprintf("Четверть %d", num),
        }
        // Add Roman numeral prefix
        roman := []string{"I", "II", "III", "IV"}
        if num >= 1 && num <= 4 {
                prefixes = append(prefixes, roman[num-1]+" четверть")
        }
        for _, prefix := range prefixes {
                if v, ok := marks[prefix]; ok {
                        return &v
                }
        }
        return nil
}

// minInt returns the smaller of two integers.
func minInt(a, b int) int {
        if a < b {
                return a
        }
        return b
}

// mapInt extracts an int from map[string]interface{}.
func mapInt(m map[string]interface{}, key string) int {
        if v, ok := m[key].(float64); ok {
                return int(v)
        }
        if v, ok := m[key].(int); ok {
                return v
        }
        return 0
}

// makeDistribution creates a text histogram of grade distribution.
func makeDistribution(grades []int) string {
        if len(grades) == 0 {
                return ""
        }
        counts := make(map[int]int)
        for _, g := range grades {
                counts[g]++
        }
        var keys []int
        for k := range counts {
                keys = append(keys, k)
        }
        sort.Ints(keys)
        var parts []string
        for _, k := range keys {
                parts = append(parts, fmt.Sprintf("%d:%d", k, counts[k]))
        }
        return strings.Join(parts, " ")
}

// makeVisualSpread creates a visual spread indicator.
func makeVisualSpread(minVal, maxVal int, avg float64, width int) string {
        if maxVal <= minVal || width <= 0 {
                return "[" + strings.Repeat("=", width) + "]"
        }
        bar := make([]rune, width)
        for i := range bar {
                bar[i] = ' '
        }
        minPos := int(float64(minVal-1) / float64(maxVal) * float64(width-1))
        if minPos < 0 {
                minPos = 0
        }
        if minPos >= width {
                minPos = width - 1
        }
        maxPos := int(float64(maxVal-1) / float64(maxVal) * float64(width-1))
        if maxPos < 0 {
                maxPos = 0
        }
        if maxPos >= width {
                maxPos = width - 1
        }
        avgPos := int((avg - 1) / float64(maxVal) * float64(width-1))
        if avgPos < 0 {
                avgPos = 0
        }
        if avgPos >= width {
                avgPos = width - 1
        }
        for i := minPos; i <= maxPos && i < width; i++ {
                bar[i] = '='
        }
        bar[avgPos] = '|'
        return string(bar)
}


