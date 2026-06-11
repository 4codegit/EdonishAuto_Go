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

// JournalPage holds the journal viewer UI components.
type JournalPage struct {
        app *App

        // Filters
        classSelect   *widget.Select
        subjectSelect *widget.Select
        quarterSelect *widget.Select

        // Display mode
        modeSelect *widget.Select

        // Output
        journalEntry *widget.Entry

        // Student analysis
        studentSelect *widget.Select
        analysisEntry *widget.Entry

        // Card references for show/hide toggling
        journalCard  *widget.Card
        analysisCard *widget.Card
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

        p.modeSelect = widget.NewSelect([]string{"Таблица", "Анализ ученика"}, func(s string) {
                p.onModeChange(s)
        })
        p.modeSelect.PlaceHolder = "Режим"
        p.modeSelect.SetSelectedIndex(0)

        loadBtn := widget.NewButtonWithIcon("Загрузить", theme.ViewRefreshIcon(), func() {
                p.loadJournal()
        })
        loadBtn.Importance = widget.HighImportance

        p.journalEntry = widget.NewMultiLineEntry()
        p.journalEntry.SetPlaceHolder("Выберите класс, предмет и четверть, затем нажмите «Загрузить»")
        p.journalEntry.Wrapping = fyne.TextWrapWord
        p.journalEntry.TextStyle = fyne.TextStyle{Monospace: true}
        p.journalEntry.SetMinRowsVisible(18)

        p.studentSelect = widget.NewSelect([]string{}, func(s string) {
                p.onStudentChange(s)
        })
        p.studentSelect.PlaceHolder = "Выберите ученика..."

        p.analysisEntry = widget.NewMultiLineEntry()
        p.analysisEntry.SetPlaceHolder("Выберите ученика для анализа оценок...")
        p.analysisEntry.Wrapping = fyne.TextWrapWord
        p.analysisEntry.TextStyle = fyne.TextStyle{Monospace: true}
        p.analysisEntry.SetMinRowsVisible(18)

        // Toolbar
        toolbar := widget.NewCard("Просмотр журнала", "", container.NewVBox(
                container.NewGridWithColumns(5,
                        p.classSelect,
                        p.subjectSelect,
                        p.quarterSelect,
                        p.modeSelect,
                        loadBtn,
                ),
        ))

        // Journal table view
        p.journalCard = widget.NewCard("Журнал оценок", "", p.journalEntry)

        // Student analysis view (hidden by default)
        p.studentSelect.Hide()
        p.analysisCard = widget.NewCard("Расброс оценок ученика", "",
                container.NewVBox(
                        p.studentSelect,
                        widget.NewSeparator(),
                        p.analysisEntry,
                ),
        )
        p.analysisCard.Hide()

        content := container.NewVBox(
                toolbar,
                p.journalCard,
                p.analysisCard,
        )

        scroll := container.NewVScroll(content)
        scroll.SetMinSize(fyne.NewSize(900, 600))

        return scroll
}

