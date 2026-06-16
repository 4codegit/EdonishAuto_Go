// Package client provides the API client for edonish.tj electronic journal.
// All HTTP interaction, authentication, and data types are encapsulated here.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// API base URLs
const (
	APIBase       = "https://api.edonish.tj"
	APILogin      = APIBase + "/auth/v1/login"
	APIRefresh    = APIBase + "/auth/v1/refresh_token"
	APIHeaderInfo = APIBase + "/auth/v1/header/info"
	LangRU        = 2
	MarkTypeID    = 8
	Signature     = "eDonish Auto by 4code"
)

// RolePrefixMap maps role names to API URL prefixes.
var RolePrefixMap = map[string]string{
	"teacher":           "/teacher/v1",
	"classroom-teacher": "/teacher/v1",
	"school_admin":      "/school_admin/v1",
	"director":          "/director/v1",
	"headteacher":       "/headteacher/v1",
}

// EdonishClient is the main API client for edonish.tj.
// It encapsulates HTTP client, auth tokens, and user data.
type EdonishClient struct {
	httpClient   *http.Client
	JWTToken     string
	RefreshToken string
	UserInfo     *UserInfo
	SchoolID     int
	Role         string
	RolePrefix   string
	Schools      []School
}

// UserInfo holds user data from the login response.
type UserInfo struct {
	UID       int    `json:"uid"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// FullName returns the user's full name (Last First).
func (u *UserInfo) FullName() string {
	if u == nil {
		return ""
	}
	return u.LastName + " " + u.FirstName
}

// LoginResponse is the API response on successful authentication.
type LoginResponse struct {
	JWTToken     string `json:"jwt_token"`
	RefreshToken string `json:"refresh_token"`
	UID          int    `json:"uid"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
}

// School represents a school/role from header/info.
type School struct {
	SchoolID   int    `json:"schoolId"`
	Name       string `json:"name"`
	SchoolName string `json:"schoolName"`
}

// HeaderInfoResponse is an element of the /auth/v1/header/info array.
type HeaderInfoResponse struct {
	SchoolID   int    `json:"schoolId"`
	Name       string `json:"name"`
	SchoolName string `json:"schoolName"`
}

// JournalOptions is the OPTIONS /journal response (groups with subjects and quarters).
type JournalOptions struct {
	Groups []JournalGroup `json:"groups"`
}

// JournalGroup represents a class group in the journal.
type JournalGroup struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Subjects []Subject `json:"subjects"`
	Quarters []Quarter `json:"quarters"`
}

// Subject represents a subject in the journal.
type Subject struct {
	SubjectID   int    `json:"subjectId"`
	SubjectName string `json:"subjectName"`
}

// Quarter represents a school quarter.
type Quarter struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	CurrentQuarter bool   `json:"currentQuarter"`
}

// Student represents a student with marks.
type Student struct {
	StudentID    int           `json:"studentId"`
	FirstName    string        `json:"firstName"`
	LastName     string        `json:"lastName"`
	SubjectMarks []SubjectMark `json:"subjectMarks"`
}

// SubjectMark is a student's mark for a specific date.
type SubjectMark struct {
	AssignmentDateID string `json:"assignmentDateId"`
	Mark             int    `json:"mark"`
	ID               int    `json:"id"`
	ShortName        string `json:"shortName"`
}

// Day represents a date in the journal.
type Day struct {
	AssignmentDateID string `json:"assignmentDateId"`
	AssignmentDate   string `json:"assignmentDate"`
}

// CreateMarkRequest is the body for creating a mark.
type CreateMarkRequest struct {
	MarkTypeGroupSubgroupStudentID int    `json:"mark_type_id"`
	GroupSubgroupStudentID         int    `json:"group_subgroup_student_id"`
	ScheduleDateID                 string `json:"schedule_date_id"`
	QuarterPropertyID              int    `json:"quarter_property_id"`
	Mark                           int    `json:"mark"`
	SignatureStr                   string `json:"signature"`
}

// CreateQuarterMarkRequest is the body for creating a quarter mark.
type CreateQuarterMarkRequest struct {
	GroupSubgroupStudentID int `json:"group_subgroup_student_id"`
	QuarterPropertyID      int `json:"quarter_property_id"`
	Mark                   int `json:"mark"`
}

// CreateSemesterMarkRequest is the body for creating a semester mark.
type CreateSemesterMarkRequest struct {
	GroupSubgroupStudentID int `json:"group_subgroup_student_id"`
	SemesterPropertyID     int `json:"semester_property_id"`
	Mark                   int `json:"mark"`
}

// CreateYearMarkRequest is the body for creating a year mark.
type CreateYearMarkRequest struct {
	GroupSubgroupStudentID int `json:"group_subgroup_student_id"`
	YearPropertyID         int `json:"year_property_id"`
	Mark                   int `json:"mark"`
}

