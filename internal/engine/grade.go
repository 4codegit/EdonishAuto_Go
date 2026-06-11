// Package engine implements the grade automation engine with concurrent workers.
package engine

import (
        "fmt"
        "log"
        "math"
        "math/rand"
        "strconv"
        "strings"
        "sync"
        "sync/atomic"
        "time"

        "github.com/4codegit/edonish-auto/internal/api"
        "github.com/4codegit/edonish-auto/internal/config"
)

// TaskStatus represents the current state of a grade task.
type TaskStatus int

const (
        StatusPending TaskStatus = iota
        StatusRunning
        StatusSuccess
        StatusError
        StatusSkipped
)

// GradeTask represents a single grade creation task.
type GradeTask struct {
        StudentID         int
        StudentName       string
        AssignmentDateID  string
        DateStr           string
        QuarterPropertyID int
        Mark              int
        SubjectName       string
        GroupName         string
        Status            TaskStatus
        Error             string
}

// GradePlan represents a complete plan for grade creation.
type GradePlan struct {
        Tasks      []*GradeTask
        TotalTasks int
        Completed  int32
        Failed     int32
        Skipped    int32
        mu         sync.Mutex
}

// NewGradePlan creates an empty grade plan.
func NewGradePlan() *GradePlan {
        return &GradePlan{}
}

// AddTask adds a task to the plan.
func (p *GradePlan) AddTask(t *GradeTask) {
        p.mu.Lock()
        defer p.mu.Unlock()
        p.Tasks = append(p.Tasks, t)
        p.TotalTasks = len(p.Tasks)
}

// Progress returns the current progress as a fraction (0.0 to 1.0).
func (p *GradePlan) Progress() float64 {
        if p.TotalTasks == 0 {
                return 0
        }
        done := int(atomic.LoadInt32(&p.Completed)) + int(atomic.LoadInt32(&p.Failed)) + int(atomic.LoadInt32(&p.Skipped))
        return float64(done) / float64(p.TotalTasks)
}

// PendingCount returns the number of tasks still pending.
func (p *GradePlan) PendingCount() int {
        p.mu.Lock()
        defer p.mu.Unlock()
        count := 0
        for _, t := range p.Tasks {
                if t.Status == StatusPending {
                        count++
                }
        }
        return count
}

// ProgressCallback is called when progress updates.
type ProgressCallback func(plan *GradePlan)

// LogCallback is called for log messages.
type LogCallback func(message, level string)

// Engine handles automated grade creation with parallel processing.
type Engine struct {
        api              *api.Client
        stopChan         chan struct{}
        progressCallback ProgressCallback
        logCallback      LogCallback
        running          atomic.Bool
}

// NewEngine creates a new grade engine.
func NewEngine(client *api.Client) *Engine {
        return &Engine{
                api:      client,
                stopChan: make(chan struct{}),
        }
}

// SetCallbacks sets the progress and log callbacks.
func (e *Engine) SetCallbacks(progressCB ProgressCallback, logCB LogCallback) {
        e.progressCallback = progressCB
        e.logCallback = logCB
}

// IsRunning returns whether the engine is currently executing.
func (e *Engine) IsRunning() bool {
        return e.running.Load()
}

// Stop signals the engine to stop processing.
func (e *Engine) Stop() {
        e.running.Store(false)
        close(e.stopChan)
        e.log("Остановка двигателя оценок...", "warning")
        e.stopChan = make(chan struct{})
}

func (e *Engine) log(message, level string) {
        if e.logCallback != nil {
                e.logCallback(message, level)
        }
        if level == "error" {
                log.Printf("[ERROR] %s", message)
        } else if level == "warning" {
                log.Printf("[WARN] %s", message)
        } else {
                log.Printf("[INFO] %s", message)
        }
}

func (e *Engine) updateProgress(plan *GradePlan) {
        if e.progressCallback != nil {
                e.progressCallback(plan)
        }
}