// onModeChange switches between table and analysis mode.
func (p *JournalPage) onModeChange(mode string) {
        switch mode {
        case "Анализ ученика":
                p.journalCard.Hide()
                p.analysisCard.Show()
                p.studentSelect.Show()
        default: // "Таблица"
                p.journalCard.Show()
                p.analysisCard.Hide()
                p.studentSelect.Hide()
        }
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

// studentGrade holds computed grade statistics for a student.
type studentGrade struct {
        name    string
        grades  []int
        min     int
        max     int
        avg     float64
        sum     int
        count   int
        missing int
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
                        fyne.Do(func() {
                                p.journalEntry.SetText("Не выбран класс")
                        })
                        return
                }

                var allLines []string
                // Collect all students across groups for student selector
                allStudents := make(map[string]studentGrade)

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
                                                allLines = append(allLines, fmt.Sprintf("Ошибка дат: %v", err))
                                                continue
                                        }

                                        days := engine.ExtractDays(datesData)
                                        if len(days) == 0 {
                                                allLines = append(allLines, fmt.Sprintf("Нет дат: %s | %s | %s", groupName, subjectName, quarterName))
                                                continue
                                        }

                                        // Get students
                                        studentsData, err := p.app.apiClient.GetJournalStudents(groupID, subjectID, qpropID)
                                        if err != nil {
                                                allLines = append(allLines, fmt.Sprintf("Ошибка студентов: %v", err))
                                                continue
                                        }

                                        students := engine.ExtractStudents(studentsData)

                                        // Build journal display
                                        allLines = append(allLines, "")
                                        allLines = append(allLines, fmt.Sprintf("═══ %s │ %s │ %s ═══", groupName, subjectName, quarterName))
                                        allLines = append(allLines, "")

                                        // Header row with dates
                                        header := "Ученик                \t"
                                        for _, day := range days {
                                                dateStr := mapStr(day, "assignmentDate")
                                                if len(dateStr) >= 10 {
                                                        dateStr = dateStr[5:10] // MM-DD
                                                }
                                                header += dateStr + "\t"
                                        }
                                        header += "Ср│Мин│Макс"
                                        allLines = append(allLines, header)
                                        allLines = append(allLines, strings.Repeat("─", len(header)+10))

                                        // Student rows
                                        for _, student := range students {
                                                studentName := fmt.Sprintf("%s %s", mapStr(student, "lastName"), mapStr(student, "firstName"))
                                                row := studentName + "\t"

                                                existingMarks := engine.ExtractExistingMarks(student)
                                                markDetails := extractMarkDetails(student)

                                                var studentGrades []int
                                                missingCount := 0

                                                for _, day := range days {
                                                        dateID := mapStr(day, "assignmentDateId")
                                                        if _, has := existingMarks[dateID]; has {
                                                                if markInfo, ok := markDetails[dateID]; ok {
                                                                        display := engine.ParseGradeDisplay(markInfo.shortName, markInfo.markValue)
                                                                        row += display + "\t"
                                                                        if markInfo.markValue > 0 {
                                                                                studentGrades = append(studentGrades, markInfo.markValue)
                                                                        }
                                                                } else {
                                                                        row += "+\t"
                                                                }
                                                        } else {
                                                                row += "—\t"
                                                                missingCount++
                                                        }
                                                }

                                                // Compute stats for this student in this subject/quarter
                                                sg := computeStudentGrades(studentName, studentGrades, missingCount)
                                                key := fmt.Sprintf("%s|%s|%s", studentName, subjectName, quarterName)
                                                allStudents[key] = sg

                                                // Append min/max/avg to the row
                                                if sg.count > 0 {
                                                        row += fmt.Sprintf("%.1f│%d│%d", sg.avg, sg.min, sg.max)
                                                } else {
                                                        row += "—│—│—"
                                                }

                                                allLines = append(allLines, row)
                                        }

                                        // Class summary
                                        allLines = append(allLines, "")
                                        allLines = append(allLines, p.buildClassSummary(allStudents, groupName, subjectName, quarterName))
                                }
                        }
                }

                if len(allLines) == 0 {
                        allLines = append(allLines, "Нет данных для отображения")
                }

                result := strings.Join(allLines, "\n")
                fyne.Do(func() {
                        p.journalEntry.SetText(result)

                        // Update student selector for analysis mode
                        p.updateStudentSelector(allStudents)
                })
        }()
}

// computeStudentGrades computes min, max, avg for a student's grades.
func computeStudentGrades(name string, grades []int, missing int) studentGrade {
        sg := studentGrade{
                name:    name,
                grades:  grades,
                missing: missing,
                count:   len(grades),
        }
        if len(grades) == 0 {
                return sg
        }
        sg.min = grades[0]
        sg.max = grades[0]
        sg.sum = 0
        for _, g := range grades {
                sg.sum += g
                if g < sg.min {
                        sg.min = g
                }
                if g > sg.max {
                        sg.max = g
                }
        }
        sg.avg = float64(sg.sum) / float64(sg.count)
        return sg
}

