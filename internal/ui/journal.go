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

// studentAnalysis holds detailed analysis for a single student.
type studentAnalysis struct {
        studentName string
        subjects    []subjectAnalysis
}

// subjectAnalysis holds per-subject analysis.
type subjectAnalysis struct {
        groupName   string
        subjectName string
        quarterName string
        grades      []int
        min         int
        max         int
        avg         float64
        missing     int
        total       int
        classAvg    float64
        classMin    int
        classMax    int
}

// ─── JournalPage ──────────────────────────────────────────────

// JournalPage holds the journal viewer UI components.
type JournalPage struct {
        app *App

        // Filters
        classSelect   *widget.Select
        subjectSelect *widget.Select
        quarterSelect *widget.Select

        // Loading state
        loadingLabel *widget.Label

        // Table view
        journalTable *widget.Table
        journalData  []journalData // one per group/subject/quarter combo

        // Analysis view
        studentSelect  *widget.Select
        analysisTable  *widget.Table
        analysisData   []studentAnalysis
        analysisDetail *widget.Entry

        // Container references for mode switching
        tableContainer   *fyne.Container
        analysisContainer *fyne.Container
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

        loadBtn := widget.NewButtonWithIcon("Загрузить", theme.ViewRefreshIcon(), func() {
                p.loadJournal()
        })
        loadBtn.Importance = widget.HighImportance

        modeSelect := widget.NewSelect([]string{"Таблица", "Анализ ученика"}, func(s string) {
                p.onModeChange(s)
        })
        modeSelect.PlaceHolder = "Режим"

        p.loadingLabel = widget.NewLabelWithStyle("Выберите класс, предмет и четверть, затем нажмите «Загрузить»",
                fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

        // ── Journal Table ─────────────────────────────────────
        // Create an empty table that will be populated after loading.
        // Columns: [N, ФИО ученика, date1, date2, ..., Ср, Мин, Макс]
        // We start with 4 columns and 1 row (placeholder).
        p.journalTable = widget.NewTable(
                func() (int, int) { return p.journalTableRowCount(), p.journalTableColCount() },
                func() fyne.CanvasObject {
                        return widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})
                },
                func(id widget.TableCellID, cell fyne.CanvasObject) {
                        p.journalTableCellUpdate(id, cell.(*widget.Label))
                },
        )
        p.journalTable.SetColumnWidth(0, 40)  // #
        p.journalTable.SetColumnWidth(1, 200) // ФИО
        // date columns default width set after loading

        // ── Analysis: student selector + detail ───────────────
        p.studentSelect = widget.NewSelect([]string{}, func(s string) {
                p.onStudentChange(s)
        })
        p.studentSelect.PlaceHolder = "Выберите ученика..."

        // Analysis table: [Предмет, Четверть, Ср, Мин, Макс, Расброс, vs Класс]
        p.analysisTable = widget.NewTable(
                func() (int, int) { return p.analysisRowCount(), 7 },
                func() fyne.CanvasObject {
                        return widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})
                },
                func(id widget.TableCellID, cell fyne.CanvasObject) {
                        p.analysisTableCellUpdate(id, cell.(*widget.Label))
                },
        )
        p.analysisTable.SetColumnWidth(0, 150) // Предмет
        p.analysisTable.SetColumnWidth(1, 100) // Четверть
        p.analysisTable.SetColumnWidth(2, 60)  // Ср
        p.analysisTable.SetColumnWidth(3, 50)  // Мин
        p.analysisTable.SetColumnWidth(4, 50)  // Макс
        p.analysisTable.SetColumnWidth(5, 60)  // Расброс
        p.analysisTable.SetColumnWidth(6, 100) // vs Класс

        p.analysisDetail = widget.NewMultiLineEntry()
        p.analysisDetail.SetPlaceHolder("Подробный анализ появится здесь...")
        p.analysisDetail.Wrapping = fyne.TextWrapWord
        p.analysisDetail.TextStyle = fyne.TextStyle{Monospace: true}
        p.analysisDetail.SetMinRowsVisible(12)

        // ── Containers ────────────────────────────────────────
        p.tableContainer = container.NewBorder(p.loadingLabel, nil, nil, nil, p.journalTable)

        analysisHeader := container.NewVBox(
                p.studentSelect,
                widget.NewSeparator(),
                p.analysisTable,
                widget.NewSeparator(),
        )
        analysisDetailCard := widget.NewCard("Подробный анализ", "", p.analysisDetail)
        p.analysisContainer = container.NewBorder(analysisHeader, nil, nil, nil, analysisDetailCard)
        p.analysisContainer.Hide()

        // ── Toolbar ───────────────────────────────────────────
        toolbar := widget.NewCard("Просмотр журнала", "", container.NewVBox(
                container.NewGridWithColumns(5,
                        p.classSelect,
                        p.subjectSelect,
                        p.quarterSelect,
                        modeSelect,
                        loadBtn,
                ),
        ))

        content := container.NewVBox(
                toolbar,
                p.tableContainer,
                p.analysisContainer,
        )

        scroll := container.NewVScroll(content)
        scroll.SetMinSize(fyne.NewSize(900, 600))

        return scroll
}