// BuildGradePlan builds a complete plan of grades to create.
func (e *Engine) BuildGradePlan(
        groups []map[string]interface{},
        subjects []map[string]interface{},
        quarters []map[string]interface{},
        minGrade, maxGrade int,
        fillEmptyOnly bool,
) *GradePlan {
        plan := NewGradePlan()
        e.log("Построение плана оценок...", "info")

        for _, group := range groups {
                groupID := intField(group, "id")
                groupName := fmt.Sprintf("%s%s", stringField(group, "number"), stringField(group, "name"))

                for _, subject := range subjects {
                        subjectID := intField(subject, "subjectId")
                        if subjectID == 0 {
                                subjectID = intField(subject, "id")
                        }
                        subjectName := stringField(subject, "subjectName")
                        if subjectName == "" {
                                subjectName = stringField(subject, "name")
                        }

                        for _, quarter := range quarters {
                                qpropID := intField(quarter, "qpropId")
                                quarterName := stringField(quarter, "name")
                                if quarterName == "" {
                                        quarterName = fmt.Sprintf("Четверть %d", qpropID)
                                }

                                e.log(fmt.Sprintf(" %s | %s | %s", groupName, subjectName, quarterName), "info")

                                datesData, err := e.api.GetJournalDates(groupID, subjectID, qpropID)
                                if err != nil {
                                        e.log(fmt.Sprintf("  Ошибка получения дат: %v", err), "error")
                                        continue
                                }

                                days := ExtractDays(datesData)
                                if len(days) == 0 {
                                        continue
                                }

                                studentsData, err := e.api.GetJournalStudents(groupID, subjectID, qpropID)
                                if err != nil {
                                        e.log(fmt.Sprintf("  Ошибка получения студентов: %v", err), "error")
                                        continue
                                }

                                students := ExtractStudents(studentsData)
                                if len(students) == 0 {
                                        continue
                                }

                                for _, student := range students {
                                        studentID := intField(student, "studentId")
                                        studentName := fmt.Sprintf("%s %s", stringField(student, "lastName"), stringField(student, "firstName"))
                                        existingMarks := ExtractExistingMarks(student)

                                        for _, day := range days {
                                                dateID := stringField(day, "assignmentDateId")
                                                dateStr := stringField(day, "assignmentDate")

                                                if fillEmptyOnly {
                                                        if _, hasMark := existingMarks[dateID]; hasMark {
                                                                task := &GradeTask{
                                                                        StudentID:         studentID,
                                                                        StudentName:       studentName,
                                                                        AssignmentDateID:  dateID,
                                                                        DateStr:           dateStr,
                                                                        QuarterPropertyID: qpropID,
                                                                        Mark:              0,
                                                                        SubjectName:       subjectName,
                                                                        GroupName:         groupName,
                                                                        Status:            StatusSkipped,
                                                                }
                                                                plan.AddTask(task)
                                                                atomic.AddInt32(&plan.Skipped, 1)
                                                                continue
                                                        }
                                                }

                                                grade := minGrade + rand.Intn(maxGrade-minGrade+1)
                                                task := &GradeTask{
                                                        StudentID:         studentID,
                                                        StudentName:       studentName,
                                                        AssignmentDateID:  dateID,
                                                        DateStr:           dateStr,
                                                        QuarterPropertyID: qpropID,
                                                        Mark:              grade,
                                                        SubjectName:       subjectName,
                                                        GroupName:         groupName,
                                                        Status:            StatusPending,
                                                }
                                                plan.AddTask(task)
                                        }
                                }
                        }
                }
        }

        e.log(fmt.Sprintf("План построен: %d задач (%d пропущено)", plan.TotalTasks, int(atomic.LoadInt32(&plan.Skipped))), "info")
        return plan
}