// buildClassSummary builds a summary block for the class showing overall stats.
func (p *JournalPage) buildClassSummary(students map[string]studentGrade, groupName, subjectName, quarterName string) string {
        var lines []string
        lines = append(lines, fmt.Sprintf("  ┌─ Сводка: %s │ %s │ %s ─┐", groupName, subjectName, quarterName))

        var allGrades []int
        studentStats := make([]studentGrade, 0)

        for _, sg := range students {
                studentStats = append(studentStats, sg)
                allGrades = append(allGrades, sg.grades...)
        }

        if len(allGrades) == 0 {
                lines = append(lines, "  │ Нет оценок")
                lines = append(lines, "  └────────────────────────────────┘")
                return strings.Join(lines, "\n")
        }

        // Overall class stats
        classMin, classMax, classSum := allGrades[0], allGrades[0], 0
        for _, g := range allGrades {
                classSum += g
                if g < classMin {
                        classMin = g
                }
                if g > classMax {
                        classMax = g
                }
        }
        classAvg := float64(classSum) / float64(len(allGrades))

        lines = append(lines, fmt.Sprintf("  │ Учеников: %d  │  Оценок: %d", len(studentStats), len(allGrades)))
        lines = append(lines, fmt.Sprintf("  │ Средняя: %.1f  │  Мин: %d  │  Макс: %d", classAvg, classMin, classMax))

        // Sort students by average descending
        sort.Slice(studentStats, func(i, j int) bool {
                return studentStats[i].avg > studentStats[j].avg
        })

        // Top 5
        lines = append(lines, "  │")
        lines = append(lines, "  │ Топ ученики:")
        for i, sg := range studentStats {
                if i >= 5 {
                        break
                }
                if sg.count > 0 {
                        spread := sg.max - sg.min
                        bar := makeGradeBar(sg.avg, sg.min, sg.max, 10)
                        lines = append(lines, fmt.Sprintf("  │  %d. %-25s ср:%.1f мин:%d макс:%d расп:%d %s",
                                i+1, sg.name, sg.avg, sg.min, sg.max, spread, bar))
                }
        }

        // Bottom 3
        if len(studentStats) > 5 {
                lines = append(lines, "  │")
                lines = append(lines, "  │ Ниже среднего:")
                start := len(studentStats) - 3
                if start < 5 {
                        start = 5
                }
                for i := start; i < len(studentStats); i++ {
                        sg := studentStats[i]
                        if sg.count > 0 {
                                spread := sg.max - sg.min
                                bar := makeGradeBar(sg.avg, sg.min, sg.max, 10)
                                lines = append(lines, fmt.Sprintf("  │  %-25s ср:%.1f мин:%d макс:%d расп:%d %s",
                                        sg.name, sg.avg, sg.min, sg.max, spread, bar))
                        }
                }
        }

        lines = append(lines, "  └────────────────────────────────┘")
        return strings.Join(lines, "\n")
}

// makeGradeBar creates a text-based visual bar representing the grade spread.
// Shows the student's grade position as a bar from min to max.
func makeGradeBar(avg float64, min, max, width int) string {
        if max <= min || width <= 0 {
                return "[" + strings.Repeat("=", width) + "]"
        }
        bar := make([]rune, width)
        for i := range bar {
                bar[i] = ' '
        }
        // Mark min position
        minPos := int(float64(min-1) / float64(max) * float64(width-1))
        if minPos < 0 {
                minPos = 0
        }
        if minPos >= width {
                minPos = width - 1
        }
        // Mark max position
        maxPos := int(float64(max-1) / float64(max) * float64(width-1))
        if maxPos < 0 {
                maxPos = 0
        }
        if maxPos >= width {
                maxPos = width - 1
        }
        // Mark avg position
        avgPos := int((avg - 1) / float64(max) * float64(width-1))
        if avgPos < 0 {
                avgPos = 0
        }
        if avgPos >= width {
                avgPos = width - 1
        }

        // Fill range
        for i := minPos; i <= maxPos && i < width; i++ {
                bar[i] = '='
        }
        // Mark average
        if avgPos >= 0 && avgPos < width {
                bar[avgPos] = 'o'
        }

        return "[" + string(bar) + "]"
}

