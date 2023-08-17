package games

const (
	CategoryArcade              = "Arcade"
	CategoryConsole             = "Console"
	CategoryComputer            = "Computer"
	CategoryHandheld            = "Handheld"
	CategoryOther               = "Other"
	ManufacturerEntex           = "Entex"
	ManufacturerEmerson         = "Emerson"
	ManufacturerMattel          = "Mattel"
	ManufacturerBally           = "Bally"
	ManufacturerAtari           = "Atari"
	ManufacturerColeco          = "Coleco"
	ManufacturerSega            = "Sega"
	ManufacturerNintendo        = "Nintendo"
	ManufacturerNEC             = "NEC"
	ManufacturerSNK             = "SNK"
	ManufacturerBandai          = "Bandai"
	ManufacturerVTech           = "VTech"
	ManufacturerCasio           = "Casio"
	ManufacturerWatara          = "Watara"
	ManufacturerMagnavox        = "Magnavox"
	ManufacturerFairchild       = "Fairchild"
	ManufacturerGCE             = "GCE"
	ManufacturerBitCorp         = "Bit Corporation"
	ManufacturerCommodore       = "Commodore"
	ManufacturerAmstrad         = "Amstrad"
	ManufacturerAcorn           = "Acorn"
	ManufacturerApple           = "Apple"
	ManufacturerBenesse         = "Benesse"
	ManufacturerSony            = "Sony"
	ManufacturerInterton        = "Interton"
	ManufacturerTandy           = "Tandy"
	ManufacturerIBM             = "IBM"
	ManufacturerApogee          = "Apogee"
	ManufacturerElektronika     = "Elektronika"
	ManufacturerCambridge       = "Cambridge"
	ManufacturerInteract        = "Interact"
	ManufacturerJupiter         = "Jupiter"
	ManufacturerVideoTechnology = "Video Technology"
	ManufacturerMicrosoft       = "Microsoft"
)

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

type System struct {
	Id           string
	Name         string // US
	Category     string
	ReleaseDate  string // US
	Manufacturer string
	Alias        []string
	SetName      string
	Folder       []string
	Rbf          string
	Slots        []Slot
}