// BuildGradePlanForQuarterMarks builds a plan for quarter marks using ceil(average) calculation.
func (e *Engine) BuildGradePlanForQuarterMarks(
        groups []map[string]interface{},
        subjects []map[string]interface{},
        quarters []map[string]interface{},
        minGrade, maxGrade int,
        fillEmptyOnly bool,
) *GradePlan {
        plan := NewGradePlan()
        e.log("Построение плана четвертных/семестровых/годовых оценок...", "info")

        for _, group := range groups {
                groupID := intField(group, "id")
                groupName := fmt.Sprintf("%s%s", stringField(group, "number"), stringField(group, "name"))

                for _, subject := range subjects {
                        subjectID := intField(subject, "subjectId")
                        if subjectID == 0 {
                                subjectID = intField(subject, "id")
                        }
                        subjectName := stringField(subject, "subjectName")
                        if subjectName == "" {
                                subjectName = stringField(subject, "name")
                        }
                        curriculumPropertyID := intField(subject, "curriculumPropertyId")

                        for _, quarter := range quarters {
                                qpropID := intField(quarter, "qpropId")
                                quarterName := stringField(quarter, "name")

                                studentsData, err := e.api.GetJournalStudents(groupID, subjectID, qpropID)
                                if err != nil {
                                        e.log(fmt.Sprintf("  Ошибка: %v", err), "error")
                                        continue
                                }

                                students := ExtractStudents(studentsData)
                                for _, student := range students {
                                        studentID := intField(student, "studentId")
                                        studentName := fmt.Sprintf("%s %s", stringField(student, "lastName"), stringField(student, "firstName"))

                                        // Check if quarter mark already exists
                                        if fillEmptyOnly {
                                                if qm := getMapField(student, "quarterMark"); qm != nil {
                                                        if arr, ok := qm.([]interface{}); ok && len(arr) > 0 {
                                                                if first, ok := arr[0].(map[string]interface{}); ok {
                                                                        sn := stringField(first, "shortName")
                                                                        mv := intField(first, "mark")
                                                                        display := ParseGradeDisplay(sn, mv)
                                                                        if display != "" && display != config.ABSENT_SHORT {
                                                                                // Quarter mark already set, skip
                                                                                continue
                                                                        }
                                                                }
                                                        }
                                                }
                                        }

                                        // Calculate ceil(average) from existing subject marks
                                        grade := calculateCeilAverage(student, minGrade, maxGrade)
                                        if grade == 0 {
                                                continue // No grades to calculate from
                                        }

                                        task := &GradeTask{
                                                StudentID:         studentID,
                                                StudentName:       studentName,
                                                AssignmentDateID:  "",
                                                DateStr:           quarterName,
                                                QuarterPropertyID: qpropID,
                                                Mark:              grade,
                                                SubjectName:       subjectName,
                                                GroupName:         groupName,
                                                Status:            StatusPending,
                                        }
                                        // Store extra data for quarter mark creation
                                        task.Error = fmt.Sprintf("subject_id:%d,curriculum_property_id:%d", subjectID, curriculumPropertyID)
                                        plan.AddTask(task)
                                }
                        }
                }
        }

        e.log(fmt.Sprintf("План построен: %d четвертных оценок", plan.TotalTasks), "info")
        return plan
}

// calculateCeilAverage calculates ceil(average) from a student's subject marks.
// Returns 0 if no valid grades found.
func calculateCeilAverage(student map[string]interface{}, minGrade, maxGrade int) int {
        var gradeValues []int
        if subjectMarks, ok := student["subjectMarks"].([]interface{}); ok {
                for _, m := range subjectMarks {
                        if mm, ok := m.(map[string]interface{}); ok {
                                sn := stringField(mm, "shortName")
                                mv := intField(mm, "mark")
                                display := ParseGradeDisplay(sn, mv)
                                if display != "" && display != config.ABSENT_SHORT {
                                        if v, err := strconv.Atoi(display); err == nil {
                                                if minGrade <= v && v <= config.MaxGradeAllow {
                                                        gradeValues = append(gradeValues, v)
                                                }
                                        }
                                }
                        }
                }
        }
        if len(gradeValues) == 0 {
                return 0
        }
        sum := 0
        for _, v := range gradeValues {
                sum += v
        }
        avg := float64(sum) / float64(len(gradeValues))
        grade := int(math.Ceil(avg))
        if grade < minGrade {
                grade = minGrade
        }
        if grade > config.MaxGradeAllow {
                grade = config.MaxGradeAllow
        }
        return grade
}

