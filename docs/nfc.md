# NFC

NFC is a service for launching games, cores and custom dynamic commands using a USB NFC card reader. 
All hardware required is inexpensive, easily available and quick to set up.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/nfc.sh"><img src="images/download.svg" alt="Download Remote" title="Download Remote" width="140"></a>

<!-- TOC -->
* [Thanks to](#thanks-to)
* [Card labels](#card-labels)
* [Hardware required](#hardware-required)
  * [Readers](#readers)
  * [Tags](#tags)
* [Install](#install)
  * [Hardware configuration](#hardware-configuration)
* [Setting up tags](#setting-up-tags)
  * [Combining commands](#combining-commands)
  * [Launching games and cores](#launching-games-and-cores)
  * [Custom commands](#custom-commands)
    * [Launch a system (system)](#launch-a-system-system)
    * [Launch a random game (random)](#launch-a-random-game-random)
    * [Change the actve MiSTer.ini file (ini)](#change-the-actve-misterini-file-ini)
    * [Make an HTTP request to a URL (get)](#make-an-http-request-to-a-url-get)
    * [Press a keyboard key (key)](#press-a-keyboard-key-key)
    * [Insert a coin/credit (coinp1/coinp2)](#insert-a-coincredit-coinp1coinp2)
    * [Run a system/Linux command (command)](#run-a-systemlinux-command-command)
  * [Mappings database](#mappings-database)
  * [Writing to tags](#writing-to-tags)
  * [Reading tags](#reading-tags)
<!-- TOC -->

## Thanks to

- [symm](https://github.com/symm) - for doing all the actual (hard) work of making the NFC scanners function on MiSTer.
- [ElRojo](https://github.com/ElRojo/MiSTerRFID) & [javiwwweb](https://github.com/javiwwweb/MisTerRFID) - for project inspiration.

## Card labels

- [Arcade & Computer core artwork collated and edited by NML32](https://mega.nz/folder/vH5WGSJI#UANuzi-5uG9XBqddPeApmw)

Feel free to [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) to add to this list.

## Hardware required

The following hardware is currently known to work. Many other devices may work, but might also require a project 
update for proper support. Please [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) if you'd like to
add a working device to this list, or troubleshoot a device that isn't working.

This project uses the [libnfc](https://nfc-tools.github.io/projects/libnfc/) library, so any device supported by it 
should work.

### Readers

**WARNING: There is a certain version of clone of the ACR122U reader which is not compatible with the script. At this stage it's impossible to tell which version to avoid from a shop listing, and no ETA on support for the clone revision. Most listings are fine, but be aware of the risk. Your best bet is to not buy the literal cheapest listing available.**

These are some known okay listings:
- https://www.amazon.com/dp/B07KRKPWYC
- https://www.ebay.co.uk/itm/145044206870

Feel free to [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) to add to this list.

| Device                 | Details                                                                                                                                            |
|------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| ACR122U USB NFC reader | <ul><li>Plug and play</li><li>Cheap</li><li>Littered on Amazon, eBay and AliExpress</li></ul>                                                      |
| PN532 NFC module       | <ul><li>Really cheap</li><li>Small</li><li>Requires a USB to TTL cable</li><li>Some manual configuration</li><li>Possibly some soldering</li></ul> |

### Tags

The form factor of the tag is up to you. Can be a card, sticker, keychain, etc.

| Device                 | Details                                            |
|------------------------|----------------------------------------------------|
| NTAG213                | 144 bytes storage                                  |
| NTAG215                | 504 bytes storage                                  |
| NTAG216                | 888 bytes storage                                  |
| MIFARE Classic 1K      | 716 bytes storage, often ships with readers        |
| Amiibo                 | Supported using the `nfc.csv` file described below |

Custom NFC commands can be written to NTAG213 without issue, but keep storage size in mind if you have a large
collection of games with deep folders. The tag may need to store the whole game path.

## Install

Download [NFC](https://github.com/wizzomafizzo/mrext/releases/latest/download/nfc.sh) and copy it to the `Scripts`
folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` or
`downloader` script:
```
[mrext/nfc]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/nfc/nfc.json
```

Once installed, run `nfc` from the MiSTer `Scripts` menu, a prompt will offer to enable NFC as a startup service, then
the service will be started in the background.

After the initial setup is complete, a status display will be shown. It's ok to exit this screen, the service will
continue to run in the background.

### Hardware configuration

Your reader may work out of the box with no extra configuration. Run `nfc` from the `Scripts` menu, plug it in, and
check if it shows as connected in the log view.

If you are using a PN532 NFC module connected with a USB to TTL cable, then the following config may be needed 
in `nfc.ini` in the `Scripts` folder:

```
[nfc]
connection_string="pn532_uart:/dev/ttyUSB0"
allow_commands=no
```

Create this file if it doesn't exist. Be aware the `ttyUSB0` part may be different if you have other devices connected
such as tty2oled.

## Configuration

The NFC script supports a `nfc.ini` file in the `Scripts` folder. This file can be used to configure the NFC service.

If one doesn't exist, create a new one. This example has all the default values:
```
[nfc]
connection_string=""
allow_commands=no
```

All lines except the `[nfc]` header are optional.

### connection_string

See [Hardware configuration](#hardware-configuration) for details. This option is for configuration of [libnfc](https://github.com/nfc-tools/libnfc)
and it currently required for the PN532 module.

### allow_commands

Enables the [command](#run-a-systemlinux-command-command) custom command to be triggered from a tag. By default this is
disabled and only works from the `nfc.csv` file described below.

## Setting up tags

The NFC Tools app is highly recommended for this. It's free and supports both
[iOS](https://apps.apple.com/us/app/nfc-tools/id1252962749) and 
[Android](https://play.google.com/store/apps/details?id=com.wakdev.wdnfc&hl=en&gl=US).

You'll want to write a *Text record* with it for all the supported NFC service features.

### Combining commands

All commands and game/core launches can be combined on a single tag if space permits using the `||` separator.

For example, to switch to MiSTer.ini number 3 and launch the SNES core:
```
**ini:3||_Console/SNES
```

Or launch a game and notify an HTTP service:
```
_Console/SNES||**get:https://example.com
```

As many of these can be combined as you like.

### Launching games and cores

The NFC script supports launching game files, core .RBF files, arcade .MRA files and .MGL shortcut files. This is
done by simply writing the path to the file to the tag.

For example, to launch a game, write something like this to the tag:
```
/media/fat/games/Genesis/1 US - Q-Z/Road Rash (USA, Europe).md
```

To save space and to handle games moving between storage devices, you can also use a relative path:
```
Genesis/1 US - Q-Z/Road Rash (USA, Europe).md
```

This will search for the file in all standard MiSTer game folder paths including CIFS.

Some other examples:
```
_Arcade/1942 (Revision B).mra
```
```
_@Favorites/Super Metroid.mgl
```

Because core filenames often change, it's supported to use the same short name as in a .MGL file to launch it:
```
_Console/PSX
```

.ZIP files are also supported natively, same as they are in MiSTer itself. Just treat the .ZIP file as a folder name:
```
Genesis/@Genesis - MegaSD Mega EverDrive 2022-05-18.zip/1 US - Q-Z/Road Rash (USA, Europe).md
```

### Custom commands

There are a small set of special commands that can be written to tags to perform dynamic actions. These are marked in
a tag by putting `**` at the start of the stored text.

#### Launch a system (system)

This command will launch a system, based on MiSTer Extensions own internal list of system IDs
[here](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md). This can be useful for "meta systems" such as
Atari 2600 and WonderSwan Color which don't have their own core .RBF file.

For example:
```
**system:Atari2600
```
```
**system:WonderSwanColor
```

It also works for any other system if you prefer this method over the standard core .RBF file one.

#### Launch a random game (random)

This command will launch a game a random for the given system. For example:
```
**random:snes
```
This will launch a random SNES game each time you scan the tag.

You can also select all systems with `**random:all`.

#### Change the actve MiSTer.ini file (ini)

Loads the specified MiSTer.ini file and relaunches the menu core if open.

Specify the .ini file with its index in the list shown in the MiSTer menu. Numbers `1` to `4`.

For example:
```
**ini:1
```

This switch will not persist after a reboot, same as loading it through the OSD.

#### Make an HTTP request to a URL (get)

Perform an HTTP GET request to the specified URL. For example:
```
**get:https://example.com
```

This is useful for triggering webhooks or other web services.

It can be combined with other commands using the `||` separator. For example:
```
**get:https://example.com||_Console/SNES
```

This does *not* check for any errors, and will not show any output. You send the request and off it goes into the ether.

#### Press a keyboard key (key)

Press a key on the keyboard using its uinput code. For example (to press F12 to bring up the OSD):
```
**key:88
```

See a full list of key codes [here](https://pkg.go.dev/github.com/bendahl/uinput@v1.6.0#pkg-constants).

#### Insert a coin/credit (coinp1/coinp2)

Insert a coin/credit for player 1 or 2. For example (to insert 1 coin for player 1):
```
**coinp1:1
```

This command presses the `5` and `6` key on the keyboard respectively, which is generally accepted as the coin insert
keys in MiSTer arcade cores. If it doesn't work, try manually mapping the coin insert keys in the OSD.

It also supports inserting multiple coins at once. For example (to insert 3 coins for player 2):
```
**coinp2:3
```

#### Run a system/Linux command (command)

**This feature is intentionally disabled for security reasons when run straight from a tag. You can still use it,
but only via the `nfc.csv` file explained below or by enabling the `allow_commands` option in `nfc.ini`.**

This command will run a MiSTer Linux command directly. For example:
```
**command:reboot
```

## Mappings database

The NFC script supports a `nfc.csv` file in the top of the SD card. This file can be used to override the text read
from a tag and map it to a different text value. This is useful for mapping Amiibos which are read-only, testing text
values before actually writing them, and is necessary for using the `command` custom command.

Create a file called `nfc.csv` in the top of the SD card, with this as the header:
```csv
match_uid,match_text,text
```

You'll then need to either power cycle your MiSTer, or restart the NFC service by running `nfc` from the `Scripts`
menu, selecting the `Stop` button, then the `Start` button.

After the file is created, the service will automatically reload it every time it's updated.

Here's an example `nfc.csv` file that maps several Amiibos to different functions:
```csv
match_uid,match_text,text
04e5c7ca024980,,**command:reboot
04078e6a724c80,,_#Favorites/Final Fantasy VII.mgl
041e6d5a983c80,,_#Favorites/Super Metroid.mgl
041ff6ea973c81,,_#Favorites/Legend of Zelda.mgl
```

Only one `match_` column is required for an entry, and the `match_uid` can include colons and uppercase characters.
You can get the UID of a tag by checking the output in the `nfc` Script display or on your phone.

## Writing to tags

The NFC script currently supports writing to NTAG tags through the command line option `-write <text>`.

For example, from the console or SSH:
```
/media/fat/Scripts/nfc.sh -write "_Console/SNES"
```
This will write the text `_Console/SNES` to the next detected tag.

This is available to any script or application on the MiSTer.

## Reading tags

Whenever a tag is successfully scanned, its UID and text contents (if available) will be written to the
file `/tmp/NFCSCAN`. The contents of the file is in the format `<uid>,<text>`.

You can monitor the file for changes to detect when a tag is scanned with the `inotifywait` command that is
shipped on the MiSTer Linux image. For example:
```
while inotifywait -e modify /tmp/NFCSCAN; do
    echo "Tag scanned"
done
```

## Service socket

When the NFC service is active, a Unix socket is created at `/tmp/nfc.sock`. This socket can be used to send
commands to the service.

Commands can be sent in a shell script like this:
```bash
echo "status" | socat - UNIX-CONNECT:/tmp/nfc.sock
```

### status

Returns the current status of the service with the following information:

- Last card scanned date in Unix epoch format
- Last card scanned UID
- Whether launching is enabled
- Last card scanned text

Each value is separated by a comma. For example:

```
1695650197,04faa9d2295880,true,**random:psx
```

### enable

Enables launching from tags (the default state).

This command has no output.

### disable

Disables launching from tags. Cards will scan and log, but no action will be triggered.

This command has no output.
