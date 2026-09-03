//go:build linux

package tui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	runApplication = func(app *tview.Application) error { return app.Run() }
	newDevTTY      = tcell.NewDevTtyFromDev
	newTTYScreen   = tcell.NewTerminfoScreenFromTty
)

func tryRunApp(app *tview.Application, builder func() (*tview.Application, error)) error {
	if err := runApplication(app); err != nil {
		retryApp, buildErr := builder()
		if buildErr != nil {
			return buildErr
		}

		ttyPath := "/dev/tty2"
		if os.Getenv("ZAPAROO_RUN_SCRIPT") == "2" {
			ttyPath = "/dev/tty4"
		}
		tty, ttyErr := newDevTTY(ttyPath)
		if ttyErr != nil {
			return fmt.Errorf("failed to create tty from device %s: %w", ttyPath, ttyErr)
		}
		screen, screenErr := newTTYScreen(tty)
		if screenErr != nil {
			return fmt.Errorf("failed to create screen from tty: %w", screenErr)
		}
		retryApp.SetScreen(screen)
		if retryErr := runApplication(retryApp); retryErr != nil {
			return fmt.Errorf("failed to run TUI application: %w", retryErr)
		}
	}
	return nil
}
