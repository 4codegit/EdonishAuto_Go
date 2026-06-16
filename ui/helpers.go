// Package ui provides helper utilities, color schemes, and custom widgets
// for the eDonish Auto application.
package ui

import (
        "fmt"
        "math"
        "math/rand"
        "strings"
        "time"

        "fyne.io/fyne/v2"
        "fyne.io/fyne/v2/canvas"
        "fyne.io/fyne/v2/container"
        "fyne.io/fyne/v2/theme"
        "fyne.io/fyne/v2/widget"
)

// ─── Diligence / Behavior Constants ────────────────────────────────────────

// DiligenceMarks maps diligence mark keys to display labels.
var DiligenceMarks = map[string]string{
        "5":  "Примерное",
        "4":  "Хорошее",
        "3":  "Удовлетворительное",
        "2":  "Неудовлетворительное",
}

// GradeCombos defines preset grade combinations for random fill.
var GradeCombos = []struct {
        Name  string
        Min   int
        Max   int
}{
        {"Отлично (9-10)", 9, 10},
        {"Хорошо-Отлично (8-10)", 8, 10},
        {"Хорошо (7-9)", 7, 9},
        {"Средне-Хорошо (6-8)", 6, 8},
        {"Средне (5-7)", 5, 7},
        {"Произвольно (2-10)", 2, 10},
}

// WeightPeriods maps period names to API weight identifiers.
var WeightPeriods = map[string]string{
        "Четверть": "quarter",
        "Семестр":  "semester",
        "Год":      "year",
}

// ─── Grade / Average Helpers ───────────────────────────────────────────────

// AvgScorer is an interface for objects that can report an average score.
type AvgScorer interface {
        AverageScore() float64
}

// AverageToGrade converts a numeric average to a 5-point grade.
func AverageToGrade(avg float64) int {
        switch {
        case avg >= 8.5:
                return 5
        case avg >= 6.5:
                return 4
        case avg >= 4.5:
                return 3
        default:
                return 2
        }
}

// ClassAverageToCategory returns a category label for a class average.
func ClassAverageToCategory(avg float64) string {
        switch {
        case avg >= 9.0:
                return "Отлично"
        case avg >= 7.0:
                return "Хорошо"
        case avg >= 5.0:
                return "Удовлетворительно"
        default:
                return "Неудовлетворительно"
        }
}

// ParseAverageScore parses a string like "8.50" into a float64.
func ParseAverageScore(s string) float64 {
        var v float64
        fmt.Sscanf(strings.TrimSpace(s), "%f", &v)
        return v
}

// CalcClassAverage computes the overall class average from student scores.
func CalcClassAverage(scores []float64) float64 {
        if len(scores) == 0 {
                return 0
        }
        sum := 0.0
        for _, s := range scores {
                sum += s
        }
        return math.Round(sum/float64(len(scores))*100) / 100
}

// ─── Behavior / Diligence Helpers ──────────────────────────────────────────

// BehaviorCategory represents a behavior assessment category.
type BehaviorCategory struct {
        Key   string
        Label string
        Color string
}

// BehaviorCategories lists the behavior assessment options.
var BehaviorCategories = []BehaviorCategory{
        {"5", "Примерное", "#22c55e"},
        {"4", "Хорошее", "#3b82f6"},
        {"3", "Удовлетворительное", "#f59e0b"},
        {"2", "Неудовлетворительное", "#ef4444"},
}

// BehaviorTemplates maps behavior keys to comment templates.
var BehaviorTemplates = map[string][]string{
        "5": {
                "Примерное поведение и дисциплина.",
                "Отличная дисциплина и активное участие.",
                "Ведёт себя примерным образом.",
        },
        "4": {
                "Хорошее поведение, старается.",
                "Дисциплинированный, активный на уроках.",
                "Хорошее отношение к учёбе.",
        },
        "3": {
                "Удовлетворительное поведение, есть замечания.",
                "Нуждается в усилении дисциплины.",
                "Не всегда внимателен на уроках.",
        },
        "2": {
                "Неудовлетворительное поведение.",
                "Частые нарушения дисциплины.",
                "Не соблюдает правила поведения.",
        },
}

// BehaviorToDiligence maps behavior key to diligence key.
var BehaviorToDiligence = map[string]string{
        "5": "5",
        "4": "4",
        "3": "3",
        "2": "2",
}

// DiligenceToBehaviorComment maps diligence key to a comment.
var DiligenceToBehaviorComment = map[string]string{
        "5": "Примерное поведение, активное участие в жизни класса.",
        "4": "Хорошее поведение, добросовестное отношение к учёбе.",
        "3": "Удовлетворительное поведение, требует внимания.",
        "2": "Неудовлетворительное поведение, частые нарушения.",
}

// ─── Sign Comment Helpers ──────────────────────────────────────────────────

// SignCommentTemplates provides templates for diary sign comments.
var SignCommentTemplates = []string{
        "Ознакомлен(а).",
        "Ознакомлен(а) с оценками.",
        "Проверено.",
        "С оценками ознакомлен(а).",
        "Подтверждено.",
}

