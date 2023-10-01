#!/usr/bin/env bash
# shellcheck disable=SC2094 # Dirty hack avoid runcommand to steal stdout

title="MiSTer NFC Writer"
scriptdir="$(dirname "$(readlink -f "${0}")")"
version="0.1"
fullFileBrowser="false"
#TODO thoroughly test this regex, being cognicent of users needs
url_regex='^(http|https|ftp)://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,4}(/.*)?$'
basedir="/media/fat/"
nfcCommand="${scriptdir}/nfc.sh"
settings="${scriptdir}/nfc.ini"
map="/media/fat/nfc.csv"
#For debugging purpouse
[[ -d "/media/fat" ]] || map="${scriptdir}/nfc.csv"
mapHeader="match_uid,match_text,text"
nfcStatus="$("${nfcCommand}" --service status)"
if [[ "${nfcStatus}" == "nfc service running" ]]; then
	nfcStatus="true"
else
	nfcStatus="false"
fi
[[ -f "/tmp/nfc.sock" ]] && nfcReadingStatus="$(echo "status" | socat - UNIX-CONNECT:/tmp/nfc.sock)"
[[ -n "${nfcReadingStatus}" ]] && nfcReadingStatus="$(cut -d ',' -f 3 <<< "${nfcReadingStatus}")"
[[ -n "${nfcReadingStatus}" ]] || nfcReadingStatus="false"
cmdPalette=(
	"system"	"Launch a system"
	"random"	"Launch a random game for the given system"
	"ini"		"Loads the specified MiSTer ini file"
	"get"		"Perform an HTTP GET request to the specified URL"
	"key"		"Press a key on the keyboard"
	"coinp1"	"Insert a coin/credit for player 1"
	"coinp2"	"Insert a coin/credit for player 2"
	"command"	"Run Linux command"
)
consoles=(
	"AdventureVision"	"Adventure Vision"
	"Amiga"			"Amiga"
	"Amstrad"		"Amstrad CPC"
	"AmstradPCW"		"Amstrad PCW"
	"Apogee"		"Apogee BK-01"
	"AppleI"		"Apple I"
	"AppleII"		"Apple IIe"
	"Arcade"		"Arcade"
	"Arcadia"		"Arcadia 2001"
	"Arduboy"		"Arduboy"
	"Atari2600"		"Atari 2600"
	"Atari5200"		"Atari 5200"
	"Atari7800"		"Atari 7800"
	"Atari800"		"Atari 800XL"
	"AtariLynx"		"Atari Lynx"
	"AcornAtom"		"Atom"
	"BBCMicro"		"BBC Micro/Master"
	"BK0011M"		"BK0011M"
	"Astrocade"		"Bally Astrocade"
	"Chip8"			"Chip-8"
	"CasioPV1000"		"Casio PV-1000"
	"CasioPV2000"		"Casio PV-2000"
	"ChannelF"		"Channel F"
	"ColecoVision"		"ColecoVision"
	"C64"			"Commodore 64"
	"PET2001"		"Commodore PET 2001"
	"VIC20"			"Commodore VIC-20"
	"EDSAC"			"EDSAC"
	"AcornElectron"		"Electron"
	"FDS"			"Famicom Disk System"
	"Galaksija"		"Galaksija"
	"Gamate"		"Gamate"
	"GameNWatch"		"Game & Watch"
	"GameGear"		"Game Gear"
	"Gameboy"		"Gameboy"
	"Gameboy2P"		"Gameboy (2 Player)"
	"GBA"			"Gameboy Advance"
	"GBA2P"			"Gameboy Advance (2 Player)"
	"GameboyColor"		"Gameboy Color"
	"Genesis"		"Genesis"
	"Sega32X"		"Genesis 32X"
	"Intellivision"		"Intellivision"
	"Interact"		"Interact"
	"Jupiter"		"Jupiter Ace"
	"Laser"			"Laser 350/500/700"
	"Lynx48"		"Lynx 48/96K"
	"SordM5"		"M5"
	"MSX"			"MSX"
	"MacPlus"		"Macintosh Plus"
	"Odyssey2"		"Magnavox Odyssey2"
	"MasterSystem"		"Master System"
	"Aquarius"		"Mattel Aquarius"
	"MegaDuck"		"Mega Duck"
	"MultiComp"		"MultiComp"
	"NES"			"NES"
	"NESMusic"		"NESMusic"
	"NeoGeo"		"Neo Geo/Neo Geo CD"
	"Nintendo64"		"Nintendo 64"
	"Orao"			"Orao"
	"Oric"			"Oric"
	"ao486"			"PC (486SX)"
	"OCXT"			"PC/XT"
	"PDP1"			"PDP-1"
	"PMD85"			"PMD 85-2A"
	"PSX"			"Playstation"
	"PocketChallengeV2"	"Pocket Challenge V2"
	"PokemonMini"		"Pokemon Mini"
	"RX78"			"RX-78 Gundam"
	"SAMCoupe"		"SAM Coupe"
	"SG1000"		"SG-1000"
	"SNES"			"SNES"
	"SNESMusic"		"SNES Music"
	"SVI328"		"SV-328"
	"Saturn"		"Saturn"
	"MegaCD"		"Sega CD"
	"QL"			"Sinclair QL"
	"Specialist"		"Specialist/MX"
	"SuperGameboy"		"Super Gameboy"
	"SuperGrafx"		"SuperGrafx"
	"SuperVision"		"SuperVision"
	"TI994A"		"TI-99/4A"
	"TRS80"			"TRS-80"
	"CoCo2"			"TRS-80 CoCo 2"
	"ZX81"			"TS-1500"
	"TSConf"		"TS-Config"
	"AliceMC10"		"Tandy MC-10"
	"TatungEinstein"	"Tatung Einstein"
	"TurboGrafx16"		"TurboGrafx-16"
	"TurboGrafx16CD"	"TurboGrafx-16 CD"
	"TomyTutor"		"Tutor"
	"UK101"			"UK101"
	"VC4000"		"VC4000"
	"CreatiVision"		"VTech CreatiVision"
	"Vector06C"		"Vector-06C"
	"Vectrex"		"Vectrex"
	"WonderSwan"		"WonderSwan"
	"WonderSwanColor"	"WonderSwan Color"
	"X68000"		"X68000"
	"ZXSpectrum"		"ZX Spectrum"
	"ZXNext"		"ZX Spectrum Next"
)