// ExecutePlan executes the grade plan with parallel workers.
func (e *Engine) ExecutePlan(plan *GradePlan, numWorkers int, taskDelay time.Duration) {
        e.running.Store(true)
        defer e.running.Store(false)

        tasks := make([]*GradeTask, 0)
        for _, t := range plan.Tasks {
                if t.Status == StatusPending {
                        tasks = append(tasks, t)
                }
        }

        if len(tasks) == 0 {
                e.log("Нет задач для выполнения", "info")
                return
        }

        e.log(fmt.Sprintf("Запуск %d задач с %d воркерами...", len(tasks), numWorkers), "info")

        workerTasks := make([][]*GradeTask, numWorkers)
        for i, t := range tasks {
                workerIdx := i % numWorkers
                workerTasks[workerIdx] = append(workerTasks[workerIdx], t)
        }

        var wg sync.WaitGroup
        for i, tasks := range workerTasks {
                if len(tasks) == 0 {
                        continue
                }
                wg.Add(1)
                go func(workerID int, tasks []*GradeTask) {
                        defer wg.Done()
                        for _, task := range tasks {
                                select {
                                case <-e.stopChan:
                                        task.Status = StatusSkipped
                                        continue
                                default:
                                }

                                task.Status = StatusRunning
                                e.updateProgress(plan)

                                result, err := e.api.CreateMark(
                                        task.StudentID,
                                        task.AssignmentDateID,
                                        task.Mark,
                                        0, // markTypeID will be set correctly in CreateMark
                                        task.QuarterPropertyID,
                                )

                                if err != nil {
                                        task.Status = StatusError
                                        task.Error = err.Error()
                                        atomic.AddInt32(&plan.Failed, 1)
                                        e.log(fmt.Sprintf("  [%d] %s: %v", workerID, task.StudentName, err), "error")
                                } else if resultMap, ok := result.(map[string]interface{}); ok {
                                        if errMsg, exists := resultMap["error"]; exists && errMsg != nil {
                                                task.Status = StatusError
                                                task.Error = fmt.Sprintf("%v", errMsg)
                                                atomic.AddInt32(&plan.Failed, 1)
                                                e.log(fmt.Sprintf("  [%d] %s: %v", workerID, task.StudentName, errMsg), "error")
                                        } else {
                                                task.Status = StatusSuccess
                                                atomic.AddInt32(&plan.Completed, 1)
                                                e.log(fmt.Sprintf("  [%d] %s -> %d (%s)", workerID, task.StudentName, task.Mark, task.DateStr), "info")
                                        }
                                } else {
                                        task.Status = StatusSuccess
                                        atomic.AddInt32(&plan.Completed, 1)
                                        e.log(fmt.Sprintf("  [%d] %s -> %d (%s)", workerID, task.StudentName, task.Mark, task.DateStr), "info")
                                }

                                e.updateProgress(plan)

                                if e.running.Load() {
                                        time.Sleep(taskDelay)
                                }
                        }
                }(i+1, tasks)
        }

        wg.Wait()

        completed := int(atomic.LoadInt32(&plan.Completed))
        failed := int(atomic.LoadInt32(&plan.Failed))
        skipped := int(atomic.LoadInt32(&plan.Skipped))
        e.log(fmt.Sprintf("Завершено! %d успешно, %d ошибок, %d пропущено", completed, failed, skipped), "info")
        e.updateProgress(plan)
}

