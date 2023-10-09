#!/usr/bin/env bash
# shellcheck disable=SC2094 # Dirty hack avoid runcommand to steal stdout

title="MiSTer NFC"
scriptdir="$(dirname "$(readlink -f "${0}")")"
version="0.1"
fullFileBrowser="false"
basedir="/media/fat"
nfcCommand="${scriptdir}/nfc.sh"
settings="${scriptdir}/nfc.ini"
map="/media/fat/nfc.csv"
#For debugging purpouse
[[ -d "/media/fat" ]] || map="${scriptdir}/nfc.csv"
mapHeader="match_uid,match_text,text"
nfcStatus="$("${nfcCommand}" --service status)"
case "${nfcStatus}" in
  "nfc service running")
    nfcStatus="true"
    nfcUnavailable="false"
    msg="Service: Enabled"
    ;;
  "nfc service not running")
    nfcStatus="false"
    nfcUnavailable="false"
    msg="Service: Disabled"
    ;;
  *)
    nfcStatus="false"
    nfcUnavailable="true"
    msg="Service: Unavailable"
esac
nfcSocket="UNIX-CONNECT:/tmp/nfc.sock"
if [[ -S "${nfcSocket#*:}" ]]; then
  nfcReadingStatus="$(echo "status" | socat - "${nfcSocket}")"
  nfcReadingStatus="$(cut -d ',' -f 3 <<< "${nfcReadingStatus}")"
  # Disable reading for the duration of the script
  # we trap the EXIT signal and execute the _exit() function to turn it on again
  echo "disable" | socat - "${nfcSocket}"
else
  nfcReadingStatus="false"
fi
# Match MiSTer theme
[[ -f "/media/fat/Scripts/.dialogrc" ]] && export DIALOGRC="/media/fat/Scripts/.dialogrc"
#dialog escape codes, requires --colors
# shellcheck disable=SC2034
black="\Z0" red="\Z1" green="\Z2" yellow="\Z3" blue="\Z4" magenta="\Z5" cyan="\Z6" white="\Z7" bold="\Zb" unbold="\ZB" reverse="\Zr" unreverse="\ZR" underline="\Zu" noUnderline="\ZU" reset="\Zn"