keycodes=(
	"Esc"			"1"
	"1"			"2"
	"2"			"3"
	"3"			"4"
	"4"			"5"
	"5"			"6"
	"6"			"7"
	"7"			"8"
	"8"			"9"
	"9"			"10"
	"0"			"11"
	"Minus"			"12"
	"Equal"			"13"
	"Backspace"		"14"
	"Tab"			"15"
	"Q"			"16"
	"W"			"17"
	"E"			"18"
	"R"			"19"
	"T"			"20"
	"Y"			"21"
	"U"			"22"
	"I"			"23"
	"O"			"24"
	"P"			"25"
	"Leftbrace"		"26"
	"Rightbrace"		"27"
	"Enter"			"28"
	"Leftctrl"		"29"
	"A"			"30"
	"S"			"31"
	"D"			"32"
	"F"			"33"
	"G"			"34"
	"H"			"35"
	"J"			"36"
	"K"			"37"
	"L"			"38"
	"Semicolon"		"39"
	"Apostrophe"		"40"
	"Grave"			"41"
	"Leftshift"		"42"
	"Backslash"		"43"
	"Z"			"44"
	"X"			"45"
	"C"			"46"
	"V"			"47"
	"B"			"48"
	"N"			"49"
	"M"			"50"
	"Comma"			"51"
	"Dot"			"52"
	"Slash"			"53"
	"Rightshift"		"54"
	"Kpasterisk"		"55"
	"Leftalt"		"56"
	"Space"			"57"
	"Capslock"		"58"
	"F1"			"59"
	"F2"			"60"
	"F3"			"61"
	"F4"			"62"
	"F5"			"63"
	"F6"			"64"
	"F7"			"65"
	"F8"			"66"
	"F9"			"67"
	"F10"			"68"
	"Numlock"		"69"
	"Scrolllock"		"70"
	"Kp7"			"71"
	"Kp8"			"72"
	"Kp9"			"73"
	"Kpminus"		"74"
	"Kp4"			"75"
	"Kp5"			"76"
	"Kp6"			"77"
	"Kpplus"		"78"
	"Kp1"			"79"
	"Kp2"			"80"
	"Kp3"			"81"
	"Kp0"			"82"
	"Kpdot"			"83"
	"Zenkakuhankaku"	"85"
	"102Nd"			"86"
	"F11"			"87"
	"F12"			"88"
	"Ro"			"89"
	"Katakana"		"90"
	"Hiragana"		"91"
	"Henkan"		"92"
	"Katakanahiragana"	"93"
	"Muhenkan"		"94"
	"Kpjpcomma"		"95"
	"Kpenter"		"96"
	"Rightctrl"		"97"
	"Kpslash"		"98"
	"Sysrq"			"99"
	"Rightalt"		"100"
	"Linefeed"		"101"
	"Home"			"102"
	"Up"			"103"
	"Pageup"		"104"
	"Left"			"105"
	"Right"			"106"
	"End"			"107"
	"Down"			"108"
	"Pagedown"		"109"
	"Insert"		"110"
	"Delete"		"111"
	"Macro"			"112"
	"Mute"			"113"
	"Volumedown"		"114"
	"Volumeup"		"115"
	"Power"			"116" #ScSystemPowerDown
	"Kpequal"		"117"
	"Kpplusminus"		"118"
	"Pause"			"119"
	"Scale"			"120" #AlCompizScale(Expose)
	"Kpcomma"		"121"
	"Hangeul"		"122"
	"Hanja"			"123"
	"Yen"			"124"
	"Leftmeta"		"125"
	"Rightmeta"		"126"
	"Compose"		"127"
	"Stop"			"128" #AcStop
	"Again"			"129"
	"Props"			"130" #AcProperties
	"Undo"			"131" #AcUndo
	"Front"			"132"
	"Copy"			"133" #AcCopy
	"Open"			"134" #AcOpen
	"Paste"			"135" #AcPaste
	"Find"			"136" #AcSearch
	"Cut"			"137" #AcCut
	"Help"			"138" #AlIntegratedHelpCenter
	"Menu"			"139" #Menu(ShowMenu)
	"Calc"			"140" #AlCalculator
	"Setup"			"141"
	"Sleep"			"142" #ScSystemSleep
	"Wakeup"		"143" #SystemWakeUp
	"File"			"144" #AlLocalMachineBrowser
	"Sendfile"		"145"
	"Deletefile"		"146"
	"Xfer"			"147"
	"Prog1"			"148"
	"Prog2"			"149"
	"Www"			"150" #AlInternetBrowser
	"Msdos"			"151"
	"Coffee"		"152" #AlTerminalLock/Screensaver
	"Direction"		"153"
	"Cyclewindows"		"154"
	"Mail"			"155"
	"Bookmarks"		"156" #AcBookmarks
	"Computer"		"157"
	"Back"			"158" #AcBack
	"Forward"		"159" #AcForward
	"Closecd"		"160"
	"Ejectcd"		"161"
	"Ejectclosecd"		"162"
	"Nextsong"		"163"
	"Playpause"		"164"
	"Previoussong"		"165"
	"Stopcd"		"166"
	"Record"		"167"
	"Rewind"		"168"
	"Phone"			"169" #MediaSelectTelephone
	"Iso"			"170"
	"Config"		"171" #AlConsumerControlConfiguration
	"Homepage"		"172" #AcHome
	"Refresh"		"173" #AcRefresh
	"Exit"			"174" #AcExit
	"Move"			"175"
	"Edit"			"176"
	"Scrollup"		"177"
	"Scrolldown"		"178"
	"Kpleftparen"		"179"
	"Kprightparen"		"180"
	"New"			"181" #AcNew
	"Redo"			"182" #AcRedo/Repeat
	"F13"			"183"
	"F14"			"184"
	"F15"			"185"
	"F16"			"186"
	"F17"			"187"
	"F18"			"188"
	"F19"			"189"
	"F20"			"190"
	"F21"			"191"
	"F22"			"192"
	"F23"			"193"
	"F24"			"194"
	"Playcd"		"200"
	"Pausecd"		"201"
	"Prog3"			"202"
	"Prog4"			"203"
	"Dashboard"		"204" #AlDashboard
	"Suspend"		"205"
	"Close"			"206" #AcClose
	"Play"			"207"
	"Fastforward"		"208"
	"Bassboost"		"209"
	"Print"			"210" #AcPrint
	"Hp"			"211"
	"Camera"		"212"
	"Sound"			"213"
	"Question"		"214"
	"Email"			"215"
	"Chat"			"216"
	"Search"		"217"
	"Connect"		"218"
	"Finance"		"219" #AlCheckbook/Finance
	"Sport"			"220"
	"Shop"			"221"
	"Alterase"		"222"
	"Cancel"		"223" #AcCancel
	"Brightnessdown"	"224"
	"Brightnessup"		"225"
	"Media"			"226"
	"Switchvideomode"	"227" #CycleBetweenAvailableVideo
	"Kbdillumtoggle"	"228"
	"Kbdillumdown"		"229"
	"Kbdillumup"		"230"
	"Send"			"231" #AcSend
	"Reply"			"232" #AcReply
	"Forwardmail"		"233" #AcForwardMsg
	"Save"			"234" #AcSave
	"Documents"		"235"
	"Battery"		"236"
	"Bluetooth"		"237"
	"Wlan"			"238"
	"Uwb"			"239"
	"Unknown"		"240"
	"VideoNext"		"241" #DriveNextVideoSource
	"VideoPrev"		"242" #DrivePreviousVideoSource
	"BrightnessCycle"	"243" #BrightnessUp,AfterMaxIsMin
	"BrightnessZero"	"244" #BrightnessOff,UseAmbient
	"DisplayOff"		"245" #DisplayDeviceToOffState
	"Wimax"			"246"
	"Rfkill"		"247" #KeyThatControlsAllRadios
	"Micmute"		"248" #Mute/UnmuteTheMicrophone

	"ButtonGamepad"		"0x130"

	"ButtonSouth"		"0x130" # A / X
	"ButtonEast"		"0x131" # X / Square
	"ButtonNorth"		"0x133" # Y / Triangle
	"ButtonWest"		"0x134" # B / Circle

	"ButtonBumperLeft"	"0x136" # L1
	"ButtonBumperRight"	"0x137" # R1
	"ButtonTriggerLeft"	"0x138" # L2
	"ButtonTriggerRight"	"0x139" # R2
	"ButtonThumbLeft"	"0x13d" # L3
	"ButtonThumbRight"	"0x13e" # R3

	"ButtonSelect"		"0x13a"
	"ButtonStart"		"0x13b"

	"ButtonDpadUp"		"0x220"
	"ButtonDpadDown"	"0x221"
	"ButtonDpadLeft"	"0x222"
	"ButtonDpadRight"	"0x223"

	"ButtonMode"		"0x13c" # This is the special button that usually bears the Xbox or Playstation logo
)