// ExecuteQuarterMarks executes quarter marks plan sequentially.
func (e *Engine) ExecuteQuarterMarks(plan *GradePlan, taskDelay time.Duration) {
        e.running.Store(true)
        defer e.running.Store(false)

        tasks := make([]*GradeTask, 0)
        for _, t := range plan.Tasks {
                if t.Status == StatusPending {
                        tasks = append(tasks, t)
                }
        }

        if len(tasks) == 0 {
                e.log("Нет четвертных оценок для выставления", "info")
                return
        }

        for _, task := range tasks {
                select {
                case <-e.stopChan:
                        return
                default:
                }

                task.Status = StatusRunning
                // Parse subject_id and curriculum_property_id from task.Error field (hack for passing extra data)
                subjectID, curriculumPropertyID := 0, 0
                if strings.HasPrefix(task.Error, "subject_id:") {
                        fmt.Sscanf(task.Error, "subject_id:%d,curriculum_property_id:%d", &subjectID, &curriculumPropertyID)
                        task.Error = ""
                }
                result, err := e.api.CreateQuarterMark(task.StudentID, task.QuarterPropertyID, task.Mark, subjectID, curriculumPropertyID)

                if err != nil {
                        task.Status = StatusError
                        task.Error = err.Error()
                        atomic.AddInt32(&plan.Failed, 1)
                        e.log(fmt.Sprintf("  %s: %v", task.StudentName, err), "error")
                } else if result != nil {
                        task.Status = StatusSuccess
                        atomic.AddInt32(&plan.Completed, 1)
                        e.log(fmt.Sprintf("  %s -> %d (%s)", task.StudentName, task.Mark, task.DateStr), "info")
                } else {
                        task.Status = StatusError
                        task.Error = "empty response"
                        atomic.AddInt32(&plan.Failed, 1)
                }

                e.updateProgress(plan)
                time.Sleep(taskDelay)
        }

        completed := int(atomic.LoadInt32(&plan.Completed))
        failed := int(atomic.LoadInt32(&plan.Failed))
        e.log(fmt.Sprintf("Четвертные оценки: %d успешно, %d ошибок", completed, failed), "info")
}

// ─── Data extraction helpers ─────────────────────────────────────────

// ExtractDays extracts days list from API dates response.
func ExtractDays(data interface{}) []map[string]interface{} {
        if data == nil {
                return nil
        }
        if arr, ok := data.([]interface{}); ok && len(arr) > 0 {
                if first, ok := arr[0].(map[string]interface{}); ok {
                        if daysRaw, ok := first["days"].([]interface{}); ok {
                                var result []map[string]interface{}
                                for _, d := range daysRaw {
                                        if dm, ok := d.(map[string]interface{}); ok {
                                                result = append(result, dm)
                                        }
                                }
                                return result
                        }
                }
        }
        return nil
}

// ExtractStudents extracts students list from API students response.
func ExtractStudents(data interface{}) []map[string]interface{} {
        if data == nil {
                return nil
        }
        if arr, ok := data.([]interface{}); ok {
                var result []map[string]interface{}
                for _, s := range arr {
                        if sm, ok := s.(map[string]interface{}); ok {
                                result = append(result, sm)
                        }
                }
                return result
        }
        return nil
}

// ExtractExistingMarks extracts existing marks indexed by assignmentDateId.
func ExtractExistingMarks(student map[string]interface{}) map[string]bool {
        marks := make(map[string]bool)
        if subjectMarks, ok := student["subjectMarks"].([]interface{}); ok {
                for _, m := range subjectMarks {
                        if mm, ok := m.(map[string]interface{}); ok {
                                dateID := stringField(mm, "assignmentDateId")
                                if dateID != "" {
                                        marks[dateID] = true
                                }
                        }
                }
        }
        return marks
}

// ParseGradeDisplay converts API shortName to display text.
// Handles fractional grades like "1/2", "0/2" by extracting the numerator.
// A numerator of 0 means Н/А (Не аттестован).
// mark_value==0 is always Н/А regardless of shortName.
func ParseGradeDisplay(shortName string, markValue int) string {
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

func stringField(m map[string]interface{}, key string) string {
        if v, ok := m[key].(string); ok {
                return v
        }
        if v, ok := m[key].(float64); ok {
                return fmt.Sprintf("%.0f", v)
        }
        return ""
}

func intField(m map[string]interface{}, key string) int {
        if v, ok := m[key].(float64); ok {
                return int(v)
        }
        if v, ok := m[key].(int); ok {
                return v
        }
        return 0
}

func getMapField(m map[string]interface{}, key string) interface{} {
        return m[key]
}
