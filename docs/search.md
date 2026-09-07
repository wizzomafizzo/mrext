# Search

> [!NOTE]
> Search works without Zaparoo. To search and launch games from your phone or browser, you can also use Zaparoo App or Web UI. If you're building a new search integration, use Core's API. See [mrext and Zaparoo](../MIGRATE.md) for details.

Search is an application to *search* for games on your MiSTer. It indexes all your games, lets you enter search queries without a keyboard, and then displays a list of results that can be launched directly.

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

The current shared index is `/media/fat/Scripts/.config/mrext/games.db`, not the legacy `/media/fat/search.db`. If Remote is installed, use its Search regenerate button to rebuild without first deleting the working index. Otherwise, exit Search and LaunchSync, stop Remote if running, then remove `games.db` and reopen Search to create a fresh index. Manual deletion gives up the previous-index fallback, so keep a backup if needed.

Full rebuilds remove stale game names and system metadata only after a successful scan. They stage bounded batches on SD before publishing the replacement, leaving an existing working index intact if regeneration fails. Mount all desired game libraries first and allow enough free space for a second index. Leave the adjacent `games.db.lock` in place while any app is indexing; it is shared writer-coordination state.

Only one indexer can write at a time. If Remote is already rebuilding the index, Search reports that another indexer is running instead of waiting; try again once it finishes.

A game folder that cannot be reached, such as a symlink into a drive that is not attached, is skipped rather than failing the whole run. Broken symlinks inside a library are skipped the same way, so one dead link no longer prevents indexing.

## Compatibility notes

Search now uses the same interface as BGM, Favorites and GamesMenu: the same border and title, the same footer help line, the same `↑↓ / ←→ / Enter / ESC` hints, and the themes, CRT mode and mouse settings from `[tui]`. Set them in `search.ini` or share them across every app; see the shared interface settings in [favorites.md](favorites.md).

The flow is unchanged. Search still opens straight onto the keyboard, and the result list still offers `Launch` (or `Select` with `-print`), `PgUp`, `PgDn`, `Options` and `Exit`, with results still shown as `[System] Game name` and duplicates removed.

Two things are better rather than merely different. The keyboard is the shared one, so it has lower case and a symbol layer; the old one accepted only digits, capitals, space and delete. And `Options` explains that updating the database rescans every games folder and takes several minutes, and asks before starting.

## Uninstall

Remove Search's Downloader subscription if configured, so updates do not reinstall it.

1. Finish any indexing operation and exit Search. It installs no startup service; remove launch commands you added manually, if any.
2. Delete `/media/fat/Scripts/search.sh`. Optionally back up and remove `/media/fat/Scripts/search.ini` if present, respecting any shared configuration override.
3. Keep `/media/fat/Scripts/.config/mrext/games.db` while Remote or LaunchSync uses it. If none of those apps remain, the index can be removed after they have stopped; it contains rebuildable game names and paths, not ROMs. Older releases may still use `/media/fat/search.db`, so check before deleting that legacy file.

Search creates no dedicated menu tree or history database. Preserve your games, existing shortcuts, shared `.LASTLAUNCH.mgl`, and `Scripts/.config/mrext/ArcadeDatabase.csv`.