_depends() {
	if ! [[ -x "$(command -v dialog)" ]]; then
		echo "dialog not installed." >"$(tty)"
		sleep 10
		_exit 1
	fi

	# commented out for testing purpouse
	#[[ -x "${nfcCommand}" ]] || _error "${nfcCommand} not found" "1"
}

main() {
	export selected
	menuOptions=(
		"Read"     "Read NFC Tag"
		"Write"    "Write ROM file paths to NFC Tag"
		"Mappings" "Edit the mappings database"
		"Settings" "Options for ${title}"
		"About"    "About this program"
	)

	selected="$(_menu \
		--cancel-label "Exit" \
		--default-item "${selected}" \
		-- "${menuOptions[@]}")"

}

_Read() {
	local nfcSCAN nfcUID nfcTXT

	nfcSCAN="$(_readTag)"
	exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
	nfcTXT="$(cut -d ',' -f 4 <<< "${nfcSCAN}" )"
	nfcUID="$(cut -d ',' -f 2 <<< "${nfcSCAN}" )"
	[[ -n "${nfcSCAN}" ]] && _yesno "Tag contents: ${nfcTXT}\n Tag UID: ${nfcUID}" --yes-label "OK" --no-label "Re-Map" --extra-button --extra-label "Clone Tag"
	case "${?}" in
		1)
			_writeTextToMap --uid "${nfcUID}" "$(_commandPalette)"
			;;
		3)
			_writeTag "${nfcTXT}"
			;;
	esac
}

