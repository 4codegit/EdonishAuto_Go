package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/4codegit/edonish-auto/internal/config"
	"github.com/4codegit/edonish-auto/internal/engine"
)

// AutoGradePage holds the auto-grade page UI components.
type AutoGradePage struct {
	app *App

	// Settings
	classSelect     *widget.Select
	subjectSelect   *widget.Select
	quarterSelect   *widget.Select
	minGradeEntry   *widget.Entry
	maxGradeEntry   *widget.Entry
	workersEntry    *widget.Entry
	fillEmptyChk    *widget.Check
	quarterMarksChk *widget.Check

	// Actions
	analyzeBtn *widget.Button
	startBtn   *widget.Button
	stopBtn    *widget.Button

	// Progress
	progressBar   *widget.ProgressBar
	progressLabel *widget.Label
	statsLabel    *widget.Label

	// Results
	resultsEntry *widget.Entry
}

// NewAutoGradePage creates a new auto-grade page.
func NewAutoGradePage(app *App) *AutoGradePage {
	return &AutoGradePage{app: app}
}

// Build creates the auto-grade view and returns the root container.
func (p *AutoGradePage) Build() fyne.CanvasObject {
	// ── Settings ────────────────────────────────────────────────
	p.classSelect = widget.NewSelect([]string{"Все классы"}, func(s string) {})
	p.classSelect.PlaceHolder = "Класс"

	p.subjectSelect = widget.NewSelect([]string{"Все предметы"}, func(s string) {})
	p.subjectSelect.PlaceHolder = "Предмет"

	p.quarterSelect = widget.NewSelect([]string{"Все четверти"}, func(s string) {})
	p.quarterSelect.PlaceHolder = "Четверть"

	p.minGradeEntry = widget.NewEntry()
	p.minGradeEntry.SetText(fmt.Sprintf("%d", config.MinGrade))

	p.maxGradeEntry = widget.NewEntry()
	p.maxGradeEntry.SetText(fmt.Sprintf("%d", config.MaxGrade))

	p.workersEntry = widget.NewEntry()
	p.workersEntry.SetText(fmt.Sprintf("%d", config.DefaultWorkers))

	p.fillEmptyChk = widget.NewCheck("Только пустые ячейки", nil)
	p.fillEmptyChk.SetChecked(true)

	p.quarterMarksChk = widget.NewCheck("Четвертные оценки", nil)
	p.quarterMarksChk.SetChecked(true)

	gradeRange := container.NewHBox(
		p.minGradeEntry,
		widget.NewLabel("—"),
		p.maxGradeEntry,
	)

	settingsCard := widget.NewCard("Настройки", "", container.NewVBox(
		container.NewGridWithColumns(2,
			container.NewVBox(
				p.classSelect,
				p.quarterSelect,
				container.NewHBox(
					widget.NewLabel("Воркеры:"),
					p.workersEntry,
				),
			),
			container.NewVBox(
				p.subjectSelect,
				container.NewHBox(
					widget.NewLabel("Оценки:"),
					gradeRange,
				),
				p.fillEmptyChk,
				p.quarterMarksChk,
			),
		),
	))

	// ── Action buttons ──────────────────────────────────────────
	p.analyzeBtn = widget.NewButton("Анализировать", func() {
		p.doAnalyze()
	})

	p.startBtn = widget.NewButton("Запустить", func() {
		p.doStart()
	})
	p.startBtn.Importance = widget.HighImportance
	p.startBtn.Disable()

	p.stopBtn = widget.NewButton("Стоп", func() {
		p.doStop()
	})
	p.stopBtn.Disable()

	actionCard := widget.NewCard("", "", container.NewHBox(
		p.analyzeBtn,
		p.startBtn,
		p.stopBtn,
		layout.NewSpacer(),
	))

	// ── Progress ────────────────────────────────────────────────
	p.progressLabel = widget.NewLabelWithStyle("Готов к работе", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	p.progressBar = widget.NewProgressBar()
	p.progressBar.Min = 0
	p.progressBar.Max = 1
	p.statsLabel = widget.NewLabel("")

	progressCard := widget.NewCard("", "", container.NewVBox(
		p.progressLabel,
		p.progressBar,
		p.statsLabel,
	))

	// ── Results ─────────────────────────────────────────────────
	p.resultsEntry = widget.NewMultiLineEntry()
	p.resultsEntry.SetPlaceHolder("Результаты анализа появятся здесь")
	p.resultsEntry.Wrapping = fyne.TextWrapWord
	p.resultsEntry.SetMinRowsVisible(12)

	resultsCard := widget.NewCard("Результаты", "", p.resultsEntry)

	// ── Main layout ─────────────────────────────────────────────
	content := container.NewVBox(
		settingsCard,
		actionCard,
		progressCard,
		resultsCard,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(900, 600))

	return scroll
}

// UpdateDropdowns populates dropdowns with loaded data.
func (p *AutoGradePage) UpdateDropdowns() {
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

// UpdateProgress updates the progress display from the engine.
func (p *AutoGradePage) UpdateProgress(plan *engine.GradePlan) {
	progress := plan.Progress()
	p.progressBar.SetValue(progress)

	completed := int(plan.Completed)
	failed := int(plan.Failed)
	skipped := int(plan.Skipped)
	total := plan.TotalTasks

	p.progressLabel.SetText(fmt.Sprintf("Выполнение: %d/%d", completed+failed, total-int(skipped)))
	p.statsLabel.SetText(fmt.Sprintf("Успешно: %d | Ошибки: %d | Пропущено: %d", completed, failed, skipped))
	p.progressBar.Refresh()
	p.progressLabel.Refresh()
	p.statsLabel.Refresh()
}

// getSelectedGroups returns the groups matching the dropdown selection.
func (p *AutoGradePage) getSelectedGroups() []map[string]interface{} {
	selected := p.classSelect.Selected
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

// getSelectedSubjects returns the subjects matching the dropdown selection.
func (p *AutoGradePage) getSelectedSubjects() []map[string]interface{} {
	selected := p.subjectSelect.Selected
	if selected == "Все предметы" || selected == "" {
		return p.app.teacherSubjects
	}
	var result []map[string]interface{}
	for _, s := range p.app.teacherSubjects {
		if name, _ := s["subjectName"].(string); name == selected {
			result = append(result, s)
		}
	}
	return result
}

// getSelectedQuarters returns the quarters matching the dropdown selection.
func (p *AutoGradePage) getSelectedQuarters() []map[string]interface{} {
	selected := p.quarterSelect.Selected
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

// doAnalyze starts the grade analysis.
func (p *AutoGradePage) doAnalyze() {
	p.analyzeBtn.Disable()
	p.analyzeBtn.SetText("Анализ...")
	p.progressBar.SetValue(0)
	p.progressLabel.SetText("Анализ...")
	p.resultsEntry.SetText("Анализ журнала...")

	go func() {
		groups := p.getSelectedGroups()
		subjects := p.getSelectedSubjects()
		quarters := p.getSelectedQuarters()

		minGrade := config.MinGrade
		maxGrade := config.MaxGrade
		if v := parseInt(p.minGradeEntry.Text); v > 0 {
			minGrade = v
		}
		if v := parseInt(p.maxGradeEntry.Text); v > 0 {
			maxGrade = v
		}

		plan := p.app.engine.BuildGradePlan(
			groups, subjects, quarters,
			minGrade, maxGrade,
			p.fillEmptyChk.Checked,
		)

		p.app.currentPlan = plan
		p.onAnalyzeComplete(plan)
	}()
}

// onAnalyzeComplete handles the analysis completion.
func (p *AutoGradePage) onAnalyzeComplete(plan *engine.GradePlan) {
	p.analyzeBtn.Enable()
	p.analyzeBtn.SetText("Анализировать")
	p.startBtn.Enable()

	toExecute := plan.PendingCount()

	lines := "════════════════════════════════════════════════════════\n"
	lines += "  ПЛАН ОЦЕНОК\n"
	lines += "════════════════════════════════════════════════════════\n\n"
	lines += fmt.Sprintf("  Всего задач:      %d\n", plan.TotalTasks)
	lines += fmt.Sprintf("  Будет выполнено:  %d\n", toExecute)
	lines += fmt.Sprintf("  Пропущено:        %d\n\n", int(plan.Skipped))

	// Group by class/subject
	type groupKey struct{ group, subject string }
	groupMap := make(map[groupKey][]*engine.GradeTask)
	for _, t := range plan.Tasks {
		if t.Status == engine.StatusPending {
			key := groupKey{t.GroupName, t.SubjectName}
			groupMap[key] = append(groupMap[key], t)
		}
	}

	for key, tasks := range groupMap {
		lines += fmt.Sprintf("  %s | %s\n", key.group, key.subject)
		lines += fmt.Sprintf("    Оценок: %d\n", len(tasks))
		for i, t := range tasks {
			if i >= 5 {
				lines += fmt.Sprintf("    ... и ещё %d\n", len(tasks)-5)
				break
			}
			lines += fmt.Sprintf("    - %s -> %d (%s)\n", t.StudentName, t.Mark, t.DateStr)
		}
		lines += "\n"
	}

	p.resultsEntry.SetText(lines)
	p.progressLabel.SetText(fmt.Sprintf("Анализ завершён: %d оценок будет добавлено", toExecute))
	p.resultsEntry.Refresh()
	p.progressLabel.Refresh()
}

// doStart starts the grade execution.
func (p *AutoGradePage) doStart() {
	if p.app.currentPlan == nil {
		p.app.LogMessage("Сначала выполните анализ!", "warning")
		return
	}

	if p.app.currentPlan.PendingCount() == 0 {
		p.app.LogMessage("Нет оценок для добавления!", "warning")
		return
	}

	p.startBtn.Disable()
	p.stopBtn.Enable()
	p.analyzeBtn.Disable()
	p.progressLabel.SetText("Заполнение...")

	numWorkers := config.DefaultWorkers
	if v := parseInt(p.workersEntry.Text); v > 0 {
		numWorkers = v
	}

	go func() {
		p.app.engine.ExecutePlan(p.app.currentPlan, numWorkers, 150*time.Millisecond)

		if p.quarterMarksChk.Checked {
			p.app.LogMessage("Заполнение четвертных оценок...", "info")
			qplan := p.app.engine.BuildGradePlanForQuarterMarks(
				p.getSelectedGroups(),
				p.getSelectedSubjects(),
				p.getSelectedQuarters(),
				parseInt(p.minGradeEntry.Text),
				parseInt(p.maxGradeEntry.Text),
				p.fillEmptyChk.Checked,
			)
			if qplan.TotalTasks > 0 {
				p.app.engine.ExecuteQuarterMarks(qplan, 200*time.Millisecond)
			}
		}

		p.onExecutionComplete()
	}()
}

// doStop stops the engine.
func (p *AutoGradePage) doStop() {
	p.app.engine.Stop()
	p.stopBtn.Disable()
	p.app.LogMessage("Остановка...", "warning")
}

// onExecutionComplete handles execution completion.
func (p *AutoGradePage) onExecutionComplete() {
	p.startBtn.Enable()
	p.stopBtn.Disable()
	p.analyzeBtn.Enable()

	if p.app.currentPlan != nil {
		plan := p.app.currentPlan
		done := int(plan.Completed) + int(plan.Failed)
		total := plan.TotalTasks - int(plan.Skipped)
		p.progressLabel.SetText(fmt.Sprintf("Завершено: %d/%d", done, total))
		p.statsLabel.SetText(fmt.Sprintf("Успешно: %d | Ошибки: %d | Пропущено: %d",
			int(plan.Completed), int(plan.Failed), int(plan.Skipped)))
	}
	p.progressLabel.Refresh()
	p.statsLabel.Refresh()
}

// parseInt safely parses an integer from a string.
func parseInt(s string) int {
	var v int
	fmt.Sscanf(s, "%d", &v)
	return v
}
