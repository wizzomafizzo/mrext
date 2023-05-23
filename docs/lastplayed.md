# LastPlayed

LastPlayed is a simple service for automatically generating a shortcut in the MiSTer menu pointing to your most recently played game.

*Just want to see your recently played games in the menu? MiSTer has a great feature for that already. Enable the `recents` option in your `MiSTer.ini` file, and press the Select button on your controller while in the menu.*

[![Download LastPlayed](images/download.png "Download LastPlayed")](https://github.com/wizzomafizzo/mrext/raw/main/releases/lastplayed/lastplayed.sh)

## Install

Download [LastPlayed](https://github.com/wizzomafizzo/mrext/raw/main/releases/lastplayed/lastplayed.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` script:
```
[mrext/lastplayed]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/lastplayed/lastplayed.json
```

Once installed, run `lastplayed` from the MiSTer `Scripts` menu, and a prompt will offer to enable LastPlayed as a startup service. You may also be asked to manually enable the `recents` option in your `MiSTer.ini` file.

From now on, a shortcut named "Last Played" will be available in the menu, launching your most recently played game.

## Configuration

LastPlayed can be configured by creating a `lastplayed.ini` file in the `/media/fat/Scripts` folder where you put `lastplayed.sh`. For example:


```
[lastplayed]
name = Some other name
```

The name of the shortcut defaults to "Last Played", but can be changed with the `name` setting. Keep in mind these characters are not allowed in a filename: `\/:*?"<>|`

## Launching Games on MiSTer Startup

By using the `bootcore` feature in MiSTer, you can make the last played game launch automatically on MiSTer startup.

In your `MiSTer.ini` file, look for the line starting with `bootcore=` and change it to `bootcore=Last Played.mgl`. If you can't find this line, just add it to the end of the file.

If you configured a custom name for the shortcut, use that instead of `Last Played.mgl`.
