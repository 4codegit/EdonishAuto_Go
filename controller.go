package main

import (
	"github.com/4codegit/edonish-auto/client"
	"github.com/4codegit/edonish-auto/ui"

	"fyne.io/fyne/v2"
)

// AppController manages the application lifecycle.
type AppController struct {
	app       fyne.App
	window    fyne.Window
	client    *client.EdonishClient
	login     *ui.LoginForm
	dashboard *ui.Dashboard
}

// NewAppController creates a new application controller.
func NewAppController(a fyne.App) *AppController {
	return &AppController{
		app:    a,
		window: a.NewWindow("eDonish Auto"),
	}
}

// Run starts the application.
func (c *AppController) Run() {
	c.client = client.NewEdonishClient()

	c.login = ui.NewLoginForm(c)
	c.window.SetContent(c.login.Container())
	c.window.Resize(fyne.NewSize(900, 650))
	c.window.Canvas().Focus(c.login.GetLoginEntry())
	c.window.ShowAndRun()
}

// GetClient implements ui.Controller.
func (c *AppController) GetClient() *client.EdonishClient {
	return c.client
}

// GetWindow implements ui.Controller.
func (c *AppController) GetWindow() fyne.Window {
	return c.window
}

// Logout implements ui.Controller.
func (c *AppController) Logout() {
	c.client = client.NewEdonishClient()
	c.login = ui.NewLoginForm(c)
	c.window.SetContent(c.login.Container())
	c.window.Canvas().Focus(c.login.GetLoginEntry())
}

// ShowDashboard switches from login screen to the main dashboard.
func (c *AppController) ShowDashboard() {
	c.dashboard = ui.NewDashboard(c)
	c.window.SetContent(c.dashboard.Container())
	c.window.Resize(fyne.NewSize(1200, 750))
}
