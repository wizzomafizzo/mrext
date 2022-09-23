package config

// TODO: should this be hardcoded? how common is usb0 setup?
const SdFolder = "/media/fat"
const CoreConfigFolder = SdFolder + "/config"

const IndexName = "index"
const ActiveGameFile = "/tmp/ACTIVEGAME"

const MisterIniFile = SdFolder + "/MiSTer.ini"
const MisterIniFile_Alt1 = SdFolder + "/MiSTer_alt_1.ini"
const MisterIniFile_Alt2 = SdFolder + "/MiSTer_alt_2.ini"
const MisterIniFile_Alt3 = SdFolder + "/MiSTer_alt_3.ini"

const StartupFile = SdFolder + "/linux/user-startup.sh"

const CoreNameFile = "/tmp/CORENAME"
const CurrentPathFile = "/tmp/CURRENTPATH"
const StartPathFile = "/tmp/STARTPATH"
const FullPathFile = "/tmp/FULLPATH"

const CmdInterface = "/dev/MiSTer_cmd"

// TODO: this can't be hardcoded if we want dynamic arcade folders
const ArcadeCoresFolder = "/media/fat/_Arcade/cores"

var GamesFolders = []string{
	"/media/usb0",
	"/media/usb0/games",
	"/media/usb1",
	"/media/usb1/games",
	"/media/usb2",
	"/media/usb2/games",
	"/media/usb3",
	"/media/usb3/games",
	"/media/usb4",
	"/media/usb4/games",
	"/media/usb5",
	"/media/usb5/games",
	"/media/fat/cifs",
	"/media/fat/cifs/games",
	"/media/fat/games",
}
