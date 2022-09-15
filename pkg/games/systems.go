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

type AltRbfOpts map[string][]string

type System struct {
	Id        string
	Name      string
	Alias     []string
	Folder    string
	Rbf       string
	AltRbf    AltRbfOpts
	FileTypes []FileType
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
	"NES":       {Systems["NES"], Systems["FDS"]},
	"Gameboy":   {Systems["Gameboy"], Systems["GameboyColor"]},
	"SMS":       {Systems["MasterSystem"], Systems["GameGear"]},
}

// FIXME: launch game > launch new game same system > not working? should it?
// TODO: setname attribute
// TODO: alternate cores
// TODO: alternate arcade folders
// TODO: custom scan function
// TODO: custom launch function
// TODO: support for multiple folders (think about symlink support here, check for dupes)
// TODO: could cut down on work scanning by folder rather than system

var Systems = map[string]System{
	"Arcade": {
		Id:     "Arcade",
		Name:   "Arcade",
		Folder: "_Arcade",
		FileTypes: []FileType{
			{
				Extensions: []string{".mra"},
				Mgl:        nil,
			},
		},
	},
	// Consoles
	// TODO: AY-3-8500
	// TODO: AVision
	// TODO: Arcadia
	// TODO: Astrocade
	"Atari2600": {
		Id:     "Atari2600",
		Name:   "Atari 2600",
		Folder: "ATARI7800",
		Rbf:    "_Console/Atari7800",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Atari7800_LLAPI"},
			AltRbfYC:    []string{"Atari7800YC"},
		},
		FileTypes: []FileType{
			{
				Extensions: []string{".a26"},
				Mgl:        nil,
			},
		},
	},
	"Atari5200": {
		Id:     "Atari5200",
		Name:   "Atari 5200",
		Folder: "ATARI5200",
		Rbf:    "_Console/Atari5200",
		AltRbf: AltRbfOpts{
			AltRbfYC: []string{"Atari5200YC"},
		},
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
		Name:   "Atari 7800",
		Folder: "ATARI7800",
		Rbf:    "_Console/Atari7800",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Atari7800_LLAPI"},
			AltRbfYC:    []string{"Atari7800YC"},
		},
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
		Name:   "Atari Lynx",
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
	// TODO: ChannelF
	// TODO: apparently indexes are wrong on this
	// TODO: probably just remove .sg from here, keep in meta
	"ColecoVision": {
		Id:     "ColecoVision",
		Name:   "ColecoVision",
		Alias:  []string{"Coleco"},
		Folder: "Coleco",
		Rbf:    "_Console/ColecoVision",
		AltRbf: AltRbfOpts{
			AltRbfYC: []string{"ColecoVisionYC"},
		},
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
		Name:   "Famicom Disk System",
		Alias:  []string{"FamicomDiskSystem"},
		Folder: "NES",
		Rbf:    "_Console/NES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"NES_LLAPI"},
			AltRbfYC:    []string{"NESYC"},
		},
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
	// TODO: Gamate
	"Gameboy": {
		Id:     "Gameboy",
		Name:   "Gameboy",
		Alias:  []string{"GB"},
		Folder: "GAMEBOY",
		Rbf:    "_Console/Gameboy",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Gameboy_LLAPI"},
			AltRbfYC:    []string{"GameboyYC"},
		},
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
		Name:   "Gameboy Color",
		Alias:  []string{"GBC"},
		Folder: "GAMEBOY",
		Rbf:    "_Console/Gameboy",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Gameboy_LLAPI"},
			AltRbfYC:    []string{"GameboyYC"},
		},
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
		Name:   "Gameboy 2P",
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
		Name:   "Game Gear",
		Folder: "SMS",
		Rbf:    "_Console/SMS",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SMS_LLAPI"},
			AltRbfYC:    []string{"SMSYC"},
		},
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
	// TODO: GameNWatch
	"GBA": {
		Id:     "GBA",
		Name:   "GBA",
		Alias:  []string{"GameboyAdvance"},
		Folder: "GBA",
		Rbf:    "_Console/GBA",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"GBA_LLAPI"},
			AltRbfYC:    []string{"GBAYC"},
		},
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
		Name:   "GBA 2P",
		Folder: "GBA2P",
		Rbf:    "_Console/GBA2P",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"GBA2P_LLAPI"},
		},
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
		Name:   "Genesis",
		Alias:  []string{"MegaDrive"},
		Folder: "Genesis",
		Rbf:    "_Console/Genesis",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"Genesis_LLAPI"},
			AltRbfYC:    []string{"GenesisYC"},
		},
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
	// TODO: Intv
	// TODO: Jaguar
	"MasterSystem": {
		Id:     "MasterSystem",
		Name:   "Master System",
		Alias:  []string{"SMS"},
		Folder: "SMS",
		Rbf:    "_Console/SMS",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SMS_LLAPI"},
			AltRbfYC:    []string{"SMSYC"},
		},
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
	"MegaCD": {
		Id:     "MegaCD",
		Name:   "Mega CD",
		Alias:  []string{"SegaCD"},
		Folder: "MegaCD",
		Rbf:    "_Console/MegaCD",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"MegaCD_LLAPI"},
			AltRbfYC:    []string{"MegaCDYC"},
		},
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
	// TODO: this also has some special handling re: zip files
	"NeoGeo": {
		Id:     "NeoGeo",
		Name:   "Neo Geo",
		Folder: "NEOGEO",
		Rbf:    "_Console/NeoGeo",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"NeoGeo_LLAPI"},
			AltRbfYC:    []string{"NeoGeoYC"},
		},
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
	// TODO: split off nsf music to separate system
	"NES": {
		Id:     "NES",
		Name:   "NES",
		Folder: "NES",
		Rbf:    "_Console/NES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"NES_LLAPI"},
			AltRbfYC:    []string{"NESYC"},
		},
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
	// TODO: Odyssey2
	"PSX": {
		Id:     "PSX",
		Name:   "Playstation",
		Alias:  []string{"Playstation", "PS1"},
		Folder: "PSX",
		Rbf:    "_Console/PSX",
		AltRbf: AltRbfOpts{
			AltRbfYC:      []string{"PSXYC"},
			AltRbfDualRAM: []string{"PSX_DualSDRAM"},
		},
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
		Name:   "Sega 32X",
		Alias:  []string{"S32X", "32X"},
		Folder: "S32X",
		Rbf:    "_Console/S32X",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"S32X_LLAPI"},
			AltRbfYC:    []string{"S32XYC"},
		},
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
		Name:   "Super Gameboy",
		Alias:  []string{"SGB"},
		Folder: "SGB",
		Rbf:    "_Console/SGB",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SGB_LLAPI"},
			AltRbfYC:    []string{"SGBYC"},
		},
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
	// TODO: Saturn
	"SNES": {
		Id:     "SNES",
		Name:   "SNES",
		Alias:  []string{"SuperNintendo"},
		Folder: "SNES",
		Rbf:    "_Console/SNES",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"SNES_LLAPI"},
			AltRbfYC:    []string{"SNESYC"},
		},
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
		Name:   "TurboGraphx-16",
		Alias:  []string{"TGFX16", "PCEngine"},
		Folder: "TGFX16",
		Rbf:    "_Console/TurboGrafx16",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"TurboGrafx16_LLAPI"},
			AltRbfYC:    []string{"TurboGrafx16YC"},
		},
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
		Name:   "TurboGraphx-16 CD",
		Alias:  []string{"TGFX16-CD", "PCEngineCD"},
		Folder: "TGFX16-CD",
		Rbf:    "_Console/TurboGrafx16",
		AltRbf: AltRbfOpts{
			AltRbfLLAPI: []string{"TurboGrafx16_LLAPI"},
			AltRbfYC:    []string{"TurboGrafx16YC"},
		},
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
	// TODO: VC4000
	"Vectrex": {
		Id:     "Vectrex",
		Name:   "Vectrex",
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
		Name:   "WonderSwan",
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
	// Computers
	// TODO: AcornAtom
	"AliceMC10": {
		Id:     "AliceMC10",
		Name:   "Alice MC-10",
		Folder: "AliceMC10",
		Rbf:    "_Computer/AliceMC10",
		FileTypes: []FileType{
			{
				Extensions: []string{".c10"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	// TODO: Altair8800
	"Amiga": {
		// TODO: new versions of MegaAGS image support launching individual games
		//       will need support for custom scan and launch functions for a core
		Id:     "Amiga",
		Name:   "Amiga",
		Folder: "Amiga",
		Rbf:    "_Computer/Minimig",
		FileTypes: []FileType{
			{
				Extensions: []string{".adf", ".hdf"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"Amstrad": {
		Id:     "Amstrad",
		Name:   "Amstrad",
		Folder: "Amstrad",
		Rbf:    "_Computer/Amstrad",
		FileTypes: []FileType{
			{
				Extensions: []string{".dsk", ".cdt"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"AmstradPCW": {
		Id:     "AmstradPCW",
		Name:   "Amstrad PCW",
		Folder: "Amstrad PCW",
		Rbf:    "_Computer/Amstrad-PCW",
		FileTypes: []FileType{
			{
				Extensions: []string{".dsk"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"ao486": {
		Id:     "ao486",
		Name:   "ao486",
		Folder: "AO486",
		Rbf:    "_Computer/ao486",
		FileTypes: []FileType{
			{
				Extensions: []string{".img", ".vhd"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	"Apogee": {
		Id:     "Apogee",
		Name:   "Apogee",
		Folder: "APOGEE",
		Rbf:    "_Computer/Apogee",
		FileTypes: []FileType{
			{
				Extensions: []string{".rka"},
				Mgl: &MglParams{
					Delay:    1,
					FileType: "f",
					Index:    0,
				},
			},
		},
	},
	// TODO: Apple-I
	// TODO: Apple-II
	// TODO: Aquarius
	// TODO: Archie
	// TODO: Atari800
	// TODO: AtariST
	// TODO: BBCMicro
	// TODO: BK0011M
	// TODO: C16
	"C64": {
		Id:     "C64",
		Name:   "Commodore 64",
		Folder: "C64",
		Rbf:    "_Computer/C64",
		AltRbf: AltRbfOpts{
			AltRbfYC: []string{"C64YC"},
		},
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
	// TODO: Chip8
	// TODO: CoCo2
	// TODO: CoCo3
	// TODO: EDSAC
	// TODO: Galaksija
	// TODO: Interac
	// TODO: Jupiter
	// TODO: Laser310
	// TODO: MacPlus
	// TODO: MSX
	// TODO: MultiComp
	// TODO: OndraSPO186
	// TODO: Orao
	// TODO: Oric
	// TODO: PC88
	// TODO: PDP1
	// TODO: PET2001
	// TODO: PMD85
	// TODO: QL
	// TODO: RX-78
	// TODO: SAM-Coupe
	// TODO: SharpMZ
	// TODO: SordM5
	// TODO: Specialist
	// TODO: SVI328
	// TODO: TI-99_4A
	// TODO: TRS-80
	// TODO: TSConf
	// TODO: UK101
	// TODO: Vector-06C
	// TODO: VIC20
	// TODO: X68000
	// TODO: ZX-Spectrum
	// TODO: ZX81
	// TODO: ZXNext
}
