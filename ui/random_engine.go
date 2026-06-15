package ui

import (
	"fmt"
	"math/rand"
	"time"
)

// RandomGradeInRange returns a random grade between min and max (inclusive).
func RandomGradeInRange(min, max int) int {
	if min > max {
		min, max = max, min
	}
	if min < 2 {
		min = 2
	}
	if max > 10 {
		max = 10
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + r.Intn(max-min+1)
}

// RandomDiligenceMark returns a random diligence mark string.
func RandomDiligenceMark() string {
	marks := []string{"5", "4", "3"}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return marks[r.Intn(len(marks))]
}

// RandomDate returns a random date string in YYYY-MM-DD format within a range.
func RandomDate() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	year := 2025 + r.Intn(2)
	month := 1 + r.Intn(12)
	day := 1 + r.Intn(28)
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

// RandomBehaviorMark returns a random behavior mark key.
func RandomBehaviorMark() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := []string{"5", "4", "3"}
	return keys[r.Intn(len(keys))]
}

// BatchRandomGrades generates a slice of random grades.
func BatchRandomGrades(count, min, max int) []int {
	grades := make([]int, count)
	for i := range grades {
		grades[i] = RandomGradeInRange(min, max)
	}
	return grades
}

// BatchRandomDiligenceMarks generates a slice of random diligence marks.
func BatchRandomDiligenceMarks(count int) []string {
	marks := make([]string, count)
	for i := range marks {
		marks[i] = RandomDiligenceMark()
	}
	return marks
}
