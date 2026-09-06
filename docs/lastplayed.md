# LastPlayed

> [!IMPORTANT]
> LastPlayed is maintained for stock-menu shortcuts. Before using it for a new MiSTer recents workflow, consider Frontend recents, Core play history, or ZapScript `launch.last`. Keep LastPlayed when dynamic `.mgl` or `bootcore` files are required. See [MIGRATE.md](../MIGRATE.md).

LastPlayed is a service for automatically generating dynamic shortcuts in the MiSTer menu.

It supports:
- Creating and auto-updating a menu folder of recently played games
- A single shortcut that always launches the last played game

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/lastplayed.sh"><img src="images/download.svg" alt="Download LastPlayed" title="Download LastPlayed" width="140"></a>

## Install

Enable the `recents` option in your `MiSTer.ini` file.

Download [LastPlayed](https://github.com/wizzomafizzo/mrext/releases/latest/download/lastplayed.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add this database to `downloader.ini` in SD card root, then run `downloader` or `update` to receive updates:

```ini
[mrext/lastplayed]
db_url = https://raw.githubusercontent.com/wizzomafizzo/mrext/main/releases/lastplayed/lastplayed.json
```

Once installed, run `lastplayed` from the MiSTer `Scripts` menu, and a prompt will offer to enable LastPlayed as a startup service.

Service status and stop commands verify the recorded PID's executable and daemon arguments before trusting it, so an unrelated process reusing a stale PID is not treated as LastPlayed. The PID-file format is unchanged.

## Configuration

LastPlayed can be configured by creating a `lastplayed.ini` file in the `/media/fat/Scripts` folder where you put `lastplayed.sh`. For example:

```
[lastplayed]
last_played_name = Last Played
disable_last_played = no
recent_folder_name = Recently Played
disable_recent_folder = no
```

These are the default settings, and you can omit any lines you don't want to change.

### Last Played Name

| Key                | Default     | 
|--------------------|-------------|
| `last_played_name` | Last Played |

The name of the shortcut which launches the last played game.

Keep in mind these characters are not allowed in a filename: `\/:*?"<>|`

### Disable Last Played

| Key                   | Default |
|-----------------------|---------|
| `disable_last_played` | no      |

If set to `yes`, the last played shortcut will not be created.

### Recent Folder Name

| Key                    | Default           |
|------------------------|-------------------|
| `recent_folder_name`   | Recently Played   |

The name of the folder which contains the recently played games.

Keep in mind these characters are not allowed in a filename: `\/:*?"<>|`

### Disable Recent Folder

| Key                      | Default |
|--------------------------|---------|
| `disable_recent_folder`  | no      |

If set to `yes`, the recent folder will not be created.

## Launching Last Played Game on MiSTer Startup

By using the `bootcore` feature in MiSTer, you can make the last played game launch automatically on MiSTer startup.

In your `MiSTer.ini` file, look for the line starting with `bootcore=` and change it to `bootcore=Last Played.mgl`. If you can't find this line, just add it to the end of the file.

If you configured a custom name for the shortcut, use that instead of `Last Played.mgl`.
