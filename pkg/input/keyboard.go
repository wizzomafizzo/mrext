package input

import (
	"github.com/bendahl/uinput"
	"time"
)

// TODO: needs delays on connect if not running as a daemon

const sleepTime = 40 * time.Millisecond

type Keyboard struct {
	Device uinput.Keyboard
}

func NewKeyboard() (Keyboard, error) {
	var kb Keyboard

	vk, err := uinput.CreateKeyboard("/dev/uinput", []byte("mrext"))
	if err != nil {
		return kb, err
	}

	kb.Device = vk

	return kb, nil
}

func (k *Keyboard) Close() {
	k.Device.Close()
}

func (k *Keyboard) Press(key int) {
	k.Device.KeyDown(key)
	time.Sleep(sleepTime)
	k.Device.KeyUp(key)
}

func (k *Keyboard) Combo(keys ...int) {
	for _, key := range keys {
		k.Device.KeyDown(key)
	}
	time.Sleep(sleepTime)
	for _, key := range keys {
		k.Device.KeyUp(key)
	}
}

func (k *Keyboard) VolumeUp() {
	k.Press(uinput.KeyVolumeup)
}

func (k *Keyboard) VolumeDown() {
	k.Press(uinput.KeyVolumedown)
}

func (k *Keyboard) VolumeMute() {
	k.Press(uinput.KeyMute)
}

func (k *Keyboard) Menu() {
	k.Press(uinput.KeyEsc)
}

func (k *Keyboard) Back() {
	k.Press(uinput.KeyBackspace)
}

func (k *Keyboard) Confirm() {
	k.Press(uinput.KeyEnter)
}

func (k *Keyboard) Cancel() {
	k.Menu()
}

func (k *Keyboard) Up() {
	k.Press(uinput.KeyUp)
}

func (k *Keyboard) Down() {
	k.Press(uinput.KeyDown)
}

func (k *Keyboard) Left() {
	k.Press(uinput.KeyLeft)
}

func (k *Keyboard) Right() {
	k.Press(uinput.KeyRight)
}

func (k *Keyboard) Osd() {
	k.Press(uinput.KeyF12)
}

func (k *Keyboard) CoreSelect() {
	k.Combo(uinput.KeyLeftalt, uinput.KeyF12)
}

func (k *Keyboard) Screenshot() {
	// TODO: for the life of me, I can't make the regular Win+PrtScn combo
	//       work. this is a hardcoded alternate combo which *does* work,
	//       but it's disabled on PS/2 keyboard or in PS/2 mode or something
	k.Combo(uinput.KeyLeftalt, uinput.KeyScrolllock)
}

func (k *Keyboard) RawScreenshot() {
	// TODO: see above
	k.Combo(uinput.KeyLeftalt, uinput.KeyLeftshift, uinput.KeyScrolllock)
}

func (k *Keyboard) User() {
	k.Combo(uinput.KeyLeftctrl, uinput.KeyLeftalt, uinput.KeyRightalt)
}

func (k *Keyboard) Reset() {
	k.Combo(uinput.KeyLeftshift, uinput.KeyLeftctrl, uinput.KeyLeftalt, uinput.KeyRightalt)
}

func (k *Keyboard) PairBluetooth() {
	k.Press(uinput.KeyF11)
}

func (k *Keyboard) ChangeBackground() {
	k.Press(uinput.KeyF1)
}

func (k *Keyboard) ToggleCoreDates() {
	k.Press(uinput.KeyF2)
}

func (k *Keyboard) Console() {
	k.Press(uinput.KeyF9)
}

func (k *Keyboard) ExitConsole() {
	k.Press(uinput.KeyF12)
}

func (k *Keyboard) ComputerOsd() {
	k.Combo(uinput.KeyLeftmeta, uinput.KeyF12)
}
