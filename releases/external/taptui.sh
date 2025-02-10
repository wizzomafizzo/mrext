#!/usr/bin/env bash
# shellcheck disable=SC2094 # Dirty hack avoid runcommand to steal stdout

# TapTo
# Copyright (C) 2023 Sigurd Bøe
#
# This file is part of TapTo.
#
# TapTo is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# TapTo is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with TapTo.  If not, see <http://www.gnu.org/licenses/>.

title="TapTo"
scriptdir="$(dirname "$(readlink -f "${0}")")"
version="0.3"
fullFileBrowser="false"
basedir="/media"
sdroot="${basedir}/fat"
searchCommand="${scriptdir}/search.sh"
nfcCommand="${scriptdir}/tapto.sh"
settings="${scriptdir}/tapto.ini"
map="${sdroot}/nfc.csv"
#For debugging purpouse
[[ -d "${sdroot}" ]] || map="${scriptdir}/nfc.csv"
[[ -d "${sdroot}" ]] && PATH="${sdroot}/linux:${sdroot}/Scripts:${PATH}"
mapHeader="match_uid,match_text,text"
nfcStatus="$("${nfcCommand}" --service status)"
case "${nfcStatus}" in
  "tapto service running")
    nfcStatus="true"
    nfcUnavailable="false"
    msg="Service: Enabled"
    ;;
  "tapto service not running")
    nfcStatus="false"
    nfcUnavailable="false"
    msg="Service: Disabled"
    ;;
  *)
    nfcStatus="false"
    nfcUnavailable="true"
    msg="Service: Unavailable"
esac
nfcSocket="UNIX-CONNECT:/tmp/tapto/tapto.sock"
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
[[ -f "${sdroot}/Scripts/.dialogrc" ]] && export DIALOGRC="${sdroot}/Scripts/.dialogrc"
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
  "Groovy"            "Groovy"
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
  "PCXT"              "PC/XT"
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

