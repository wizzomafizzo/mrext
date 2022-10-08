# LaunchSync

LaunchSync allows people to create, share and maintain live-updating game playlists for the MiSTer.

You create a [sync file](#sync-files) with a list of games, someone copies the file to their MiSTer, and LaunchSync will use it to generate a working list of game shortcuts in the MiSTer main menu. Sync files are subscriptable, so you can publish changes to your playlist and people will see your updates on their own system.

[![Download LaunchSync](download.png "Download LaunchSync")](https://github.com/wizzomafizzo/mrext/raw/main/releases/launchsync/launchsync.sh)

## Install

Download [LaunchSync](https://github.com/wizzomafizzo/mrext/raw/main/releases/launchsync/launchsync.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` script:
```
[mrext/launchsync]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/launchsync/launchsync.json
```

## Usage

1. Place at least one sync file in the root of your SD card or in a menu folder. Menu folders are folders which start with an underscore (`_`)
2. Run `launchsync` from the MiSTer `Scripts` menu

LaunchSync will search for all sync files on the MiSTer, check for sync file updates online, create folders for new sync files, and then create or update all shortcuts for listed games.

*NOTE: Currently LaunchSync must be run manually to update. In the future, it will be possible to have it run on MiSTer startup and automatically sync shortcuts in the background.*

### Sync Files

LaunchSync requires sync files to actually do anything. These are text files ending in `.sync` which define the name of a playlist, the games in it and how to find them on your own system. You can create your own or find sync files other people have created. An example is the [Discord Game of the Month](https://raw.githubusercontent.com/wizzomafizzo/mrext/main/cmd/launchsync/examples/Discord%20Game%20of%20the%20Month.sync) playlist hosted here.

## Creating Sync Files

*NOTE: Check the [Systems](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md) page to see what cores are supported. Most consoles are, most computers aren't. Use the ID or Alias listed on that page for the `system` field.*

Currently sync files must be created manually, though in most cases they're quite simple. If you want your sync file to auto-update, it also needs to hosted somewhere publicly. GitHub is a good choice for this but anywhere will work. The [Discord Game of the Month](https://raw.githubusercontent.com/wizzomafizzo/mrext/main/cmd/launchsync/examples/Discord%20Game%20of%20the%20Month.sync) file is a good base to edit and make your own.

These next sections will go through each part of the [template.sync](https://github.com/wizzomafizzo/mrext/blob/main/cmd/launchsync/examples/template.sync) file and explain in detail how each field works. It isn't a requirement to read this to create your own files, but it will show some more advanced features.

As you create a sync file, you can test it with this command:

`/media/fat/Scripts/launchsync.sh -test /path/to/my/file.sync`

This will make sure all fields are correct and show you a summary of search results for each game. It won't write any changes to disk.

### Header

The header section is the fields at the top of the files that don't have a `[Section]` line.

Field names are all case-sensitive.

```
name = My Awesome List
```

The `name` field is required. It's both for information in the UI and will also be the name of the folder created that will contain all the game shortcuts. It can't have any of these characters in the name, they'll be stripped out: `/ \ : * ? " < > |`

```
author = Me
```

The `author` field is also required. It's only used to show information in the UI.

```
url = https://example.com/example.sync
```

The `url` field is optional, but is required if you want your sync file to auto-update itself online. It should link back to itself. The `-test` flag will report if the link is accessible.

```
updated = 2022-09-01
```

The `updated` field is optional unless the `url` field exists. It's used to check if an update is actually required. It can be in the format `YYYY-MM-DD` or `YYYY-MM-DD hh:mm`.

### Game Sections

Each game in a sync file is defined with a section header:

```
[My Favorite Game]
```

This starts the section, and also defines the name of the shortcut that will be created in the MiSTer menu. It should be unique but there can be as many as you want. Same as the `name` field in the header, it can't contain any of these characters: `/ \ : * ? " < > |`

```
system = NES
```

The `system` field specifies where to look for games for the game entry and what launch arguments will be used for the core. It is required. See the [Systems](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md) page for a list of what system IDs are valid. The ID or alias will work here.

It's also ok to just enter the core's folder name as an ID which will always match up correctly. But you can be more specific using an ID from the list for cores that support multiple systems.

```
match = Cool Game (USA)
```

The `match` field is also required. This field acts as a search query for finding the game on the end user's MiSTer. The value entered is case-insensitive and only matches on a game's filename excluding the extension. This is usually straightforward, but you'll need to use your best judgement on a query that will match on the correct file but also work on different setups. This is basically the most important part of a sync file.

```
match = Cool Game (Europe)
match = Cool Game
```

A single game entry can have multiple `match` fields. When searching for a game, LaunchSync will try each query top to bottom in sequence until a match is found. This is useful if you have a very specific file in mind, but are ok with a fallback option.

If a game is not found, a placeholder shortcut is created in the menu. It won't work but it will let the user know it's missing.

```
[Another Game]
system = PSX
; Starts with Another Game
match = ~^Another Game
; Ends with Game
match = ~Game$
; Matches anything in the parentheses
match = ~Another Game \(.+\)
; Matches exactly Another Game
match = ~^Another Game$
```

This example shows how `match` fields can contain [regular expressions](https://quickref.me/regex). Just add a tilde (`~`) character to the start of a match and the rest of the string will be used as a regular expression.