// updateStudentSelector populates the student selector for analysis mode.
func (p *JournalPage) updateStudentSelector(students map[string]studentGrade) {
        var names []string
        for key, sg := range students {
                if sg.count > 0 {
                        names = append(names, key)
                }
        }
        sort.Strings(names)
        p.studentSelect.Options = names
        if len(names) > 0 {
                p.studentSelect.SetSelectedIndex(0)
        }
        p.studentSelect.Refresh()
}

// onStudentChange handles student selection in analysis mode.
func (p *JournalPage) onStudentChange(selected string) {
        if selected == "" {
                p.analysisEntry.SetText("")
                return
        }

        // Re-fetch data for this specific student to show detailed analysis
        go p.loadStudentAnalysis(selected)
}

// loadStudentAnalysis loads detailed grade analysis for a specific student.
func (p *JournalPage) loadStudentAnalysis(studentKey string) {
        // Parse the key: "StudentName|SubjectName|QuarterName"
        parts := strings.SplitN(studentKey, "|", 3)
        if len(parts) < 3 {
                p.analysisEntry.SetText("Ошибка: неверный формат данных ученика")
                return
        }
        studentName := parts[0]

        p.analysisEntry.SetText("Загрузка анализа...")

        classSelected := p.classSelect.Selected
        subjectSelected := p.subjectSelect.Selected
        quarterSelected := p.quarterSelect.Selected

        groups := p.getSelectedGroups(classSelected)
        var allLines []string

        // Header
        allLines = append(allLines, "============================================================")
        allLines = append(allLines, fmt.Sprintf("  АНАЛИЗ ОЦЕНОК: %s", studentName))
        allLines = append(allLines, "============================================================")
        allLines = append(allLines, "")

        // Collect per-subject stats
        type subjectStats struct {
                subjectName string
                quarterName string
                groupName   string
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

        var allSubjectStats []subjectStats

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

                                // Find target student and collect all grades for class comparison
                                var targetGrades []int
                                var targetMissing int
                                var allClassGrades []int
                                found := false

                                for _, student := range students {
                                        sName := fmt.Sprintf("%s %s", mapStr(student, "lastName"), mapStr(student, "firstName"))
                                        existingMarks := engine.ExtractExistingMarks(student)
                                        markDetails := extractMarkDetails(student)

                                        var studentGrades []int

                                        for _, day := range days {
                                                dateID := mapStr(day, "assignmentDateId")
                                                if _, has := existingMarks[dateID]; has {
                                                        if markInfo, ok := markDetails[dateID]; ok && markInfo.markValue > 0 {
                                                                studentGrades = append(studentGrades, markInfo.markValue)
                                                        }
                                                }
                                        }

                                        allClassGrades = append(allClassGrades, studentGrades...)

                                        if sName == studentName {
                                                found = true
                                                targetGrades = studentGrades
                                                targetMissing = len(days) - len(existingMarks)
                                        }
                                }

                                if !found {
                                        continue
                                }

                                // Compute target student stats
                                ss := subjectStats{
                                        subjectName: subjectName,
                                        quarterName: quarterName,
                                        groupName:   groupName,
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

                                // Compute class stats
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

        if len(allSubjectStats) == 0 {
                allLines = append(allLines, "Нет данных для анализа")
        } else {
                // Overall student summary
                var allStudentGrades []int
                for _, ss := range allSubjectStats {
                        allStudentGrades = append(allStudentGrades, ss.grades...)
                }

                if len(allStudentGrades) > 0 {
                        overallMin, overallMax, overallSum := allStudentGrades[0], allStudentGrades[0], 0
                        for _, g := range allStudentGrades {
                                overallSum += g
                                if g < overallMin {
                                        overallMin = g
                                }
                                if g > overallMax {
                                        overallMax = g
                                }
                        }
                        overallAvg := float64(overallSum) / float64(len(allStudentGrades))

                        allLines = append(allLines, "  ОБЩИЙ РЕЗУЛЬТАТ:")
                        allLines = append(allLines, fmt.Sprintf("  ├─ Средняя оценка:     %.1f", overallAvg))
                        allLines = append(allLines, fmt.Sprintf("  ├─ Минимальная оценка: %d", overallMin))
                        allLines = append(allLines, fmt.Sprintf("  ├─ Максимальная оценка:%d", overallMax))
                        allLines = append(allLines, fmt.Sprintf("  ├─ Разброс:            %d", overallMax-overallMin))
                        allLines = append(allLines, fmt.Sprintf("  └─ Всего оценок:       %d", len(allStudentGrades)))
                        allLines = append(allLines, "")
                        allLines = append(allLines, fmt.Sprintf("  Расброс:  %s", makeVisualSpread(overallMin, overallMax, overallAvg, 10)))
                        allLines = append(allLines, "")
                }

                // Per-subject breakdown
                allLines = append(allLines, "  ПО ПРЕДМЕТАМ:")
                allLines = append(allLines, strings.Repeat("-", 60))

                for _, ss := range allSubjectStats {
                        allLines = append(allLines, "")
                        allLines = append(allLines, fmt.Sprintf("  ┌─ %s │ %s │ %s", ss.groupName, ss.subjectName, ss.quarterName))

                        if len(ss.grades) > 0 {
                                spread := ss.max - ss.min
                                allLines = append(allLines, fmt.Sprintf("  │ Средняя:    %.1f", ss.avg))
                                allLines = append(allLines, fmt.Sprintf("  │ Мин:        %d", ss.min))
                                allLines = append(allLines, fmt.Sprintf("  │ Макс:       %d", ss.max))
                                allLines = append(allLines, fmt.Sprintf("  │ Расброс:    %d", spread))
                                allLines = append(allLines, fmt.Sprintf("  │ Оценок:     %d из %d (пропущено: %d)", len(ss.grades), ss.total, ss.missing))

                                // Visual spread bar
                                allLines = append(allLines, fmt.Sprintf("  │ %s", makeVisualSpread(ss.min, ss.max, ss.avg, 10)))

                                // Comparison with class
                                if ss.classAvg > 0 {
                                        diff := ss.avg - ss.classAvg
                                        diffSign := "+"
                                        if diff < 0 {
                                                diffSign = ""
                                        }
                                        allLines = append(allLines, fmt.Sprintf("  │ Класс ср: %.1f  %s%.1f от класса", ss.classAvg, diffSign, diff))
                                        allLines = append(allLines, fmt.Sprintf("  │ Класс мин: %d  макс: %d", ss.classMin, ss.classMax))
                                }

                                // Grade distribution
                                allLines = append(allLines, fmt.Sprintf("  │ Распределение: %s", makeDistribution(ss.grades)))

                                // Individual grades list
                                gradeStrs := make([]string, len(ss.grades))
                                for i, g := range ss.grades {
                                        gradeStrs[i] = fmt.Sprintf("%d", g)
                                }
                                allLines = append(allLines, fmt.Sprintf("  │ Оценки: %s", strings.Join(gradeStrs, ", ")))
                        } else {
                                allLines = append(allLines, "  │ Нет оценок")
                        }

                        allLines = append(allLines, "  └────────────────────────────────────")
                }
        }

        result := strings.Join(allLines, "\n")
        fyne.Do(func() {
                p.analysisEntry.SetText(result)
        })
}

// makeVisualSpread creates a visual text-based spread indicator.
// Shows a bar from min to max with the average marked.
func makeVisualSpread(min, max int, avg float64, scale int) string {
        if max < min || scale <= 0 {
                return "[---]"
        }

        barWidth := scale * 2 // double resolution
        bar := make([]rune, barWidth)
        for i := range bar {
                bar[i] = '.'
        }

        // Fill the range from min to max
        for i := 0; i < barWidth; i++ {
                pos := float64(i) / float64(barWidth-1) * float64(max)
                if pos >= float64(min) && pos <= float64(max) {
                        bar[i] = '='
                }
        }

        // Mark average position
        avgPos := int((avg - 1) / float64(max) * float64(barWidth-1))
        if avgPos < 0 {
                avgPos = 0
        }
        if avgPos >= barWidth {
                avgPos = barWidth - 1
        }
        bar[avgPos] = 'o'

        // Mark min and max
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

        // Count grades
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

        var parts []string
        // Sort grade values
        var gradeValues []int
        for g := range counts {
                gradeValues = append(gradeValues, g)
        }
        sort.Ints(gradeValues)

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