cmdPalette=(
  "system"  "Launch a system"
  "random"  "Launch a random game for a system"
  "ini"     "Change to the specified MiSTer ini file"
  "get"     "Perform an HTTP GET request to the specified URL"
  "key"     "Press a key on the keyboard"
  "coinp1"  "Insert a coin/credit for player 1"
  "coinp2"  "Insert a coin/credit for player 2"
  "command" "Run Linux command"
)
consoles=(
  "AdventureVision"   "Adventure Vision"
  "Amiga"             "Amiga"
  "Amstrad"           "Amstrad CPC"
  "AmstradPCW"        "Amstrad PCW"
  "Apogee"            "Apogee BK-01"
  "AppleI"            "Apple I"
  "AppleII"           "Apple IIe"
  "Arcade"            "Arcade"
  "Arcadia"           "Arcadia 2001"
  "Arduboy"           "Arduboy"
  "Atari2600"         "Atari 2600"
  "Atari5200"         "Atari 5200"
  "Atari7800"         "Atari 7800"
  "Atari800"          "Atari 800XL"
  "AtariLynx"         "Atari Lynx"
  "AcornAtom"         "Atom"
  "BBCMicro"          "BBC Micro/Master"
  "BK0011M"           "BK0011M"
  "Astrocade"         "Bally Astrocade"
  "Chip8"             "Chip-8"
  "CasioPV1000"       "Casio PV-1000"
  "CasioPV2000"       "Casio PV-2000"
  "ChannelF"          "Channel F"
  "ColecoVision"      "ColecoVision"
  "C64"               "Commodore 64"
  "PET2001"           "Commodore PET 2001"
  "VIC20"             "Commodore VIC-20"
  "EDSAC"             "EDSAC"
  "AcornElectron"     "Electron"
  "FDS"               "Famicom Disk System"
  "Galaksija"         "Galaksija"
  "Gamate"            "Gamate"
  "GameNWatch"        "Game & Watch"
  "GameGear"          "Game Gear"
  "Gameboy"           "Gameboy"
  "Gameboy2P"         "Gameboy (2 Player)"
  "GBA"               "Gameboy Advance"
  "GBA2P"             "Gameboy Advance (2 Player)"
  "GameboyColor"      "Gameboy Color"
  "Genesis"           "Genesis"
  "Sega32X"           "Genesis 32X"
  "Intellivision"     "Intellivision"
  "Interact"          "Interact"
  "Jupiter"           "Jupiter Ace"
  "Laser"             "Laser 350/500/700"
  "Lynx48"            "Lynx 48/96K"
  "SordM5"            "M5"
  "MSX"               "MSX"
  "MacPlus"           "Macintosh Plus"
  "Odyssey2"          "Magnavox Odyssey2"
  "MasterSystem"      "Master System"
  "Aquarius"          "Mattel Aquarius"
  "MegaDuck"          "Mega Duck"
  "MultiComp"         "MultiComp"
  "NES"               "NES"
  "NESMusic"          "NESMusic"
  "NeoGeo"            "Neo Geo/Neo Geo CD"
  "Nintendo64"        "Nintendo 64"
  "Orao"              "Orao"
  "Oric"              "Oric"
  "ao486"             "PC (486SX)"
  "OCXT"              "PC/XT"
  "PDP1"              "PDP-1"
  "PMD85"             "PMD 85-2A"
  "PSX"               "Playstation"
  "PocketChallengeV2" "Pocket Challenge V2"
  "PokemonMini"       "Pokemon Mini"
  "RX78"              "RX-78 Gundam"
  "SAMCoupe"          "SAM Coupe"
  "SG1000"            "SG-1000"
  "SNES"              "SNES"
  "SNESMusic"         "SNES Music"
  "SVI328"            "SV-328"
  "Saturn"            "Saturn"
  "MegaCD"            "Sega CD"
  "QL"                "Sinclair QL"
  "Specialist"        "Specialist/MX"
  "SuperGameboy"      "Super Gameboy"
  "SuperGrafx"        "SuperGrafx"
  "SuperVision"       "SuperVision"
  "TI994A"            "TI-99/4A"
  "TRS80"             "TRS-80"
  "CoCo2"             "TRS-80 CoCo 2"
  "ZX81"              "TS-1500"
  "TSConf"            "TS-Config"
  "AliceMC10"         "Tandy MC-10"
  "TatungEinstein"    "Tatung Einstein"
  "TurboGrafx16"      "TurboGrafx-16"
  "TurboGrafx16CD"    "TurboGrafx-16 CD"
  "TomyTutor"         "Tutor"
  "UK101"             "UK101"
  "VC4000"            "VC4000"
  "CreatiVision"      "VTech CreatiVision"
  "Vector06C"         "Vector-06C"
  "Vectrex"           "Vectrex"
  "WonderSwan"        "WonderSwan"
  "WonderSwanColor"   "WonderSwan Color"
  "X68000"            "X68000"
  "ZXSpectrum"        "ZX Spectrum"
  "ZXNext"            "ZX Spectrum Next"
)