nfcjokes=(
"Why did the NFC tag break up with the Wi-Fi router?
  Because it wanted a closer connection!"
"What did the smartphone say to the NFC tag?
  I'm just a tap away!"
"Why did the smartphone break up with the NFC tag?
  Because it felt like it didn't have enough range!"
"Why does the NFC tag need to be scanned to be updated?
  Because it needs to touch base!"
"What did the NFC reader say to the card?
  Tag! You're it!"
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

  [[ -x "${nfcCommand}" ]] || _error "${nfcCommand} not found\n\nRead more at ${underline}github.com/wizzomafizzo/tapto${noUnderline}" "1" --colors

  [[ -x "$(command -v rg)" ]] && grep() { rg "${@}"; }
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
${bold}Tag UID:${unbold} ${nfcUID}
${bold}Tag contents:${unbold}
${nfcTXT}
_EOF_
  [[ -f "${map}" ]] && mappedMatch="$(grep -i "^${nfcUID}" "${map}")"
  [[ -n "${mappedMatch}" ]] && read -rd '' message <<_EOF_
${message}

${bold}Mapped match by UID:${unbold}
${mappedMatch}
_EOF_

  [[ -f "${map}" ]] && matchedEntry="$(_searchMatchText "${nfcTXT}")"
  [[ -n "${matchedEntry}" ]] && read -rd '' message <<_EOF_
${message}

${bold}Mapped match by match_text:${unbold}
${matchedEntry}
_EOF_
  [[ -n "${nfcSCAN}" ]] && _yesno "${message}" \
    --colors --ok-label "OK" --yes-label "OK" \
    --no-label "Re-Map" --cancel-label "Re-Map" \
    --extra-button --extra-label "Clone Tag"
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
  local message txtSize text
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
  _yesno "${message}" --colors --ok-label "Write to Tag" --yes-label "Write to Tag" \
    --extra-button --extra-label "Write to Map" \
    --no-label "Cancel" --help-button --help-label "Chain Commands"
  answer="${?}"
  [[ -z "${text}" ]] && { _msgbox "Nothing selected for writing." ; return ; }
  case "${answer}" in
    0)
      # Yes button (Write to Tag)
      # if allow_commands is not set to yes (default no), and if text either starts with "**command:" or if it contains "||**command:" display an error instead of writing to tag
      if ! grep -q "^allow_commands=yes" "${settings}" && [[ "${text}" =~ (^\\*\\*|\\|\\|\\*\\*)command:* ]]; then
        _yesno "You are trying to write a linux command to a physical tag.\nWriting system commands to NFC tags is disabled.\nThis can be enabled in the Settings\n\nOffending command:\n${text}" --yes-label "Write Anyway" --no-label "Back" --defaultno || { text="${text}" "${FUNCNAME[0]}" ; return ; }
      fi
      _writeTag "${text}" || { text="${text}" "${FUNCNAME[0]}" ; return ; }
      ;;
    2)
      # Help Button (Chain Commands)
      _Write "${text}||"
      ;;
    3)
      # Extra button (Write to Map)
      _writeTextToMap "${text}" || { text="${text}" "${FUNCNAME[0]}" ; return ; }
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
  local menuOptions selected recursion fileSelected gamePath
  menuOptions=(
    "Pick"      "Pick a game, core or arcade file (supports .zip files)"
    "Commands"  "Craft a custom command using the command palette"
    "Search"    "Search for a game"
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
      fileSelected="$(_fselect "$(_gameLocation)")"
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
      [[ ! -f "${fileSelected//.zip\/*/.zip}" ]] && { _error "No file was selected." ; return ; }
      # shellcheck disable=SC2001
      fileSelected="$(sed -E "s#/media/(usb[0-7]|fat|network)(/cifs)?(/games)?/##i" <<< "${fileSelected}")"

      echo "${fileSelected}"
      ;;
    Commands)
      text="$(recursion="${recursion}" _craftCommand)"
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
      echo "${text}"
      ;;
    Search)
      gamePath="$("${searchCommand}" -print 2>&1 >"$(tty)")"
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
      gamePath="$(sed -E "s#/media/(usb[0-7]|fat|network)(/cifs)?(/games)?/##i" <<< "${gamePath}")"
      [[ -z "${gamePath}" ]] && { "${FUNCNAME[0]}" ; return ; }
      echo "${gamePath}"
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
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
      command="${command}:${console}"
      ;;
    ini)
      ini="$(_radiolist -- \
        1 one on 2 two off 3 three off 4 four off )"
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
      command="${command}:${ini}"
      ;;
    get)
      http="$(_inputbox "Enter URL" "https://")"
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
      command="${command}:${http}"
      ;;
    key)
      key="$(_menu -- "${keycodes[@]}")"
      exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
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
        exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
        [[ "${coin}" == ?(-)+([0-9]) ]] && break
        _error "${coin} is not a number"
      done
      command="${command}:${coin}"
      ;;
    command)
      while true; do
        linuxcmd="$(_inputbox "Enter Linux command" "reboot" || return )"
        exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { "${FUNCNAME[0]}" ; return ; }
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
    "Service"     "Start/stop the TapTo service"
    "Commands"    "Toggles the ability to run Linux commands from NFC tags"
    "Sounds"      "Toggles sounds played when a tag is scanned"
    "Connection"  "Hardware configuration for certain NFC readers"
    "Probe"       "Auto detection of a serial based reader device"
  )

  while true; do
    selected="$(_menu --cancel-label "Back" -- "${menuOptions[@]}")"
    exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"
    case "${selected}" in
      Service) _serviceSetting ;;
      Commands) _commandSetting ;;
      Sounds) _soundSetting ;;
      Connection) _connectionSetting ;;
      Probe) _probeSetting ;;
    esac
  done
}

_serviceSetting() {
  local menuOptions selected

  "${nfcUnavailable}" && { _error "TapTo Service Unavailable!\n\nIs TapTo installed?"; return; }

  menuOptions=(
    "Enable"   "Enable TapTo service"  "off"
    "Disable"  "Disable TapTo service" "off"
  )
  "${nfcStatus}" && menuOptions[2]="on"
  "${nfcStatus}" || menuOptions[5]="on"

  selected="$(_radiolist -- "${menuOptions[@]}" )"
  case "${selected}" in
    Enable)
      "${nfcCommand}" -service start || { _error "Unable to start the TapTo service"; return; }
      export nfcStatus="true" msg="Service: Enabled"
      _msgbox "The TapTo service started"
      ;;
    Disable)
      "${nfcCommand}" -service stop || { _error "Unable to stop the TapTo service"; return; }
      export nfcStatus="false" msg="Service: Disabled"
      _msgbox "The TapTo service stopped"
      ;;
  esac
}

