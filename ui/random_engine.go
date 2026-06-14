package ui

import (
        "math/rand"
        "strings"
)

// Note: In Go 1.20+, the global random generator is automatically
// seeded with a random value. No explicit rand.Seed() or init() needed.

// RandomDiligenceMark picks a random diligence mark from the pool.
// combo can be: "Отличный", "Хорошо", "Удовлетворительный", "Неудовлетворительный"
// or "random" for fully random.
func RandomDiligenceMark(combo string) string {
        if combo == "random" || combo == "" {
                return DiligenceMarks[rand.Intn(len(DiligenceMarks))]
        }
        return combo
}

// RandomGradeInRange returns a random grade between min and max (inclusive).
func RandomGradeInRange(minVal, maxVal int) int {
        if minVal > maxVal {
                minVal, maxVal = maxVal, minVal
        }
        return minVal + rand.Intn(maxVal-minVal+1)
}

// RandomGradeForCombo returns a random grade for a named combo.
func RandomGradeForCombo(comboName string) int {
        for _, c := range GradeCombos {
                if c.Name == comboName {
                        return RandomGradeInRange(c.MinVal, c.MaxVal)
                }
        }
        // Default: Good and Excellent
        return RandomGradeInRange(7, 10)
}

// RandomTopicForDiligence picks a random topic template for a given diligence level.
// Used by the Diaries tab for random signing.
func RandomTopicForDiligence(diligence string) string {
        topics, ok := TopicTemplates[diligence]
        if !ok || len(topics) == 0 {
                // Pick from all
                for _, t := range TopicTemplates {
                        topics = append(topics, t...)
                }
        }
        if len(topics) == 0 {
                return "Урок"
        }
        return topics[rand.Intn(len(topics))]
}

// SequentialTopicForDiligence picks a topic template sequentially for a given diligence level.
// idx is the sequential index; it cycles through the pool using modulo.
// This ensures topics are filled in order, not randomly.
func SequentialTopicForDiligence(diligence string, idx int) string {
        topics, ok := TopicTemplates[diligence]
        if !ok || len(topics) == 0 {
                // Fall back to all topics
                for _, t := range TopicTemplates {
                        topics = append(topics, t...)
                }
        }
        if len(topics) == 0 {
                return "Урок"
        }
        return topics[idx%len(topics)]
}

// RandomDiligenceCombo picks a random diligence combination.
// Returns one of: "Отличный", "Хорошо", "Удовлетворительный", "Неудовлетворительный"
func RandomDiligenceCombo() string {
        // Weighted: more likely to pick "Хорошо" and "Отличный"
        weights := []int{35, 35, 20, 10} // Отличный, Хорошо, Удовлетворительный, Неудовлетворительный
        total := 0
        for _, w := range weights {
                total += w
        }
        r := rand.Intn(total)
        cumulative := 0
        for i, w := range weights {
                cumulative += w
                if r < cumulative {
                        return DiligenceMarks[i]
                }
        }
        return DiligenceMarks[1] // default "Хорошо"
}

// ShouldFillDate determines if a date should be filled based on the weight period.
// period: "Полугодие 1", "Полугодие 2", "Весь год", "До текущей даты"
// quarterName: the quarter name the date belongs to
// currentDate: today's date in "YYYY-MM-DD" format
// assignmentDate: the date of the assignment in "YYYY-MM-DD" format
func ShouldFillDate(period, quarterName, currentDate, assignmentDate string) bool {
        switch period {
        case "Полугодие 1":
                return quarterName == "Четверть 1" || quarterName == "Четверть 2" || quarterName == "Полугодие 1"
        case "Полугодие 2":
                return quarterName == "Четверть 3" || quarterName == "Четверть 4" || quarterName == "Полугодие 2"
        case "До текущей даты":
                // Compare date strings (YYYY-MM-DD format compares lexicographically correctly)
                return assignmentDate <= currentDate
        case "Весь год":
                fallthrough
        default:
                return true
        }
}

// GenerateTopicLine generates a complete topic line with subject context.
// subject: the subject name
// topicBase: the base topic from templates
// lineNum: the lesson number
func GenerateTopicLine(subject, topicBase string, lineNum int) string {
        if strings.TrimSpace(topicBase) == "" {
                return topicBase
        }
        return topicBase
}

// BatchRandomGrades generates a map of studentID -> random grade for a given combo.
func BatchRandomGrades(studentIDs []int, comboName string) map[int]int {
        result := make(map[int]int)
        for _, id := range studentIDs {
                result[id] = RandomGradeForCombo(comboName)
        }
        return result
}

// PickRandomComboName returns a random combo name from GradeCombos.
func PickRandomComboName() string {
        if len(GradeCombos) == 0 {
                return "Хорошо и Отлично"
        }
        return GradeCombos[rand.Intn(len(GradeCombos))].Name
}