keycodes=(
  "Esc"              "1"
  "1"                "2"
  "2"                "3"
  "3"                "4"
  "4"                "5"
  "5"                "6"
  "6"                "7"
  "7"                "8"
  "8"                "9"
  "9"                "10"
  "0"                "11"
  "Minus"            "12"
  "Equal"            "13"
  "Backspace"        "14"
  "Tab"              "15"
  "Q"                "16"
  "W"                "17"
  "E"                "18"
  "R"                "19"
  "T"                "20"
  "Y"                "21"
  "U"                "22"
  "I"                "23"
  "O"                "24"
  "P"                "25"
  "Leftbrace"        "26"
  "Rightbrace"       "27"
  "Enter"            "28"
  "Leftctrl"         "29"
  "A"                "30"
  "S"                "31"
  "D"                "32"
  "F"                "33"
  "G"                "34"
  "H"                "35"
  "J"                "36"
  "K"                "37"
  "L"                "38"
  "Semicolon"        "39"
  "Apostrophe"       "40"
  "Grave"            "41"
  "Leftshift"        "42"
  "Backslash"        "43"
  "Z"                "44"
  "X"                "45"
  "C"                "46"
  "V"                "47"
  "B"                "48"
  "N"                "49"
  "M"                "50"
  "Comma"            "51"
  "Dot"              "52"
  "Slash"            "53"
  "Rightshift"       "54"
  "Kpasterisk"       "55"
  "Leftalt"          "56"
  "Space"            "57"
  "Capslock"         "58"
  "F1"               "59"
  "F2"               "60"
  "F3"               "61"
  "F4"               "62"
  "F5"               "63"
  "F6"               "64"
  "F7"               "65"
  "F8"               "66"
  "F9"               "67"
  "F10"              "68"
  "Numlock"          "69"
  "Scrolllock"       "70"
  "Kp7"              "71"
  "Kp8"              "72"
  "Kp9"              "73"
  "Kpminus"          "74"
  "Kp4"              "75"
  "Kp5"              "76"
  "Kp6"              "77"
  "Kpplus"           "78"
  "Kp1"              "79"
  "Kp2"              "80"
  "Kp3"              "81"
  "Kp0"              "82"
  "Kpdot"            "83"
  "Zenkakuhankaku"   "85"
  "102Nd"            "86"
  "F11"              "87"
  "F12"              "88"
  "Ro"               "89"
  "Katakana"         "90"
  "Hiragana"         "91"
  "Henkan"           "92"
  "Katakanahiragana" "93"
  "Muhenkan"         "94"
  "Kpjpcomma"        "95"
  "Kpenter"          "96"
  "Rightctrl"        "97"
  "Kpslash"          "98"
  "Sysrq"            "99"
  "Rightalt"         "100"
  "Linefeed"         "101"
  "Home"             "102"
  "Up"               "103"
  "Pageup"           "104"
  "Left"             "105"
  "Right"            "106"
  "End"              "107"
  "Down"             "108"
  "Pagedown"         "109"
  "Insert"           "110"
  "Delete"           "111"
  "Macro"            "112"
  "Mute"             "113"
  "Volumedown"       "114"
  "Volumeup"         "115"
  "Power"            "116" #ScSystemPowerDown
  "Kpequal"          "117"
  "Kpplusminus"      "118"
  "Pause"            "119"
  "Scale"            "120" #AlCompizScale(Expose)
  "Kpcomma"          "121"
  "Hangeul"          "122"
  "Hanja"            "123"
  "Yen"              "124"
  "Leftmeta"         "125"
  "Rightmeta"        "126"
  "Compose"          "127"
  "Stop"             "128" #AcStop
  "Again"            "129"
  "Props"            "130" #AcProperties
  "Undo"             "131" #AcUndo
  "Front"            "132"
  "Copy"             "133" #AcCopy
  "Open"             "134" #AcOpen
  "Paste"            "135" #AcPaste
  "Find"             "136" #AcSearch
  "Cut"              "137" #AcCut
  "Help"             "138" #AlIntegratedHelpCenter
  "Menu"             "139" #Menu(ShowMenu)
  "Calc"             "140" #AlCalculator
  "Setup"            "141"
  "Sleep"            "142" #ScSystemSleep
  "Wakeup"           "143" #SystemWakeUp
  "File"             "144" #AlLocalMachineBrowser
  "Sendfile"         "145"
  "Deletefile"       "146"
  "Xfer"             "147"
  "Prog1"            "148"
  "Prog2"            "149"
  "Www"              "150" #AlInternetBrowser
  "Msdos"            "151"
  "Coffee"           "152" #AlTerminalLock/Screensaver
  "Direction"        "153"
  "Cyclewindows"     "154"
  "Mail"             "155"
  "Bookmarks"        "156" #AcBookmarks
  "Computer"         "157"
  "Back"             "158" #AcBack
  "Forward"          "159" #AcForward
  "Closecd"          "160"
  "Ejectcd"          "161"
  "Ejectclosecd"     "162"
  "Nextsong"         "163"
  "Playpause"        "164"
  "Previoussong"     "165"
  "Stopcd"           "166"
  "Record"           "167"
  "Rewind"           "168"
  "Phone"            "169" #MediaSelectTelephone
  "Iso"              "170"
  "Config"           "171" #AlConsumerControlConfiguration
  "Homepage"         "172" #AcHome
  "Refresh"          "173" #AcRefresh
  "Exit"             "174" #AcExit
  "Move"             "175"
  "Edit"             "176"
  "Scrollup"         "177"
  "Scrolldown"       "178"
  "Kpleftparen"      "179"
  "Kprightparen"     "180"
  "New"              "181" #AcNew
  "Redo"             "182" #AcRedo/Repeat
  "F13"              "183"
  "F14"              "184"
  "F15"              "185"
  "F16"              "186"
  "F17"              "187"
  "F18"              "188"
  "F19"              "189"
  "F20"              "190"
  "F21"              "191"
  "F22"              "192"
  "F23"              "193"
  "F24"              "194"
  "Playcd"           "200"
  "Pausecd"          "201"
  "Prog3"            "202"
  "Prog4"            "203"
  "Dashboard"        "204" #AlDashboard
  "Suspend"          "205"
  "Close"            "206" #AcClose
  "Play"             "207"
  "Fastforward"      "208"
  "Bassboost"        "209"
  "Print"            "210" #AcPrint
  "Hp"               "211"
  "Camera"           "212"
  "Sound"            "213"
  "Question"         "214"
  "Email"            "215"
  "Chat"             "216"
  "Search"           "217"
  "Connect"          "218"
  "Finance"          "219" #AlCheckbook/Finance
  "Sport"            "220"
  "Shop"             "221"
  "Alterase"         "222"
  "Cancel"           "223" #AcCancel
  "Brightnessdown"   "224"
  "Brightnessup"     "225"
  "Media"            "226"
  "Switchvideomode"  "227" #CycleBetweenAvailableVideo
  "Kbdillumtoggle"   "228"
  "Kbdillumdown"     "229"
  "Kbdillumup"       "230"
  "Send"             "231" #AcSend
  "Reply"            "232" #AcReply
  "Forwardmail"      "233" #AcForwardMsg
  "Save"             "234" #AcSave
  "Documents"        "235"
  "Battery"          "236"
  "Bluetooth"        "237"
  "Wlan"             "238"
  "Uwb"              "239"
  "Unknown"          "240"
  "VideoNext"        "241" #DriveNextVideoSource
  "VideoPrev"        "242" #DrivePreviousVideoSource
  "BrightnessCycle"  "243" #BrightnessUp,AfterMaxIsMin
  "BrightnessZero"   "244" #BrightnessOff,UseAmbient
  "DisplayOff"       "245" #DisplayDeviceToOffState
  "Wimax"            "246"
  "Rfkill"           "247" #KeyThatControlsAllRadios
  "Micmute"          "248" #Mute/UnmuteTheMicrophone

## Valid uinput keycodes, not supported, may be supported in the future
#  "ButtonGamepad"      "0x130"
#
#  "ButtonSouth"        "0x130" # A / X
#  "ButtonEast"         "0x131" # X / Square
#  "ButtonNorth"        "0x133" # Y / Triangle
#  "ButtonWest"         "0x134" # B / Circle
#
#  "ButtonBumperLeft"   "0x136" # L1
#  "ButtonBumperRight"  "0x137" # R1
#  "ButtonTriggerLeft"  "0x138" # L2
#  "ButtonTriggerRight" "0x139" # R2
#  "ButtonThumbLeft"    "0x13d" # L3
#  "ButtonThumbRight"   "0x13e" # R3
#
#  "ButtonSelect"       "0x13a"
#  "ButtonStart"        "0x13b"
#
#  "ButtonDpadUp"       "0x220"
#  "ButtonDpadDown"     "0x221"
#  "ButtonDpadLeft"     "0x222"
#  "ButtonDpadRight"    "0x223"
#
#  "ButtonMode"         "0x13c" # This is the special button that usually bears the Xbox or Playstation logo
)

