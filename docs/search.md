# Search

Search is an application to *search* for games on your MiSTer. It indexes all your games, lets you enter search queries without a keyboard, and then displays a list of results that can be launched directly.

*NOTE: Search is still a work in progress. Core functionality of indexing and searching works great, but the GUI is missing a lot of features like filtering, sorting, re-indexing etc. which will come later. Feel free to use it now though.*

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/search.sh"><img src="images/download.svg" alt="Download Search" title="Download Search" width="140"></a>

## Install

Download [Search](https://github.com/wizzomafizzo/mrext/releases/latest/download/search.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` script:
```
[mrext/search]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/search/search.json
```

## Usage

1. Run `search` from the MiSTer `Scripts` menu
2. Wait for Search to index your games (only happens on first launch)
3. Enter a search query and search (controller or keyboard works)
4. Select a game to launch from the list of results

## Updating the Index

At the moment, re-indexing of games must be triggered manually. You might need to do this if you've made changes to the games on your MiSTer.

Just delete this file: `/media/fat/search.db`

Search will create a new database next time you launch it.