// ─── Mode switching ───────────────────────────────────────────

func (p *JournalPage) onModeChange(mode string) {
        if p.tableContainer == nil || p.analysisContainer == nil {
                return
        }
        switch mode {
        case "Анализ ученика":
                p.tableContainer.Hide()
                p.analysisContainer.Show()
        default:
                p.tableContainer.Show()
                p.analysisContainer.Hide()
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

// ─── Journal Table helpers ────────────────────────────────────

func (p *JournalPage) journalTableRowCount() int {
        rows := 0
        for _, jd := range p.journalData {
                rows += 1 // header row
                rows += len(jd.students)
                rows += 1 // spacer
        }
        if rows == 0 {
                rows = 1 // placeholder
        }
        return rows
}

func (p *JournalPage) journalTableColCount() int {
        maxCols := 4 // #, ФИО, Ср, Мин/Макс
        for _, jd := range p.journalData {
                cols := 3 + len(jd.dates) + 3 // #, ФИО, dates..., Ср, Мин, Макс
                if cols > maxCols {
                        maxCols = cols
                }
        }
        return maxCols
}

func (p *JournalPage) journalTableCellUpdate(id widget.TableCellID, label *widget.Label) {
        if len(p.journalData) == 0 {
                if id.Row == 0 && id.Col == 1 {
                        label.SetText("Нет данных — нажмите «Загрузить»")
                } else {
                        label.SetText("")
                }
                return
        }

        // Map row index to the correct journalData + local row
        rowIdx := 0
        for di, jd := range p.journalData {
                // Header row
                if id.Row == rowIdx {
                        p.journalHeaderCell(id.Col, jd, label, di)
                        return
                }
                rowIdx++

                // Student rows
                if id.Row < rowIdx+len(jd.students) {
                        si := id.Row - rowIdx
                        p.journalStudentCell(id.Col, jd, jd.students[si], label, di)
                        return
                }
                rowIdx += len(jd.students)

                // Spacer row
                if id.Row == rowIdx {
                        p.journalSpacerCell(id.Col, label, di)
                        return
                }
                rowIdx++
        }

        label.SetText("")
}

func (p *JournalPage) journalHeaderCell(col int, jd journalData, label *widget.Label, dataIdx int) {
        totalCols := 3 + len(jd.dates) + 3
        if col >= totalCols {
                label.SetText("")
                return
        }

        label.TextStyle = fyne.TextStyle{Bold: true}

        switch {
        case col == 0:
                label.SetText("#")
        case col == 1:
                // Show group/subject/quarter as header
                label.SetText(fmt.Sprintf("%s │ %s │ %s", jd.groupName, jd.subjectName, jd.quarterName))
        case col >= 2 && col < 2+len(jd.dates):
                label.SetText(jd.dates[col-2].shortStr)
        case col == 2+len(jd.dates):
                label.SetText("Ср")
        case col == 2+len(jd.dates)+1:
                label.SetText("Мин")
        case col == 2+len(jd.dates)+2:
                label.SetText("Макс")
        }
}

func (p *JournalPage) journalStudentCell(col int, jd journalData, sr studentRow, label *widget.Label, dataIdx int) {
        totalCols := 3 + len(jd.dates) + 3
        if col >= totalCols {
                label.SetText("")
                return
        }

        label.TextStyle = fyne.TextStyle{}

        switch {
        case col == 0:
                // Find student index
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
                        // Color code by mark value
                        if val, okv := sr.markValues[dateID]; okv {
                                if val >= 9 {
                                        label.TextStyle = fyne.TextStyle{Bold: true}
                                } else if val <= 3 {
                                        label.TextStyle = fyne.TextStyle{Italic: true}
                                }
                        }
                } else {
                        label.SetText("—")
                }
        case col == 2+len(jd.dates):
                if sr.gradeCount > 0 {
                        label.SetText(fmt.Sprintf("%.1f", sr.avg))
                } else {
                        label.SetText("—")
                }
        case col == 2+len(jd.dates)+1:
                if sr.gradeCount > 0 {
                        label.SetText(fmt.Sprintf("%d", sr.min))
                } else {
                        label.SetText("—")
                }
        case col == 2+len(jd.dates)+2:
                if sr.gradeCount > 0 {
                        label.SetText(fmt.Sprintf("%d", sr.max))
                } else {
                        label.SetText("—")
                }
        }
}

func (p *JournalPage) journalSpacerCell(col int, label *widget.Label, dataIdx int) {
        label.SetText("")
}

// ─── Load journal data ────────────────────────────────────────

func (p *JournalPage) loadJournal() {
        classSelected := p.classSelect.Selected
        subjectSelected := p.subjectSelect.Selected
        quarterSelected := p.quarterSelect.Selected

        p.loadingLabel.SetText("Загрузка журнала...")
        p.app.LogMessage(fmt.Sprintf("Загрузка журнала: %s / %s / %s", classSelected, subjectSelected, quarterSelected), "info")

        go func() {
                groups := p.getSelectedGroups(classSelected)
                if len(groups) == 0 {
                        fyne.Do(func() {
                                p.loadingLabel.SetText("Не выбран класс")
                        })
                        return
                }

                var allData []journalData
                // Collect student names for analysis selector
                studentNames := make(map[string]bool)

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
                                                continue
                                        }
                                        days := engine.ExtractDays(datesData)
                                        if len(days) == 0 {
                                                continue
                                        }

                                        // Get students
                                        studentsData, err := p.app.apiClient.GetJournalStudents(groupID, subjectID, qpropID)
                                        if err != nil {
                                                continue
                                        }
                                        students := engine.ExtractStudents(studentsData)
                                        if len(students) == 0 {
                                                continue
                                        }

                                        // Build date columns
                                        var dateCols []dateCol
                                        for _, day := range days {
                                                dateID := mapStr(day, "assignmentDateId")
                                                dateStr := mapStr(day, "assignmentDate")
                                                shortStr := dateStr
                                                if len(dateStr) >= 10 {
                                                        shortStr = dateStr[5:10] // MM-DD
                                                }
                                                dateCols = append(dateCols, dateCol{
                                                        dateID:   dateID,
                                                        dateStr:  dateStr,
                                                        shortStr: shortStr,
                                                })
                                        }

                                        // Build student rows
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

                                                // Compute stats
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
                                                studentNames[studentName] = true
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
                                        p.journalTable.SetColumnWidth(2+i, 50)
                                }
                                // Ср, Мин, Макс columns
                                cols := len(jd.dates)
                                p.journalTable.SetColumnWidth(2+cols, 50)   // Ср
                                p.journalTable.SetColumnWidth(2+cols+1, 45) // Мин
                                p.journalTable.SetColumnWidth(2+cols+2, 45) // Макс
                        }

                        if len(allData) == 0 {
                                p.loadingLabel.SetText("Нет данных для отображения")
                        } else {
                                totalStudents := 0
                                for _, jd := range allData {
                                        totalStudents += len(jd.students)
                                }
                                p.loadingLabel.SetText(fmt.Sprintf("Загружено: %d класс/предмет/четверть, %d учеников",
                                        len(allData), totalStudents))
                        }

                        // Update student selector for analysis mode
                        var names []string
                        for n := range studentNames {
                                names = append(names, n)
                        }
                        sort.Strings(names)
                        p.studentSelect.Options = names
                        if len(names) > 0 {
                                p.studentSelect.SetSelectedIndex(0)
                        }
                        p.studentSelect.Refresh()
                })
        }()
}