_Write() {
	local fileSelected message txtSize
	text="$(_commandPalette)"
	[[ "${?}" -eq 1 || "${?}" -eq 255 ]] && return
	txtSize="$(echo -n "${text}" | wc --bytes)"
	read -rd '' message <<_EOF_
The following file or command (without quotes) is to be written:

"${text:0:144}\Z4${text:144:504}\Z2${text:504:716}\Z3${text:716:888}\Z1${text:888}\Zn"

The NFC Tag needs to be able to fit at least ${txtSize} Bytes to write this tag
Common tag sizes:
NTAG213 		144 bytes storage
\Z4NTAG215 		504 bytes storage
\Z2MIFARE Classic 1K 	716 bytes storage
\Z3NTAG216 		888 bytes storage
\Z1Text over this size will be colored red\Zn
_EOF_
	_yesno "${message}" --colors --yes-label "Write to Tag" --extra-button --extra-label "Write to Map" --no-label "Cancel"
	answer="${?}"
	[[ -z "${text}" ]] && { _msgbox "Nothing selected for writing" ; return ; }
	[[ "${text}" =~ ^\*\*command:* ]] && { _msgbox "Writing system commands to NFC tags are disabled" ; return ; }
	case "${answer}" in
		0)
			_writeTag "${text}"
			;;
		3)
			# Extra button
			_writeTextToMap "${text}"
			;;
		1|255)
			return
			;;
	esac
}

# Gives the user the ability to enter text manually, pick a file, or use a command palette
# Usage: _commandPalette [-r]
# Returns a text string
# Example: text="$(_commandPalette)"
_commandPalette() {
	local menuOptions selected recursion
	recursion=false
	[[ "${1}" == "-r" ]] && recursion="true"
	menuOptions=(
		"Input"    "Input text manually, requires a keyboard"
		"Pick"     "Pick a file, including files inside zip files"
		"Commands" "Craft a custom command using a command palette"
	)

	selected="$(_menu \
		--cancel-label "Exit" \
		--default-item "${selected}" \
		-- "${menuOptions[@]}" )"
	exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"

	case "${selected}" in
		Input)
			inputText="$( _inputbox "Replace match text" "${match_text}")"
			exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"

			echo "${inputText}"
			;;
		Pick)
			#TODO refactor here a bit, because we may want to run commands after launching a file
			fileSelected="$(_fselect "${basedir}")"
			exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
			[[ ! -f "${fileSelected//.zip\/*/.zip}" ]] && { _error "No file was selected." ; return ; }
			fileSelected="${fileSelected//$basedir}"
			fileSelected="${fileSelected#/}"

			echo "${fileSelected}"
			;;
		Commands)
			text="$(recursion="${recursion}" _craftCommand)"
			exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
			echo "${text}"
			;;
	esac

}

# Build a command using a command palette
# Usage: _craftCommand
_craftCommand(){
	local command selected console recursion
	"${recursion}" || command="**"
	## Test if function is called recursively
	#if [[ "${FUNCNAME[0]}" != "${FUNCNAME[1]}" ]]; then
	#	command="**"
	#else
	#	command="||"
	#fi
	selected="$(_menu \
		--cancel-label "Exit" \
		-- "${cmdPalette[@]}" )"
	exitcode="${?}"
	[[ "${exitcode}" -ge 1 ]] && return "${exitcode}"

	command="${command}${selected}"

	case "${selected}" in
		system | random)
			console="$(_menu \
				--backtitle "${title}" \
				-- "${consoles[@]}" )"
			exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return 1
			command="${command}:${console}"
			;;
		ini)
			ini="$(_radiolist -- \
				1 one off 2 two off 3 three off 4 four off )"
			exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
			command="${command}:${ini}"
			;;
		get)
			while true; do
				http="$(_inputbox "Enter URL" "https://")"
				exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
				[[ "${http}" =~ ${url_regex} ]] && break
				_error "${http} doesn't appear to be a valid URL\n\nIf you believe this is a mistake please open an issue at\n\Zuhttps://github.com/wizzomafizzo/mrext/issues\ZU"
			done
			command="${command}:${http}"
			;;
		key)
			key="$(_menu -- "${keycodes[@]}")"
			for ((i=0; i<${#keycodes[@]}; i++)); do
				if [[ "${keycodes[$i]}" == "${key}" ]]; then
					index=$((i + 1))
					#workaround, for the shape of the array
					[[ "${keycodes[$index]}" == "${key}" ]] && index=$(( index + 1))
					key="${keycodes[$index]}"
					break
				fi
			done
			command="${command}:${key}"
			;;
		coinp1 | coinp2)
			while true; do
				coin="$(_inputbox "Enter number" "1")"
				exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
				[[ "${coin}" == ?(-)+([0-9]) ]] && break
				_error "${coin} is not a number"
			done
			command="${command}:${coin}"
			;;
		command)
			while true; do
				linuxcmd="$(_inputbox "Enter Linux command" "reboot" || return )"
				exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
				command -v "${linuxcmd%% *}" >/dev/null && break
				_error "${linuxcmd%% *} from ${linuxcmd} does not seam to be a valid command"
			done
			command="${command}:${linuxcmd}"
			;;
	esac
	_yesno "Do you wish to add an additional command?" --defaultno && command="${command}||$(_commandPalette -r)"
	echo "${command}"

}

