package config

// TODO: should this be hardcoded? how common is usb0 setup?
const SdFolder = "/media/fat"
const CoreConfigFolder = SdFolder + "/config"
const FontFolder = SdFolder + "/font"
const TempFolder = "/tmp"

const MenuConfigFile = CoreConfigFolder + "/MENU.CFG"

const MisterIniFile = SdFolder + "/MiSTer.ini"
const MisterIniFileAlt1 = SdFolder + "/MiSTer_alt_1.ini"
const MisterIniFileAlt2 = SdFolder + "/MiSTer_alt_2.ini"
const MisterIniFileAlt3 = SdFolder + "/MiSTer_alt_3.ini"

const StartupFile = SdFolder + "/linux/user-startup.sh"

const CoreNameFile = TempFolder + "/CORENAME"
const CurrentPathFile = TempFolder + "/CURRENTPATH"
const StartPathFile = TempFolder + "/STARTPATH"
const FullPathFile = TempFolder + "/FULLPATH"

const CoresRecentFile = CoreConfigFolder + "/cores_recent.cfg"

const MenuCore = "MENU"

const CmdInterface = "/dev/MiSTer_cmd"

// TODO: this can't be hardcoded if we want dynamic arcade folders
const ArcadeCoresFolder = "/media/fat/_Arcade/cores"

// TODO: this is technically not the order mister checks, it does games folders
//
//	second, but right now this is simpler for checking for a prefix
var GamesFolders = []string{
	"/media/usb0/games",
	"/media/usb0",
	"/media/usb1/games",
	"/media/usb1",
	"/media/usb2/games",
	"/media/usb2",
	"/media/usb3/games",
	"/media/usb3",
	"/media/usb4/games",
	"/media/usb4",
	"/media/usb5/games",
	"/media/usb5",
	"/media/fat/cifs/games",
	"/media/fat/cifs",
	"/media/fat/games",
	"/media/fat",
}
