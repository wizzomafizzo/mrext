package games

type mglParams struct {
	delay    int
	fileType string
	index    int
}

type fileType struct {
	extensions []string
	mgl        *mglParams
}

type System struct {
	folder    string
	rbf       string
	fileTypes []fileType
}

var SYSTEMS = map[string]System{
	"Arcade": {
		folder: "_Arcade",
		rbf:    "",
		fileTypes: []fileType{
			{
				extensions: []string{".mra"},
				mgl:        nil,
			},
		},
	},
	"Atari7800": {
		folder: "ATARI7800",
		rbf:    "_Console/Atari7800",
		fileTypes: []fileType{
			{
				extensions: []string{".a78", ".a26", ".bin"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
		},
	},
	"AtariLynx": {
		folder: "AtariLynx",
		rbf:    "_Console/AtariLynx",
		fileTypes: []fileType{
			{
				extensions: []string{".lnx"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"C64": {
		folder: "C64",
		rbf:    "_Computer/C64",
		fileTypes: []fileType{
			{
				extensions: []string{".prg", ".crt", ".reu", ".tap"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
		},
	},
	"ColecoVision": {
		folder: "Coleco",
		rbf:    "_Console/ColecoVision",
		fileTypes: []fileType{
			{
				extensions: []string{".col", ".bin", ".rom", ".sg"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"FDS": {
		folder: "NES",
		rbf:    "_Console/NES",
		fileTypes: []fileType{
			{
				extensions: []string{".fds"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"Gameboy": {
		folder: "GAMEBOY",
		rbf:    "_Console/Gameboy",
		fileTypes: []fileType{
			{
				extensions: []string{".gb", ".gbc"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
		},
	},
	"Gameboy2P": {
		folder: "GAMEBOY2P",
		rbf:    "_Console/Gameboy2P",
		fileTypes: []fileType{
			{
				extensions: []string{".gb", ".gbc"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
		},
	},
	"GBA": {
		folder: "GBA",
		rbf:    "_Console/GBA",
		fileTypes: []fileType{
			{
				extensions: []string{".gba"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"GBA2P": {
		folder: "GBA2P",
		rbf:    "_Console/GBA2P",
		fileTypes: []fileType{
			{
				extensions: []string{".gba"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"Genesis": {
		folder: "Genesis",
		rbf:    "_Console/Genesis",
		fileTypes: []fileType{
			{
				extensions: []string{".bin", ".gen", ".md"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"MegaCD": {
		folder: "MegaCD",
		rbf:    "_Console/MegaCD",
		fileTypes: []fileType{
			{
				extensions: []string{".cue", ".chd"},
				mgl: &mglParams{
					delay:    1,
					fileType: "s",
					index:    0,
				},
			},
		},
	},
	"NeoGeo": {
		folder: "NEOGEO",
		rbf:    "_Console/NeoGeo",
		fileTypes: []fileType{
			{
				extensions: []string{".neo"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
			{
				extensions: []string{".iso"},
				mgl: &mglParams{
					delay:    1,
					fileType: "s",
					index:    1,
				},
			},
		},
	},
	"NES": {
		folder: "NES",
		rbf:    "_Console/NES",
		fileTypes: []fileType{
			{
				extensions: []string{".nes", ".fds", ".nsf"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"PSX": {
		folder: "PSX",
		rbf:    "_Console/PSX",
		fileTypes: []fileType{
			{
				extensions: []string{".cue", ".chd"},
				mgl: &mglParams{
					delay:    1,
					fileType: "s",
					index:    1,
				},
			},
		},
	},
	"Sega32X": {
		folder: "S32X",
		rbf:    "_Console/S32X",
		fileTypes: []fileType{
			{
				extensions: []string{".32x"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"SuperGameboy": {
		folder: "SGB",
		rbf:    "_Console/SGB",
		fileTypes: []fileType{
			{
				extensions: []string{".gb", ".gbc"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
		},
	},
	"MasterSystem": {
		folder: "SMS",
		rbf:    "_Console/SMS",
		fileTypes: []fileType{
			{
				extensions: []string{".sms", ".sg"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
			{
				extensions: []string{".gg"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    2,
				},
			},
		},
	},
	"SNES": {
		folder: "SNES",
		rbf:    "_Console/SNES",
		fileTypes: []fileType{
			{
				extensions: []string{".smc", ".sfc"},
				mgl: &mglParams{
					delay:    2,
					fileType: "f",
					index:    0,
				},
			},
		},
	},
	"TurboGraphx16": {
		folder: "TGFX16",
		rbf:    "_Console/TGFX16",
		fileTypes: []fileType{
			{
				extensions: []string{".bin", ".pce"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    0,
				},
			},
			{
				extensions: []string{".sgx"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
		},
	},
	"TurboGraphx16CD": {
		folder: "TGFX16-CD",
		rbf:    "_Console/TGFX16",
		fileTypes: []fileType{
			{
				extensions: []string{".cue", ".chd"},
				mgl: &mglParams{
					delay:    1,
					fileType: "s",
					index:    0,
				},
			},
		},
	},
	"Vectrex": {
		folder: "VECTREX",
		rbf:    "_Console/Vectrex",
		fileTypes: []fileType{
			{
				extensions: []string{".ovr", ".vec", ".bin", ".rom"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
				},
			},
		},
	},
	"WonderSwan": {
		folder: "WonderSwan",
		rbf:    "_Console/WonderSwan",
		fileTypes: []fileType{
			{
				extensions: []string{".ws", ".wsc"},
				mgl: &mglParams{
					delay:    1,
					fileType: "f",
					index:    1,
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