// CoreGroups is a list of common MiSTer aliases that map back to a system.
// First in list takes precendence for simple attributes in case there's a
// conflict in the future.
var CoreGroups = map[string][]System{
	"Atari7800": {Systems["Atari7800"], Systems["Atari2600"]},
	"Coleco":    {Systems["ColecoVision"], Systems["SG1000"]},
	"Gameboy":   {Systems["Gameboy"], Systems["GameboyColor"]},
	"NES":       {Systems["NES"], Systems["NESMusic"], Systems["FDS"]},
	"SMS": {Systems["MasterSystem"], Systems["GameGear"], System{
		Name: "SG-1000",
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
// TODO: alternate cores (user core override)
// TODO: alternate arcade folders
// TODO: custom scan function
// TODO: custom launch function
// TODO: support globbing on extensions

var Systems = map[string]System{
	// Consoles
	"AdventureVision": {
		Id:           "AdventureVision",
		Name:         "Adventure Vision",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerEntex,
		ReleaseDate:  "1982-01-01",
		Alias:        []string{"AVision"},
		Folder:       []string{"AVision"},
		Rbf:          "_Console/AdventureVision",
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
		Id:           "Arcadia",
		Name:         "Arcadia 2001",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerEmerson,
		ReleaseDate:  "1982-01-01",
		Folder:       []string{"Arcadia"},
		Rbf:          "_Console/Arcadia",
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
		Id:           "Astrocade",
		Name:         "Bally Astrocade",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerBally,
		ReleaseDate:  "1978-04-01",
		Folder:       []string{"Astrocade"},
		Rbf:          "_Console/Astrocade",
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
		Id:           "Atari2600",
		Name:         "Atari 2600",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerAtari,
		ReleaseDate:  "1977-09-11",
		Folder:       []string{"ATARI7800", "Atari2600"},
		SetName:      "Atari2600",
		Rbf:          "_Console/Atari7800",
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
		Id:           "Atari5200",
		Name:         "Atari 5200",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerAtari,
		ReleaseDate:  "1982-11-01",
		Folder:       []string{"ATARI5200"},
		Rbf:          "_Console/Atari5200",
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
		Id:           "Atari7800",
		Name:         "Atari 7800",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerAtari,
		ReleaseDate:  "1986-05-01",
		Folder:       []string{"ATARI7800"},
		Rbf:          "_Console/Atari7800",
		Slots: []Slot{
			{
				Exts: []string{".a78", ".bin"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			//{
			//	Label: "BIOS",
			//	Exts:  []string{".rom", ".bin"},
			//	Mgl: &MglParams{
			//		Delay:  1,
			//		Method: "f",
			//		Index:  2,
			//	},
			//},
		},
	},
	"AtariLynx": {
		Id:           "AtariLynx",
		Name:         "Atari Lynx",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerAtari,
		ReleaseDate:  "1989-09-01",
		Folder:       []string{"AtariLynx"},
		Rbf:          "_Console/AtariLynx",
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
		Id:           "CasioPV1000",
		Name:         "Casio PV-1000",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerCasio,
		ReleaseDate:  "1983-10-01",
		Alias:        []string{"Casio_PV-1000"},
		Folder:       []string{"Casio_PV-1000"},
		Rbf:          "_Console/Casio_PV-1000",
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
		Id:           "ChannelF",
		Name:         "Channel F",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerFairchild,
		ReleaseDate:  "1976-11-01",
		Folder:       []string{"ChannelF"},
		Rbf:          "_Console/ChannelF",
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
		Id:           "ColecoVision",
		Name:         "ColecoVision",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerColeco,
		ReleaseDate:  "1982-08-01",
		Alias:        []string{"Coleco"},
		Folder:       []string{"Coleco"},
		Rbf:          "_Console/ColecoVision",
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
		Id:           "CreatiVision",
		Name:         "VTech CreatiVision",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerVTech,
		ReleaseDate:  "1981-01-01",
		Folder:       []string{"CreatiVision"},
		Rbf:          "_Console/CreatiVision",
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
			//{
			//	Label: "Bios",
			//	Exts:  []string{".rom", ".bin"},
			//	Mgl: &MglParams{
			//		Delay:  1,
			//		Method: "f",
			//		Index:  2,
			//	},
			//},
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
		Id:           "FDS",
		Name:         "Famicom Disk System",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1986-02-21",
		SetName:      "FDS",
		Alias:        []string{"FamicomDiskSystem"},
		Folder:       []string{"NES", "FDS"},
		Rbf:          "_Console/NES",
		Slots: []Slot{
			{
				Exts: []string{".fds"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
			//{
			//	Label: "FDS BIOS",
			//	Exts:  []string{".bin"},
			//	Mgl: &MglParams{
			//		Delay:  1,
			//		Method: "f",
			//		Index:  2,
			//	},
			//},
		},
	},
	"Gamate": {
		Id:           "Gamate",
		Name:         "Gamate",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerBitCorp,
		ReleaseDate:  "1990-01-01",
		Folder:       []string{"Gamate"},
		Rbf:          "_Console/Gamate",
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
		Id:           "Gameboy",
		Name:         "Gameboy",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1989-04-21",
		Alias:        []string{"GB"},
		Folder:       []string{"GAMEBOY"},
		Rbf:          "_Console/Gameboy",
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
		Id:           "GameboyColor",
		Name:         "Gameboy Color",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1998-10-21",
		Alias:        []string{"GBC"},
		Folder:       []string{"GAMEBOY", "GBC"},
		SetName:      "GBC",
		Rbf:          "_Console/Gameboy",
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
		Id:           "Gameboy2P",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1989-04-21",
		Name:         "Gameboy (2 Player)",
		Folder:       []string{"GAMEBOY2P"},
		Rbf:          "_Console/Gameboy2P",
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
		Id:           "GameGear",
		Name:         "Game Gear",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerSega,
		ReleaseDate:  "1990-10-06",
		Alias:        []string{"GG"},
		Folder:       []string{"SMS", "GameGear"},
		SetName:      "GameGear",
		Rbf:          "_Console/SMS",
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
		Id:           "GameNWatch",
		Name:         "Game & Watch",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1980-04-28",
		Folder:       []string{"GameNWatch"},
		Rbf:          "_Console/GnW",
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
		Id:           "GBA",
		Name:         "Gameboy Advance",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "2001-03-21",
		Alias:        []string{"GameboyAdvance"},
		Folder:       []string{"GBA"},
		Rbf:          "_Console/GBA",
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
		Id:           "GBA2P",
		Name:         "Gameboy Advance (2 Player)",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "2001-03-21",
		Folder:       []string{"GBA2P"},
		Rbf:          "_Console/GBA2P",
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
		Id:           "Genesis",
		Name:         "Genesis",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSega,
		ReleaseDate:  "1988-10-29",
		Alias:        []string{"MegaDrive"},
		Folder:       []string{"Genesis"},
		Rbf:          "_Console/Genesis",
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
		Id:           "Intellivision",
		Name:         "Intellivision",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerMattel,
		ReleaseDate:  "1979-12-03",
		Folder:       []string{"Intellivision"},
		Rbf:          "_Console/Intellivision",
		Slots: []Slot{
			{
				//Exts: []string{".rom", ".int", ".bin"},
				Exts: []string{".int", ".bin"},
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
		Id:           "MasterSystem",
		Name:         "Master System",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSega,
		ReleaseDate:  "1985-10-20",
		Alias:        []string{"SMS"},
		Folder:       []string{"SMS"},
		Rbf:          "_Console/SMS",
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
		Id:           "MegaCD",
		Name:         "Sega CD",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSega,
		ReleaseDate:  "1991-12-12",
		Alias:        []string{"SegaCD"},
		Folder:       []string{"MegaCD"},
		Rbf:          "_Console/MegaCD",
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
	"MegaDuck": {
		Id:           "MegaDuck",
		Name:         "Mega Duck",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerWatara,
		ReleaseDate:  "1993-01-01",
		Folder:       []string{"GAMEBOY", "MegaDuck"},
		Rbf:          "_Console/Gameboy",
		Slots: []Slot{
			{
				Exts: []string{".bin"},
				Mgl: &MglParams{
					Delay:  2,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"NeoGeo": {
		Id:           "NeoGeo",
		Name:         "Neo Geo MVS/AES",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSNK,
		ReleaseDate:  "1990-01-01",
		Folder:       []string{"NEOGEO"},
		Rbf:          "_Console/NeoGeo",
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
		},
	},
	"NeoGeoCD": {
		Id:           "NeoGeo",
		Name:         "Neo Geo CD",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSNK,
		ReleaseDate:  "1994-09-09",
		Folder:       []string{"NeoGeo-CD", "NEOGEO"},
		Rbf:          "_Console/NeoGeo",
		Slots: []Slot{
			{
				Label: "CD Image",
				Exts:  []string{".cue", ".chd"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  1,
				},
			},
		},
	},
	"NES": {
		Id:           "NES",
		Name:         "NES",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1985-10-18",
		Folder:       []string{"NES"},
		Rbf:          "_Console/NES",
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
		Id:       "NESMusic",
		Name:     "NES Music",
		Category: CategoryOther,
		Folder:   []string{"NES"},
		Rbf:      "_Console/NES",
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
	"Nintendo64": {
		Id:           "Nintendo64",
		Name:         "Nintendo 64",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1996-06-23",
		Folder:       []string{"N64"},
		Rbf:          "_Console/N64",
		Slots: []Slot{
			{
				Exts: []string{".n64", ".z64"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"Odyssey2": {
		Id:           "Odyssey2",
		Name:         "Magnavox Odyssey2",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerMagnavox,
		ReleaseDate:  "1978-09-01",
		Folder:       []string{"ODYSSEY2"},
		Rbf:          "_Console/Odyssey2",
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
			//{
			//	Label: "XROM",
			//	Exts:  []string{".rom"},
			//	Mgl: &MglParams{
			//		Delay:  1,
			//		Method: "f",
			//		Index:  2,
			//	},
			//},
		},
	},
	"PocketChallengeV2": {
		Id:           "PocketChallengeV2",
		Name:         "Pocket Challenge V2",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerBenesse,
		ReleaseDate:  "2000-01-01",
		Folder:       []string{"WonderSwan", "PocketChallengeV2"},
		SetName:      "PocketChallengeV2",
		Rbf:          "_Console/WonderSwan",
		Slots: []Slot{
			{
				Label: "ROM",
				Exts:  []string{".pc2"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"PokemonMini": {
		Id:           "PokemonMini",
		Name:         "Pokemon Mini",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "2001-11-16",
		Folder:       []string{"PokemonMini"},
		Rbf:          "_Console/PokemonMini",
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
		Id:           "PSX",
		Name:         "Playstation",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSony,
		ReleaseDate:  "1994-12-03",
		Alias:        []string{"Playstation", "PS1"},
		Folder:       []string{"PSX"},
		Rbf:          "_Console/PSX",
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
		Id:           "Sega32X",
		Name:         "Genesis 32X",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSega,
		ReleaseDate:  "1994-11-21",
		Alias:        []string{"S32X", "32X"},
		Folder:       []string{"S32X"},
		Rbf:          "_Console/S32X",
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
		Id:           "SG1000",
		Name:         "SG-1000",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSega,
		ReleaseDate:  "1983-07-15",
		SetName:      "SG1000",
		Folder:       []string{"SG1000", "Coleco", "SMS"},
		Rbf:          "_Console/ColecoVision",
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
		Id:           "SuperGameboy",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1994-06-14",
		Name:         "Super Gameboy",
		Alias:        []string{"SGB"},
		Folder:       []string{"SGB"},
		Rbf:          "_Console/SGB",
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
		Id:           "SuperVision",
		Name:         "SuperVision",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerWatara,
		ReleaseDate:  "1992-01-01",
		Folder:       []string{"SuperVision"},
		Rbf:          "_Console/SuperVision",
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
	"Saturn": {
		Id:           "Saturn",
		Name:         "Saturn",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerSega,
		ReleaseDate:  "1994-11-22",
		Folder:       []string{"Saturn"},
		Rbf:          "_Console/Saturn",
		Slots: []Slot{
			{
				Label: "Disk",
				Exts:  []string{".cue"}, // TODO: .chd support later
				Mgl: &MglParams{
					Delay:  1,
					Method: "s",
					Index:  0,
				},
			},
		},
	},
	"SNES": {
		Id:           "SNES",
		Name:         "SNES",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNintendo,
		ReleaseDate:  "1990-11-21",
		Alias:        []string{"SuperNintendo"},
		Folder:       []string{"SNES"},
		Rbf:          "_Console/SNES",
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
		Id:       "SNESMusic",
		Name:     "SNES Music",
		Category: CategoryOther,
		Folder:   []string{"SNES"},
		Rbf:      "_Console/SNES",
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
		Id:           "SuperGrafx",
		Name:         "SuperGrafx",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNEC,
		ReleaseDate:  "1989-12-08",
		Folder:       []string{"TGFX16"},
		Rbf:          "_Console/TurboGrafx16",
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
	"TurboGrafx16": {
		Id:           "TurboGrafx16",
		Name:         "TurboGrafx-16",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNEC,
		ReleaseDate:  "1987-10-30",
		Alias:        []string{"TGFX16", "PCEngine"},
		Folder:       []string{"TGFX16"},
		Rbf:          "_Console/TurboGrafx16",
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
	"TurboGrafx16CD": {
		Id:           "TurboGrafx16CD",
		Name:         "TurboGrafx-16 CD",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerNEC,
		ReleaseDate:  "1989-11-01",
		Alias:        []string{"TGFX16-CD", "PCEngineCD"},
		Folder:       []string{"TGFX16-CD"},
		Rbf:          "_Console/TurboGrafx16",
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
		Id:           "VC4000",
		Name:         "VC4000",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerInterton,
		ReleaseDate:  "1978-01-01",
		Folder:       []string{"VC4000"},
		Rbf:          "_Console/VC4000",
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
		Id:           "Vectrex",
		Name:         "Vectrex",
		Category:     CategoryConsole,
		Manufacturer: ManufacturerGCE,
		ReleaseDate:  "1982-11-01",
		Folder:       []string{"VECTREX"},
		Rbf:          "_Console/Vectrex",
		Slots: []Slot{
			{
				Exts: []string{".vec", ".bin", ".rom"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
			//{
			//	Label: "Overlay",
			//	Exts:  []string{".ovr"},
			//	Mgl: &MglParams{
			//		Delay:  1,
			//		Method: "f",
			//		Index:  2,
			//	},
			//},
		},
	},
	"WonderSwan": {
		Id:           "WonderSwan",
		Name:         "WonderSwan",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerBandai,
		ReleaseDate:  "1999-03-04",
		Folder:       []string{"WonderSwan"},
		Rbf:          "_Console/WonderSwan",
		Slots: []Slot{
			{
				Label: "ROM",
				Exts:  []string{".ws"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  1,
				},
			},
		},
	},
	"WonderSwanColor": {
		Id:           "WonderSwanColor",
		Name:         "WonderSwan Color",
		Category:     CategoryHandheld,
		Manufacturer: ManufacturerBandai,
		ReleaseDate:  "1999-12-30",
		Folder:       []string{"WonderSwan", "WonderSwanColor"},
		SetName:      "WonderSwanColor",
		Rbf:          "_Console/WonderSwan",
		Slots: []Slot{
			{
				Label: "ROM",
				Exts:  []string{".wsc"},
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
		Id:           "AcornAtom",
		Name:         "Atom",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerAcorn,
		ReleaseDate:  "1979-01-01",
		Folder:       []string{"AcornAtom"},
		Rbf:          "_Computer/AcornAtom",
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
		Id:           "AcornElectron",
		Name:         "Electron",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerAcorn,
		ReleaseDate:  "1983-08-01",
		Folder:       []string{"AcornElectron"},
		Rbf:          "_Computer/AcornElectron",
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
		Id:           "AliceMC10",
		Name:         "Tandy MC-10",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerTandy,
		ReleaseDate:  "1983-01-01",
		Folder:       []string{"AliceMC10"},
		Rbf:          "_Computer/AliceMC10",
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
		Id:           "Amiga",
		Name:         "Amiga",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerCommodore,
		ReleaseDate:  "1985-07-23",
		Folder:       []string{"Amiga"},
		Alias:        []string{"Minimig"},
		Rbf:          "_Computer/Minimig",
		Slots: []Slot{
			{
				Label: "df0",
				Exts:  []string{".adf"},
				Mgl: &MglParams{
					Delay:  1,
					Method: "f",
					Index:  0,
				},
			},
		},
	},
	"Amstrad": {
		Id:           "Amstrad",
		Name:         "Amstrad CPC",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerAmstrad,
		ReleaseDate:  "1984-06-21",
		Folder:       []string{"Amstrad"},
		Rbf:          "_Computer/Amstrad",
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
		Id:           "AmstradPCW",
		Name:         "Amstrad PCW",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerAmstrad,
		ReleaseDate:  "1985-09-01",
		Alias:        []string{"Amstrad-PCW"},
		Folder:       []string{"Amstrad PCW"},
		Rbf:          "_Computer/Amstrad-PCW",
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
		Id:           "ao486",
		Name:         "PC (486SX)",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerIBM,
		ReleaseDate:  "1989-04-10",
		Folder:       []string{"AO486"},
		Rbf:          "_Computer/ao486",
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
		Id:           "Apogee",
		Name:         "Apogee BK-01",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerApogee,
		ReleaseDate:  "1992-01-01",
		Folder:       []string{"APOGEE"},
		Rbf:          "_Computer/Apogee",
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
		Id:           "AppleI",
		Name:         "Apple I",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerApple,
		ReleaseDate:  "1976-04-01",
		Alias:        []string{"Apple-I"},
		Folder:       []string{"Apple-I"},
		Rbf:          "_Computer/Apple-I",
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
		Id:           "AppleII",
		Name:         "Apple IIe",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerApple,
		ReleaseDate:  "1983-01-01",
		Alias:        []string{"Apple-II"},
		Folder:       []string{"Apple-II"},
		Rbf:          "_Computer/Apple-II",
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
		Id:           "Aquarius",
		Name:         "Mattel Aquarius",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerMattel,
		ReleaseDate:  "1983-06-01",
		Folder:       []string{"AQUARIUS"},
		Rbf:          "_Computer/Aquarius",
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
		Id:           "Atari800",
		Name:         "Atari 800XL",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerAtari,
		ReleaseDate:  "1983-01-01",
		Folder:       []string{"ATARI800"},
		Rbf:          "_Computer/Atari800",
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
		Id:           "BBCMicro",
		Name:         "BBC Micro/Master",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerAcorn,
		ReleaseDate:  "1981-12-01",
		Folder:       []string{"BBCMicro"},
		Rbf:          "_Computer/BBCMicro",
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
		Id:           "BK0011M",
		Name:         "BK0011M",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerElektronika,
		Folder:       []string{"BK0011M"},
		Rbf:          "_Computer/BK0011M",
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
		Id:           "C16",
		Name:         "Commodore 16",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerCommodore,
		ReleaseDate:  "1984-07-01",
		Folder:       []string{"C16"},
		Rbf:          "_Computer/C16",
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
		Id:           "C64",
		Name:         "Commodore 64",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerCommodore,
		ReleaseDate:  "1982-08-01",
		Folder:       []string{"C64"},
		Rbf:          "_Computer/C64",
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
		Id:           "CasioPV2000",
		Name:         "Casio PV-2000",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerCasio,
		ReleaseDate:  "1983-01-01",
		Alias:        []string{"Casio_PV-2000"},
		Folder:       []string{"Casio_PV-2000"},
		Rbf:          "_Computer/Casio_PV-2000",
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
		Id:           "CoCo2",
		Name:         "TRS-80 CoCo 2",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerTandy,
		ReleaseDate:  "1983-01-01",
		Folder:       []string{"CoCo2"},
		Rbf:          "_Computer/CoCo2",
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
		Id:           "EDSAC",
		Name:         "EDSAC",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerCambridge,
		ReleaseDate:  "1949-01-01",
		Folder:       []string{"EDSAC"},
		Rbf:          "_Computer/EDSAC",
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
		Id:          "Galaksija",
		Name:        "Galaksija",
		Category:    CategoryComputer,
		ReleaseDate: "1983-01-01",
		Folder:      []string{"Galaksija"},
		Rbf:         "_Computer/Galaksija",
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
		Id:           "Interact",
		Name:         "Interact",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerInteract,
		ReleaseDate:  "1981-01-01",
		Folder:       []string{"Interact"},
		Rbf:          "_Computer/Interact",
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
		Id:           "Jupiter",
		Name:         "Jupiter Ace",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerJupiter,
		ReleaseDate:  "1982-01-01",
		Folder:       []string{"Jupiter"},
		Rbf:          "_Computer/Jupiter",
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
		Id:           "Laser",
		Name:         "Laser 350/500/700",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerVideoTechnology,
		ReleaseDate:  "1984-01-01",
		Alias:        []string{"Laser310"},
		Folder:       []string{"Laser"},
		Rbf:          "_Computer/Laser310",
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
		Id:           "Lynx48",
		Name:         "Lynx 48/96K",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerCambridge,
		ReleaseDate:  "1983-01-01",
		Folder:       []string{"Lynx48"},
		Rbf:          "_Computer/Lynx48",
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
		Id:           "MacPlus",
		Name:         "Macintosh Plus",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerApple,
		ReleaseDate:  "1986-01-01",
		Folder:       []string{"MACPLUS"},
		Rbf:          "_Computer/MacPlus",
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
		Id:           "MSX",
		Name:         "MSX",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerMicrosoft,
		ReleaseDate:  "1983-06-01",
		Folder:       []string{"MSX"},
		Rbf:          "_Computer/MSX",
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
		Id:       "MultiComp",
		Name:     "MultiComp",
		Category: CategoryComputer,
		Folder:   []string{"MultiComp"},
		Rbf:      "_Computer/MultiComp",
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
		Id:       "Orao",
		Name:     "Orao",
		Category: CategoryComputer,
		Folder:   []string{"ORAO"},
		Rbf:      "_Computer/ORAO",
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
		Id:       "Oric",
		Name:     "Oric",
		Category: CategoryComputer,
		Folder:   []string{"Oric"},
		Rbf:      "_Computer/Oric",
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
		Id:           "PCXT",
		Name:         "PC/XT",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerIBM,
		ReleaseDate:  "1983-01-01",
		Folder:       []string{"PCXT"},
		Rbf:          "_Computer/PCXT",
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
		Id:          "PDP1",
		Name:        "PDP-1",
		Category:    CategoryComputer,
		ReleaseDate: "1960-01-01",
		Folder:      []string{"PDP1"},
		Rbf:         "_Computer/PDP1",
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
		Id:           "PET2001",
		Name:         "Commodore PET 2001",
		Category:     CategoryComputer,
		Manufacturer: ManufacturerCommodore,
		ReleaseDate:  "1977-01-01",
		Folder:       []string{"PET2001"},
		Rbf:          "_Computer/PET2001",
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
		Id:       "PMD85",
		Name:     "PMD 85-2A",
		Category: CategoryComputer,
		Folder:   []string{"PMD85"},
		Rbf:      "_Computer/PMD85",
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
		Id:       "QL",
		Name:     "Sinclair QL",
		Category: CategoryComputer,
		Folder:   []string{"QL"},
		Rbf:      "_Computer/QL",
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
		Id:       "RX78",
		Name:     "RX-78 Gundam",
		Category: CategoryComputer,
		Folder:   []string{"RX78"},
		Rbf:      "_Computer/RX78",
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
		Id:       "SAMCoupe",
		Name:     "SAM Coupe",
		Category: CategoryComputer,
		Folder:   []string{"SAMCOUPE"},
		Rbf:      "_Computer/SAMCoupe",
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
		Id:       "SordM5",
		Name:     "M5",
		Category: CategoryComputer,
		Alias:    []string{"Sord M5"},
		Folder:   []string{"Sord M5"},
		Rbf:      "_Computer/SordM5",
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
		Id:       "Specialist",
		Name:     "Specialist/MX",
		Category: CategoryComputer,
		Alias:    []string{"SPMX"},
		Folder:   []string{"SPMX"},
		Rbf:      "_Computer/Specialist",
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
		Id:       "SVI328",
		Name:     "SV-328",
		Folder:   []string{"SVI328"},
		Category: CategoryComputer,
		Rbf:      "_Computer/Svi328",
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
		Id:       "TatungEinstein",
		Name:     "Tatung Einstein",
		Category: CategoryComputer,
		Folder:   []string{"TatungEinstein"},
		Rbf:      "_Computer/TatungEinstein",
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
		Id:       "TI994A",
		Name:     "TI-99/4A",
		Category: CategoryComputer,
		Alias:    []string{"TI-99_4A"},
		Folder:   []string{"TI-99_4A"},
		Rbf:      "_Computer/Ti994a",
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
		Id:       "TomyTutor",
		Name:     "Tutor",
		Category: CategoryComputer,
		Folder:   []string{"TomyTutor"},
		Rbf:      "_Computer/TomyTutor",
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
		Id:       "TRS80",
		Name:     "TRS-80",
		Category: CategoryComputer,
		Folder:   []string{"TRS-80"},
		Rbf:      "_Computer/TRS-80",
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
		Id:       "TSConf",
		Name:     "TS-Config",
		Category: CategoryComputer,
		Folder:   []string{"TSConf"},
		Rbf:      "_Computer/TSConf",
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
		Id:       "UK101",
		Name:     "UK101",
		Category: CategoryComputer,
		Folder:   []string{"UK101"},
		Rbf:      "_Computer/UK101",
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
		Id:       "Vector06C",
		Name:     "Vector-06C",
		Category: CategoryComputer,
		Alias:    []string{"Vector06"},
		Folder:   []string{"VECTOR06"},
		Rbf:      "_Computer/Vector-06C",
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
		Id:       "VIC20",
		Name:     "Commodore VIC-20",
		Category: CategoryComputer,
		Folder:   []string{"VIC20"},
		Rbf:      "_Computer/VIC20",
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
		Id:       "X68000",
		Name:     "X68000",
		Category: CategoryComputer,
		Folder:   []string{"X68000"},
		Rbf:      "_Computer/X68000",
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
		Id:       "ZX81",
		Name:     "TS-1500",
		Category: CategoryComputer,
		Folder:   []string{"ZX81"},
		Rbf:      "_Computer/ZX81",
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
		Id:       "ZXSpectrum",
		Name:     "ZX Spectrum",
		Category: CategoryComputer,
		Alias:    []string{"Spectrum"},
		Folder:   []string{"Spectrum"},
		Rbf:      "_Computer/ZX-Spectrum",
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
		Id:       "ZXNext",
		Name:     "ZX Spectrum Next",
		Category: CategoryComputer,
		Folder:   []string{"ZXNext"},
		Rbf:      "_Computer/ZXNext",
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
		Id:       "Arcade",
		Name:     "Arcade",
		Category: CategoryArcade,
		Folder:   []string{"_Arcade"},
		Slots: []Slot{
			{
				Exts: []string{".mra"},
				Mgl:  nil,
			},
		},
	},
	"Arduboy": {
		Id:       "Arduboy",
		Name:     "Arduboy",
		Category: CategoryOther,
		Folder:   []string{"Arduboy"},
		Rbf:      "_Other/Arduboy",
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
		Id:       "Chip8",
		Name:     "CHIP-8",
		Category: CategoryOther,
		Folder:   []string{"Chip8"},
		Rbf:      "_Other/Chip8",
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