_Settings() {
	local menuOptions selected
	menuOptions=(
		"Service"	"Toggle the NFC Service"
		"Commands"	"Toggles the ability to run Linux commands from NFC tags"
		"Sounds" 	"Toggles sounds played when a tag is scanned"
		"Connection"	"Hardware configuration for certain NFC readers"
	)

	while true; do
		selected="$(_menu -- "${menuOptions[@]}")"
		exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
		case "${selected}" in
			Service) _serviceSetting ;;
			Commands) _commandSetting ;;
			Sounds) _soundSetting ;;
			Connection) _connectionSetting ;;
		esac
	done
}

_serviceSetting() {
	local menuOptions selected
	menuOptions=(
		"Enable"     "Enable NFC service"  "off"
		"Disable"    "Disable NFC service" "off"
	)
	"${nfcStatus}" && menuOptions[2]="on"
	"${nfcStatus}" || menuOptions[5]="on"

	selected="$(_radiolist -- "${menuOptions[@]}" )"
	case "${selected}" in
		Enable)
			"${nfcCommand}" -service start || { _error "Unable to start the NFC service"; return; }
			nfcStatus="true"
			export nfcStatus
			_msgbox "The NFC service started"
			;;
		Disable)
			"${nfcCommand}" -service stop || { _error "Unable to stop the NFC service"; return; }
			nfcStatus="false"
			export nfcStatus
			_msgbox "The NFC service stopped"
			;;
	esac
}

_commandSetting() {
	local menuOptions selected
	menuOptions=(
		"Enable"     "Enable Linux commands"  "off"
		"Disable"    "Disable Linux commands" "off"
	)

	[[ -f "${settings}" ]] || echo "[nfc]" > "${settings}" || { _error "Can't create settings file" ; return 1 ; }

	if grep -q "^allow_commands=yes" "${settings}"; then
		menuOptions[2]="on"
	else
		menuOptions[5]="on"
	fi

	selected="$(_radiolist -- "${menuOptions[@]}" )"
	case "${selected}" in
		Enable)
			if grep -q "^allow_commands=" "${settings}"; then
			    sed -i "s/^allow_commands=.*/allow_commands=yes/" "${settings}"
			else
			    echo "allow_commands=yes" >> "${settings}"
			fi
			;;
		Disable)
			if grep -q "^allow_commands=" "${settings}"; then
			    sed -i "s/^allow_commands=.*/allow_commands=no/" "${settings}"
			else
			    echo "allow_commands=no" >> "${settings}"
			fi
			;;
	esac
}

_soundSetting() {
	local menuOptions selected
	menuOptions=(
		"Enable"     "Enable sounds played when a tag is scanned" "off"
		"Disable"    "Disable sounds played when a tag is scanned" "off"
	)

	[[ -f "${settings}" ]] || echo "[nfc]" > "${settings}" || { _error "Can't create settings file" ; return 1 ; }

	if grep -q "^disable_sounds=no" "${settings}"; then
		menuOptions[5]="on"
	else
		menuOptions[2]="on"
	fi

	selected="$(_radiolist -- "${menuOptions[@]}" )"
	case "${selected}" in
		Enable)
			if grep -q "^disable_sounds=" "${settings}"; then
			    sed -i "s/^disable_sounds=.*/disable_sounds=yes/" "${settings}"
			else
			    echo "disable_sounds=yes" >> "${settings}"
			fi
			;;
		Disable)
			if grep -q "^disable_sounds=" "${settings}"; then
			    sed -i "s/^disable_sounds=.*/disable_sounds=no/" "${settings}"
			else
			    echo "disable_sounds=no" >> "${settings}"
			fi
			;;
	esac
}

_connectionSetting() {
	local menuOptions selected customString
	menuOptions=(
		"Default"   "Automatically detect hardware (recommended)"  "off"
		"PN532"     "Select this option if you are using a PN532 UART module"       "off"
		"Custom"    "Manually enter a custom connection string"                     "off"
	)

	[[ -f "${settings}" ]] || echo "[nfc]" > "${settings}" || { _error "Can't create settings file" ; return 1 ; }

	if ! grep -q "^connection_string=.*" "${settings}"; then
		menuOptions[2]="on"
	elif grep -q "^connection_string=\"\"" "${settings}"; then
		menuOptions[2]="on"
	elif grep -q "^connection_string=\"pn532_uart:/dev/ttyUSB0\"" "${settings}"; then
		menuOptions[5]="on"
	elif grep -q "^connection_string=\".*\"" "${settings}"; then
		menuOptions[8]="on"
		customString="$(grep "^connection_string=" "${settings}" | cut -d '=' -f 2)"
		menuOptions[7]="Current custom option: ${customString}"
	fi

	selected="$(_radiolist -- "${menuOptions[@]}" )"
	case "${selected}" in
		Default)
			if grep -q "^connection_string=" "${settings}"; then
				sed -i "s/^connection_string=.*/connection_string=\"\"/" "${settings}"
			else
				echo 'connection_string=""' >> "${settings}"
			fi
			;;
		PN532)
			if grep -q "^connection_string=" "${settings}"; then
				sed -i 's/^connection_string=.*/connection_string="pn532_uart:\/dev\/ttyUSB0"/' "${settings}"
			else
				echo 'connection_string="pn532_uart:/dev/ttyUSB0"' >> "${settings}"
			fi
			;;
		Custom)
			customString="$(_inputbox "Enter connection string" "${customString}")"
			#TODO sanitize input
			if grep -q "^connection_string=" "${settings}"; then
				sed -i "s/^connection_string=.*/connection_string=\"${customString}\"/" "${settings}"
			else
				echo "connection_string=\"${customString}\"" >> "${settings}"
			fi
			;;
	esac
}