// NewEdonishClient creates a new API client with default settings.
func NewEdonishClient() *EdonishClient {
	return &EdonishClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Login authenticates the user via the edonish.tj API.
func (c *EdonishClient) Login(login, password string) error {
	body := map[string]string{
		"login":    login,
		"password": password,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	resp, err := c.httpClient.Post(APILogin, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth error (code %d): %s", resp.StatusCode, string(respBody))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	c.JWTToken = loginResp.JWTToken
	c.RefreshToken = loginResp.RefreshToken
	c.UserInfo = &UserInfo{
		UID:       loginResp.UID,
		FirstName: loginResp.FirstName,
		LastName:  loginResp.LastName,
	}

	return nil
}

// FetchHeaderInfo retrieves user's schools/roles after login.
func (c *EdonishClient) FetchHeaderInfo() error {
	u, err := url.Parse(APIHeaderInfo)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	q := u.Query()
	q.Set("lang", fmt.Sprintf("%d", LangRU))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.JWTToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch header info (code %d): %s", resp.StatusCode, string(respBody))
	}

	var headerInfo []HeaderInfoResponse
	if err := json.Unmarshal(respBody, &headerInfo); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	c.Schools = make([]School, 0, len(headerInfo))
	for _, hi := range headerInfo {
		c.Schools = append(c.Schools, School{
			SchoolID:   hi.SchoolID,
			Name:       hi.Name,
			SchoolName: hi.SchoolName,
		})
	}

	return nil
}

// SelectSchool selects a school and sets the role/prefix.
func (c *EdonishClient) SelectSchool(schoolID int) error {
	for _, school := range c.Schools {
		if school.SchoolID == schoolID {
			c.SchoolID = schoolID
			c.Role = school.Name
			if prefix, ok := RolePrefixMap[school.Name]; ok {
				c.RolePrefix = prefix
			} else {
				c.RolePrefix = "/teacher/v1"
			}
			return nil
		}
	}
	return fmt.Errorf("school with ID %d not found", schoolID)
}

// RefreshJWT refreshes the JWT token using the refresh token.
func (c *EdonishClient) RefreshJWT() error {
	req, err := http.NewRequest("GET", APIRefresh, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.RefreshToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh token error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh token (code %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		JWTToken string `json:"jwt_token"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	c.JWTToken = result.JWTToken
	return nil
}

// apiURL builds the full URL for journal API requests.
func (c *EdonishClient) apiURL(path string, params map[string]string) string {
	base := APIBase + c.RolePrefix + path
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("school_id", fmt.Sprintf("%d", c.SchoolID))
	q.Set("lang", fmt.Sprintf("%d", LangRU))
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// doRequest performs an HTTP request with auth, auto-refreshes on 401.
func (c *EdonishClient) doRequest(method, reqURL string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("encode body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.JWTToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.RefreshJWT(); err != nil {
			return nil, resp.StatusCode, fmt.Errorf("token expired, refresh failed: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.JWTToken)
		if body != nil {
			jsonData, _ := json.Marshal(body)
			req.Body = io.NopCloser(bytes.NewBuffer(jsonData))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewBuffer(jsonData)), nil
			}
		}
		resp2, err := c.httpClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("retry request error: %w", err)
		}
		defer resp2.Body.Close()
		respBody, err = io.ReadAll(resp2.Body)
		if err != nil {
			return nil, resp2.StatusCode, fmt.Errorf("read retry response: %w", err)
		}
		return respBody, resp2.StatusCode, nil
	}

	return respBody, resp.StatusCode, nil
}

// GetJournalOptions retrieves groups (classes) with subjects and quarters.
func (c *EdonishClient) GetJournalOptions() (*JournalOptions, error) {
	u := c.apiURL("/journal", nil)
	respBody, statusCode, err := c.doRequest("OPTIONS", u, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get journal options (code %d): %s", statusCode, string(respBody))
	}

	var opts JournalOptions
	if err := json.Unmarshal(respBody, &opts); err != nil {
		return nil, fmt.Errorf("parse journal options: %w", err)
	}

	return &opts, nil
}

// GetJournalDates retrieves dates for a subject in a quarter.
func (c *EdonishClient) GetJournalDates(groupID, subjectID, quarterID string) ([]Day, error) {
	params := map[string]string{
		"group_id":            groupID,
		"subject_id":          subjectID,
		"quarter_property_id": quarterID,
	}
	u := c.apiURL("/journal/dates", params)

	respBody, statusCode, err := c.doRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get dates (code %d): %s", statusCode, string(respBody))
	}

	var dates []Day
	if err := json.Unmarshal(respBody, &dates); err != nil {
		return nil, fmt.Errorf("parse dates: %w", err)
	}

	return dates, nil
}

// GetJournalStudents retrieves students with marks.
func (c *EdonishClient) GetJournalStudents(groupID, subjectID, quarterID string) ([]Student, error) {
	params := map[string]string{
		"group_id":            groupID,
		"subject_id":          subjectID,
		"quarter_property_id": quarterID,
	}
	u := c.apiURL("/journal/students", params)

	respBody, statusCode, err := c.doRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get students (code %d): %s", statusCode, string(respBody))
	}

	var students []Student
	if err := json.Unmarshal(respBody, &students); err != nil {
		return nil, fmt.Errorf("parse students: %w", err)
	}

	return students, nil
}

// CreateMark creates a grade for a student on a date.
func (c *EdonishClient) CreateMark(studentID int, dateID string, quarterID, mark int) error {
	reqBody := CreateMarkRequest{
		MarkTypeGroupSubgroupStudentID: MarkTypeID,
		GroupSubgroupStudentID:         studentID,
		ScheduleDateID:                 dateID,
		QuarterPropertyID:              quarterID,
		Mark:                           mark,
		SignatureStr:                   Signature,
	}

	u := c.apiURL("/journal/10_point_mark/create", nil)
	respBody, statusCode, err := c.doRequest("POST", u, reqBody)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("create mark (code %d): %s", statusCode, string(respBody))
	}

	return nil
}

// DeleteMark deletes a mark by its ID.
func (c *EdonishClient) DeleteMark(markID string) error {
	params := map[string]string{
		"mark_id": markID,
	}
	u := c.apiURL("/journal/mark/delete", params)

	respBody, statusCode, err := c.doRequest("POST", u, nil)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("delete mark (code %d): %s", statusCode, string(respBody))
	}

	return nil
}

// CreateQuarterMark creates a quarter mark.
func (c *EdonishClient) CreateQuarterMark(studentID, quarterID, mark int) error {
	reqBody := CreateQuarterMarkRequest{
		GroupSubgroupStudentID: studentID,
		QuarterPropertyID:      quarterID,
		Mark:                   mark,
	}

	u := c.apiURL("/journal/10_point_quarter_mark/create", nil)
	respBody, statusCode, err := c.doRequest("POST", u, reqBody)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("create quarter mark (code %d): %s", statusCode, string(respBody))
	}

	return nil
}

// CreateSemesterMark creates a semester mark.
func (c *EdonishClient) CreateSemesterMark(studentID, semesterID, mark int) error {
	reqBody := CreateSemesterMarkRequest{
		GroupSubgroupStudentID: studentID,
		SemesterPropertyID:     semesterID,
		Mark:                   mark,
	}

	u := c.apiURL("/journal/10_point_semester/create", nil)
	respBody, statusCode, err := c.doRequest("POST", u, reqBody)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("create semester mark (code %d): %s", statusCode, string(respBody))
	}

	return nil
}

// CreateYearMark creates a year mark.
func (c *EdonishClient) CreateYearMark(studentID, yearID, mark int) error {
	reqBody := CreateYearMarkRequest{
		GroupSubgroupStudentID: studentID,
		YearPropertyID:         yearID,
		Mark:                   mark,
	}

	u := c.apiURL("/journal/10_point_year/create", nil)
	respBody, statusCode, err := c.doRequest("POST", u, reqBody)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("create year mark (code %d): %s", statusCode, string(respBody))
	}

	return nil
}

// GetDiaryStudents retrieves students for diary operations.
func (c *EdonishClient) GetDiaryStudents(groupID, quarterID string) ([]Student, error) {
	params := map[string]string{
		"group_id":            groupID,
		"quarter_property_id": quarterID,
	}
	u := c.apiURL("/journal/students", params)

	respBody, statusCode, err := c.doRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get diary students (code %d): %s", statusCode, string(respBody))
	}

	var students []Student
	if err := json.Unmarshal(respBody, &students); err != nil {
		return nil, fmt.Errorf("parse diary students: %w", err)
	}

	return students, nil
}

// WriteDiaryComment writes a comment for a student's diary.
func (c *EdonishClient) WriteDiaryComment(studentID int, quarterID, comment string) error {
	reqBody := map[string]interface{}{
		"group_subgroup_student_id": studentID,
		"quarter_property_id":       quarterID,
		"comment":                   comment,
	}

	u := c.apiURL("/diary/comment/create", nil)
	respBody, statusCode, err := c.doRequest("POST", u, reqBody)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("write diary comment (code %d): %s", statusCode, string(respBody))
	}

	return nil
}

// SignDiary signs a student's diary.
func (c *EdonishClient) SignDiary(studentID int, quarterID string) error {
	reqBody := map[string]interface{}{
		"group_subgroup_student_id": studentID,
		"quarter_property_id":       quarterID,
	}

	u := c.apiURL("/diary/sign", nil)
	respBody, statusCode, err := c.doRequest("POST", u, reqBody)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("sign diary (code %d): %s", statusCode, string(respBody))
	}

	return nil
}

// GetTopics retrieves topics for a subject in a quarter.
func (c *EdonishClient) GetTopics(groupID, subjectID, quarterID string) ([]Topic, error) {
	params := map[string]string{
		"group_id":            groupID,
		"subject_id":          subjectID,
		"quarter_property_id": quarterID,
	}
	u := c.apiURL("/journal/topics", params)

	respBody, statusCode, err := c.doRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get topics (code %d): %s", statusCode, string(respBody))
	}

	var topics []Topic
	if err := json.Unmarshal(respBody, &topics); err != nil {
		return nil, fmt.Errorf("parse topics: %w", err)
	}

	return topics, nil
}

// Topic represents a topic/homework entry.
type Topic struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	TopicText   string `json:"topic"`
	Homework    string `json:"homework"`
	SubjectName string `json:"subjectName"`
}
