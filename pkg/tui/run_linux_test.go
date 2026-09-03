//go:build linux

package tui

import (
	"errors"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestTryRunAppRetriesOnMiSTerTTY(t *testing.T) {
	initialRun := errors.New("no controlling terminal")
	originalRun := runApplication
	originalTTY := newDevTTY
	originalScreen := newTTYScreen
	t.Cleanup(func() {
		runApplication = originalRun
		newDevTTY = originalTTY
		newTTYScreen = originalScreen
	})

	for _, tt := range []struct {
		name      string
		runScript string
		wantTTY   string
	}{
		{name: "MiSTer script", wantTTY: "/dev/tty2"},
		{name: "ZapScript", runScript: "2", wantTTY: "/dev/tty4"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ZAPAROO_RUN_SCRIPT", tt.runScript)
			runs := 0
			runApplication = func(_ *tview.Application) error {
				runs++
				if runs == 1 {
					return initialRun
				}
				return nil
			}
			ttyPath := ""
			newDevTTY = func(path string) (tcell.Tty, error) {
				ttyPath = path
				return nil, nil //nolint:nilnil // Test stub does not exercise returned TTY.
			}
			newTTYScreen = func(tcell.Tty) (tcell.Screen, error) {
				return nil, nil //nolint:nilnil // Test stub does not exercise returned screen.
			}
			builds := 0
			builder := func() (*tview.Application, error) {
				builds++
				return tview.NewApplication(), nil
			}

			if err := tryRunApp(tview.NewApplication(), builder); err != nil {
				t.Fatal(err)
			}
			if runs != 2 || builds != 1 {
				t.Fatalf("runs=%d builds=%d", runs, builds)
			}
			if ttyPath != tt.wantTTY {
				t.Fatalf("retry TTY = %q, want %q", ttyPath, tt.wantTTY)
			}
		})
	}
}
