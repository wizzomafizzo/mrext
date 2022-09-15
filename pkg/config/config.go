package config

const SdRoot = "/media/fat"

const IndexName = "index"

const CoreName = "/tmp/CORENAME"
const CurrentPath = "/tmp/CURRENTPATH"
const StartPath = "/tmp/STARTPATH"
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