_depends() {
  if ! [[ -x "$(command -v dialog)" ]]; then
    echo "dialog not installed." >"$(tty)"
    sleep 10
    _exit 1
  fi

  [[ -x "${nfcCommand}" ]] || _error "${nfcCommand} not found" "1"
}

main() {
  export selected
  menuOptions=(
    "Read"      "Read NFC tag contents"
    "Write"     "Write game or command to NFC tag"
    "Mappings"  "Edit the mappings database"
    "Settings"  "Options for NFC script"
    "About"     "About this program"
  )


  selected="$(_menu \
    --cancel-label "Exit" --colors \
    --default-item "${selected}" \
    -- "${menuOptions[@]}")"

}

_Read() {
  local nfcSCAN nfcUID nfcTXT mappedMatch message

  nfcSCAN="$(_readTag)"
  exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
  nfcTXT="$(cut -d ',' -f 4 <<< "${nfcSCAN}" )"
  nfcUID="$(cut -d ',' -f 2 <<< "${nfcSCAN}" )"
  read -rd '' message <<_EOF_
Tag contents: ${nfcTXT}
Tag UID: ${nfcUID}
_EOF_
  [[ -f "${map}" ]] && mappedMatch="$(grep -i "^${nfcUID}" "${map}")"
  [[ -n "${mappedMatch}" ]] && read -rd '' message <<_EOF_
${message}

Mapped match by UID:
${mappedMatch}
_EOF_

  [[ -f "${map}" ]] && matchedEntry="$(_searchMatchText "${nfcTXT}")"
  [[ -n "${matchedEntry}" ]] && read -rd '' message <<_EOF_
${message}

Mapped match by match_text:
${matchedEntry}
_EOF_
  [[ -n "${nfcSCAN}" ]] && _yesno "${message}" --ok-label "OK" --yes-label "OK" --no-label "Re-Map" --cancel-label "Re-Map" --extra-button --extra-label "Clone Tag"
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
  local fileSelected message txtSize text
  # We can decide text via environment, and we can extend the command via argument
  # but since extending the command is done recursively it inherits the environemnt
  # so we do this check
  if [[ -z "${text}" ]] || [[ -n "${1}" ]]; then
    text="${1}$(_commandPalette)"
  fi
  [[ "${?}" -eq 1 || "${?}" -eq 255 ]] && return
  txtSize="$(echo -n "${text}" | wc --bytes)"
  read -rd '' message <<_EOF_
The following file or command will be written:

${text:0:144}${blue}${text:144:504}${green}${text:504:716}${yellow}${text:716:888}${red}${text:888}${reset}

The NFC tag needs to be able to fit at least ${txtSize} bytes.
Common tag sizes:
NTAG213     144 bytes storage
${blue}NTAG215    504 bytes storage
${green}MIFARE Classic 1K  716 bytes storage
${yellow}NTAG216    888 bytes storage
${red}Text over this size will be colored red.${reset}
_EOF_
  _yesno "${message}" --colors --ok-label "Write to Tag" --yes-label "Write to Tag" --extra-button --extra-label "Write to Map" --no-label "Cancel" --help-button --help-label "Chain Commands"
  answer="${?}"
  [[ -z "${text}" ]] && { _msgbox "Nothing selected for writing." ; return ; }
  case "${answer}" in
    0)
      # Yes button (Write to Tag)
      # if allow_commands is not set to yes (default no), and if text either starts with "**command:" or if it contains "||**command:" display an error instead of writing to tag
      if ! grep -q "^allow_commands=yes" "${settings}" && [[ "${text}" =~ (^\\*\\*|\\|\\|\\*\\*)command:* ]]; then
        _msgbox "You are trying to write a linux command to a physical tag.\nWriting system commands to NFC tags is disabled.\nThis can be enabled in the Settings\n\nOffending command:\n${text}"
        return
      fi
      _writeTag "${text}"
      ;;
    2)
      # Help Button (Chain Commands)
      _Write "${text}||"
      ;;
    3)
      # Extra button (Write to Map)
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
  menuOptions=(
    "Pick"      "Pick a game, core or arcade file (supports .zip files)"
    "Commands"  "Craft a custom command using the command palette"
    "Input"     "Input text manually (requires a keyboard)"
  )

  selected="$(_menu \
    --cancel-label "Back" \
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
  command="**"
  selected="$(_menu \
    --cancel-label "Back" \
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
      http="$(_inputbox "Enter URL" "https://")"
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
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
  echo "${command}"

}

