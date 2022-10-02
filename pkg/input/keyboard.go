package input

import (
	"github.com/bendahl/uinput"
)

// TODO: needs delays on connect and between presses

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

func (k *Keyboard) VolumeUp() {
	k.Device.KeyPress(uinput.KeyVolumeup)
}

func (k *Keyboard) VolumeDown() {
	k.Device.KeyPress(uinput.KeyVolumedown)
}

func (k *Keyboard) VolumeMute() {
	k.Device.KeyPress(uinput.KeyMute)
}

func (k *Keyboard) Menu() {
	k.Device.KeyPress(uinput.KeyEsc)
}

func (k *Keyboard) Back() {
	k.Device.KeyPress(uinput.KeyBackspace)
}

func (k *Keyboard) Confirm() {
	k.Device.KeyPress(uinput.KeyEnter)
}

func (k *Keyboard) Cancel() {
	k.Menu()
}

func (k *Keyboard) Up() {
	k.Device.KeyPress(uinput.KeyUp)
}

func (k *Keyboard) Down() {
	k.Device.KeyPress(uinput.KeyDown)
}

func (k *Keyboard) Left() {
	k.Device.KeyPress(uinput.KeyLeft)
}

func (k *Keyboard) Right() {
	k.Device.KeyPress(uinput.KeyRight)
}

func (k *Keyboard) Osd() {
	k.Device.KeyPress(uinput.KeyF12)
}

func (k *Keyboard) CoreSelect() {
	k.Device.KeyDown(uinput.KeyLeftalt)
	k.Osd()
	k.Device.KeyUp(uinput.KeyLeftalt)
}

func (k *Keyboard) Screenshot() {
	k.Device.KeyDown(uinput.KeyLeftmeta)
	k.Device.KeyPress(uinput.KeyPrint)
	k.Device.KeyUp(uinput.KeyLeftmeta)
}

func (k *Keyboard) RawScreenshot() {
	k.Device.KeyDown(uinput.KeyLeftshift)
	k.Device.KeyDown(uinput.KeyLeftmeta)
	k.Device.KeyPress(uinput.KeyPrint)
	k.Device.KeyUp(uinput.KeyLeftmeta)
	k.Device.KeyUp(uinput.KeyLeftshift)
}

func (k *Keyboard) User() {
	k.Device.KeyDown(uinput.KeyLeftctrl)
	k.Device.KeyDown(uinput.KeyLeftalt)
	k.Device.KeyPress(uinput.KeyRightalt)
	k.Device.KeyUp(uinput.KeyLeftalt)
	k.Device.KeyUp(uinput.KeyLeftctrl)
}

func (k *Keyboard) Reset() {
	k.Device.KeyDown(uinput.KeyLeftshift)
	k.Device.KeyDown(uinput.KeyLeftctrl)
	k.Device.KeyDown(uinput.KeyLeftalt)
	k.Device.KeyPress(uinput.KeyRightalt)
	k.Device.KeyUp(uinput.KeyLeftalt)
	k.Device.KeyUp(uinput.KeyLeftctrl)
	k.Device.KeyUp(uinput.KeyLeftshift)
}

func (k *Keyboard) PairBluetooth() {
	k.Device.KeyPress(uinput.KeyF11)
}

func (k *Keyboard) ChangeBackground() {
	k.Device.KeyPress(uinput.KeyF1)
}

func (k *Keyboard) ToggleCoreDates() {
	k.Device.KeyPress(uinput.KeyF2)
}

func (k *Keyboard) Console() {
	k.Device.KeyPress(uinput.KeyF3)
}

func (k *Keyboard) ComputerOsd() {
	k.Device.KeyDown(uinput.KeyLeftmeta)
	k.Device.KeyPress(uinput.KeyF12)
	k.Device.KeyUp(uinput.KeyLeftmeta)
}