_About() {
	local about githash builddate gitbranch
	githash="$(git --git-dir="${scriptdir}/.git" rev-parse --short HEAD)"
	gitbranch="$(git --git-dir="${scriptdir}/.git" rev-parse --abbrev-ref HEAD)"
	builddate="$(git --git-dir="${scriptdir}/.git" log -1 --date=short --pretty=format:%cd)"
	#TODO actually write an about page
	read -rd '' about <<_EOF_
${title} ${version}-${gitbranch}-${builddate} + ${githash}

Add useful description here!
_EOF_
	_msgbox "${about}" --title "About"
}

# dialog --fselect broken out to a function,
# the purpouse is that
# if the screen is smaller then what --fselec can handle
# I can do somethig else
# Usage: _fselect "${fullPath}"
# returns the file that is selected including the full path, if full path is used.
_fselect() {
	local termh windowh dirList selected extension fileName fullPath newDir
	fullPath="${1}"
	[[ -f "${fullPath}" ]] && { echo "${fullPath}"; return; }
	termh="$(tput lines)"
	((windowh = "${termh}" - 10))
	[[ "${windowh}" -gt "22" ]] && windowh="22"
	if "${fullFileBrowser}" && [[ "${windowh}" -ge "8" ]]; then
		dialog \
			--backtitle "${title}" \
			--title "${fullPath}" \
			--fselect "${fullPath}/" \
			"${windowh}" 77 3>&1 1>&2 2>&3 >"$(tty)"
		exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"

	else
		# in case of a very tiny terminal window
		# make an array of the filenames and put them into --menu instead
		dirList=(
			"goto" "Go to directory (keyboard required)"
			".." "Up one directory"
		)

		while read -r folderName; do
			dirList+=("$(basename "${folderName}")" "Directory")

		done < <(find "${fullPath}" -mindepth 1 -maxdepth 1 ! -name '.*' -type d)

		while read -r fileName; do
			extension="${fileName##*.}"
			case "${extension,,}" in
			"")
				dirList+=("$(basename "${fileName}")")
				dirList+=("")
				;;

			*)
				dirList+=("$(basename "${fileName}")")
				dirList+=("File")
				;;
			esac

		done < <(find "${fullPath}" -maxdepth 1 -type f)

		selected="$(msg="Pick a game to write to NFC Tag" \
			_menu  --title "${fullPath}" -- "${dirList[@]}")"
		exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"

		case "${selected,,}" in
		"goto")
			newDir="$(_inputbox "Input a directory to go to" "${fullPath}")"
			_fselect "${newDir%/}"
			;;
		"..")
			_fselect "${fullPath%/*}"
			;;
		*.zip)
			echo "${fullPath}/${selected}/$(_browseZip "${fullPath}/${selected}")"
			;;
		*)
			_fselect "${fullPath}/${selected}"
			;;
		esac

	fi

}

# Browse contents of zip file as if it was a folder
# Usage: _browseZip "file.zip"
# returns a file path of a file inside the zip file
_browseZip() {
	local zipFile zipContents dirList currentDir relativeComponents currentDirTree relativePath
	zipFile="${1}"
	currentDir=""
	mapfile -t zipContents < <(unzip -l "${zipFile}" | awk 'NR > 3 {for (i=4; i<=NF; i++) {printf "%s", $i; if (i != NF) printf " ";} printf "\n"}')
	unset "zipContents[-1]" "zipContents[-1]"
	relativeComponents=(
		".." "Up one directory"
	)
	while true; do

		unset currentDirTree
		unset currentDirList
		for entry in "${zipContents[@]}"; do
			if [[ "${entry}" == "$currentDir" ]]; then
				true
			elif [[ "${entry}" == "$currentDir"* ]]; then
				currentDirTree+=( "${entry}" )
			fi
		done

		declare -a currentDirList
		for entry in "${currentDirTree[@]}"; do
			if [[ "${entry%/}" != "${currentDir}/"* ]]; then
				relativePath="${entry#"$currentDir"}"
				if [[ ${relativePath} == *"/"* ]]; then
					[[ "${currentDirList[-2]}" == "${relativePath%%/*}/" ]] && continue
					currentDirList+=( "${relativePath%%/*}/" )
				else
					[[ "${currentDirList[-2]}" == "${relativePath}" ]] && continue
					currentDirList+=( "${relativePath}" )
				fi


				if [[ "${currentDirList[-1]}" == */ ]]; then
					currentDirList+=( "Directory" )
				else
					currentDirList+=( "File" )
				fi
			fi
		done

		dirList=( "${relativeComponents[@]}" )
		dirList+=( "${currentDirList[@]}" )
		selected="$(dialog \
			--backtitle "${title}" \
			--title "${zipFile}" \
			--menu "${currentDir}" \
			22 77 16 "${dirList[@]}" 3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)")"
		exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"

		case "${selected,,}" in
		"..")
			currentDir="${currentDir%/}"
			[[ "${currentDir}" != *"/"* ]] && currentDir=""
			currentDir="${currentDir%/*}"
			[[ -n ${currentDir} ]] && currentDir="${currentDir}/"
			;;
		*/)
			currentDir="${currentDir}${selected}"
			;;
		*)
			echo "${currentDir}${selected}"
			break
			;;
		esac
	done
}

# Map or remap filepath or command for a given NFC tag (written to local database)
# Usage: _map "UID" "Text"
_map() {
	local uid txt
	uid="${1}"
	txt="${2}"
	[[ -e "${map}" ]] ||  printf "%s\n" "${mapHeader}" >> "${map}" || { _error "Can't initialize mappings database!" ; return 1 ; }
	grep -q "^${uid}" "${map}" && sed -i "/^${uid}/d" "${map}"
	printf "%s,,%s\n" "${uid}" "${txt}" >> "${map}"
}