_commandSetting() {
  local menuOptions selected helpmsg
  menuOptions=(
    "Enable"   "Enable Linux commands"  "off"
    "Disable"  "Disable Linux commands" "off"
  )

  [[ -f "${settings}" ]] || echo "[tapto]" > "${settings}" || { _error "Can't create settings file" ; return 1 ; }

  if grep -q "^allow_commands=yes" "${settings}"; then
    menuOptions[2]="on"
  else
    menuOptions[5]="on"
  fi

  read -rd '' helpmsg <<_EOF_
This option enables the execution of Linux commands stored on NFC tags. If you disable this option, you will be limited to running Linux commands that are stored in the mappings database.

Running Linux commands stored in NFC tags on your MiSTer FPGA device may pose security risks as it opens the door to potential vulnerabilities and unintended consequences. While this approach offers convenience, it lacks the security safeguards inherent in the Mappings database, which is designed to ensure that only known, verified commands are executed. By relying on NFC tags, unauthorized users could potentially create and execute commands with unverified origins, increasing the risk of malicious actions and compromising the integrity of your MiSTer system. Therefore, using the Mappings database remains the more secure option for safeguarding your device and its functionality.
_EOF_

  selected="$(_radiolist --help-button -- "${menuOptions[@]}" )"
  [[ "${?}" -eq 2 ]] && { _msgbox "${helpmsg}" ; "${FUNCNAME[0]}" ; return ; }
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

  [[ -f "${settings}" ]] || echo "[tapto]" > "${settings}" || { _error "Can't create settings file" ; return 1 ; }

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

  [[ -f "${settings}" ]] || echo "[tapto]" > "${settings}" || { _error "Can't create settings file" ; return 1 ; }

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

_probeSetting() {
  local menuOptions selected
  menuOptions=(
    "Enable"   "Enable detection of a serial based reader device"   "off"
    "Disable"  "Disable detection of a serial based reader device"  "off"
  )

  [[ -f "${settings}" ]] || echo "[tapto]" > "${settings}" || { _error "Can't create settings file" ; return 1 ; }

  # Check if probe_device is set to "no" in the settings
  if grep -q "^probe_device=no" "${settings}"; then
    # If probe_device is "no", set the corresponding radio button to "on"
    menuOptions[5]="on" # Disable option is selected
  else
    # If probe_device is not "no", set the corresponding radio button to "on"
    menuOptions[2]="on" # Enable option is selected by default
  fi

  selected="$(_radiolist -- "${menuOptions[@]}" )"
  case "${selected}" in
    Enable)
      if grep -q "^probe_device=" "${settings}"; then
        sed -i "s/^probe_device=.*/probe_device=yes/" "${settings}"
      else
        echo "probe_device=yes" >> "${settings}"
      fi
      ;;
    Disable)
      if grep -q "^probe_device=" "${settings}"; then
        sed -i "s/^probe_device=.*/probe_device=no/" "${settings}"
      else
        echo "probe_device=no" >> "${settings}"
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
  ${underline}github.com/wizzomafizzo/tapto${noUnderline}

