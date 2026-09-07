# Random

> [!NOTE]
> Random works without Zaparoo. For more selection options, Zaparoo's `launch.random` supports search patterns and tags. See [mrext and Zaparoo](../MIGRATE.md) to compare features.

Random is a simple application for launching a game at random from your MiSTer's collection.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/random.sh"><img src="images/download.svg" alt="Download Random" title="Download Random" width="140"></a>

## Install

Download [Random](https://github.com/wizzomafizzo/mrext/releases/latest/download/random.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add this database to `downloader.ini` in SD card root, then run `downloader` or `update` to receive updates:

```ini
[mrext/random]
db_url = https://raw.githubusercontent.com/wizzomafizzo/mrext/main/releases/random/random.json
```

## Usage

For basic usage, just run `random` from the MiSTer `Scripts` menu. A game will start running immediately.

By default, Random will pick a game from any system available with valid files. This can be configured to include or exclude certain systems. Each system has an equal chance of being selected, rather than it being based on the number of games in a system.

Random offers 2 command line flags to customise which systems are included during a scan of games to launch: `-filter` and `-ignore`.

Both arguments take a comma-separated list of system IDs from the [supported systems](systems.md) documentation.

The `-filter` flag will restrict the systems searched to only those specified. The `-ignore` flag does the opposite. Both flags can be used at the same time if desired.

Example of searching only Game Boy Advance, PSX, and NES: `random.sh -filter gba,psx,nes`

Example of Commodore 64 being ignored: `random.sh -filter all -ignore c64`

A `-noscan` flag is also available which will use a slightly faster but less random method to pick a game. It instead traverses folders at random until it finds a game, meaning results will be weighted by folder depth.

## Uninstall

Remove Random's Downloader subscription if configured, so updates do not reinstall it.

1. Let any Random invocation finish. It installs no startup service; remove custom startup, scheduled, or wrapper-script invocations you added yourself.
2. Delete `/media/fat/Scripts/random.sh`. Optionally back up and remove `/media/fat/Scripts/random.ini` if present, respecting any shared configuration override.
3. Remove only custom Random wrapper scripts you created and no longer need. Random has no dedicated persistent game database or generated menu tree to delete; keep original games and other apps' shortcuts.

Keep shared `.LASTLAUNCH.mgl` and `Scripts/.config/mrext/` files for other mrext apps. Removing Random does not stop a core it already launched.

## Custom Launchers

Random can be customised by creating your own shell scripts which call `random.sh` with the above arguments.

To create, for example, a launcher which only picks random PSX games:

1. Create a new file in `/media/fat/Scripts` called `random_psx.sh` (or anything with `.sh` on the end)
2. Set the contents of the file to:

   ```sh
   #!/bin/bash
   /media/fat/Scripts/random.sh -filter psx
   ```

Use the absolute path so the launcher works regardless of the current directory or whether `Scripts` is in `PATH`.

And that's it, you'll have a new entry in your `Scripts` menu to launch a random PSX game.