_Mappings() {
	local oldMap arrayIndex line lineNumber match_uid match_text text menuOptions selected replacement_match_text replacement_match_uid replacement_text message new_match_uid new_text
	unset replacement_match_uid replacement_text

	[[ -e "${map}" ]] || printf "%s\n" "${mapHeader}" >> "${map}" || { _error "Can't initialize mappings database!" ; return 1 ; }

	mapfile -t -O 1 -s 1 oldMap < "${map}"
	echo "${oldMap[@]}"

	mapfile -t arrayIndex < <( _numberedArray "${oldMap[@]}" )

	line="$(msg="${mapHeader}" _menu \
		--extra-button --extra-label "New" \
		-- "${arrayIndex[@]//\"/}" )"

	if [[ "${?}" == "3" ]]; then
		new_match_uid="$(_readTag | cut -d ',' -f 2)"
		exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
		new_match_uid="$(cut -d ',' -f 2 <<< "${new_match_uid}")"
		new_text="$(_commandPalette)"
		exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
		_map "${new_match_uid}" "${new_text}"
		_Mappings
		return
	fi

	[[ -z "${line}" ]] && return
	lineNumber=$((line + 1))
	match_uid="$(cut -d ',' -f 1 <<< "${oldMap[$line]}")"
	match_text="$(cut -d ',' -f 2 <<< "${oldMap[$line]}")"
	text="$(cut -d ',' -f 3 <<< "${oldMap[$line]}")"

	menuOptions=(
		"UID" "${match_uid}"
		"Match" "${match_text}"
		"Text" "${text}"
		"Delete" "Remove entry from mappings database"
	)

	selected="$(_menu \
		-- "${menuOptions[@]}" )"
	exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { _Mappings ; return ; }

	case "${selected}" in
	UID)
		# Replace match_uid
		replacement_match_uid="$(_readTag | cut -d ',' -f 2)"
		[[ -z "${replacement_match_uid}" ]] && return
		;;
	Match)
		# Replace match_text
		replacement_match_text="$( _inputbox "Replace match text" "${match_text}" || return )"
		exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
		;;
	Text)
		# Replace text
		replacement_text="$(_commandPalette)"
		[[ -z "${replacement_text}" ]] && { _msgbox "Nothing selected for writing" ; return ; }
		;;
	Delete)
		# Delete line from Mappings database
		sed -i "${lineNumber}d" "${map}"
		return
		;;
	esac

	read -rd '' message <<_EOF_
Replace:
${match_uid},${match_text},${text}
With:
${replacement_match_uid:-$match_uid},${replacement_match_text:-$match_text},${replacement_text:-$text}
_EOF_
	_yesno "${message}" || return
	sed -i "${lineNumber}c\\${replacement_match_uid:-$match_uid},${replacement_match_text:-$match_text},${replacement_text:-$text}" "${map}"

}

# Returns array in a numbered fashion
# Usage: _numberedArray "${array[@]}"
# Returns:
# 1 first_element 2 second_element ....
_numberedArray() {
	local array index
	array=("$@")
	index=1

	for element in "${array[@]}"; do
		printf "%s\n" "${index}"
		printf "%s\n" "\"${element}\""
		((index++))
	done
}

# Write text string to physical NFC tag
# Usage: _writeTag "Text"
_writeTag() {
	local txt
	txt="${1}"

	"${nfcStatus}" && { "${nfcCommand}" -service stop || { _error "Unable to stop the NFC service" ; return ; } ; }
	"${nfcCommand}" -write "${txt}" || { _error "Unable to write the NFC Tag"; "${nfcCommand}" -service start ;  return; }
	"${nfcStatus}" && { "${nfcCommand}" -service start || { _error "Unable to start the NFC service" ; return ; } ; }

	_msgbox "${txt} \n successfully written to NFC tag"
}

# Write text string to NFC map (overrides physical NFC Tag contents)
# Usage: _writeTextToMap [--uid "UID"] <"Text">
_writeTextToMap() {
	local txt uid oldMapEntry

	while [[ "${#}" -gt "0" ]]; do
		case "${1}" in
		--uid)
			uid="${2}"
			shift 2
			;;
		*)
			txt="${1}"
			shift
			;;
		esac
	done

	# Check if UID is provided
	[[ -z "${uid}" ]] && uid="$(_readTag | cut -d ',' -f 2 )"

	# Check if the map file exists and read the existing entry for the given UID
	if [[ -f "${map}" ]]; then
		oldMapEntry=$(grep "^${uid}," "${map}")
	else
		echo "${mapHeader}" > "${map}"
	fi

	# If an existing entry is found, ask to replace it
	if [[ -n "$oldMapEntry" ]] && _yesno "UID:${uid}\nText:${txt}\n\nAdd entry to map? This will replace:\n${oldMapEntry}"; then
		sed -i "s|^${uid},.*|${uid},,${txt}|g" "${map}"
	elif _yesno "UID:${uid}\nText:${txt}\n\nAdd entry to map?"; then
		echo "${uid},,${txt}" >> "${map}"
	fi
}