_Settings() {
  local menuOptions selected
  menuOptions=(
    "Service"     "Start/stop the NFC service"
    "Commands"    "Toggles the ability to run Linux commands from NFC tags"
    "Sounds"      "Toggles sounds played when a tag is scanned"
    "Connection"  "Hardware configuration for certain NFC readers"
  )

  while true; do
    selected="$(_menu --cancel-label "Back" -- "${menuOptions[@]}")"
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

  "${nfcUnavailable}" && { _error "NFC Service Unavailable!\n\nIs the NFC script installed?"; return; }

  menuOptions=(
    "Enable"   "Enable NFC service"  "off"
    "Disable"  "Disable NFC service" "off"
  )
  "${nfcStatus}" && menuOptions[2]="on"
  "${nfcStatus}" || menuOptions[5]="on"

  selected="$(_radiolist -- "${menuOptions[@]}" )"
  case "${selected}" in
    Enable)
      "${nfcCommand}" -service start || { _error "Unable to start the NFC service"; return; }
      export nfcStatus="true" msg="Service: Enabled"
      _msgbox "The NFC service started"
      ;;
    Disable)
      "${nfcCommand}" -service stop || { _error "Unable to stop the NFC service"; return; }
      export nfcStatus="false" msg="Service: Disabled"
      _msgbox "The NFC service stopped"
      ;;
  esac
}

