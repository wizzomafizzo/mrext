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

// TODO: consider just hardcoding this in the paths list
const GamesFolderSubfolder = "games"

var GamesFolders = []string{
	"/media/fat",
	"/media/usb0",
	"/media/usb1",
	"/media/usb2",
	"/media/usb3",
	"/media/usb4",
	"/media/usb5",
	"/media/fat/cifs",
}