// RandomSignComment returns a random sign comment.
func RandomSignComment() string {
        r := rand.New(rand.NewSource(time.Now().UnixNano()))
        return SignCommentTemplates[r.Intn(len(SignCommentTemplates))]
}

// ─── Color Helpers ─────────────────────────────────────────────────────────

// Modern dark-accent color palette (primary definitions in dashboard.go)

// getDiligenceColor returns a color name for a diligence mark.
func getDiligenceColor(key string) string {
        switch key {
        case "5":
                return "#22c55e"
        case "4":
                return "#3b82f6"
        case "3":
                return "#f59e0b"
        case "2":
                return "#ef4444"
        default:
                return "#6b7280"
        }
}

// getBehaviorColor returns a color name for a behavior key.
func getBehaviorColor(key string) string {
        switch key {
        case "5":
                return "#22c55e"
        case "4":
                return "#3b82f6"
        case "3":
                return "#f59e0b"
        case "2":
                return "#ef4444"
        default:
                return "#6b7280"
        }
}

// GradeColor returns a color string for a grade value.
func GradeColor(grade int) string {
        switch {
        case grade >= 9:
                return "#22c55e" // green
        case grade >= 7:
                return "#3b82f6" // blue
        case grade >= 5:
                return "#f59e0b" // yellow
        case grade >= 3:
                return "#f97316" // orange
        default:
                return "#ef4444" // red
        }
}

// ─── UI Helpers ────────────────────────────────────────────────────────────

// MakeFixedHeader creates a non-scrolling header bar with the given content.
func MakeFixedHeader(title string, objects ...fyne.CanvasObject) *fyne.Container {
        titleText := canvas.NewText(title, theme.Color(theme.ColorNameForeground))
        titleText.TextStyle = fyne.TextStyle{Bold: true}
        titleText.TextSize = 16

        items := []fyne.CanvasObject{titleText}
        items = append(items, objects...)
        return container.NewHBox(items...)
}

// FormatSignedStatus returns a formatted string for signed/unsigned status.
func FormatSignedStatus(signed bool) string {
        if signed {
                return "✓ Подписано"
        }
        return "✗ Не подписано"
}

// FormatStudentName formats a student's full name.
func FormatStudentName(lastName, firstName string) string {
        return lastName + " " + firstName
}

// ─── tapOverlay Widget ─────────────────────────────────────────────────────

// tapOverlay is a transparent widget that captures taps and can pass them
// through to underlying widgets. Useful for dismissing popups.
type tapOverlay struct {
        widget.BaseWidget
        onTapped func()
}

// NewTapOverlay creates a new tap overlay widget.
func NewTapOverlay(onTapped func()) *tapOverlay {
        t := &tapOverlay{onTapped: onTapped}
        t.ExtendBaseWidget(t)
        return t
}

// CreateRenderer implements fyne.Widget.
func (t *tapOverlay) CreateRenderer() fyne.WidgetRenderer {
        return widget.NewSimpleRenderer(canvas.NewRectangle(colorSurface))
}

// Tapped implements fyne.Tappable.
func (t *tapOverlay) Tapped(_ *fyne.PointEvent) {
        if t.onTapped != nil {
                t.onTapped()
        }
}

// ─── KeyboardScrollTable Widget ────────────────────────────────────────────

// KeyboardScrollTable wraps a widget.Table inside a scroll container and
// implements fyne.Focusable to handle keyboard navigation.
type KeyboardScrollTable struct {
        widget.BaseWidget
        scroll   *container.Scroll
        table    *widget.Table
        onKeyDown func(*fyne.KeyEvent)
        onRune   func(rune)
        focused  bool
}

// NewKeyboardScrollTable creates a new KeyboardScrollTable wrapping the given table.
func NewKeyboardScrollTable(tbl *widget.Table) *KeyboardScrollTable {
        k := &KeyboardScrollTable{
                table: tbl,
                scroll: container.NewScroll(tbl),
        }
        k.ExtendBaseWidget(k)
        return k
}

// CreateRenderer implements fyne.Widget.
func (k *KeyboardScrollTable) CreateRenderer() fyne.WidgetRenderer {
        return widget.NewSimpleRenderer(k.scroll)
}

// FocusGained implements fyne.Focusable.
func (k *KeyboardScrollTable) FocusGained() {
        k.focused = true
}

// FocusLost implements fyne.Focusable.
func (k *KeyboardScrollTable) FocusLost() {
        k.focused = false
}

// TypedKey implements fyne.Focusable.
func (k *KeyboardScrollTable) TypedKey(e *fyne.KeyEvent) {
        if k.onKeyDown != nil {
                k.onKeyDown(e)
        }
}

// TypedRune implements fyne.Focusable.
func (k *KeyboardScrollTable) TypedRune(r rune) {
        if k.onRune != nil {
                k.onRune(r)
        }
}

// Table returns the wrapped widget.Table.
func (k *KeyboardScrollTable) Table() *widget.Table {
        return k.table
}

// Scroll returns the scroll container.
func (k *KeyboardScrollTable) Scroll() *container.Scroll {
        return k.scroll
}

// Focused returns whether the widget currently has focus.
func (k *KeyboardScrollTable) Focused() bool {
        return k.focused
}
