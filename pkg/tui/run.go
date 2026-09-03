package tui

import "github.com/rivo/tview"

// BuildAndRetry builds and runs an application, retrying through MiSTer's
// alternate TTY path when the normal terminal cannot initialize.
func BuildAndRetry(builder func() (*tview.Application, error)) error {
	app, err := builder()
	if err != nil {
		return err
	}
	return tryRunApp(app, builder)
}
