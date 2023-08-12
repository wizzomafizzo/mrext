# NFC

NFC is a service for launching games, cores and custom dynamic commands using a USB NFC card reader. 
All hardware required is inexpensive, easily available and quick to set up.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/nfc.sh"><img src="images/download.svg" alt="Download Remote" title="Download Remote" width="140"></a>

## Hardware required

The following hardware is currently known to work. Many other devices may work, but might also require a project 
update for proper support. Please [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) if you'd like to
add a working device to this list, or troubleshoot a device that isn't working.

This project uses the [libnfc](https://nfc-tools.github.io/projects/libnfc/) library, so any device supported by it 
should work.

### Readers

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
```

Create this file if it doesn't exist. Be aware the `ttyUSB0` part may be different if you have other devices connected
such as tty2oled.

## Setting up tags

Currently, the NFC script does not support writing to tags. You can use your phone instead to write data to them until
this support is added.

The NFC Tools app is highly recommended for this. It's free and supports both
[iOS](https://apps.apple.com/us/app/nfc-tools/id1252962749) and 
[Android](https://play.google.com/store/apps/details?id=com.wakdev.wdnfc&hl=en&gl=US).

You'll want to write a *Text record* with it for all the supported NFC service features.

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

#### system

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

#### random

This command will launch a game a random for the given system. For example:
```
**random:snes
```
This will launch a random SNES game each time you scan the tag.

You can also select all systems with `**random:all`.

#### command

**This feature is intentionally disabled for security reasons when run straight from a tag. You can still use it,
but only via the `nfc.csv` file explained below.**

This command will run a MiSTer Linux command directly. For example:
```
**command:reboot
```

### Mappings database

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