// ─── Analysis mode ────────────────────────────────────────────

func (p *JournalPage) analysisRowCount() int {
        if len(p.analysisData) == 0 {
                return 1
        }
        rows := 1 // header
        for _, sa := range p.analysisData {
                rows += len(sa.subjects)
        }
        return rows
}

func (p *JournalPage) analysisTableCellUpdate(id widget.TableCellID, label *widget.Label) {
        if len(p.analysisData) == 0 {
                if id.Row == 0 && id.Col == 0 {
                        label.SetText("Выберите ученика")
                } else {
                        label.SetText("")
                }
                return
        }

        // Header row
        if id.Row == 0 {
                headers := []string{"Предмет", "Четверть", "Ср", "Мин", "Макс", "Расброс", "vs Класс"}
                if id.Col < len(headers) {
                        label.TextStyle = fyne.TextStyle{Bold: true}
                        label.SetText(headers[id.Col])
                }
                return
        }

        // Find the correct subject analysis
        row := id.Row - 1
        for _, sa := range p.analysisData {
                if row < len(sa.subjects) {
                        ss := sa.subjects[row]
                        switch id.Col {
                        case 0:
                                label.SetText(ss.subjectName)
                        case 1:
                                label.SetText(ss.quarterName)
                        case 2:
                                if len(ss.grades) > 0 {
                                        label.SetText(fmt.Sprintf("%.1f", ss.avg))
                                } else {
                                        label.SetText("—")
                                }
                        case 3:
                                if len(ss.grades) > 0 {
                                        label.SetText(fmt.Sprintf("%d", ss.min))
                                } else {
                                        label.SetText("—")
                                }
                        case 4:
                                if len(ss.grades) > 0 {
                                        label.SetText(fmt.Sprintf("%d", ss.max))
                                } else {
                                        label.SetText("—")
                                }
                        case 5:
                                if len(ss.grades) > 0 {
                                        spread := ss.max - ss.min
                                        bar := makeGradeBar(ss.avg, ss.min, ss.max, 8)
                                        label.SetText(fmt.Sprintf("%d %s", spread, bar))
                                } else {
                                        label.SetText("—")
                                }
                        case 6:
                                if ss.classAvg > 0 && len(ss.grades) > 0 {
                                        diff := ss.avg - ss.classAvg
                                        sign := "+"
                                        if diff < 0 {
                                                sign = ""
                                        }
                                        label.SetText(fmt.Sprintf("%s%.1f (кл:%.1f)", sign, diff, ss.classAvg))
                                } else {
                                        label.SetText("—")
                                }
                        }
                        return
                }
                row -= len(sa.subjects)
        }

        label.SetText("")
}

