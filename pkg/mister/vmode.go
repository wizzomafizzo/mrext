package mister

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"os"
)

const (
	VideoModeScaleFull    = "1"
	VideoModeScaleHalf    = "2"
	VideoModeScaleThird   = "3"
	VideoModeScaleQuarter = "4"
	VideoModeFormatRGB32  = "18888"
	VideoModeFormatRGB15  = "11555"
	VideoModeFormatRGB16  = "1565"
	VideoModeFormatBGR32  = "08888"
	VideoModeFormatBGR15  = "01555"
	VideoModeFormatBGR16  = "0565"
	VideoModeFormatIDX8   = "08"
	ResCountPath          = "/sys/module/MiSTer_fb/parameters/res_count"
)

// fb_cmd0 = scaled = fb_cmd0 $fmt $rb $scale
// fb_cmd1 = exact = fb_cmd1 $fmt $rb $width $height

// in vmode script, checks for rescount contents at start, sets mode,
// then polls until it's the same value (up to 5 times)

func SetVideoMode(width int, height int) error {
	if _, err := os.Stat(config.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer func(cmd *os.File) {
		_ = cmd.Close()
	}(cmd)

	cmdStr := fmt.Sprintf(
		"%s %d %d %d",
		VideoModeFormatRGB32[1:],
		VideoModeFormatRGB32[0],
		width,
		height,
	)

	fmt.Println(cmdStr)

	_, err = cmd.WriteString(cmdStr)
	if err != nil {
		return err
	}

	return nil
}
