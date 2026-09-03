//go:build !linux

package tui

import "github.com/rivo/tview"

func tryRunApp(app *tview.Application, _ func() (*tview.Application, error)) error {
	return app.Run()
}