_commandSetting() {
  local menuOptions selected
  menuOptions=(
    "Enable"   "Enable Linux commands"  "off"
    "Disable"  "Disable Linux commands" "off"
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
    "Enable"   "Enable sounds played when a tag is scanned"   "off"
    "Disable"  "Disable sounds played when a tag is scanned"  "off"
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
    "Default"   "Automatically detect hardware (recommended)"               "off"
    "PN532"     "Select this option if you are using a PN532 UART module"   "off"
    "Custom"    "Manually enter a custom connection string"                 "off"
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
  local about
  read -rd '' about <<_EOF_
${bold}${title}${unbold}
${version}
A tool for making and working with NFC tags on MiSTer FPGA

Whats New? Get involved? Need help?
  ${underline}github.com/wizzomafizzo/mrext${noUnderline}

Why did the NFC tag break up with the Wi-Fi router?
  Because it wanted a closer connection!

Gaz       ${underline}github.com/symm${noUnderline}
Wizzo     ${underline}github.com/wizzomafizzo${noUnderline}
Ziggurat  ${underline}github.com/sigboe${noUnderline}

License: GPL v3.0
  ${underline}github.com/wizzomafizzo/mrext/blob/main/LICENSE${noUnderline}
_EOF_
  _msgbox "${about}" --no-collapse --colors --title "About"
}

# dialog --fselect broken out to a function,
# the purpouse is that
# if the screen is smaller then what --fselec can handle
# I can do somethig else
# Usage: _fselect "${fullPath}"
# returns the file that is selected including the full path, if full path is used.
_fselect() {
  local termh windowh relativeComponents selected fullPath newDir currentDirDirs currentDirFiles
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
    relativeComponents=(
      "goto"  "Go to directory (keyboard required)"
      ".."    "Up one directory"
    )

    # Get all folders in the current dir, and put them into an array Dialog likes,
    # then do the same for all the files. They are added with full path names
    # we could remove them here, but feels snappier for the user when we use
    # bash variable expansion with ##*/ when we expand the array
    readarray -t currentDirDirs <<< "$(find "${fullPath}" -mindepth 1 -maxdepth 1 -type d | while read -r line; do echo -e "${line}\nDirectory"; done)"
    readarray -t currentDirFiles <<< "$(find "${fullPath}" -mindepth 1 -maxdepth 1 -type f | while read -r line; do echo -e "${line}\nFile"; done)"

    selected="$(msg="Pick a game to write to NFC Tag" \
      _menu  --title "${fullPath}" -- "${relativeComponents[@]}" "${currentDirDirs[@]##*/}" "${currentDirFiles[@]##*/}" )"
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
  local zipFile currentDir relativeComponents currentDirDirs currentDirFiles
  zipFile="${1}"
  currentDir=""

  relativeComponents=(
    ".." "Up one directory"
  )
  while true; do

    # Lets do some black magic. We do this because its many many times faster than previous methods used. I don't know if it can be optimized more in bash.
    # use readarray to make an array for current direcotires in the current directory we are browsing. We read the whole zipFile every time in this wile true loop
    # as it's faster. remove first and last line, remove leading whitespace. filter out elements not in currentDir, remove current dir it self,
    # remove the current dir path from elements, filter out any elements not inclduing a / as they are files, remove elements in subdirectories, and filter duplicates
    # lastly interleave the elements with the element "Directory" because dialog --menu expects a description for each element.
    #
    # Then do the same for currentDirFiles, but filtering out any element that has a / in them as they are folders. And use "File" as description instead.
    _infobox "Loading."
    readarray -t currentDirDirs <<< "$(zip -sf "${zipFile}"  | tail -n +2 | head -n -1 | sed 's/^[[:space:]]*//' | grep "${currentDir}" | sed -e "/^${currentDir//\//\\/}$/d" -e "s|^${currentDir}||" | grep "/" | sed 's/\/.*$/\//' | uniq | while read -r line; do echo -e "${line}\nDirectory"; done)"
    _infobox "Loading.."
    readarray -t currentDirFiles <<< "$(zip -sf "${zipFile}"  | tail -n +2 | head -n -1 | sed 's/^[[:space:]]*//' | grep "${currentDir}" | sed -e "/^${currentDir//\//\\/}$/d" -e "s|^${currentDir}||" | grep -v "/" | while read -r line; do echo -e "${line}\nFile"; done)"
    _infobox "Loading..."
    [[ "${#currentDirDirs[@]}" -le "1" ]] && unset currentDirDirs
    [[ "${#currentDirFiles[@]}" -le "1" ]] && unset currentDirFiles

    selected="$(msg="${currentDir}" _menu --backtitle "${title}" \
      --title "${zipFile}" -- "${relativeComponents[@]}" "${currentDirDirs[@]}" "${currentDirFiles[@]}")"
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
# Usage: _map "UID" "Match Text" "Text"
# Values may be empty
_map() {
  local uid match txt
  uid="${1}"
  match="${2}"
  txt="${3}"
  [[ -e "${map}" ]] ||  printf "%s\n" "${mapHeader}" >> "${map}" || { _error "Can't initialize mappings database!" ; return 1 ; }
  [[ -z "${uid}" ]] || { grep -q "^${uid}" "${map}" && sed -i "/^${uid}/d" "${map}" ; }
  printf "%s,%s,%s\n" "${uid}" "${match}" "${txt}" >> "${map}"
}

_Mappings() {
  local oldMap arrayIndex line lineNumber match_uid match_text text menuOptions selected replacement_match_text replacement_match_uid replacement_text message new_match_uid new_text
  unset replacement_match_uid replacement_text

  [[ -e "${map}" ]] || printf "%s\n" "${mapHeader}" >> "${map}" || { _error "Can't initialize mappings database!" ; return 1 ; }

  mapfile -t -O 1 -s 1 oldMap < "${map}"

  mapfile -t arrayIndex < <( _numberedArray "${oldMap[@]}" )

  # Display something useful if the file is empty
  [[ "${#arrayIndex[@]}" -eq 0 ]] && arrayIndex=( "File Empty" "" )

  line="$(msg="${mapHeader}" _menu \
    --extra-button --extra-label "New" \
    --cancel-label "Back" \
    -- "${arrayIndex[@]//\"/}" )"
  exitcode="${?}"

  # Cancel button (Back) or Esc hit
  [[ "${exitcode}" -eq "1" ]] || [[ "${exitcode}" -eq "255" ]] && return "${exitcode}"

  # Extra button (New) pressed
  if [[ "${exitcode}" == "3" ]]; then
    _yesno "Read tag or type match text?" \
      --ok-label "Read tag" --yes-label "Read tag" \
      --no-label "Cancel" \
      --extra-button --extra-label "Match text"
    case "${?}" in
      0)
        # Yes button (Read tag)
        new_match_uid="$(_readTag)"
        exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
        new_match_uid="$(cut -d ',' -f 2 <<< "${new_match_uid}")"
        ;;
      3)
        # Extra button (Match text)
        new_match_text="$( _inputbox "Replace match text" "${match_text}")"
        exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
        ;;
      1|255)
        # No button (Cancel)
        _Mappings
        return
        ;;
    esac
    while true; do
      [[ -z "${new_text}" ]] && new_text="$(_commandPalette)"
      [[ -z "${new_text}" ]] || new_text="${new_text}||$(_commandPalette)"
      _yesno "Do you want to chain more commands?" --defaultno || break
    done
    exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
    _map "${new_match_uid}" "${new_match_text}" "${new_text}"
    _Mappings
    return
  fi

  [[ ${line} == "File Empty" ]] && return
  lineNumber=$((line + 1))
  match_uid="$(cut -d ',' -f 1 <<< "${oldMap[$line]}")"
  match_text="$(cut -d ',' -f 2 <<< "${oldMap[$line]}")"
  text="$(cut -d ',' -f 3 <<< "${oldMap[$line]}")"

  menuOptions=(
    "UID"     "${match_uid}"
    "Match"   "${match_text}"
    "Text"    "${text}"
    "Write"   "Write text to physical tag"
    "Delete"  "Remove entry from mappings database"
  )

  selected="$(_menu \
    --cancel-label "Done" \
    -- "${menuOptions[@]}" )"
  exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { _Mappings ; return ; }

  case "${selected}" in
  UID)
    # Replace match_uid
    replacement_match_uid="$(_readTag | cut -d ',' -f 2)"
    [[ -z "${replacement_match_uid}" ]] && return
    replacement_match_text="${match_text}"
    replacement_text="${text}"

    ;;
  Match)
    # Replace match_text
    replacement_match_text="$( _inputbox "Replace match text" "${match_text}")"
    exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
    replacement_match_uid="${match_uid}"
    replacement_text="${text}"
    ;;
  Text)
    # Replace text
    replacement_text="$(_commandPalette)"
    [[ -z "${replacement_text}" ]] && { _msgbox "Nothing selected for writing" ; return ; }
    replacement_match_uid="${match_uid}"
    replacement_match_text="${match_text}"
    ;;
  Write)
    # Write to physical tag
    #_writeTag "${text}"
    text="${text}" _Write
    return
    ;;
  Delete)
    # Delete line from Mappings database
    sed -i "${lineNumber}d" "${map}"
    _Mappings
    return
    ;;
  esac

  read -rd '' message <<_EOF_
