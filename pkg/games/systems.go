package games

type MglParams struct {
	Delay    int
	FileType string
	Index    int
}

type FileType struct {
	Extensions []string
	Mgl        *MglParams
}

type System struct {
	Id        string
	Folder    string
	Rbf       string
	FileTypes []FileType
}

var SYSTEMS = map[string]System{
	"Amiga": {
		// TODO: amiga will require a custom scan function
		Id:     "Amiga",
		Folder: "Amiga",
		Rbf:    "_Computer/Minimig",
		FileTypes: []FileType{
			{
				Extensions: nil,
				Mgl:        nil,
			},
		},
	},
	"Arcade": {
		Id:     "Arcade",
		Folder: "_Arcade",
		Rbf:    "",
		FileTypes: []FileType{
			{
				Extensions: []string{".mra"},
				Mgl:        nil,
			},
		},
	},
	// TODO: support for multiple folders?
	// TODO: could cut down on work scanning by folder rather than system
	"Atari2600": {
		Id:     "Atari2600",
		Folder: "ATARI7800",
		Rbf:    "_Console/Atari7800",
		FileTypes: []FileType{
			{
				Extensions: []string{".a26"},
				Mgl:        nil,
			},
		},
	},
	"Atari5200": {
		Id:     "Atari5200",
		Folder: "ATARI5200",
		Rbf:    "_Console/Atari5200",
		FileTypes: []FileType{
			{
				Extensions: []string{".a52", ".car"},
				// TODO: this probably supports mgl launching
				Mgl: nil,
			},
		},
	},
	"Atari7800": {
		Id:     "Atari7800",
		Folder: "ATARI7800",
		Rbf:    "_Console/Atari7800",
		FileTypes: []FileType{
			{
				Extensions: []string{".a78", ".bin"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"AtariLynx": {
		Id:     "AtariLynx",
		Folder: "AtariLynx",
		Rbf:    "_Console/AtariLynx",
		FileTypes: []FileType{
			{
				Extensions: []string{".lnx"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"C64": {
		Id:     "C64",
		Folder: "C64",
		Rbf:    "_Computer/C64",
		FileTypes: []FileType{
			{
				Extensions: []string{".prg", ".crt", ".reu", ".tap"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"ColecoVision": {
		Id:     "ColecoVision",
		Folder: "Coleco",
		Rbf:    "_Console/ColecoVision",
		FileTypes: []FileType{
			{
				Extensions: []string{".col", ".bin", ".rom", ".sg"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"FDS": {
		Id:     "FDS",
		Folder: "NES",
		Rbf:    "_Console/NES",
		FileTypes: []FileType{
			{
				Extensions: []string{".fds"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"Gameboy": {
		Id:     "Gameboy",
		Folder: "GAMEBOY",
		Rbf:    "_Console/Gameboy",
		FileTypes: []FileType{
			{
				Extensions: []string{".gb"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"GameboyColor": {
		Id:     "GameboyColor",
		Folder: "GAMEBOY",
		Rbf:    "_Console/Gameboy",
		FileTypes: []FileType{
			{
				Extensions: []string{".gbc"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"Gameboy2P": {
		Id:     "Gameboy2P",
		Folder: "GAMEBOY2P",
		Rbf:    "_Console/Gameboy2P",
		FileTypes: []FileType{
			{
				Extensions: []string{".gb", ".gbc"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"GameGear": {
		Id:     "GameGear",
		Folder: "SMS",
		Rbf:    "_Console/SMS",
		FileTypes: []FileType{
			{
				Extensions: []string{".gg"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    2,
				},
			},
		},
	},
	"GBA": {
		Id:     "GBA",
		Folder: "GBA",
		Rbf:    "_Console/GBA",
		FileTypes: []FileType{
			{
				Extensions: []string{".gba"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"GBA2P": {
		Id:     "GBA2P",
		Folder: "GBA2P",
		Rbf:    "_Console/GBA2P",
		FileTypes: []FileType{
			{
				Extensions: []string{".gba"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"Genesis": {
		Id:     "Genesis",
		Folder: "Genesis",
		Rbf:    "_Console/Genesis",
		FileTypes: []FileType{
			{
				Extensions: []string{".bin", ".gen", ".md"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"MegaCD": {
		Id:     "MegaCD",
		Folder: "MegaCD",
		Rbf:    "_Console/MegaCD",
		FileTypes: []FileType{
			{
				Extensions: []string{".cue", ".chd"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "s",
					Index:    0,
				},
			},
		},
	},
	"NeoGeo": {
		Id:     "NeoGeo",
		Folder: "NEOGEO",
		Rbf:    "_Console/NeoGeo",
		FileTypes: []FileType{
			{
				Extensions: []string{".neo"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
			{
				Extensions: []string{".iso"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "s",
					Index:    1,
				},
			},
		},
	},
	"NES": {
		Id:     "NES",
		Folder: "NES",
		Rbf:    "_Console/NES",
		FileTypes: []FileType{
			{
				Extensions: []string{".nes", ".nsf"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"PSX": {
		Id:     "PSX",
		Folder: "PSX",
		Rbf:    "_Console/PSX",
		FileTypes: []FileType{
			{
				Extensions: []string{".cue", ".chd"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "s",
					Index:    1,
				},
			},
		},
	},
	"Sega32X": {
		Id:     "Sega32X",
		Folder: "S32X",
		Rbf:    "_Console/S32X",
		FileTypes: []FileType{
			{
				Extensions: []string{".32x"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"SuperGameboy": {
		Id:     "SuperGameboy",
		Folder: "SGB",
		Rbf:    "_Console/SGB",
		FileTypes: []FileType{
			{
				Extensions: []string{".gb", ".gbc"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"MasterSystem": {
		Id:     "MasterSystem",
		Folder: "SMS",
		Rbf:    "_Console/SMS",
		FileTypes: []FileType{
			{
				Extensions: []string{".sms", ".sg"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"SNES": {
		Id:     "SNES",
		Folder: "SNES",
		Rbf:    "_Console/SNES",
		FileTypes: []FileType{
			{
				Extensions: []string{".smc", ".sfc"},
				Mgl: &MglParams{
					Delay:    2,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"TurboGraphx16": {
		Id:     "TurboGraphx16",
		Folder: "TGFX16",
		Rbf:    "_Console/TGFX16",
		FileTypes: []FileType{
			{
				Extensions: []string{".bin", ".pce"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
			{
				Extensions: []string{".sgx"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"TurboGraphx16CD": {
		Id:     "TurboGraphx16CD",
		Folder: "TGFX16-CD",
		Rbf:    "_Console/TGFX16",
		FileTypes: []FileType{
			{
				Extensions: []string{".cue", ".chd"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "s",
					Index:    0,
				},
			},
		},
	},
	"Vectrex": {
		Id:     "Vectrex",
		Folder: "VECTREX",
		Rbf:    "_Console/Vectrex",
		FileTypes: []FileType{
			{
				Extensions: []string{".ovr", ".vec", ".bin", ".rom"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
	"WonderSwan": {
		Id:     "WonderSwan",
		Folder: "WonderSwan",
		Rbf:    "_Console/WonderSwan",
		FileTypes: []FileType{
			{
				Extensions: []string{".ws", ".wsc"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    1,
				},
			},
		},
	},
}

var GAMES_FOLDERS = []string{
	"/media/fat",
	"/media/usb0",
	"/media/usb1",
	"/media/usb2",
	"/media/usb3",
	"/media/usb4",
	"/media/usb5",
	"/media/fat/cifs",
	".",
}
