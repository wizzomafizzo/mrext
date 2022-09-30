package games

type MglParams struct {
	Delay  int
	Method string
	Index  int
}

type Slot struct {
	Label string
	Exts  []string
	Mgl   *MglParams
}

type AltRbfOpts map[string][]string

type System struct {
	Id     string
	Name   string
	Alias  []string
	Folder []string
	Rbf    string
	AltRbf AltRbfOpts
	Slots  []Slot
}

const (
	AltRbfLLAPI   = "LLAPI"
	AltRbfYC      = "YC"
	AltRbfDualRAM = "DualRAM"
)

// First in list takes precendence for simple attributes in case there's a
// conflict in the future.
var CoreGroups = map[string][]System{
	"Atari7800": {Systems["Atari7800"], Systems["Atari2600"]},
	"Coleco":    {Systems["Coleco"], Systems["SG1000"]},
	"Gameboy":   {Systems["Gameboy"], Systems["GameboyColor"]},
	"NES":       {Systems["NES"], Systems["NESMusic"], Systems["FDS"]},
	"SMS": {Systems["MasterSystem"], Systems["GameGear"], System{
		Slots: []Slot{
			{
				Exts: []string{".sg"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	}},
	"SNES":   {Systems["SNES"], Systems["SNESMusic"]},
	"TGFX16": {Systems["TurboGrafx16"], Systems["SuperGrafx"]},
}

// FIXME: launch game > launch new game same system > not working? should it?
// TODO: setname attribute
// TODO: alternate cores
// TODO: alternate arcade folders
// TODO: custom scan function
// TODO: custom launch function
// TODO: support for multiple folders (think about symlink support here, check for dupes)
// TODO: could cut down on work scanning by folder rather than system
// TODO: support globbing on extensions

var Systems = map[string]System{
	// Consoles
	"AdventureVision": {
		Id:     "AdventureVision",
		Name:   "Adventure Vision",
		Alias:  []string{"AVision"},
		Folder: []string{"AVision"},
		Rbf:    "_Console/AdventureVision",
		Slots: []Slot{
			{
				Label: "Game",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Arcadia": {
		Id:     "Arcadia",
		Name:   "Arcadia 2001",
		Folder: []string{"Arcadia"},
		Rbf:    "_Console/Arcadia",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Astrocade": {
		Id:     "Astrocade",
		Name:   "Bally Astrocade",
		Folder: []string{"Astrocade"},
		Rbf:    "_Console/Astrocade",
		Slots: []Slot{
			{
				Exts: []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Atari2600": {
		Id:     "Atari2600",
		Name:   "Atari 2600",
		Folder: []string{"ATARI7800"},
		Rbf:    "_Console/Atari7800",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Atari7800_LLAPI"},
			AltRbfYC:    []string{"Atari7800YC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".a26"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Atari5200": {
		Id:     "Atari5200",
		Name:   "Atari 5200",
		Folder: []string{"ATARI5200"},
		Rbf:    "_Console/Atari5200",
		AltRbf: AltRbfOpts{
			AltRbfYC: []string{"Atari5200YC"},
		},
		Slots: []Slot{
			{
				Label: "Cart",
				Exts:  []string{".car", ".a52", ".bin", ".rom"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"Atari7800": {
		Id:     "Atari7800",
		Name:   "Atari 7800",
		Folder: []string{"ATARI7800"},
		Rbf:    "_Console/Atari7800",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Atari7800_LLAPI"},
			AltRbfYC:    []string{"Atari7800YC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".a78", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "BIOS",
				Exts:  []string{".rom", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"AtariLynx": {
		Id:     "AtariLynx",
		Name:   "Atari Lynx",
		Folder: []string{"AtariLynx"},
		Rbf:    "_Console/AtariLynx",
		Slots: []Slot{
			{
				Exts: []string{".lnx"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	// TODO: AY-3-8500
	//       Doesn't appear to have roms even though it has a folder.
	// TODO: C2650
	//       Not in official repos, think it comes with update_all.
	//       https://github.com/Grabulosaure/C2650_MiSTer
	"CasioPV1000": {
		Id:     "CasioPV1000",
		Name:   "Casio PV-1000",
		Alias:  []string{"Casio_PV-1000"},
		Folder: []string{"Casio_PV-1000"},
		Rbf:    "_Console/Casio_PV-1000",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"ChannelF": {
		Id:     "ChannelF",
		Name:   "Channel F",
		Folder: []string{"ChannelF"},
		Rbf:    "_Console/ChannelF",
		Slots: []Slot{
			{
				Exts: []string{".rom", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"ColecoVision": {
		Id:     "ColecoVision",
		Name:   "ColecoVision",
		Alias:  []string{"Coleco"},
		Folder: []string{"Coleco"},
		Rbf:    "_Console/ColecoVision",
		AltRbf: AltRbfOpts{
			AltRbfYC: []string{"ColecoVisionYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".col", ".bin", ".rom"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"CreatiVision": {
		Id:     "CreatiVision",
		Name:   "VTech CreatiVision",
		Folder: []string{"CreatiVision"},
		Rbf:    "_Console/CreatiVision",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".rom", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "Bios",
				Exts:  []string{".rom", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
			{
				Label: "BASIC",
				Exts:  []string{".bas"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  3,
				},
			},
		},
	},
	// TODO: EpochGalaxy2
	//       Has a folder and mount entry but commented as "remove".
	"FDS": {
		Id:     "FDS",
		Name:   "Famicom Disk System",
		Alias:  []string{"FamicomDiskSystem"},
		Folder: []string{"NES"},
		Rbf:    "_Console/NES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"NES_LLAPI"},
			AltRbfYC:    []string{"NESYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".fds"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "FDS BIOS",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"Gamate": {
		Id:     "Gamate",
		Name:   "Gamate",
		Folder: []string{"Gamate"},
		Rbf:    "_Console/Gamate",
		Slots: []Slot{
			{
				Exts: []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Gameboy": {
		Id:     "Gameboy",
		Name:   "Gameboy",
		Alias:  []string{"GB"},
		Folder: []string{"GAMEBOY"},
		Rbf:    "_Console/Gameboy",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Gameboy_LLAPI"},
			AltRbfYC:    []string{"GameboyYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".gb"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"GameboyColor": {
		Id:     "GameboyColor",
		Name:   "Gameboy Color",
		Alias:  []string{"GBC"},
		Folder: []string{"GAMEBOY"},
		Rbf:    "_Console/Gameboy",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Gameboy_LLAPI"},
			AltRbfYC:    []string{"GameboyYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".gbc"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Gameboy2P": {
		// TODO: Split 2P core into GB and GBC?
		Id:     "Gameboy2P",
		Name:   "Gameboy (2 Player)",
		Folder: []string{"GAMEBOY2P"},
		Rbf:    "_Console/Gameboy2P",
		Slots: []Slot{
			{
				Exts: []string{".gb", ".gbc"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"GameGear": {
		Id:     "GameGear",
		Name:   "Game Gear",
		Alias:  []string{"GG"},
		Folder: []string{"SMS"},
		Rbf:    "_Console/SMS",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SMS_LLAPI"},
			AltRbfYC:    []string{"SMSYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".gg"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"GameNWatch": {
		Id:     "GameNWatch",
		Name:   "Game & Watch",
		Folder: []string{"GameNWatch"},
		Rbf:    "_Console/GnW",
		Slots: []Slot{
			{
				Exts: []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"GBA": {
		Id:     "GBA",
		Name:   "Gameboy Advance",
		Alias:  []string{"GameboyAdvance"},
		Folder: []string{"GBA"},
		Rbf:    "_Console/GBA",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"GBA_LLAPI"},
			AltRbfYC:    []string{"GBAYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".gba"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"GBA2P": {
		Id:     "GBA2P",
		Name:   "Gameboy Advance (2 Player)",
		Folder: []string{"GBA2P"},
		Rbf:    "_Console/GBA2P",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"GBA2P_LLAPI"},
		},
		Slots: []Slot{
			{
				Exts: []string{".gba"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Genesis": {
		Id:     "Genesis",
		Name:   "Genesis",
		Alias:  []string{"MegaDrive"},
		Folder: []string{"Genesis"},
		Rbf:    "_Console/Genesis",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Genesis_LLAPI"},
			AltRbfYC:    []string{"GenesisYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".bin", ".gen", ".md"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Intellivision": {
		Id:     "Intellivision",
		Name:   "Intellivision",
		Folder: []string{"Intellivision"},
		Rbf:    "_Console/Intellivision",
		Slots: []Slot{
			{
				Exts: []string{".rom", ".int", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	// TODO: Jaguar
	"MasterSystem": {
		Id:     "MasterSystem",
		Name:   "Master System",
		Alias:  []string{"SMS"},
		Folder: []string{"SMS"},
		Rbf:    "_Console/SMS",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SMS_LLAPI"},
			AltRbfYC:    []string{"SMSYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".sms"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"MegaCD": {
		Id:     "MegaCD",
		Name:   "Sega CD",
		Alias:  []string{"SegaCD"},
		Folder: []string{"MegaCD"},
		Rbf:    "_Console/MegaCD",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"MegaCD_LLAPI"},
			AltRbfYC:    []string{"MegaCDYC"},
		},
		Slots: []Slot{
			{
				Label: "Disk",
				Exts:  []string{".cue", ".chd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"NeoGeo": {
		Id:     "NeoGeo",
		Name:   "Neo Geo MVS/AES",
		Folder: []string{"NEOGEO"},
		Rbf:    "_Console/NeoGeo",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"NeoGeo_LLAPI"},
			AltRbfYC:    []string{"NeoGeoYC"},
		},
		Slots: []Slot{
			{
				// TODO: This also has some special handling re: zip files (darksoft pack).
				// Exts: []strings{".*"}
				Label: "ROM set",
				Exts:  []string{".neo"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "CD Image",
				Exts:  []string{".iso", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"NES": {
		Id:     "NES",
		Name:   "NES",
		Folder: []string{"NES"},
		Rbf:    "_Console/NES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"NES_LLAPI"},
			AltRbfYC:    []string{"NESYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".nes"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"NESMusic": {
		Id:     "NESMusic",
		Name:   "NES Music",
		Folder: []string{"NES"},
		Rbf:    "_Console/NES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"NES_LLAPI"},
			AltRbfYC:    []string{"NESYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".nsf"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Odyssey2": {
		Id:     "Odyssey2",
		Name:   "Magnavox Odyssey2",
		Folder: []string{"ODYSSEY2"},
		Rbf:    "_Console/Odyssey2",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "XROM",
				Exts:  []string{".rom"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"PokemonMini": {
		Id:     "PokemonMini",
		Name:   "Pokemon Mini",
		Folder: []string{"PokemonMini"},
		Rbf:    "_Console/PokemonMini",
		Slots: []Slot{
			{
				Label: "ROM",
				Exts:  []string{".min"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"PSX": {
		Id:     "PSX",
		Name:   "Playstation",
		Alias:  []string{"Playstation", "PS1"},
		Folder: []string{"PSX"},
		Rbf:    "_Console/PSX",
		AltRbf: AltRbfOpts{
			AltRbfYC:      []string{"PSXYC"},
			AltRbfDualRAM: []string{"PSX_DualSDRAM"},
		},
		Slots: []Slot{
			{
				Label: "CD",
				Exts:  []string{".cue", ".chd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "Exe",
				Exts:  []string{".exe"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Sega32X": {
		Id:     "Sega32X",
		Name:   "Genesis 32X",
		Alias:  []string{"S32X", "32X"},
		Folder: []string{"S32X"},
		Rbf:    "_Console/S32X",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"S32X_LLAPI"},
			AltRbfYC:    []string{"S32XYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".32x"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"SG1000": {
		Id:     "SG1000",
		Name:   "SG-1000",
		Folder: []string{"SG1000", "Coleco", "SMS"},
		Rbf:    "_Console/ColecoVision",
		AltRbf: AltRbfOpts{
			AltRbfYC: []string{"ColecoVisionYC"},
		},
		Slots: []Slot{
			{
				Label: "SG-1000",
				Exts:  []string{".sg"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"SuperGameboy": {
		Id:     "SuperGameboy",
		Name:   "Super Gameboy",
		Alias:  []string{"SGB"},
		Folder: []string{"SGB"},
		Rbf:    "_Console/SGB",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SGB_LLAPI"},
			AltRbfYC:    []string{"SGBYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".gb", ".gbc"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"SuperVision": {
		Id:     "SuperVision",
		Name:   "SuperVision",
		Folder: []string{"SuperVision"},
		Rbf:    "_Console/SuperVision",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin", ".sv"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	// TODO: Saturn
	"SNES": {
		Id:     "SNES",
		Name:   "SNES",
		Alias:  []string{"SuperNintendo"},
		Folder: []string{"SNES"},
		Rbf:    "_Console/SNES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SNES_LLAPI"},
			AltRbfYC:    []string{"SNESYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".sfc", ".smc", ".bin", ".bs"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  0,
				},
			},
		},
	},
	"SNESMusic": {
		Id:     "SNESMusic",
		Name:   "SNES Music",
		Folder: []string{"SNES"},
		Rbf:    "_Console/SNES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SNES_LLAPI"},
			AltRbfYC:    []string{"SNESYC"},
		},
		Slots: []Slot{
			{
				Exts: []string{".spc"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"SuperGrafx": {
		Id:     "SuperGrafx",
		Name:   "SuperGrafx",
		Folder: []string{"TGFX16"},
		Rbf:    "_Console/TurboGrafx16",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"TurboGrafx16_LLAPI"},
			AltRbfYC:    []string{"TurboGrafx16YC"},
		},
		Slots: []Slot{
			{
				Label: "SuperGrafx",
				Exts:  []string{".sgx"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"TurboGraphx16": {
		Id:     "TurboGraphx16",
		Name:   "TurboGraphx-16",
		Alias:  []string{"TGFX16", "PCEngine"},
		Folder: []string{"TGFX16"},
		Rbf:    "_Console/TurboGrafx16",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"TurboGrafx16_LLAPI"},
			AltRbfYC:    []string{"TurboGrafx16YC"},
		},
		Slots: []Slot{
			{
				Label: "TurboGrafx",
				Exts:  []string{".bin", ".pce"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  0,
				},
			},
		},
	},
	"TurboGraphx16CD": {
		Id:     "TurboGraphx16CD",
		Name:   "TurboGraphx-16 CD",
		Alias:  []string{"TGFX16-CD", "PCEngineCD"},
		Folder: []string{"TGFX16-CD"},
		Rbf:    "_Console/TurboGrafx16",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"TurboGrafx16_LLAPI"},
			AltRbfYC:    []string{"TurboGrafx16YC"},
		},
		Slots: []Slot{
			{
				Label: "CD",
				Exts:  []string{".cue", ".chd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"VC4000": {
		Id:     "VC4000",
		Name:   "VC4000",
		Folder: []string{"VC4000"},
		Rbf:    "_Console/VC4000",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Vectrex": {
		Id:     "Vectrex",
		Name:   "Vectrex",
		Folder: []string{"VECTREX"},
		Rbf:    "_Console/Vectrex",
		Slots: []Slot{
			{
				Exts: []string{".vec", ".bin", ".rom"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "Overlay",
				Exts:  []string{".ovr"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"WonderSwan": {
		Id:     "WonderSwan",
		Name:   "WonderSwan",
		Folder: []string{"WonderSwan"},
		Rbf:    "_Console/WonderSwan",
		Slots: []Slot{
			{
				Label: "ROM",
				Exts:  []string{".ws", ".wsc"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	// Computers
	"AcornAtom": {
		Id:     "AcornAtom",
		Name:   "Atom",
		Folder: []string{"AcornAtom"},
		Rbf:    "_Computer/AcornAtom",
		Slots: []Slot{
			{
				Exts: []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"AcornElectron": {
		Id:     "AcornElectron",
		Name:   "Electron",
		Folder: []string{"AcornElectron"},
		Rbf:    "_Computer/AcornElectron",
		Slots: []Slot{
			{
				Exts: []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"AliceMC10": {
		Id:     "AliceMC10",
		Name:   "Tandy MC-10",
		Folder: []string{"AliceMC10"},
		Rbf:    "_Computer/AliceMC10",
		Slots: []Slot{
			{
				Label: "Tape",
				Exts:  []string{".c10"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	// TODO: Altair8800
	//       Has a folder but roms are built in.
	"Amiga": {
		// TODO: New versions of MegaAGS image support launching individual games,
		//       will need support for custom scan and launch functions for a core.
		// TODO: This core has 2 .adf drives and 4 .hdf drives. No CONF_STR.
		Id:     "Amiga",
		Name:   "Amiga",
		Folder: []string{"Amiga"},
		Alias:  []string{"Minimig"},
		Rbf:    "_Computer/Minimig",
		Slots:  nil,
	},
	"Amstrad": {
		Id:     "Amstrad",
		Name:   "Amstrad CPC",
		Folder: []string{"Amstrad"},
		Rbf:    "_Computer/Amstrad",
		Slots: []Slot{
			{
				Label: "A:",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "B:",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "Expansion",
				Exts:  []string{".e??"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  3,
				},
			},
			{
				Label: "Tape",
				Exts:  []string{".cdt"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  4,
				},
			},
		},
	},
	"AmstradPCW": {
		Id:     "AmstradPCW",
		Name:   "Amstrad PCW",
		Alias:  []string{"Amstrad-PCW"},
		Folder: []string{"Amstrad PCW"},
		Rbf:    "_Computer/Amstrad-PCW",
		Slots: []Slot{
			{
				Label: "A:",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "B:",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"ao486": {
		Id:     "ao486",
		Name:   "PC (486SX)",
		Folder: []string{"AO486"},
		Rbf:    "_Computer/ao486",
		Slots: []Slot{
			{
				Label: "Floppy A:",
				Exts:  []string{".img", ".ima", ".vfd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "Floppy B:",
				Exts:  []string{".img", ".ima", ".vfd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "IDE 0-0",
				Exts:  []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  2,
				},
			},
			{
				Label: "IDE 0-1",
				Exts:  []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  3,
				},
			},
			{
				Label: "IDE 1-0",
				Exts:  []string{".vhd", ".iso", ".cue", ".chd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  4,
				},
			},
			{
				Label: "IDE 1-1",
				Exts:  []string{".vhd", ".iso", ".cue", ".chd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  5,
				},
			},
		},
	},
	"Apogee": {
		Id:     "Apogee",
		Name:   "Apogee BK-01",
		Folder: []string{"APOGEE"},
		Rbf:    "_Computer/Apogee",
		Slots: []Slot{
			{
				Exts: []string{".rka", ".rkr", ".gam"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"AppleI": {
		Id:     "AppleI",
		Name:   "Apple I",
		Alias:  []string{"Apple-I"},
		Folder: []string{"Apple-I"},
		Rbf:    "_Computer/Apple-I",
		Slots: []Slot{
			{
				Label: "ASCII",
				Exts:  []string{".txt"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"AppleII": {
		Id:     "AppleII",
		Name:   "Apple IIe",
		Alias:  []string{"Apple-II"},
		Folder: []string{"Apple-II"},
		Rbf:    "_Computer/Apple-II",
		Slots: []Slot{
			{
				Exts: []string{".nib", ".dsk", ".do", ".po"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Exts: []string{".hdv"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"Aquarius": {
		Id:     "Aquarius",
		Name:   "Mattel Aquarius",
		Folder: []string{"AQUARIUS"},
		Rbf:    "_Computer/Aquarius",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "Tape",
				Exts:  []string{".caq"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	// TODO: Archie
	//       Can't see anything in CONF_STR. Mentioned explicitly in menu.
	"Atari800": {
		Id:     "Atari800",
		Name:   "Atari 800XL",
		Folder: []string{"ATARI800"},
		Rbf:    "_Computer/Atari800",
		Slots: []Slot{
			{
				Label: "D1",
				Exts:  []string{".atr", ".xex", ".xfd", ".atx"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "D2",
				Exts:  []string{".atr", ".xex", ".xfd", ".atx"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "Cartridge",
				Exts:  []string{".car", ".rom", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  2,
				},
			},
		},
	},
	// TODO: AtariST
	//       CONF_STR does not have any information about the file types.
	"BBCMicro": {
		Id:     "BBCMicro",
		Name:   "BBC Micro/Master",
		Folder: []string{"BBCMicro"},
		Rbf:    "_Computer/BBCMicro",
		Slots: []Slot{
			{
				Exts: []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Exts: []string{".ssd", ".dsd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Exts: []string{".ssd", ".dsd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  2,
				},
			},
		},
	},
	"BK0011M": {
		Id:     "BK0011M",
		Name:   "BK0011M",
		Folder: []string{"BK0011M"},
		Rbf:    "_Computer/BK0011M",
		Slots: []Slot{
			{
				Exts: []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "FDD(A)",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "FDD(B)",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  2,
				},
			},
			{
				Label: "HDD",
				Exts:  []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"C16": {
		Id:     "C16",
		Name:   "Commodore 16",
		Folder: []string{"C16"},
		Rbf:    "_Computer/C16",
		Slots: []Slot{
			{
				Label: "#8",
				Exts:  []string{".d64", ".g64"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "#9",
				Exts:  []string{".d64", ".g64"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				// TODO: This has a hidden option with only .prg and .tap.
				Exts: []string{".prg", ".tap", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"C64": {
		Id:     "C64",
		Name:   "Commodore 64",
		Folder: []string{"C64"},
		Rbf:    "_Computer/C64",
		AltRbf: AltRbfOpts{
			AltRbfYC: []string{"C64YC"},
		},
		Slots: []Slot{
			{
				Label: "#8",
				Exts:  []string{".d64", ".g64", ".t64", ".d81"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "#9",
				Exts:  []string{".d64", ".g64", ".t64", ".d81"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Exts: []string{".prg", ".crt", ".reu", ".tap"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"CasioPV2000": {
		Id:     "CasioPV2000",
		Name:   "Casio PV-2000",
		Alias:  []string{"Casio_PV-2000"},
		Folder: []string{"Casio_PV-2000"},
		Rbf:    "_Computer/Casio_PV-2000",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"CoCo2": {
		Id:     "CoCo2",
		Name:   "TRS-80 CoCo 2",
		Folder: []string{"CoCo2"},
		Rbf:    "_Computer/CoCo2",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".rom", ".ccc"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "Disk Drive 0",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "Disk Drive 1",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "Disk Drive 2",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  2,
				},
			},
			{
				Label: "Disk Drive 3",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  3,
				},
			},
			{
				Label: "Cassette",
				Exts:  []string{".cas"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	// TODO: CoCo3
	//       This core has several menu states for different combinations of
	//       files to load. Unsure if MGL is compatible with it.
	// TODO: ColecoAdam
	//       Unsure what folder this uses. Coleco?
	"EDSAC": {
		Id:     "EDSAC",
		Name:   "EDSAC",
		Folder: []string{"EDSAC"},
		Rbf:    "_Computer/EDSAC",
		Slots: []Slot{
			{
				Label: "Tape",
				Exts:  []string{".tap"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Galaksija": {
		Id:     "Galaksija",
		Name:   "Galaksija",
		Folder: []string{"Galaksija"},
		Rbf:    "_Computer/Galaksija",
		Slots: []Slot{
			{
				Exts: []string{".tap"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Interact": {
		Id:     "Interact",
		Name:   "Interact",
		Folder: []string{"Interact"},
		Rbf:    "_Computer/Interact",
		Slots: []Slot{
			{
				Label: "Tape",
				Exts:  []string{".cin", ".k7"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Jupiter": {
		Id:     "Jupiter",
		Name:   "Jupiter Ace",
		Folder: []string{"Jupiter"},
		Rbf:    "_Computer/Jupiter",
		Slots: []Slot{
			{
				Exts: []string{".ace"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Laser": {
		Id:     "Laser",
		Name:   "Laser 350/500/700",
		Alias:  []string{"Laser310"},
		Folder: []string{"Laser"},
		Rbf:    "_Computer/Laser310",
		Slots: []Slot{
			{
				Label: "VZ Image",
				Exts:  []string{".vz"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Lynx48": {
		Id:     "Lynx48",
		Name:   "Lynx 48/96K",
		Folder: []string{"Lynx48"},
		Rbf:    "_Computer/Lynx48",
		Slots: []Slot{
			{
				Label: "Cassette",
				Exts:  []string{".tap"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"MacPlus": {
		Id:     "MacPlus",
		Name:   "Macintosh Plus",
		Folder: []string{"MACPLUS"},
		Rbf:    "_Computer/MacPlus",
		Slots: []Slot{
			{
				Label: "Pri Floppy",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "Sec Floppy",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
			{
				Label: "SCSI-6",
				Exts:  []string{".img", ".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "SCSI-5",
				Exts:  []string{".img", ".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"MSX": {
		Id:     "MSX",
		Name:   "MSX",
		Folder: []string{"MSX"},
		Rbf:    "_Computer/MSX",
		Slots: []Slot{
			{
				Exts: []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"MultiComp": {
		Id:     "MultiComp",
		Name:   "MultiComp",
		Folder: []string{"MultiComp"},
		Rbf:    "_Computer/MultiComp",
		Slots: []Slot{
			{
				Exts: []string{".img"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	// TODO: OndraSPO186
	//       Nothing listed in CONF_STR but docs do mention loading files.
	"Orao": {
		Id:     "Orao",
		Name:   "Orao",
		Folder: []string{"ORAO"},
		Rbf:    "_Computer/ORAO",
		Slots: []Slot{
			{
				Exts: []string{".tap"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Oric": {
		Id:     "Oric",
		Name:   "Oric",
		Folder: []string{"Oric"},
		Rbf:    "_Computer/Oric",
		Slots: []Slot{
			{
				Label: "Drive A:",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	// TODO: PC88
	//       Nothing listed in CONF_STR.
	"PCXT": {
		Id:     "PCXT",
		Name:   "PC/XT",
		Folder: []string{"PCXT"},
		Rbf:    "_Computer/PCXT",
		Slots: []Slot{
			{
				Label: "FDD Image",
				Exts:  []string{".img", ".ima"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "HDD Image",
				Exts:  []string{".img"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"PDP1": {
		Id:     "PDP1",
		Name:   "PDP-1",
		Folder: []string{"PDP1"},
		Rbf:    "_Computer/PDP1",
		Slots: []Slot{
			{
				Exts: []string{".pdp", ".rim", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"PET2001": {
		Id:     "PET2001",
		Name:   "Commodore PET 2001",
		Folder: []string{"PET2001"},
		Rbf:    "_Computer/PET2001",
		Slots: []Slot{
			{
				Exts: []string{".prg", ".tap"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"PMD85": {
		Id:     "PMD85",
		Name:   "PMD 85-2A",
		Folder: []string{"PMD85"},
		Rbf:    "_Computer/PMD85",
		Slots: []Slot{
			{
				Label: "ROM Pack",
				Exts:  []string{".rmm"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"QL": {
		Id:     "QL",
		Name:   "Sinclair QL",
		Folder: []string{"QL"},
		Rbf:    "_Computer/QL",
		Slots: []Slot{
			{
				Label: "HD Image",
				Exts:  []string{".win"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "MDV Image",
				Exts:  []string{".mdv"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"RX78": {
		Id:     "RX78",
		Name:   "RX-78 Gundam",
		Folder: []string{"RX78"},
		Rbf:    "_Computer/RX78",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"SAMCoupe": {
		Id:     "SAMCoupe",
		Name:   "SAM Coupe",
		Folder: []string{"SAMCOUPE"},
		Rbf:    "_Computer/SAMCoupe",
		Slots: []Slot{
			{
				Label: "Drive 1",
				Exts:  []string{".dsk", ".mgt", ".img"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "Drive 2",
				Exts:  []string{".dsk", ".mgt", ".img"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	// TODO: SharpMZ
	//       Nothing listed in CONF_STR.
	"SordM5": {
		Id:     "SordM5",
		Name:   "M5",
		Alias:  []string{"Sord M5"},
		Folder: []string{"Sord M5"},
		Rbf:    "_Computer/SordM5",
		Slots: []Slot{
			{
				Label: "ROM",
				Exts:  []string{".bin", ".rom"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "Tape",
				Exts:  []string{".cas"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"Specialist": {
		Id:     "Specialist",
		Name:   "Specialist/MX",
		Alias:  []string{"SPMX"},
		Folder: []string{"SPMX"},
		Rbf:    "_Computer/Specialist",
		Slots: []Slot{
			{
				Label: "Tape",
				Exts:  []string{".rks"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  0,
				},
			},
			{
				Label: "Disk",
				Exts:  []string{".odi"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"SVI328": {
		Id:     "SVI328",
		Name:   "SV-328",
		Folder: []string{"SVI328"},
		Rbf:    "_Computer/Svi328",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin", ".rom"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "CAS File",
				Exts:  []string{".cas"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
		},
	},
	"TatungEinstein": {
		Id:     "TatungEinstein",
		Name:   "Tatung Einstein",
		Folder: []string{"TatungEinstein"},
		Rbf:    "_Computer/TatungEinstein",
		Slots: []Slot{
			{
				Label: "Disk 0",
				Exts:  []string{".dsk"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"TI994A": {
		Id:     "TI994A",
		Name:   "TI-99/4A",
		Alias:  []string{"TI-99_4A"},
		Folder: []string{"TI-99_4A"},
		Rbf:    "_Computer/Ti994a",
		Slots: []Slot{
			{
				Label: "Full Cart",
				Exts:  []string{".m99", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "ROM Cart",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
			{
				Label: "GROM Cart",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  3,
				},
			},
			// TODO: Also 3 .dsk entries, inactive on first load.
		},
	},
	"TomyTutor": {
		Id:     "TomyTutor",
		Name:   "Tutor",
		Folder: []string{"TomyTutor"},
		Rbf:    "_Computer/TomyTutor",
		Slots: []Slot{
			{
				Label: "Cartridge",
				Exts:  []string{".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
			{
				Label: "Tape Image",
				Exts:  []string{".cas"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"TRS80": {
		Id:     "TRS80",
		Name:   "TRS-80",
		Folder: []string{"TRS-80"},
		Rbf:    "_Computer/TRS-80",
		Slots: []Slot{
			{
				Label: "Disk 0",
				Exts:  []string{".dsk", ".jvi"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "Disk 1",
				Exts:  []string{".dsk", ".jvi"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "Program",
				Exts:  []string{".cmd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
			{
				Label: "Cassette",
				Exts:  []string{".cas"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"TSConf": {
		Id:     "TSConf",
		Name:   "TS-Config",
		Folder: []string{"TSConf"},
		Rbf:    "_Computer/TSConf",
		Slots: []Slot{
			{
				Label: "Virtual SD",
				Exts:  []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"UK101": {
		Id:     "UK101",
		Name:   "UK101",
		Folder: []string{"UK101"},
		Rbf:    "_Computer/UK101",
		Slots: []Slot{
			{
				Label: "ASCII",
				Exts:  []string{".txt", ".bas", ".lod"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Vector06C": {
		Id:     "Vector06C",
		Name:   "Vector-06C",
		Alias:  []string{"Vector06"},
		Folder: []string{"VECTOR06"},
		Rbf:    "_Computer/Vector-06C",
		Slots: []Slot{
			{
				Exts: []string{".rom", ".com", ".c00", ".edd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			{
				Label: "Disk A",
				Exts:  []string{".fdd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "Disk B",
				Exts:  []string{".fdd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"VIC20": {
		Id:     "VIC20",
		Name:   "Commodore VIC-20",
		Folder: []string{"VIC20"},
		Rbf:    "_Computer/VIC20",
		Slots: []Slot{
			{
				Label: "#8",
				Exts:  []string{".d64", ".g64"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "#9",
				Exts:  []string{".d64", ".g64"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Exts: []string{".prg", ".crt", ".ct?", ".tap"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"X68000": {
		Id:     "X68000",
		Name:   "X68000",
		Folder: []string{"X68000"},
		Rbf:    "_Computer/X68000",
		Slots: []Slot{
			{
				Label: "FDD0",
				Exts:  []string{".d88"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "FDD1",
				Exts:  []string{".d88"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "SASI Hard Disk",
				Exts:  []string{".hdf"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  2,
				},
			},
			{
				Label: "RAM",
				Exts:  []string{".ram"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  3,
				},
			},
		},
	},
	// TODO: zx48
	//       https://github.com/Kyp069/zx48-MiSTer
	"ZX81": {
		Id:     "ZX81",
		Name:   "TS-1500",
		Folder: []string{"ZX81"},
		Rbf:    "_Computer/ZX81",
		Slots: []Slot{
			{
				Label: "Tape",
				Exts:  []string{".0", ".p"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"ZXSpectrum": {
		Id:     "ZXSpectrum",
		Name:   "ZX Spectrum",
		Alias:  []string{"Spectrum"},
		Folder: []string{"Spectrum"},
		Rbf:    "_Computer/ZX-Spectrum",
		Slots: []Slot{
			{
				Label: "Disk",
				Exts:  []string{".trd", ".img", ".dsk", ".mgt"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "Tape",
				Exts:  []string{".tap", ".csw", ".tzx"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  2,
				},
			},
			{
				Label: "Snapshot",
				Exts:  []string{".z80", ".sna"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  4,
				},
			},
			{
				Label: "DivMMC",
				Exts:  []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"ZXNext": {
		Id:     "ZXNext",
		Name:   "ZX Spectrum Next",
		Folder: []string{"ZXNext"},
		Rbf:    "_Computer/ZXNext",
		Slots: []Slot{
			{
				Label: "C:",
				Exts:  []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
			{
				Label: "D:",
				Exts:  []string{".vhd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
			{
				Label: "Tape",
				Exts:  []string{".tzx", ".csw"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	// Other
	"Arcade": {
		Id:     "Arcade",
		Name:   "Arcade",
		Folder: []string{"_Arcade"},
		Slots: []Slot{
			{
				Exts: []string{".mra"},
				Mgl:  nil,
			},
		},
	},
	"Arduboy": {
		Id:     "Arduboy",
		Name:   "Arduboy",
		Folder: []string{"Arduboy"},
		Rbf:    "_Other/Arduboy",
		Slots: []Slot{
			{
				Exts: []string{".bin", ".hex"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  0,
				},
			},
		},
	},
	"Chip8": {
		Id:     "Chip8",
		Name:   "CHIP-8",
		Folder: []string{"Chip8"},
		Rbf:    "_Other/Chip8",
		Slots: []Slot{
			{
				Exts: []string{".ch8"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	// TODO: Life
	//       Has loadable files, but no folder?
	// TODO: ScummVM
	//       Requires a custom scan and launch function.
	// TODO: SuperJacob
	//       A custom computer?
	//       https://github.com/dave18/MiSTER-SuperJacob
	// TODO: TomyScramble
	//       Has loadable files and a folder but is marked as "remove"?
}
