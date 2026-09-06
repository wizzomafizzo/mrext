# LaunchSync

> [!IMPORTANT]
> LaunchSync is maintained for `.sync` files and stock-menu shortcuts. Before creating a new shared MiSTer list with it, consider Online cards and decks, Zap Links, self-hosted Zap Link servers, or Zaparoo playlists. Keep LaunchSync when subscribed `.mgl` folders are required. See [MIGRATE.md](../MIGRATE.md).

LaunchSync allows people to create, share and maintain live-updating game playlists for the MiSTer.

You create a [sync file](#sync-files) with a list of games, someone copies the file to their MiSTer, and LaunchSync uses it to generate working game shortcuts in MiSTer main menu. Sync files can subscribe to a published source, so playlist changes can reach other systems on their next LaunchSync run.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/launchsync.sh"><img src="images/download.svg" alt="Download LaunchSync" title="Download LaunchSync" width="140"></a>

## Install

Download [LaunchSync](https://github.com/wizzomafizzo/mrext/releases/latest/download/launchsync.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add this database to `downloader.ini` in SD card root, then run `downloader` or `update` to receive updates:

```ini
[mrext/launchsync]
db_url = https://raw.githubusercontent.com/wizzomafizzo/mrext/main/releases/launchsync/launchsync.json
```

## Usage

1. Place at least one sync file in the root of your SD card or in a menu folder. Menu folders are folders which start with an underscore (`_`)
2. Run `launchsync` from the MiSTer `Scripts` menu

LaunchSync will search for all sync files on the MiSTer, check for sync file updates online, create folders for new sync files, and then create or update all shortcuts for listed games.

LaunchSync must be run manually whenever you want to update subscribed files or regenerate shortcuts.

### Sync Files

LaunchSync requires sync files to actually do anything. These are text files ending in `.sync` which define the name of a playlist, the games in it and how to find them on your own system. You can create your own or find sync files other people have created. An example is the [Discord Game of the Month](https://raw.githubusercontent.com/wizzomafizzo/mrext/main/cmd/launchsync/examples/Discord%20Game%20of%20the%20Month.sync) playlist hosted here.

## Uninstall

Remove LaunchSync's Downloader subscription if configured, so updates do not reinstall it.

1. Let any update finish and exit LaunchSync. It installs no startup service; remove any scheduled or startup invocations you added yourself.
2. Delete `/media/fat/Scripts/launchsync.sh`. Optionally back up and remove `/media/fat/Scripts/launchsync.ini` if present, respecting any shared configuration override.
3. Preserve your `.sync` files by default: they contain subscriptions or authored game lists. To unsubscribe permanently, remove only the `.sync` files you no longer want from their actual locations.
4. Generated folders live beside each `.sync` file, named from its top-level `name` field with an underscore prefix. Keep them to retain static shortcuts, or inspect and remove their generated `.mgl` files, arcade `.mra` links, and `[NOT FOUND].mgl` placeholders. Do not remove unrelated files or follow `cores` symlinks into the real arcade cores directory. Removing a `.sync` file does not by itself remove its generated folder.

The game index at `/media/fat/Scripts/.config/mrext/games.db` is shared with Search and Remote; leave it in place when either remains installed. Keep original games, arcade MRAs, and shared metadata. There is no separate LaunchSync service database to delete.

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

The `updated` field is only required if a `url` is set. It's used to check if an update is actually necessary. It can be in the format `YYYY-MM-DD` or `YYYY-MM-DD hh:mm`.

### Game Sections

Each game in a sync file is defined with a section header:

```
[My Favorite Game]
```

This starts the section, and also defines the name of the shortcut that will be created in the MiSTer menu. It must be unique and there can be as many as you want. Similar to the `name` field in the header, it can't contain any of these characters: `\ : * ? " < > |`

A game's name *can* contain forward slashes (`/`), which will allow you to specify subfolders for a game's launcher to be placed in. Any folders contained in the name will be automatically created and prefixed with a `_` character so they will display in the menu. You can specify the same folder for multiple games.

For example, these are all valid section headers: `[Game]`, `[PSX/Game]` and `[Some/Deep Folder/Game]`.

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