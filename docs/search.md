# Search

> [!IMPORTANT]
> Search is maintained for its controller-driven stock-OSD workflow. Before using it for new MiSTer search or library work, consider Core's media index and `media.search`, `zaparoo-cli media search`, or the search interfaces in App, Web UI, and Frontend. See [MIGRATE.md](../MIGRATE.md).

Search is an application to *search* for games on your MiSTer. It indexes all your games, lets you enter search queries without a keyboard, and then displays a list of results that can be launched directly.

*NOTE: Search is still a work in progress. Core functionality of indexing and searching works great, but the GUI is missing a lot of features like filtering, sorting, re-indexing etc. which will come later. Feel free to use it now though.*

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/search.sh"><img src="images/download.svg" alt="Download Search" title="Download Search" width="140"></a>

## Install

Download [Search](https://github.com/wizzomafizzo/mrext/releases/latest/download/search.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add this database to `downloader.ini` in SD card root, then run `downloader` or `update` to receive updates:

```ini
[mrext/search]
db_url = https://raw.githubusercontent.com/wizzomafizzo/mrext/main/releases/search/search.json
```

## Usage

1. Run `search` from the MiSTer `Scripts` menu
2. Wait for Search to index your games (only happens on first launch)
3. Enter a search query and search (controller or keyboard works)
4. Select a game to launch from the list of results

## Updating the Index

To force a rebuild, first exit Search and LaunchSync and stop Remote if it is running. After all users have stopped, remove the current shared index at `/media/fat/Scripts/.config/mrext/games.db`, then launch Search to rebuild it. This affects Search, Remote, and LaunchSync, not your game files. `/media/fat/search.db` is a legacy index, not the current database.

## Uninstall

Remove Search's Downloader subscription if configured, so updates do not reinstall it.

1. Finish any indexing operation and exit Search. It installs no startup service; remove launch commands you added manually, if any.
2. Delete `/media/fat/Scripts/search.sh`. Optionally back up and remove `/media/fat/Scripts/search.ini` if present, respecting any shared configuration override.
3. Keep `/media/fat/Scripts/.config/mrext/games.db` while Remote or LaunchSync uses it. If none of those apps remain, the index can be removed after they have stopped; it contains rebuildable game names and paths, not ROMs. Older releases may still use `/media/fat/search.db`, so check before deleting that legacy file.

Search creates no dedicated menu tree or history database. Preserve your games, existing shortcuts, shared `.LASTLAUNCH.mgl`, and `Scripts/.config/mrext/ArcadeDatabase.csv`.
