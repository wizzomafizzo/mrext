# PlayLog

PlayLog is an application to track and store stats of what games and cores you play on your MiSTer.

*NOTE: Still a work in progress. Core functionality works well but reporting is very basic. If you're into this idea, you can start using it right now to track stats until more interesting reports are created. You won't lose any of your stats with future updates.*

<a href="https://github.com/wizzomafizzo/mrext/raw/main/releases/playlog/playlog.sh"><img src="images/download.svg" alt="Download PlayLog" title="Download PlayLog" width="140"></a>

## Install

Enable the `recents` option in your `MiSTer.ini` file and reboot your MiSTer.

Download [PlayLog](https://github.com/wizzomafizzo/mrext/raw/main/releases/playlog/playlog.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` script:
```
[mrext/playlog]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/playlog/playlog.json
```

## Usage

*WARNING: As a power loss protection feature, by default, PlayLog will write a small amount of data to the SD card every 5 minutes while you're playing a game. This shouldn't be a problem with modern SD cards, but you can configure this value to be longer or disable it entirely. See further down for instructions.*

1. Run `playlog` from the MiSTer `Scripts` menu
2. Select Yes when asked to add PlayLog to the MiSTer boot script (only happens once)

From this point, PlayLog will always run on boot and silently track game playing stats in the background. At any point you can run `playlog` again and see a summary report of the stats.

## Configuration

PlayLog can be configured by creating a `playlog.ini` file in the `/media/fat/Scripts` folder where you put `playlog.sh`. For example:


```
[playlog]
save_every = 5
```

### Save Interval

| Key          | Default | 
| ------------ | ------- |
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