func (p *JournalPage) onStudentChange(selected string) {
        if selected == "" {
                p.analysisData = nil
                p.analysisTable.Refresh()
                p.analysisDetail.SetText("")
                return
        }
        go p.loadStudentAnalysis(selected)
}

func (p *JournalPage) loadStudentAnalysis(studentName string) {
        classSelected := p.classSelect.Selected
        subjectSelected := p.subjectSelect.Selected
        quarterSelected := p.quarterSelect.Selected

        groups := p.getSelectedGroups(classSelected)

        var allSubjectStats []subjectAnalysis

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

                                var targetGrades []int
                                var targetMissing int
                                var allClassGrades []int
                                found := false

                                for _, student := range students {
                                        sName := fmt.Sprintf("%s %s", mapStr(student, "lastName"), mapStr(student, "firstName"))
                                        existingMarks := engine.ExtractExistingMarks(student)
                                        markDetails := extractMarkDetails(student)

                                        var sGrades []int
                                        for _, day := range days {
                                                dateID := mapStr(day, "assignmentDateId")
                                                if _, has := existingMarks[dateID]; has {
                                                        if mi, ok := markDetails[dateID]; ok && mi.markValue > 0 {
                                                                sGrades = append(sGrades, mi.markValue)
                                                        }
                                                }
                                        }

                                        allClassGrades = append(allClassGrades, sGrades...)

                                        if sName == studentName {
                                                found = true
                                                targetGrades = sGrades
                                                targetMissing = len(days) - len(existingMarks)
                                        }
                                }

                                if !found {
                                        continue
                                }

                                ss := subjectAnalysis{
                                        groupName:   groupName,
                                        subjectName: subjectName,
                                        quarterName: quarterName,
                                        grades:      targetGrades,
                                        missing:     targetMissing,
                                        total:       len(days),
                                }

                                if len(targetGrades) > 0 {
                                        ss.min = targetGrades[0]
                                        ss.max = targetGrades[0]
                                        sum := 0
                                        for _, g := range targetGrades {
                                                sum += g
                                                if g < ss.min {
                                                        ss.min = g
                                                }
                                                if g > ss.max {
                                                        ss.max = g
                                                }
                                        }
                                        ss.avg = float64(sum) / float64(len(targetGrades))
                                }

                                if len(allClassGrades) > 0 {
                                        classMin, classMax, classSum := allClassGrades[0], allClassGrades[0], 0
                                        for _, g := range allClassGrades {
                                                classSum += g
                                                if g < classMin {
                                                        classMin = g
                                                }
                                                if g > classMax {
                                                        classMax = g
                                                }
                                        }
                                        ss.classAvg = float64(classSum) / float64(len(allClassGrades))
                                        ss.classMin = classMin
                                        ss.classMax = classMax
                                }

                                allSubjectStats = append(allSubjectStats, ss)
                        }
                }
        }

        // Build detail text
        var detailLines []string
        detailLines = append(detailLines, fmt.Sprintf("АНАЛИЗ: %s", studentName))
        detailLines = append(detailLines, strings.Repeat("=", 50))

        // Overall stats
        var allGrades []int
        for _, ss := range allSubjectStats {
                allGrades = append(allGrades, ss.grades...)
        }
        if len(allGrades) > 0 {
                min, max, sum := allGrades[0], allGrades[0], 0
                for _, g := range allGrades {
                        sum += g
                        if g < min {
                                min = g
                        }
                        if g > max {
                                max = g
                        }
                }
                avg := float64(sum) / float64(len(allGrades))
                detailLines = append(detailLines, fmt.Sprintf("Общая средняя: %.1f  |  Мин: %d  |  Макс: %d  |  Расброс: %d",
                        avg, min, max, max-min))
                detailLines = append(detailLines, fmt.Sprintf("Всего оценок: %d", len(allGrades)))
                detailLines = append(detailLines, "")
                detailLines = append(detailLines, fmt.Sprintf("Расброс: %s", makeVisualSpread(min, max, avg, 10)))
                detailLines = append(detailLines, fmt.Sprintf("Распределение: %s", makeDistribution(allGrades)))
        } else {
                detailLines = append(detailLines, "Нет оценок")
        }

        fyne.Do(func() {
                p.analysisData = []studentAnalysis{{
                        studentName: studentName,
                        subjects:    allSubjectStats,
                }}
                p.analysisTable.Refresh()
                p.analysisDetail.SetText(strings.Join(detailLines, "\n"))
        })
}

// ─── Helper functions ─────────────────────────────────────────

// makeGradeBar creates a text-based visual bar representing the grade spread.
func makeGradeBar(avg float64, min, max, width int) string {
        if max <= min || width <= 0 {
                return "[" + strings.Repeat("=", width) + "]"
        }
        bar := make([]rune, width)
        for i := range bar {
                bar[i] = ' '
        }
        minPos := int(float64(min-1) / float64(max) * float64(width-1))
        if minPos < 0 {
                minPos = 0
        }
        if minPos >= width {
                minPos = width - 1
        }
        maxPos := int(float64(max-1) / float64(max) * float64(width-1))
        if maxPos < 0 {
                maxPos = 0
        }
        if maxPos >= width {
                maxPos = width - 1
        }
        avgPos := int((avg - 1) / float64(max) * float64(width-1))
        if avgPos < 0 {
                avgPos = 0
        }
        if avgPos >= width {
                avgPos = width - 1
        }
        for i := minPos; i <= maxPos && i < width; i++ {
                bar[i] = '='
        }
        if avgPos >= 0 && avgPos < width {
                bar[avgPos] = 'o'
        }
        return "[" + string(bar) + "]"
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

