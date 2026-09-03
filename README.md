# MiSTer Extensions

> [!IMPORTANT]
> mrext is maintained for legacy users. Before starting or extending a MiSTer project with mrext, review current Zaparoo replacements in [MIGRATE.md](MIGRATE.md). Existing installations can keep working, and mrext can remain appropriate where Zaparoo does not replace required stock-MiSTer workflows.

Extensions and utilities for [MiSTer](https://github.com/MiSTer-devel/Main_MiSTer/wiki).

Check linked documentation for each script. Most work immediately, but some require manual setup.

[Remote](#remote) • [BGM](#bgm) • [Favorites](#favorites) • [GamesMenu](#gamesmenu) • [LastPlayed](#lastplayed) • [LaunchSync](#launchsync) • [PlayLog](#playlog) • [Random](#random) • [Search](#search)

[Supported Systems](docs/systems.md) • [Developer Guide](docs/dev.md)

## Install

### Update All

Run [Update All](https://github.com/theypsilon/Update_All_MiSTer), press Up during its startup countdown to open Settings, then enable **MiSTer Extensions (wizzo)** under **Other Tools & Scripts**. Select **Save** before leaving Settings.

### Downloader

To install everything through MiSTer Downloader, add this database to `downloader.ini` in SD card root:

```ini
[mrext/all]
db_url = https://raw.githubusercontent.com/wizzomafizzo/mrext/main/releases/all.json
```

Run either `downloader` or `update` from MiSTer Scripts menu. Both names launch MiSTer Downloader.

Each application also has an individual database listed in its documentation.

### Manual

Use Download button for an application, copy downloaded script to SD card's `Scripts` folder, then run it from MiSTer Scripts menu.

## Remote

Control MiSTer from any device on your network. Remote provides a web interface for input, launching, menu management, search, screenshots, settings, wallpapers, and other system functions.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/remote.sh"><img src="docs/images/download.svg" alt="Download Remote" title="Download Remote" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/blob/main/docs/remote.md"><img src="docs/images/readme.svg" alt="Readme Remote" title="Readme Remote" width="140"></a>

## BGM

Play music in MiSTer menu. BGM supports common audio formats and internet radio streams, and can pause automatically while a core is running.

<a href="https://github.com/wizzomafizzo/MiSTer_BGM/raw/main/bgm.sh"><img src="docs/images/download.svg" alt="Download BGM" title="Download BGM" width="140"></a>
<a href="https://github.com/wizzomafizzo/MiSTer_BGM"><img src="docs/images/readme.svg" alt="Readme BGM" title="Readme BGM" width="140"></a>

## Favorites

Create and manage shortcuts for favorite games and cores in MiSTer menu.

<a href="https://github.com/wizzomafizzo/MiSTer_Favorites/raw/main/favorites.sh"><img src="docs/images/download.svg" alt="Download Favorites" title="Download Favorites" width="140"></a>
<a href="https://github.com/wizzomafizzo/MiSTer_Favorites"><img src="docs/images/readme.svg" alt="Readme Favorites" title="Readme Favorites" width="140"></a>

## GamesMenu

Browse a game collection from MiSTer menu. GamesMenu scans games and creates launchers matching collection folder layout.

<a href="https://github.com/wizzomafizzo/MiSTer_GamesMenu/raw/main/gamesmenu.sh"><img src="docs/images/download.svg" alt="Download GamesMenu" title="Download GamesMenu" width="140"></a>
<a href="https://github.com/wizzomafizzo/MiSTer_GamesMenu"><img src="docs/images/readme.svg" alt="Readme GamesMenu" title="Readme GamesMenu" width="140"></a>

## LastPlayed

Generate auto-updating shortcuts for recently played games and last played game.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/lastplayed.sh"><img src="docs/images/download.svg" alt="Download LastPlayed" title="Download LastPlayed" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/blob/main/docs/lastplayed.md"><img src="docs/images/readme.svg" alt="Readme LastPlayed" title="Readme LastPlayed" width="140"></a>

## LaunchSync

Create shareable game playlists that generate MiSTer menu shortcuts and can update from published `.sync` files.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/launchsync.sh"><img src="docs/images/download.svg" alt="Download LaunchSync" title="Download LaunchSync" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/blob/main/docs/launchsync.md"><img src="docs/images/readme.svg" alt="Readme LaunchSync" title="Readme LaunchSync" width="140"></a>

## PlayLog

Track play history and time for games and cores on MiSTer.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/playlog.sh"><img src="docs/images/download.svg" alt="Download PlayLog" title="Download PlayLog" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/blob/main/docs/playlog.md"><img src="docs/images/readme.svg" alt="Readme PlayLog" title="Readme PlayLog" width="140"></a>

## Random

Launch a random game from MiSTer Scripts menu.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/random.sh"><img src="docs/images/download.svg" alt="Download Random" title="Download Random" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/blob/main/docs/random.md"><img src="docs/images/readme.svg" alt="Readme Random" title="Readme Random" width="140"></a>

## Search

Search and launch games from a controller-friendly interface.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/search.sh"><img src="docs/images/download.svg" alt="Download Search" title="Download Search" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/blob/main/docs/search.md"><img src="docs/images/readme.svg" alt="Readme Search" title="Readme Search" width="140"></a>
