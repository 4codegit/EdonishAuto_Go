// Package config holds all application constants and configuration.
package config

import (
	"os"
	"path/filepath"
)

// Application metadata
const (
	AppName    = "eDonish Auto"
	AppVersion = "0.4.0"
)

// API base configuration
const (
	APIBase       = "https://api.edonish.tj"
	APILogin      = APIBase + "/auth/v1/login"
	APIRefresh    = APIBase + "/auth/v1/refresh_token"
	APIHeaderInfo = APIBase + "/auth/v1/header/info"
)

// Role-based API prefixes
var APIPrefixes = map[string]string{
	"teacher":           "/teacher/v1",
	"classroom-teacher": "/teacher/v1",
	"school_admin":      "/school_admin/v1",
	"director":          "/director/v1",
	"headteacher":       "/headteacher/v1",
	"chief_curator":     "/chief_curator/v1",
	"regional_curator":  "/regional_curator/v1",
	"parent":            "/parent/v1",
	"student":           "/student/v1",
}

// Journal API endpoints (relative to role prefix)
const (
	JournalOptions        = "/journal"
	JournalDates          = "/journal/dates"
	JournalStudents       = "/journal/students"
	JournalStudentsFinal  = "/journal/students/final"
	JournalDatesFinal     = "/journal/dates/final"
	JournalMarkCreate     = "/journal/10_point_mark/create"
	JournalMarkDelete     = "/journal/mark/delete"
	JournalQuarterCreate  = "/journal/10_point_quarter_mark/create"
	JournalSemesterCreate = "/journal/10_point_semester/create"
	JournalYearCreate     = "/journal/10_point_year/create"
	JournalAssignmentUpd  = "/journal/assignment/update"
	JournalComment        = "/journal/comment"
	PeriodQuarters        = "/period/quaters"
	GroupsList            = "/groups/list"
	TeacherSubject        = "/teacher/subject"
	Subgroups             = "/subgroups"
)

// Language codes
const (
	LangTJ = 1 // Тоҷикӣ
	LangRU = 2 // Русский
	LangEN = 3 // English
)

// Grade settings
const (
	MinGrade       = 5   // Minimum valid grade
	MaxGrade       = 10  // Maximum grade for random/auto
	MaxGradeAllow  = 11  // Allow manual entry up to 11
	DefaultWorkers = 4
)

// Absent grade (mark=0, mark_type_id=1 in edonish API)
const (
	ABSENT_MARK    = 0         // Mark value sent to API for absent (Н/А)
	ABSENT_SHORT   = "Н/А"    // Short display in journal cells (edonish standard)
	ABSENT_DISPLAY = "Отсутствует" // UI label for Н/А grade
	ABSENT_MarkTypeID = 1     // mark_type_id for absent in edonish API
)

// SessionFile returns the path to the session persistence file.
func SessionFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".edonish_session.json")
}

// LogFile returns the path to the log file.
func LogFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".edonish_auto.log")
}
