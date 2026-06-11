package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// LoginPage holds the login screen UI components.
type LoginPage struct {
	app         *App
	loginEntry  *widget.Entry
	passEntry   *widget.Entry
	rememberChk *widget.Check
	statusLabel *widget.Label
	loginBtn    *widget.Button
}

// NewLoginPage creates a new login page.
func NewLoginPage(app *App) *LoginPage {
	return &LoginPage{app: app}
}

// Build creates the login view and returns the root container.
func (p *LoginPage) Build() fyne.CanvasObject {
	p.loginEntry = widget.NewEntry()
	p.loginEntry.SetPlaceHolder("Логин (ID)")

	p.passEntry = widget.NewPasswordEntry()
	p.passEntry.SetPlaceHolder("Пароль")
	p.passEntry.OnSubmitted = func(_ string) { p.doLogin() }

	p.rememberChk = widget.NewCheck("Запомнить меня", nil)

	p.statusLabel = widget.NewLabel("")

	p.loginBtn = widget.NewButton("Войти", func() {
		p.doLogin()
	})
	p.loginBtn.Importance = widget.HighImportance

	// Logo and title
	icon := canvas.NewImageFromResource(nil)
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(64, 64))

	title := widget.NewLabelWithStyle("eDonish Auto", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel("Автоматизация электронного журнала")
	subtitle.Alignment = fyne.TextAlignCenter

	shortcut := widget.NewLabel("Enter для быстрого входа")
	shortcut.Alignment = fyne.TextAlignCenter
	shortcut.TextStyle = fyne.TextStyle{Italic: true}

	form := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(
			container.NewVBox(
				container.NewCenter(title),
				container.NewCenter(subtitle),
				widget.NewSeparator(),
				p.loginEntry,
				p.passEntry,
				p.rememberChk,
				p.loginBtn,
				container.NewCenter(shortcut),
				p.statusLabel,
			),
		),
		layout.NewSpacer(),
	)

	return form
}

// LoadSession loads saved session data into the form.
func (p *LoginPage) LoadSession() {
	loginID, password, remember, _ := p.app.LoadSessionData()
	if loginID != "" {
		p.loginEntry.SetText(loginID)
	}
	if remember && password != "" {
		p.passEntry.SetText(password)
		p.rememberChk.SetChecked(true)
	}
}

// doLogin handles the login button press.
func (p *LoginPage) doLogin() {
	loginID := p.loginEntry.Text
	password := p.passEntry.Text

	if loginID == "" || password == "" {
		p.statusLabel.SetText("Введите логин и пароль")
		return
	}

	p.loginBtn.Disable()
	p.loginBtn.SetText("Вход...")
	p.statusLabel.SetText("Подключение...")

	// Save session
	p.app.SaveSession(loginID, password, p.rememberChk.Checked)

	go func() {
		userInfo, err := p.app.apiClient.Login(loginID, password)
		if err != nil {
			p.loginBtn.Enable()
			p.loginBtn.SetText("Войти")
			p.statusLabel.SetText(err.Error())
			p.statusLabel.Refresh()
			p.loginBtn.Refresh()
			return
		}

		// Apply saved school selection if available
		_, _, _, savedSchoolID := p.app.LoadSessionData()
		if savedSchoolID > 0 && p.app.apiClient.HasMultipleSchools() {
			p.app.apiClient.SetSchool(savedSchoolID)
			p.app.LogMessage(fmt.Sprintf("Восстановлена школа ID: %d", savedSchoolID), "info")
		}

		p.app.onLoginSuccess(userInfo)
	}()
}