Replace:
${match_uid},${match_text},${text}
With:
${replacement_match_uid},${replacement_match_text},${replacement_text}
_EOF_
  _yesno "${message}" || return
  sed -i "${lineNumber}c\\${replacement_match_uid},${replacement_match_text},${replacement_text}" "${map}"

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

  "${nfcCommand}" -write "${txt}" || { _error "Unable to write the NFC Tag"; return; }
  # Workaround for -write enabling launching games again
  echo "disable" | socat - "${nfcSocket}"

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
  local lastScanTime currentScan currentScanTime scanSuccess
  lastScanTime="$(echo "status" | socat - "${nfcSocket}" | cut -d ',' -f 1)"
  _infobox "Scan NFC Tag to continue...\n\nPress any key to go back"
  while true; do
    currentScan="$(echo "status" | socat - "${nfcSocket}" 2>/dev/null)"
    currentScanTime="$(cut -d ',' -f 1 <<< "${currentScan}")"
    [[ "${lastScanTime}" != "${currentScanTime}" ]] && { scanSuccess="true" ; break; }
    sleep 1
    read -t 1 -n 1 -r  && return 1
  done
  currentScan="$(echo "status" | socat - "${nfcSocket}")"
  if [[ ! "${scanSuccess}" ]]; then
    _yesno "Tag not read" --yes-label "Retry" && _readTag
    return
  fi
  #TODO determin if we need a message here saying the scan was successful
  [[ -n "${currentScan}" ]] && echo "${currentScan}"
}

# Search for possible matches by match_text in the mappings database
# Usage: _searchMatchText "Text"
# Returns lines that match
_searchMatchText() {
  local nfcTxt
  nfcTxt="${1}"

  [[ -f "${map}" ]] || return
  [[ $(head -n 1 "${map}") == "${mapHeader}" ]] || return

  sed 1d "${map}" | while IFS=, read -r match_uid match_text text; do
    [[ -z "${match_text}" ]] && continue
    if [[ "${nfcTxt}" == *"${match_text}"* ]]; then
      echo "${match_uid},${match_text},${text}"
    fi
  done
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
# Usage: [msg="message"] _menu [--optional-arguments] -- [ tag item ] ...
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
# Usage: [msg="message"] _radiolist [--optional-arguments] -- [ tag item status ] ...
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
    --radiolist "${msg:-Chose one}" \
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
# Usage: _error "My error" [<number>] [--optional-arguments]
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
  "${nfcReadingStatus}" && echo "enable" | socat - "${nfcSocket}"
  exit "${1:-0}"
}
trap _exit EXIT

_depends

while true; do
  main
  "_${selected:-exit}"
done

# vim: set expandtab ts=2 sw=2:
