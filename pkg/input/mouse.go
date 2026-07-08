package input

import (
	"sync"
	"time"

	"github.com/bendahl/uinput"
)

const doubleClickDelay = 100 * time.Millisecond

// clickHoldTime is how long a button is held down during a click. uinput's own
// click sends the press and release back-to-back with no gap, which is too fast
// for MiSTer cores that poll the mouse (e.g. classic Mac) to reliably register,
// so we hold the button briefly like the keyboard does for key presses.
const clickHoldTime = 60 * time.Millisecond

// AbsMax is the maximum value of the virtual touchpad's axes. Absolute
// positions are given as per-mille (0-1000) of the screen and scaled into
// this range, so positioning is independent of the actual screen resolution.
const AbsMax = 1023

type Mouse struct {
	// Rel is a standard relative movement mouse device.
	Rel uinput.Mouse
	// Abs is a touchpad device used for absolute positioning.
	Abs uinput.TouchPad
	// mu serializes access to the shared uinput devices. It is a pointer so
	// value copies of Mouse (as handed to each HTTP/WebSocket handler) share
	// the same lock and can't interleave events on the devices.
	mu *sync.Mutex
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
	ms.mu = &sync.Mutex{}

	return ms, nil
}

func (m *Mouse) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	_ = m.Rel.Close()
	_ = m.Abs.Close()
}

// Move moves the mouse cursor relative to its current position.
func (m *Mouse) Move(x int32, y int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
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

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Abs.MoveTo(absX, absY)
}

// click presses a button, holds it briefly so the core can register it, then
// releases it. The caller must hold m.mu.
func click(press func() error, release func() error) error {
	if err := press(); err != nil {
		return err
	}
	time.Sleep(clickHoldTime)
	return release()
}

// leftClick performs a left click without taking the lock, for callers that
// already hold m.mu (e.g. DoubleClick).
func (m *Mouse) leftClick() error {
	return click(m.Rel.LeftPress, m.Rel.LeftRelease)
}

func (m *Mouse) LeftClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.leftClick()
}

func (m *Mouse) DoubleClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.leftClick(); err != nil {
		return err
	}
	time.Sleep(doubleClickDelay)
	return m.leftClick()
}

func (m *Mouse) RightClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return click(m.Rel.RightPress, m.Rel.RightRelease)
}

func (m *Mouse) MiddleClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return click(m.Rel.MiddlePress, m.Rel.MiddleRelease)
}

func (m *Mouse) LeftDown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Rel.LeftPress()
}

func (m *Mouse) LeftUp() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Rel.LeftRelease()
}

func (m *Mouse) RightDown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Rel.RightPress()
}

func (m *Mouse) RightUp() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Rel.RightRelease()
}

func (m *Mouse) MiddleDown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Rel.MiddlePress()
}

func (m *Mouse) MiddleUp() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Rel.MiddleRelease()
}

func (m *Mouse) Wheel(delta int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Rel.Wheel(false, delta)
}