# Read UID and Text from tag, returns comma separated values below
# Usage: _readTag
# Returns: "Unix epoch time","UID","core launch status","text"
_readTag() {
	local nfcSCAN nfcUID nfcTXT
	nfcScanTime="$(echo "status" | socat - UNIX-CONNECT:/tmp/nfc.sock | cut -d ',' -f 1)"
	[[ "${nfcReadingStatus}" ]] && echo "disable" | socat - UNIX-CONNECT:/tmp/nfc.sock
	_infobox "Scan NFC Tag to continue...\n\nPress any key to go back"
	while true; do
		[[ "${nfcScanTime}" != "$(echo "status" | socat - UNIX-CONNECT:/tmp/nfc.sock 2>/dev/null | cut -d ',' -f 1)" ]] && break
		sleep 1
		read -t 1 -n 1 -r  && return 1
	done
	nfcSCAN="$(echo "status" | socat - UNIX-CONNECT:/tmp/nfc.sock)"
	[[ "${nfcReadingStatus}" ]] && echo "enable" | socat - UNIX-CONNECT:/tmp/nfc.sock
	#[[ -z "${nfcSCAN}" ]] && { _error "Tag not read" ; _readTag ; }
	[[ -n "${nfcSCAN}" ]] && echo "${nfcSCAN}"
}


# Ask user for a string
# Usage: _inputbox "My message" "Initial text" [--optional-arguments]
# You can pass additioal arguments to the dialog program
# Backtitle is already set
_inputbox() {
	local msg opts init
	msg="${1}"
	init="${2}"
	shift 2
	opts=("${@}")
	dialog \
		--backtitle "${title}" \
		"${opts[@]}" \
		--inputbox "${msg}" \
		22 77 "${init}" 3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
	return "${?}"
}

# Display a menu
# Usage: _menu [--optional-arguments] -- [tag item]
# You can pass additioal arguments to the dialog program
# Backtitle is already set
_menu() {
	local menu_items optional_args

	# Separate optional arguments from menu items
	while [[ $# -gt 0 ]]; do
		if [[ "$1" == "--" ]]; then
			shift
			break
		else
			optional_args+=("$1")
			shift
		fi
	done

	# Collect menu items
	while [[ $# -gt 0 ]]; do
		menu_items+=("$1")
		shift
	done

	dialog \
		--backtitle "${title}" \
		"${optional_args[@]}" \
		--menu "${msg:-Chose one}" \
		22 77 16 "${menu_items[@]}" 3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
	return "${?}"
}

# Display a radio menu
# Usage: _radiolist [--optional-arguments] -- [tag item status]
# You can pass additioal arguments to the dialog program
# Backtitle is already set
_radiolist() {
	local menu_items optional_args

	# Separate optional arguments from menu items
	while [[ $# -gt 0 ]]; do
		if [[ "$1" == "--" ]]; then
			shift
			break
		else
			optional_args+=("$1")
			shift
		fi
	done

	# Collect menu items
	while [[ $# -gt 0 ]]; do
		menu_items+=("$1")
		shift
	done

	dialog \
		--backtitle "${title}" \
		"${optional_args[@]}" \
		--radiolist "Chose one" \
		22 77 16 "${menu_items[@]}" 3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
	return "${?}"
}

# Display an infobox, this exits immediately without clearing the screen
# Usage: _msgbox "My message" [--optional-arguments]
# You can pass additioal arguments to the dialog program
# Backtitle is already set
_infobox() {
	local msg opts
	msg="${1}"
	shift
	opts=("${@}")
	dialog \
		--backtitle "${title}" \
		--aspect 0 "${opts[@]}" \
		--infobox "${msg}" \
		0 0  3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
	return "${?}"
}

# Display a message
# Usage: _msgbox "My message" [--optional-arguments]
# You can pass additioal arguments to the dialog program
# Backtitle is already set
_msgbox() {
	local msg opts
	msg="${1}"
	shift
	opts=("${@}")
	dialog \
		--backtitle "${title}" \
		"${opts[@]}" \
		--msgbox "${msg}" \
		22 77  3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
	return "${?}"
}

# Request user input
# Usage: _yesno "My question" [--optional-arguments]
# You can pass additioal arguments to the dialog program
# Backtitle is already set
# returns the exit code from dialog which depends on the user answer
_yesno() {
	local msg opts
	msg="${1}"
	shift
	opts=("${@}")
	dialog \
		--backtitle "${title}" \
		"${opts[@]}" \
		--yesno "${msg}" \
		22 77 3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
	return "${?}"
}

# Display an error
# Usage: _error "My error" [1] [--optional-arguments]
# If the second argument is a number, the program will exit with that number as an exit code.
# You can pass additioal arguments to the dialog program
# Backtitle and title are already set
# Returns the exit code of the dialog program
_error() {
	local msg opts answer exitcode
	msg="${1}"
	shift
	[[ "${1}" =~ ^[0-9]+$ ]] && exitcode="${1}" && shift
	opts=("${@}")

	dialog \
		--backtitle "${title}" \
		--title "\Z1ERROR:\Zn" \
		--colors \
		"${opts[@]}" \
		--msgbox "${msg}" \
		22 77 3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
	answer="${?}"
	[[ -n "${exitcode}" ]] && exit "${exitcode}"
	return "${answer}"
}

_exit() {
	clear
	exit "${1:-0}"
}

# Check if element is in array
# Usage: _isInArray "element" "${array[@]}"
# returns exit code 0 if element is array, returns exitcode 1 if element is in array
_isInArray() {
	local string="${1}"
	shift
	local array=("${@}")
	[[ "${#array}" -eq 0 ]] && return 1

	for item in "${array[@]}"; do
		if [[ "${string}" == "${item}" ]]; then
			return 0
		fi
	done

	return 1
}

_depends

while true; do
	main
	"_${selected:-exit}"
done
