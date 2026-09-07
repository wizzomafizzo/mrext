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

## Arcade games

LastPlayed creates `.mra` links for arcade games in Last Played and Recently Played, including MRAs in nested installed arcade folders. The shared tracker uses Arcade Database to recognize the active set name, then confirms it against each candidate MRA's `<setname>`; display names alone do not select a launcher. Symlink aliases to the same file count once. Multiple distinct matching MRAs are treated as ambiguous rather than choosing one arbitrarily.

If Arcade Database is unavailable, or no unique readable MRA can be found, existing shortcuts remain untouched. Successful resolutions are cached until the tracker's name map is reloaded or the service restarts, and cached files are revalidated before use. Restart after adding duplicate or replacement MRAs to force a fresh scan.

Arcade shortcuts are `.mra` links, so the Last Played shortcut becomes `Last Played.mra` after an arcade game and `Last Played.mgl` after any other game. Only one is kept: the other extension is removed so a stale shortcut cannot linger in the menu or be launched by `bootcore`. Set `bootcore` to the extension you actually use.

Recently Played numbers `.mra` and `.mgl` shortcuts together. Replaying a game refreshes its entry without deleting the newly numbered shortcut. `/tmp/ACTIVEGAME` continues to contain the legacy arcade set name; tracker events carry the resolved launchable path separately. This works on stock MiSTer without Zaparoo Core.

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

The name of the folder which contains the recently played games. LastPlayed adds the menu prefix `_` on disk: `recent_folder_name = Game History` creates `/media/fat/_Game History`. Restart the LastPlayed service after changing `lastplayed.ini`; it is not reloaded while running. Changing this setting creates/uses the new folder and does not move or delete the old folder.

Keep in mind these characters are not allowed in a filename: `\/:*?"<>|`

### Disable Recent Folder

| Key                      | Default |
|--------------------------|---------|
| `disable_recent_folder`  | no      |

If set to `yes`, the recent folder will not be created.

## Troubleshooting recent launchers

Current MiSTer resolves MGL media paths from the core's games directory, not from the directory containing the MGL. RBF paths are resolved from the MiSTer root. Moving a current generated launcher into Recently Played therefore does not require adding another `../` to its paths. Root and recent shortcuts use the same shared generator.

After updating LastPlayed, launch an affected game normally to regenerate its Last Played and recent shortcuts. This refreshes that game's entry, not every existing shortcut. Back up old entries before manual changes; there is no need to delete the entire recent folder. Unnumbered user launchers are left alone.

If a recent shortcut still loads only the core, keep a copy of the failing MGL before replaying the game. Compare it with the regenerated root shortcut, and report both files, their locations, the full media path, and the MiSTer/core versions. Local regression fixtures cover NES/PSX, renamed folders, deep media paths, spaces/punctuation, and replay regeneration; they do not replace a device reproduction of a location-sensitive failure.

## Launching Last Played Game on MiSTer Startup

By using the `bootcore` feature in MiSTer, you can make the last played game launch automatically on MiSTer startup.

In your `MiSTer.ini` file, look for the line starting with `bootcore=` and change it to `bootcore=Last Played.mgl`. If you can't find this line, just add it to the end of the file.

If you configured a custom name for the shortcut, use that instead of `Last Played.mgl`.

## Uninstall

Remove LastPlayed's Downloader subscription if configured, so updates do not reinstall it.

1. Run `/media/fat/Scripts/lastplayed.sh -service stop` and wait for shutdown. Remove the `# mrext/lastplayed` block and its `lastplayed.sh -service $1` command from `/media/fat/linux/user-startup.sh`.
2. If `bootcore` in any MiSTer INI points at the Last Played shortcut, disable or change that setting before removing the shortcut. Do not delete the INI or disable `recents` just to uninstall LastPlayed; other apps may use it.
3. Delete `/media/fat/Scripts/lastplayed.sh`. Optionally back up and remove `/media/fat/Scripts/lastplayed.ini` to discard settings.
4. Optionally remove `/media/fat/Last Played.mgl` and generated entries in `/media/fat/_Recently Played/`. Use the actual configured names (`last_played_name`, legacy `name`, and `recent_folder_name`) if customized. Inspect the folder for user-added files before deleting anything. Keeping the shortcuts is allowed; they will stop updating.

LastPlayed has no separate database. Keep MiSTer's recents files under `config/` and shared `/tmp/ACTIVEGAME`. After shutdown, `/tmp/lastplayed.pid`, `/tmp/lastplayed.log`, and `/tmp/lastplayed.sh` are optional cleanup.
