package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/4codegit/edonish-auto/client"
)

// Controller interface provides access to the client, window, and logout.
type Controller interface {
	GetClient() *client.EdonishClient
	GetWindow() fyne.Window
	Logout()
}

// LoginForm is the login screen with school/username/password fields.
type LoginForm struct {
	controller Controller
	container  *fyne.Container
	loginEntry *widget.Entry
	passEntry  *widget.Entry
	schoolSel  *widget.Select
	loginBtn   *widget.Button
}

// NewLoginForm creates a new login form.
func NewLoginForm(ctrl Controller) *LoginForm {
	lf := &LoginForm{controller: ctrl}

	// Title
	titleText := canvas.NewText("eDonish Auto v5.5", theme.Color(theme.ColorNameForeground))
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.TextSize = 28

	// Subtitle
	subtitleText := canvas.NewText("Автоматизация электронного журнала edonish.tj", theme.Color(theme.ColorNameForeground))
	subtitleText.TextSize = 14

	// Login entry
	lf.loginEntry = widget.NewEntry()
	lf.loginEntry.SetPlaceHolder("Введите логин")

	// Password entry
	lf.passEntry = widget.NewPasswordEntry()
	lf.passEntry.SetPlaceHolder("Введите пароль")

	// Login button
	lf.loginBtn = widget.NewButton("Войти", lf.onLogin)
	lf.loginBtn.Importance = widget.HighImportance

	// School selector (hidden until login succeeds)
	lf.schoolSel = widget.NewSelect([]string{}, lf.onSchoolSelected)
	lf.schoolSel.PlaceHolder = "Выберите школу/роль..."

	// Handle Enter key in entries
	lf.loginEntry.OnSubmitted = func(_ string) {
		lf.passEntry.FocusGained()
	}
	lf.passEntry.OnSubmitted = func(_ string) {
		lf.loginBtn.OnTapped()
	}

	// Build form
	form := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), titleText, layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), subtitleText, layout.NewSpacer()),
		widget.NewSeparator(),
		widget.NewForm(
			&widget.FormItem{Text: "Логин", Widget: lf.loginEntry},
			&widget.FormItem{Text: "Пароль", Widget: lf.passEntry},
		),
		lf.loginBtn,
		lf.schoolSel,
		layout.NewSpacer(),
	)

	lf.container = container.NewPadded(
		container.NewCenter(
			container.NewVBox(
				container.NewHBox(layout.NewSpacer(), form, layout.NewSpacer()),
			),
		),
	)

	return lf
}

// Container returns the root container.
func (lf *LoginForm) Container() fyne.CanvasObject {
	return lf.container
}

// GetLoginEntry returns the login entry widget for focus.
func (lf *LoginForm) GetLoginEntry() *widget.Entry {
	return lf.loginEntry
}

// onLogin handles the login button press.
func (lf *LoginForm) onLogin() {
	login := lf.loginEntry.Text
	password := lf.passEntry.Text

	if login == "" || password == "" {
		dialog.ShowInformation("Ошибка", "Введите логин и пароль", lf.controller.GetWindow())
		return
	}

	lf.loginBtn.Disable()
	lf.loginBtn.SetText("Вход...")

	go func() {
		err := lf.controller.GetClient().Login(login, password)
		fyne.Do(func() {
			lf.loginBtn.Enable()
			lf.loginBtn.SetText("Войти")

			if err != nil {
				dialog.ShowError(err, lf.controller.GetWindow())
				return
			}

			// Fetch schools/roles
			err = lf.controller.GetClient().FetchHeaderInfo()
			if err != nil {
				dialog.ShowError(err, lf.controller.GetWindow())
				return
			}

			// Show school selector
			schoolNames := make([]string, len(lf.controller.GetClient().Schools))
			for i, school := range lf.controller.GetClient().Schools {
				schoolNames[i] = fmt.Sprintf("%s (%s)", school.SchoolName, school.Name)
			}
			lf.schoolSel.Options = schoolNames
			lf.schoolSel.Refresh()

			// Auto-select if only one school
			if len(lf.controller.GetClient().Schools) == 1 {
				lf.schoolSel.SetSelectedIndex(0)
			}
		})
	}()
}

// onSchoolSelected handles school selection and navigates to dashboard.
func (lf *LoginForm) onSchoolSelected(selected string) {
	if selected == "" {
		return
	}

	idx := -1
	for i, opt := range lf.schoolSel.Options {
		if opt == selected {
			idx = i
			break
		}
	}

	if idx < 0 || idx >= len(lf.controller.GetClient().Schools) {
		return
	}

	school := lf.controller.GetClient().Schools[idx]
	err := lf.controller.GetClient().SelectSchool(school.SchoolID)
	if err != nil {
		dialog.ShowError(err, lf.controller.GetWindow())
		return
	}

	// Navigate to dashboard
	if dash, ok := lf.controller.(interface{ ShowDashboard() }); ok {
		dash.ShowDashboard()
	}
}
