# Random

Random is a simple application for launching a game at random from your MiSTer's collection.

[![Download Random](images/download.png "Download Random")](https://github.com/wizzomafizzo/mrext/raw/main/releases/random/random.sh)

## Install

Download [Random](https://github.com/wizzomafizzo/mrext/raw/main/releases/random/random.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` script:
```
[mrext/random]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/random/random.json
```

## Usage

For basic usage, just run `random` from the MiSTer `Scripts` menu. A game will start running immediately.

By default, Random will pick a game from any system available with valid files. This can be configured to include or exclude certain systems. Each system has an equal chance of being selected, rather than it being based on the number of games in a system.

Random offers 2 command line flags to customise which systems are included during a scan of games to launch: `-filter` and `-ignore`.

Both arguments take a comma-separated list of system IDs from the [supported systems](systems.md) documentation.

The `-filter` flag will restrict the systems searched to only those specified. The `-ignore` flag does the opposite. Both flags can be used at the same time if desired.

Example of only Gameboy Advance, PSX and NES being search: `random.sh -filter gba,psx,nes`

Example of Commodore 64 being ignored: `random.sh -filter all -ignore c64`

A `-noscan` flag is also available which will use a slightly faster but less random method to pick a game. It instead traverses folders at random until it finds a game, meaning results will be weighted by folder depth.

## Custom Launchers

Random can be customised by creating your own shell scripts which call `random.sh` with the above arguments.

To create, for example, a launcher which only picks random PSX games:

1. Create a new file in `/media/fat/Scripts` called `random_psx.sh` (or anything with `.sh` on the end)
2. Set the contents of the file to:

   ```
   #!/bin/bash
   random.sh -filter psx
   ```

And that's it, you'll have a new entry in your `Scripts` menu to launch a random PSX game.