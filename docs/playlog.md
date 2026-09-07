# PlayLog

> [!IMPORTANT]
> PlayLog is maintained for legacy and stock-MiSTer workflows. Zaparoo Core also creates and maintains `/tmp/ACTIVEGAME` on MiSTer, so existing file readers can continue working. New MiSTer state and history integrations can use Core history, `media.active`, media notifications, or MQTT for richer cross-platform state. Executable state hooks have no direct Core replacement. See [MIGRATE.md](../MIGRATE.md).

PlayLog is an application to track and store stats of what games and cores you play on your MiSTer.

*NOTE: Still a work in progress. Core functionality works well but reporting is very basic. If you're into this idea, you can start using it right now to track stats until more interesting reports are created. You won't lose any of your stats with future updates.*

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/playlog.sh"><img src="images/download.svg" alt="Download PlayLog" title="Download PlayLog" width="140"></a>

## Arcade tracking

When Arcade Database recognizes a running arcade set name, the shared tracker resolves a unique installed MRA by its XML `<setname>`, including nested arcade folders. Game-start events and game state include that path while retaining the set name as the game identifier. Ambiguous or unreadable matches leave the path empty rather than guessing. The legacy arcade value in `/tmp/ACTIVEGAME` remains the set name. Existing history is not backfilled.

## Install

Enable the `recents` option in your `MiSTer.ini` file and reboot your MiSTer.

Download [PlayLog](https://github.com/wizzomafizzo/mrext/releases/latest/download/playlog.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add this database to `downloader.ini` in SD card root, then run `downloader` or `update` to receive updates:

```ini
[mrext/playlog]
db_url = https://raw.githubusercontent.com/wizzomafizzo/mrext/main/releases/playlog/playlog.json
```

## Usage

*WARNING: As a power loss protection feature, by default, PlayLog will write a small amount of data to the SD card every 5 minutes while you're playing a game. This shouldn't be a problem with modern SD cards, but you can configure this value to be longer or disable it entirely. See further down for instructions.*

1. Run `playlog` from the MiSTer `Scripts` menu
2. Select Yes when asked to add PlayLog to the MiSTer boot script (only happens once)

From this point, PlayLog will always run on boot and silently track game playing stats in the background. At any point you can run `playlog` again and see a summary report of the stats.

Service status and stop commands verify the recorded PID's executable and daemon arguments before trusting it, so an unrelated process reusing a stale PID is not treated as PlayLog. The PID-file format is unchanged.

PlayLog does not record menu navigation. Earlier versions wrote a database row for every cursor movement in the MiSTer menu, which put an SD card write behind each one and grew the `events` table without bound, for rows PlayLog never read back. Those rows are deleted and the database compacted the first time this version opens it; play history and totals are untouched. Remote still receives menu navigation over its websocket, which is the only place it was ever used.

PlayLog waits for MiSTer's state files to appear when it starts, so launching it from `user-startup.sh` before MiSTer main has created them no longer makes the service exit. It also recovers when another script replaces `/tmp/ACTIVEGAME` or `/tmp/CORENAME` instead of writing over them, which previously left tracking silent until a restart.

## Configuration

PlayLog can be configured by creating a `playlog.ini` file in the `/media/fat/Scripts` folder where you put `playlog.sh`. For example:


```
[playlog]
save_every = 5
```

### Save Interval

| Key          | Default | 
|--------------|---------|
| `save_every` | 5       |

This setting changes how often PlayLog will save time data to disk while in a core.

Change `5` to whatever number of minutes you want PlayLog to wait between saves. For example: `1` for every minute, `60` for every hour and so on. Set it to `0` to disable this feature entirely and only update stats during core change. Smaller values means less time "lost" after power loss.

PlayLog always saves the current stats when you exit to the MiSTer menu or launch a new game/core.

## Integrating with Scripts

A common requirement for scripts is to detect the currently running game which is somewhat complex to do reliably on MiSTer. PlayLog can do this for you, and offers a couple of ways to integrate your own scripts with it.

## Active Game

PlayLog creates a file called `/tmp/ACTIVEGAME` when it is running. All different ways of detecting a game change are funneled through to this file, which PlayLog itself uses to trigger updates. You can safely monitor this file for changes to see what the path to the current game is.

At this stage, MiSTer does not offer a way to detect the currently playing game if it has been launched directly through the `/dev/MiSTer_cmd` interface. If you have a script which does this, and you'd like to integrate it with PlayLog, you can simply check if `/tmp/ACTIVEGAME` exists at the time of game launch and then write the game's path to this file if it does. An absolute path is preferred, but PlayLog will do its best to resolve relative paths since they're very common in MiSTer.

An example of doing this with Bash: `[[ -e /tmp/ACTIVEGAME ]] && echo "/path/to/game" > /tmp/ACTIVEGAME`

## State Hooks

PlayLog offers 4 hooks which can launch a custom script or application:

- `on_core_start`: when a core is first started
- `on_core_stop`: when that core is stopped
- `on_game_start`: when a game is first started
- `on_game_stop`: when that game is stopped

These can be configured in the `playlog.ini` file by setting the same key name as above and a path to an executable file. For example: `on_game_start = /media/fat/Scripts/my_script.sh`

When the hook's condition is met, PlayLog will run the given executable with either a core's internal name or a game's absolute path as its first argument.

Be aware that stop hooks will be unreliable when an end user shuts down their MiSTer via power switch.

## Uninstall

Remove PlayLog's Downloader subscription if configured, so updates do not reinstall it.

1. Run `/media/fat/Scripts/playlog.sh -service stop` and wait for shutdown and database writes to finish. Remove the `# mrext/playlog` block and its `playlog.sh -service $1` command from `/media/fat/linux/user-startup.sh`.
2. Delete `/media/fat/Scripts/playlog.sh`. Optionally back up and remove `/media/fat/Scripts/playlog.ini` to discard settings and hook configuration.
3. **Preserve `/media/fat/playlog.db` unless you intend to erase your play history and totals.** After the service stops, back up the database together with any `playlog.db-wal`, `playlog.db-shm`, or `playlog.db-journal` sidecars present. Keep that set together; do not delete sidecars independently. Only discard the database and remaining sidecars when intentionally removing all history.

PlayLog generates no menu folder. Custom scripts referenced by its state hooks belong to you and may be shared; removing PlayLog does not require deleting them. Keep MiSTer recents/configuration and shared `/tmp/ACTIVEGAME`. After shutdown, `/tmp/playlog.pid`, `/tmp/playlog.log`, and `/tmp/playlog.sh` are optional cleanup.