${nfcjokes[$((RANDOM % ${#nfcjokes[@]}))]}

Gaz       ${underline}github.com/symm${noUnderline}
Wizzo     ${underline}github.com/wizzomafizzo${noUnderline}
Ziggurat  ${underline}github.com/sigboe${noUnderline}

License: GPL v3.0
  ${underline}github.com/wizzomafizzo/tapto/blob/main/LICENSE${noUnderline}
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
  local termh windowh relativeComponents selected fullPath newDir
  fullPath="${1}"
  if [[ -f "${fullPath}" ]] && [[ -h "${fullPath}" ]]; then
      read -rd '' message <<_EOF_
The following file is a symlink:
${fullPath}
The path is $(echo -n "${fullPath}" | wc --bytes) bytes


Do you want to use the symlink, or the path to the actual file?
$(readlink -f "${fullPath}")
The path is $(echo -n "$(readlink -f "${fullPath}")" | wc --bytes) bytes
_EOF_
    _yesno "${message}" \
    --yes-label "Use Link" --no-label "Use File" && \
    { readlink -f "${fullPath}"; return; }
  elif [[ -f "${fullPath}" ]]; then
    echo "${fullPath}"
    return
  fi
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

    readarray -t currentDirContents <<< "$( \
      find "${fullPath}" -mindepth 1 -maxdepth 1 \
      \( -type d -printf '%P\nDirectory\n' \) \
      -o \( -type f -printf '%P\nFile\n' \))"

    selected="$(msg="Pick a game" \
      _menu  --title "${fullPath}" -- "${relativeComponents[@]}" "${currentDirContents[@]}" )"
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
      zippedPath="$(_browseZip "${fullPath}/${selected}")"
      if [[ -n "${zippedPath}" ]]; then
        echo "${fullPath}/${selected}/${zippedPath}"
      else
        _fselect "${fullPath}"
      fi
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
  local zipFile currentDir relativeComponents currentDirDirs currentDirFiles tmpFile
  zipFile="${1}"
  currentDir=""

  relativeComponents=(
    ".." "Up one directory"
  )
  tmpFile="$(mktemp -t "$(basename "${zipFile}").XXXXXXXXXX")"
  trap 'rm "${tmpFile}"; _exit' SIGINT # Trap Ctrl+C (SIGINT) to clean up tmp file
  _infobox "Loading."
  zip -sf "${zipFile}" > "${tmpFile}"

  while true; do

    readarray -t currentDirDirs <<< "$( \
      grep -x "^  ${currentDir}[^/]*/$"  "${tmpFile}" |
      while read -r line; do
        line="${line#  }"
        echo -e "${line#"$currentDir"}\nDirectory"
      done )"
    [[ "${#currentDirDirs[@]}" -le "1" ]] && unset currentDirDirs

    readarray -t currentDirFiles <<< "$( \
      grep -x "^  ${currentDir}[^[:space:]][^/]*" "${tmpFile}" |
      while read -r line; do
        line="${line#  }"
        echo -e "${line#"$currentDir"}\nFile"
      done )"
    [[ "${#currentDirFiles[@]}" -le "1" ]] && unset currentDirFiles

    selected="$(msg="${currentDir}" _menu --backtitle "${title}" \
      --title "${zipFile}" -- "${relativeComponents[@]}" "${currentDirDirs[@]}" "${currentDirFiles[@]}")"
    exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && { rm "${tmpFile}" ; return "${exitcode}" ; }

    case "${selected,,}" in
    "..")
      [[ -z "${currentDir}" ]] && break
      [[ "${currentDir%/}" != *"/"* ]] && currentDir=""
      [[ -n "${currentDir}" ]] && currentDir="${currentDir%/*/}/"
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
  rm "${tmpFile}"
}

# Search for game folders, and list them
# Usage: _gameLocation
# Returns: path to selected folder
_gameLocation() {
  local gameLocations tag location item selected exitcode dir
  readarray -t gameLocations <<< "$( \
    find "${basedir}" -maxdepth 3 -type d -iregex '.*/\(usb[0-7]\|fat\|network\)\(/cifs\)?\(/games\|/_Arcade\)' |
      while read -r line; do
        tag="${line#"$basedir/"}"
        location="${tag%/*}"
        location="${location#/}"
        location="${location/fat\/cifs/Network}"
        location="${location/fat/SD Card}"
        item="${tag##*/} (${location%/*})"
        echo -e "${tag}\n${item#_}"
      done )"
  gameLocations=(
    "goto"  "Go to directory (keyboard required)"
    "fat"   "${underline}${bold}SD Card${reset}"
    "${gameLocations[@]}"
  )

  selected="$(_menu --default-item "fat" --colors -- "${gameLocations[@]}")"
  exitcode="${?}"; [[ "${exitcode}" -ge 1 ]] && return "${exitcode}"

  case "${selected}" in
    "goto")
      dir="$(_inputbox "Input a directory to go to" "${sdroot}")"
      echo "${dir%/}"
      ;;
    *)
      echo "${basedir}/${selected}"
      ;;
  esac
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

  _infobox "Present NFC tag to begin writing..."
  "${nfcCommand}" -write "${txt}" || { _error "Unable to write the NFC Tag"; return 1; }
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
    return 1
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
  local msg opts height width
  msg="${1}"
  shift
  opts=("${@}")
  dialog \
    --backtitle "${title}" \
    --aspect 0 "${opts[@]}" \
    --infobox "${msg}" \
    "${height:-0}" "${width:-0}"  3>&1 1>&2 2>&3 >"$(tty)" <"$(tty)"
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
