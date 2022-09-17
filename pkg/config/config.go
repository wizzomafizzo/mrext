package config

const SdFolder = "/media/fat"
const CoreConfigFolder = SdFolder + "/config"

const IndexName = "index"
const ActiveGameFile = "/tmp/ACTIVEGAME"

const CoreNameFile = "/tmp/CORENAME"
const CurrentPathFile = "/tmp/CURRENTPATH"
const StartPathFile = "/tmp/STARTPATH"
const FullPathFile = "/tmp/FULLPATH"

const CmdInterface = "/dev/MiSTer_cmd"

// TODO: this can't be hardcoded if we want dynamic arcade folders
const ArcadeCoresFolder = "/media/fat/_Arcade/cores"

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
