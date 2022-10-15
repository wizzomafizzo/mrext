# PlayLog

PlayLog is an application to track and store stats of what games and cores you play on your MiSTer.

*NOTE: Still a work in progress. Core functionality works well but reporting is very basic. If you're into this idea, you can start using it right now to track stats until more interesting reports are created. You won't lose any of your stats with future updates.*

[![Download PlayLog](images/download.png "Download PlayLog")](https://github.com/wizzomafizzo/mrext/raw/main/releases/playlog/playlog.sh)

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

PlayLog can be configured by creating a `playlog.ini` file in the `/media/fat/Scripts` folder where you put `playlog.sh`.

At the moment, the only configurable option is the interval at which it saves data while playing. You can set that by changing the contents of `playlog.ini` to this:

```
[playlog]
save_every = 5
```

Change `5` to whatever number of minutes you want PlayLog to wait between saves. For example: `1` for every minute, `60` for every hour and so on. Set it to `0` to disable this feature entirely and only update stats during core change. Smaller values means less time "lost" after power loss.

PlayLog always saves the current stats when you exit to the MiSTer menu or launch a new game/core.
