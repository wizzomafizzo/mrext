package input

import (
	"time"

	"github.com/bendahl/uinput"
)

const doubleClickDelay = 100 * time.Millisecond

// AbsMax is the maximum value of the virtual touchpad's axes. Absolute
// positions are given as per-mille (0-1000) of the screen and scaled into
// this range, so positioning is independent of the actual screen resolution.
const AbsMax = 1023

type Mouse struct {
	// Rel is a standard relative movement mouse device.
	Rel uinput.Mouse
	// Abs is a touchpad device used for absolute positioning.
	Abs uinput.TouchPad
}

func NewMouse() (Mouse, error) {
	var ms Mouse

	rel, err := uinput.CreateMouse("/dev/uinput", []byte("mrext-mouse"))
	if err != nil {
		return ms, err
	}

	abs, err := uinput.CreateTouchPad(
		"/dev/uinput", []byte("mrext-touchpad"),
		0, AbsMax, 0, AbsMax,
	)
	if err != nil {
		_ = rel.Close()
		return ms, err
	}

	ms.Rel = rel
	ms.Abs = abs

	return ms, nil
}

func (m *Mouse) Close() {
	_ = m.Rel.Close()
	_ = m.Abs.Close()
}

// Move moves the mouse cursor relative to its current position.
func (m *Mouse) Move(x int32, y int32) error {
	return m.Rel.Move(x, y)
}

// MoveToPermille moves the mouse cursor to an absolute position given as
// per-mille (0-1000) of the screen along each axis. This is independent of the
// screen resolution: the caller sends where on the screen to point as a
// fraction, and it maps onto the touchpad's absolute range.
func (m *Mouse) MoveToPermille(x int, y int) error {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 1000 {
			return 1000
		}
		return v
	}

	absX := int32(clamp(x) * AbsMax / 1000)
	absY := int32(clamp(y) * AbsMax / 1000)

	return m.Abs.MoveTo(absX, absY)
}

func (m *Mouse) LeftClick() error {
	return m.Rel.LeftClick()
}

func (m *Mouse) DoubleClick() error {
	err := m.Rel.LeftClick()
	if err != nil {
		return err
	}
	time.Sleep(doubleClickDelay)
	return m.Rel.LeftClick()
}

func (m *Mouse) RightClick() error {
	return m.Rel.RightClick()
}

func (m *Mouse) MiddleClick() error {
	return m.Rel.MiddleClick()
}

func (m *Mouse) LeftDown() error {
	return m.Rel.LeftPress()
}

func (m *Mouse) LeftUp() error {
	return m.Rel.LeftRelease()
}

func (m *Mouse) RightDown() error {
	return m.Rel.RightPress()
}

func (m *Mouse) RightUp() error {
	return m.Rel.RightRelease()
}

func (m *Mouse) MiddleDown() error {
	return m.Rel.MiddlePress()
}

func (m *Mouse) MiddleUp() error {
	return m.Rel.MiddleRelease()
}

func (m *Mouse) Wheel(delta int32) error {
	return m.Rel.Wheel(false, delta)
}
