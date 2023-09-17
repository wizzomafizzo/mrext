# Remote API

<!-- TOC -->
* [Remote API](#remote-api)
  * [REST](#rest)
    * [Screenshots](#screenshots)
      * [List screenshots](#list-screenshots)
      * [Take new screenshot](#take-new-screenshot)
      * [View a screenshot](#view-a-screenshot)
      * [Delete a screenshot](#delete-a-screenshot)
    * [Systems](#systems)
      * [List systems](#list-systems)
      * [Launch system](#launch-system)
    * [Wallpapers](#wallpapers)
      * [List wallpapers](#list-wallpapers)
      * [Clear active wallpaper](#clear-active-wallpaper)
      * [View a wallpaper](#view-a-wallpaper)
      * [Set active wallpaper](#set-active-wallpaper)
    * [Music](#music)
      * [Get music service status](#get-music-service-status)
      * [Play music](#play-music)
      * [Stop music](#stop-music)
      * [Skip current track](#skip-current-track)
      * [Set playback type](#set-playback-type)
      * [List playlists](#list-playlists)
      * [Set active playlist](#set-active-playlist)
    * [Games](#games)
      * [Search for games](#search-for-games)
      * [List indexed systems](#list-indexed-systems)
      * [Launch game](#launch-game)
      * [Generate search index](#generate-search-index)
      * [Check current playing game and system](#check-current-playing-game-and-system)
    * [Launchers](#launchers)
      * [Launch token data](#launch-token-data)
      * [Launch games, cores, arcade and .mgl](#launch-games-cores-arcade-and-mgl)
      * [Launch menu](#launch-menu)
      * [Create shortcuts (.mgl files)](#create-shortcuts-mgl-files)
    * [Controls (keyboard)](#controls-keyboard)
      * [Send named keyboard key or combo](#send-named-keyboard-key-or-combo)
      * [Send raw keyboard key](#send-raw-keyboard-key)
    * [Menu](#menu)
      * [List menu folder](#list-menu-folder)
      * [Create menu folder](#create-menu-folder)
      * [Rename menu item](#rename-menu-item)
      * [Delete menu item](#delete-menu-item)
    * [Scripts](#scripts)
      * [Launch a script](#launch-a-script)
      * [List scripts](#list-scripts)
      * [Open framebuffer console](#open-framebuffer-console)
      * [Kill active script](#kill-active-script)
    * [Settings](#settings)
      * [List .ini files](#list-ini-files)
      * [Set active .ini file](#set-active-ini-file)
      * [Get .ini file values](#get-ini-file-values)
      * [Set .ini file values](#set-ini-file-values)
      * [Set menu background mode](#set-menu-background-mode)
      * [Restart Remote service](#restart-remote-service)
      * [Download Remote log file](#download-remote-log-file)
      * [List Remote peers on network](#list-remote-peers-on-network)
      * [Get custom Remote logo](#get-custom-remote-logo)
      * [Reboot MiSTer](#reboot-mister)
      * [Generate a MAC address](#generate-a-mac-address)
    * [Get system information](#get-system-information)
  * [WebSocket](#websocket)
    * [Connection](#connection)
      * [Indexing status](#indexing-status)
      * [Core status](#core-status)
      * [Game status](#game-status)
    * [Events](#events)
    * [Commands](#commands)
      * [Get indexing status](#get-indexing-status)
      * [Send named keyboard key or combo](#send-named-keyboard-key-or-combo-1)
      * [Send raw keyboard key](#send-raw-keyboard-key-1)
      * [Send raw keyboard key down](#send-raw-keyboard-key-down)
      * [Send raw keyboard key up](#send-raw-keyboard-key-up)
<!-- TOC -->

## REST

The REST API is accessible at `http://<ip or hostname>:8182/api` on a default Remote install. It can be used through any standard HTTP client. Examples below use [curl](https://curl.se/).

See the [supported systems](systems.md) page for a list of system IDs referred to throughout this document.

### Screenshots

Methods related to viewing, manageing and taking screenshots.

#### List screenshots

Returns a list of all screenshot files in all core subfolders.

```plaintext
GET /screenshots
```

This method takes no arguments.

On success, returns `200` and a list of objects with attributes:

| Attribute  | Type   | Description                                                                           |
|------------|--------|---------------------------------------------------------------------------------------|
| `game`     | string | Name of game taken from filename. Depends on core support.                            |
| `filename` | string | Full filename of screenshot.                                                          |
| `path`     | string | Relative path of screenshot file within screenshots folder, including core subfolder. |
| `core`     | string | The `setname` ID of core screenshot was taken from.                                   |
| `modified` | string | Screenshot file modified time. Format: `YYYY-MM-DDThh:mm:ss+TZ`                       |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/screenshots"
```

Example response:

```json
[
  {
    "game": "screen",
    "filename": "20230721_212410-screen.png",
    "path": "AO486/20230721_212410-screen.png",
    "core": "AO486",
    "modified": "2023-07-21T21:24:10+08:00"
  },
  {
    "game": "Shahmaty (1987)(-)",
    "filename": "20230721_215005-Shahmaty (1987)(-).png",
    "path": "APOGEE/20230721_215005-Shahmaty (1987)(-).png",
    "core": "APOGEE",
    "modified": "2023-07-21T21:50:06+08:00"
  }
]
```

#### Take new screenshot

Request main to take a new screenshot of the current core using the `/dev/MiSTer_cmd` interface.

*Due to limitations with main, this method cannot report on errors during screenshot, when the screenshot is complete or
which file is the taken screenshot. Check for newest screenshot from the full list after a delay to see taken
screenshot. The method includes an artificial delay which has been reliable.*

```plaintext
POST /screenshots
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/screenshots"
```

#### View a screenshot

Returns raw binary data of specified screenshot file. Can be used as `src` to embed image in documents. Arguments match
the `path` attribute of object in list screenshots method.

```plaintext
GET /screenshots/{core}/{filename}
```

Arguments:

| Attribute  | Type   | Required | Description                                                                                                |
|------------|--------|----------|------------------------------------------------------------------------------------------------------------|
| `core`     | string | Yes      | The `setname` of core screenshot was taken, same as the screenshot's subfolder in main screenshots folder. |
| `filename` | string | Yes      | Filename of screenshot.                                                                                    |

On success, returns `200` and raw screenshot file data with appropriate HTTP headers.

If screenshot does not exist, returns `404`.

Example request:

```shell
curl --request GET --url "http://mister:8182/api/screenshots/AO486/20230721_212410-screen.png" > 20230721_212410-screen.png
```

#### Delete a screenshot

Deletes specified screenshot from disk. Arguments match the `path` attribute of object in list screenshots method.

```plaintext
DELETE /screenshots/{core}/{filename}
```

Arguments:

| Attribute  | Type   | Required | Description                                                                                                |
|------------|--------|----------|------------------------------------------------------------------------------------------------------------|
| `core`     | string | Yes      | The `setname` of core screenshot was taken, same as the screenshot's subfolder in main screenshots folder. |
| `filename` | string | Yes      | Filename of screenshot.                                                                                    |

On success, returns `200`.

If screenshot does not exist, returns `404`.

Example request:

```shell
curl --request DELETE --url "http://mister:8182/api/screenshots/AO486/20230721_212410-screen.png"
```

### Systems

#### List systems

Returns a list of all available systems on device.

```plaintext
GET /systems
```

This method takes no arguments.

On success, returns `200` and a list of objects with attributes:

| Attribute  | Type   | Description                                                              |
|------------|--------|--------------------------------------------------------------------------|
| `id`       | string | Remote's internal ID of the system. See [systems](systems.md). |
| `name`     | string | Friendly name of system. Prefers using `names.txt` file.                 |
| `category` | string | Name of subfolder core .rbf file is contained in.                        |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/systems"
```

Example response:

```json
[
  {
    "id": "Intellivision",
    "name": "Intellivision",
    "category": "Console"
  },
  {
    "id": "MacPlus",
    "name": "Macintosh Plus",
    "category": "Computer"
  }
]
```

#### Launch system

Launch a system based on its ID.

```plaintext
POST /systems/{id}
```

Arguments:

| Attribute | Type   | Required | Description            |
|-----------|--------|----------|------------------------|
| `id`      | string | Yes      | System's internal ID. See [systems](systems.md). |

On success, returns `200`.

If system does not exist, returns `404`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/systems/SNES"
```

### Wallpapers

Remote has its own mechanism of setting wallpapers as "active" on the MiSTer menu by managing a symlink to the wallpaper
file in the root of the SD card.

#### List wallpapers

Returns a list of all wallpapers.

```plaintext
GET /wallpapers
```

This method takes no arguments.

On success, returns `200` and object:

| Attribute        | Type        | Description                                                                                             |
|------------------|-------------|---------------------------------------------------------------------------------------------------------|
| `active`         | string      | Filename of active wallpaper. Empty string if none set.                                                 |
| `backgroundMode` | number      | The current index of "background mode" set in MiSTer menu. This is the mode changed when F1 is pressed. |
| `wallpapers`     | Wallpaper[] | See below.                                                                                              |

Wallpaper object:

| Attribute  | Type    | Description                                 |
|------------|---------|---------------------------------------------|
| `name`     | string  | Filename of wallpaper without extension.    |
| `filename` | string  | Full filename of wallpaper.                 |
| `width`    | number  | *Not in use.*                               |
| `height`   | number  | *Not in use.*                               |
| `active`   | boolean | `true` if wallpaper has been set as active. |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/wallpapers"
```

Example response:

```json
{
  "active": "",
  "backgroundMode": 2,
  "wallpapers": [
    {
      "name": "MiSTer 3D World Runner",
      "filename": "MiSTer 3D World Runner.jpg",
      "width": 0,
      "height": 0,
      "active": false
    },
    {
      "name": "MiSTer AMIGA",
      "filename": "MiSTer AMIGA.jpg",
      "width": 0,
      "height": 0,
      "active": false
    }
  ]
}
```

#### Clear active wallpaper

Clears any current active wallpaper and reverts to default behaviour of displaying a random wallpaper.

```plaintext
DELETE /wallpapers
```

This method takes no arguments.

On success, returns `200`.

If no wallpaper is active, returns `500`.

Example request:

```shell
curl --request DELETE --url "http://mister:8182/api/wallpapers"
```

#### View a wallpaper

Returns raw binary data of specified wallpaper file. Can be used as `src` to embed image in documents.

```plaintext
GET /wallpapers/{filename}
```

Arguments:

| Attribute  | Type   | Required | Description                |
|------------|--------|----------|----------------------------|
| `filename` | string | Yes      | Filename of the wallpaper. |

On success, returns `200` and raw wallpaper file data with appropriate HTTP headers.

If wallpaper does not exist, returns `404`.

Example request:

```shell
curl --request GET --url "http://mister:8182/api/wallpapers/snatcher.png" > snatcher.png
```

#### Set active wallpaper

Sets the specified wallpaper as active. This method also changes the current menu "background mode" to "wallpaper" and
will relaunch the menu core if it's open.

```plaintext
POST /wallpapers/{filename}
```

Arguments:

| Attribute  | Type   | Required | Description                |
|------------|--------|----------|----------------------------|
| `filename` | string | Yes      | Filename of the wallpaper. |

On success, returns `200`.

If wallpaper does not exist or an error occurs handling symlink, returns `500`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/wallpapers/snatcher.png"
```

### Music

All methods in the music endpoint depend on the BGM service to be installed and running.

#### Get music service status

Polls music service and returns current status.

```plaintext
GET /music/status
```

This method takes no arguments.

On success, returns `200` and object with attributes:

| Attribute  | Type    | Description                                                  |
|------------|---------|--------------------------------------------------------------|
| `running`  | boolean | `true` if the BGM service is active.                         |
| `playing`  | boolean | `true` if music is currently playing.                        |
| `playback` | string  | Current playlist playback type: `random`, `loop`, `disabled` |
| `playlist` | string  | Name of playlist (folder containing tracks).                 |
| `track`    | string  | Filename of current track.                                   |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/music/status"
```

Example response:

```json
{
  "running": true,
  "playing": true,
  "playback": "random",
  "playlist": "Vidya",
  "track": "Final Fantasy VI - The Mines of Narshe.mp3"
}
```

#### Play music

Play the active playlist.

```plaintext
POST /music/play
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/music/play"
```

#### Stop music

Stop playing music.

```plaintext
POST /music/stop
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/music/stop"
```

#### Skip current track

Skips to the next track in playlist.

```plaintext
POST /music/next
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/music/next"
```

#### Set playback type

Set the playback type of the playlists. This setting doesn't not persist between service restarts.

```plaintext
POST /music/playback/{type}
```

Arguments:

| Attribute | Type   | Required | Description                                       |
|-----------|--------|----------|---------------------------------------------------|
| `type`    | string | Yes      | ID of playback type: `random`, `loop`, `disabled` |

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/music/playback/loop"
```

#### List playlists

Returns a list of all playlists available to play.

*The `none` playlist is a special virtual playlist in BGM, which instructs it to play only tracks in the first level of
the root of the music folder.*

```plaintext
GET /music/playlist
```

This method takes no arguments.

On success, returns `200` and a simple list of strings matching subfolder names in the music folder.

Example request:

```shell
curl --request GET --url "http://mister:8182/api/music/playlist"
```

Example response:

```json
[
  "none",
  "Arcade Ambience",
  "Menu Music"
]
```

#### Set active playlist

Set the active playlist being played.

```plaintext
POST /music/playlist/{name}
```

Arguments:

| Attribute | Type   | Required | Description                   |
|-----------|--------|----------|-------------------------------|
| `name`    | string | Yes      | Name of playlist (subfolder). |

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/music/playback/Arcade%20Ambiance"
```

### Games

Methods relating to searching games require an index to be generated in advance (through the indexing method). Search
results are currently not paginated.

#### Search for games

Search for games on device by name (filename). Query strings are split by whitespace and each subsequent keyword token
must be contained in the filename. For example, query "crash bandicoot" matches on games containing "crash" AND "
bandicoot" somewhere.

```plaintext
POST /games/search
```

Arguments (JSON):

| Attribute | Type   | Required | Description                                                                                               |
|-----------|--------|----------|-----------------------------------------------------------------------------------------------------------|
| `data`    | string | Yes      | Query to search for in game filename (by word).                                                           |
| `system`  | string | Yes      | System ID to search in. `all` or empty string to search all systems. Must be an exact match of system ID. |

On success, returns `200` and object:

| Attribute  | Type     | Description                                                                                  |
|------------|----------|----------------------------------------------------------------------------------------------|
| `data`     | Result[] | List of result objects (see below).                                                          |
| `total`    | number   | Total number of results.                                                                     |
| `pageSize` | number   | Max number of results per page. *Accurate, but multiple pages aren't currently implemented.* |
| `page`     | number   | Current page number.                                                                         |

Result object:

| Attribute | Type   | Description                              |
|-----------|--------|------------------------------------------|
| `system`  | System | Information of system game is linked to. |
| `name`    | string | Filename of game excluding extension.    |
| `path`    | string | Absolute path to game file.              |

System object:

| Attribute | Type   | Description                    |
|-----------|--------|--------------------------------|
| `id`      | string | Internal ID of linked system. See [systems](systems.md). |
| `name`    | string | Friendly name of system.       |

Example request:

```shell
curl --request POST --url "http://mister:8182/api/games/search" --data '{"query":"crash bandicoot","system":"PSX"}'
```

Example response:

```json
{
  "data": [
    {
      "system": {
        "id": "PSX",
        "name": "Playstation"
      },
      "name": "Crash Bandicoot (USA)",
      "path": "/media/fat/games/PSX/1 USA - A-D/Crash Bandicoot (USA).chd"
    },
    {
      "system": {
        "id": "PSX",
        "name": "Playstation"
      },
      "name": "Crash Bandicoot - Warped (USA)",
      "path": "/media/fat/games/PSX/1 USA - A-D/Crash Bandicoot - Warped (USA).chd"
    },
    {
      "system": {
        "id": "PSX",
        "name": "Playstation"
      },
      "name": "Crash Bandicoot 2 - Cortex Strikes Back (USA)",
      "path": "/media/fat/games/PSX/1 USA - A-D/Crash Bandicoot 2 - Cortex Strikes Back (USA).chd"
    }
  ],
  "total": 3,
  "pageSize": 500,
  "page": 1
}
```

#### List indexed systems

Returns a list of all systems with indexed games.

**This method is know to have significant resource usage, and can cause screen flickering in high resolution
framebuffers, and potentially affect access timing for cores mounting CD images. It won't damage anything, but has
noticeable effects.**

```plaintext
GET /games/search/systems
```

This method takes no arguments.

On success, returns `200` and object:

| Attribute | Type     | Description                         |
|-----------|----------|-------------------------------------|
| `systems` | System[] | List of system objects (see below). |

System object:

| Attribute | Type   | Description                   |
|-----------|--------|-------------------------------|
| `id`      | string | Internal ID of linked system. |
| `name`    | string | Friendly name of system.      |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/games/search/systems"
```

Example response:

```json
{
  "systems": [
    {
      "id": "AcornAtom",
      "name": "Atom"
    },
    {
      "id": "AcornElectron",
      "name": "Electron"
    },
    {
      "id": "AdventureVision",
      "name": "Adventure Vision"
    }
  ]
}
```

#### Launch game

Launch a game given an absolute path to the game file. System is auto-detected from path and file type.

```plaintext
POST /games/launch
```

Arguments (JSON):

| Attribute | Type   | Required | Description                 |
|-----------|--------|----------|-----------------------------|
| `path`    | string | Yes      | Absolute path to game file. |

On success, returns `200`.

If system cannot be detected from path, returns `500`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/games/search" --data '{"query":"crash bandicoot","system":"PSX"}'
```

#### Generate search index

Trigger an asynchronous request to generate a new search index on disk. Will overwrite any existing index.

*Currently status of index must be monitored through WebSocket endpoint.*

**This method can cause screen flickering on high resolution framebuffers and read timing access issues for cores that
mount CD images. Take care when it's run.**

```plaintext
POST /games/index
```

This method takes no arguments.

Always returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/games/index"
```

#### Check current playing game and system

Returns the current running game, system and core. Values are empty strings if nothing is running or if in menu. If the
core is an arcade core, the system is set to `Arcade` and attempt is made to look up information from Arcade DB.

```plaintext
GET /games/playing
```

This method takes no arguments.

On success, returns `200` and object with attributes:

| Attribute    | Type   | Description                         |
|--------------|--------|-------------------------------------|
| `core`       | string | `setname` of active core.           |
| `system`     | string | System ID of active core.           |
| `systemName` | string | Friendly name of system.            |
| `game`       | string | System ID and filename of game.     |
| `gameName`   | string | Filename of game without extension. |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/games/playing"
```

Example response:

```json
{
  "core": "NES",
  "system": "NES",
  "systemName": "NES",
  "game": "NES/2022-04 Crystalis.mgl",
  "gameName": "2022-04 Crystalis"
}
```

### Launchers

#### Launch token data

Launch encoded data matching format of
the [NFC script](nfc.md#setting-up-tags) which includes cores,
games, .mras, .mgls and custom commands. This method is intended for QR code launching or any other devices with limited
REST support. Data is encoded in [base64url](https://simplycalc.com/base64url-encode.php).

```plaintext
GET /l/{data}
```

Arguments:

| Attribute | Type   | Required | Description               |
|-----------|--------|----------|---------------------------|
| `data`    | string | Yes      | base64url encoded string. |

On success, returns `200`.

Example request (data is `menu.rbf`):

```shell
curl --request GET --url "http://mister:8182/api/l/bWVudS5yYmY="
```

#### Launch games, cores, arcade and .mgl

This is a general purpose method which will attempt to detect and launch any given path. It supports game files, .rbfs,
.mras and .mgls.

```plaintext
POST /launch
```

Arguments (JSON):

| Attribute | Type   | Required | Description   |
|-----------|--------|----------|---------------|
| `path`    | string | Yes      | Path to file. |

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/launch" --data '{"path":"/media/fat/menu.rbf"}'
```

#### Launch menu

Launches the menu core, exiting the current core.

```plaintext
POST /launch/menu
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/launch/menu"
```

#### Create shortcuts (.mgl files)

Creates a .mgl file at the specified location for a game file. System and all necessary .mgl configuration is
auto-detected.

```plaintext
POST /launch/new
```

Arguments (JSON):

| Attribute  | Type   | Required | Description                                                                        |
|------------|--------|----------|------------------------------------------------------------------------------------|
| `gamePath` | string | Yes      | Path to the game file.                                                             |
| `folder`   | string | Yes      | Folder .mgl file will be created (including underscores before menu folder names). |
| `name`     | string | Yes      | Name of .mgl file, excluding extension.                                            |

On success, returns `200` and object:

| Attribute | Type   | Description                  |
|-----------|--------|------------------------------|
| `path`    | string | Final path of new .mgl file. |

Example request:

```shell
curl --request POST --url "http://mister:8182/api/launch/new" --data '{"gamePath":"/media/fat/games/PSX/1 USA - A-D/Crash Bandicoot (USA).chd","folder":"_@Favorites","name":"Crash Bandicoot"}'
```

Example response:

```json
{
  "path": "/media/fat/_@Favorites/Crash Bandicoot.mgl"
}
```

### Controls (keyboard)

#### Send named keyboard key or combo

Sends a keyboard key or combo to the MiSTer based on a predefined list of names that describe its function.
See [here](https://github.com/wizzomafizzo/mrext/blob/f03acd3b2cab6950037a83f73bdd37af3b63510e/cmd/remote/control/control.go#L49)
for the full list of names available.

```plaintext
POST /controls/keyboard/{name}
```

Arguments:

| Attribute | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| `name`    | string | Yes      | Name of keyboard key. |

On success, returns `200`.

If the name is not recognised, returns `500`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/controls/keyboard/confirm"
```

#### Send raw keyboard key

Send a keyboard key based on its uinput code.
See [here](https://github.com/bendahl/uinput/blob/600101208cf24d14eff079d2478ac1f8cad4ae8a/keycodes.go#L5) for full list
of possible codes.

*This method does not allow for sending combos, but a `-` can be prepended to the code to hold shift while sending.*

```plaintext
POST /controls/keyboard-raw/{code}
```

Arguments:

| Attribute | Type   | Required | Description         |
|-----------|--------|----------|---------------------|
| `code`    | number | Yes      | uinput code of key. |

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/controls/keyboard-raw/-16"
```

### Menu

#### List menu folder

List the contents of a menu folder.

```plaintext
POST /menu/view
```

Arguments (JSON):

| Attribute | Type   | Required | Description                                |
|-----------|--------|----------|--------------------------------------------|
| `path`    | string | Yes      | Path of menu folder (relative to SD card). |

On success, returns `200` and object:

| Attribute | Type   | Description                       |
|-----------|--------|-----------------------------------|
| `items`   | Item[] | List of Item objects (see below). |

Item object:

| Attribute   | Type    | Description                                                                        |
|-------------|---------|------------------------------------------------------------------------------------|
| `name`      | string  | Filename of item excluding extension (prefers `names.txt`).                        |
| `path`      | string  | Absolute path to file.                                                             |
| `parent`    | string  | Path to parent folder.                                                             |
| `filename`  | string  | Full filename of item.                                                             |
| `extension` | string  | File extension of item.                                                            |
| `type`      | string  | Type of file: `folder`, `mra`, `rbf`, `mgl`, `unknown`                             |
| `modified`  | string  | File modified date. Format: `YYYY-MM-DDThh:mm:ss+TZ`                               |
| `version`   | string? | *Cores only.* Release date of core from filename. Format: `YYYY-MM-DDThh:mm:ss+TZ` |
| `size`      | number  | Size of file in bytes.                                                             |

Example request:

```shell
curl --request POST --url "http://mister:8182/api/menu/view" --data '{"path":"."}'
```

Example response:

```json
{
  "items": [
    {
      "name": "10-Yard Fight (USA, Europe)",
      "path": "/media/fat/10-Yard Fight (USA, Europe).mgl",
      "parent": ".",
      "filename": "10-Yard Fight (USA, Europe).mgl",
      "extension": ".mgl",
      "type": "mgl",
      "modified": "2023-08-12T08:30:38+08:00",
      "size": 259
    },
    {
      "name": "3 Count Bout",
      "path": "/media/fat/3 Count Bout.mgl",
      "parent": ".",
      "filename": "3 Count Bout.mgl",
      "extension": ".mgl",
      "type": "mgl",
      "modified": "2023-08-14T09:59:04+08:00",
      "size": 130
    }
  ]
}
```

#### Create menu folder

Create a new menu folder. Underscore (`_`) is automatically prepended to the folder name.

```plaintext
POST /menu/files/create
```

Arguments (JSON):

| Attribute | Type   | Required | Description                                       |
|-----------|--------|----------|---------------------------------------------------|
| `type`    | string | Yes      | Must be: `folder`                                 |
| `folder`  | string | Yes      | Path containing new folder (relative to SD card). |
| `name`    | string | Yes      | Name of new folder.                               |

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/menu/files/create" --data '{"type":"folder","folder":".","name":"New Folder"}'
```

#### Rename menu item

Rename a menu item.

```plaintext
POST /menu/files/rename
```

Arguments (JSON):

| Attribute  | Type   | Required | Description                     |
|------------|--------|----------|---------------------------------|
| `fromPath` | string | Yes      | Absolute path to existing file. |
| `toPath`   | string | Yes      | Absolute path to new file.      |

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/menu/files/rename" --data '{"fromPath":"/media/fat/New Folder","toPath":"/media/fat/New Folder 2"}'
```

#### Delete menu item

Delete a menu item.

```plaintext
POST /menu/files/delete
```

Arguments (JSON):

| Attribute | Type   | Required | Description                         |
|-----------|--------|----------|-------------------------------------|
| `path`    | string | Yes      | Path of file (relative to SD card). |

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/menu/files/delete" --data '{"path":"/media/fat/New Folder 2"}'
```

### Scripts

Scripts are located in the `Scripts` folder on the SD card. These methods currently do not support scripts in subfolders
or on external mounts.

#### Launch a script

Launch a script by filename. This method is intended to replicate exactly how launching a script from the MiSTer menu
works. It attempts to switch to the framebuffer console and then launches the script using the same wrapper tmp script.

```plaintext
POST /scripts/launch/{filename}
```

Arguments:

| Attribute  | Type   | Required | Description         |
|------------|--------|----------|---------------------|
| `filename` | string | Yes      | Filename of script. |

Always returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/scripts/launch/update_all.sh"
```

#### List scripts

List all scripts in the `Scripts` folder.

```plaintext
GET /scripts/list
```

This method takes no arguments.

On success, returns `200` and object:

| Attribute   | Type     | Description                                   |
|-------------|----------|-----------------------------------------------|
| `canLaunch` | boolean  | Reports if the menu core is currently active. |
| `scripts`   | Script[] | List of Script objects (see below).           |

Script object:

| Attribute  | Type   | Description                             |
|------------|--------|-----------------------------------------|
| `name`     | string | Filename of script excluding extension. |
| `filename` | string | Full filename of script.                |
| `path`     | string | Absolute path to script.                |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/scripts/list"
```

Example response:

```json
{
  "canLaunch": true,
  "scripts": [
    {
      "name": "Install_ScummVM",
      "filename": "Install_ScummVM.sh",
      "path": "/media/fat/Scripts/Install_ScummVM.sh"
    },
    {
      "name": "MiSTer_SAM_off",
      "filename": "MiSTer_SAM_off.sh",
      "path": "/media/fat/Scripts/MiSTer_SAM_off.sh"
    },
    {
      "name": "MiSTer_SAM_on",
      "filename": "MiSTer_SAM_on.sh",
      "path": "/media/fat/Scripts/MiSTer_SAM_on.sh"
    }
  ]
}
```

#### Open framebuffer console

Switch to the framebuffer console using the keyboard. This method works even if the menu is asleep.

```plaintext
POST /scripts/console
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/scripts/console"
```

#### Kill active script

Kill the active script.

```plaintext
POST /scripts/kill
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/scripts/kill"
```

### Settings

Most configuration is done through the `inis` endpoint. This is a low-level interface to the `MiSTer.ini` files that
makes them accessible as a dictionary of string keys to string values.

Notes on usage:

- Keys are edited as needed, you don't need to send the entire state of the file to keep everything there, just the keys
  you want to change.
- This interface has very minimal validation, make sure to validate settings yourself.
- Set a value to an empty string to remove it from the file.
- Sections, like `[SNES]`, are not currently supported. All values edited are in the `[MiSTer]` section, but it's safe
  to edit a file containing these sections, they remain untouched.
- Some internal keys are added with a leading double underscore (`__`). These map back to system configuration files not
  included in the `MiSTer.ini` files. Currently these are: `__hostname`, `__ethernetMacAddress`

#### List .ini files

List all available MiSTer.ini files on the SD card including alternate files.

```plaintext
GET /settings/inis
```

This method takes no arguments.

On success, returns `200` and object:

| Attribute | Type   | Description                                                                                                                                          |
|-----------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| `active`  | number | `0` to `4`. This is the current active .ini file's ID in the list. `0` is a special value meaning no value has been set, which falls back on ID `1`. |
| `inis`    | Ini[]  | List of Ini objects (see below).                                                                                                                     |

Ini object:

| Attribute     | Type   | Description                                        |
|---------------|--------|----------------------------------------------------|
| `id`          | number | ID of .ini file.                                   |
| `displayName` | string | Name of the .ini file as it would show in the OSD. |
| `filename`    | string | Filename of the .ini file.                         |
| `path`        | string | Absolute path to the .ini file.                    |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/settings/inis"
```

Example response:

```json
{
  "active": 0,
  "inis": [
    {
      "id": 1,
      "displayName": "Main",
      "filename": "MiSTer.ini",
      "path": "/media/fat/MiSTer.ini"
    }
  ]
}
```

#### Set active .ini file

Set the active .ini file by ID. Like MiSTer itself, this change is not persistent, and will be lost on reboot.

```plaintext
PUT /settings/inis
```

Arguments (JSON):

| Attribute | Type   | Required | Description      |
|-----------|--------|----------|------------------|
| `ini`     | number | Yes      | ID of .ini file. |

On success, returns `200`.

Example request:

```shell
curl --request PUT --url "http://mister:8182/api/settings/inis" --data '{"ini":1}'
```

#### Get .ini file values

Get all values from the specified .ini file.

```plaintext
GET /settings/inis/{id}
```

Arguments:

| Attribute | Type   | Required | Description                            |
|-----------|--------|----------|----------------------------------------|
| `id`      | number | Yes      | ID of .ini file. `1`, `2`, `3` or `4`. |

On success, returns `200` a dictionary of string keys to string values. Keys are case sensitive and map back to the
appropriate key in the .ini file.

Example request:

```shell
curl --request GET --url "http://mister:8182/api/settings/inis/1"
```

Example reponse:

```json
{
  "__ethernetMacAddress": "",
  "__hostname": "MiSTuh",
  "bootcore_timeout": "10",
  "bt_auto_disconnect": "0",
  "bt_reset_before_pair": "0",
  "composite_sync": "0",
  "controller_info": "6",
  "direct_video": "0",
  "disable_autofire": "0",
  "fb_size": "0",
  "fb_terminal": "1",
  "font": "font/myfont.pf",
  "forced_scandoubler": "0",
  "gamepad_defaults": "0"
}
```

#### Set .ini file values

Set values in the specified .ini file.

```plaintext
PUT /settings/inis/{id}
```

Arguments (JSON):

A dictionary of string keys to string values. Keys are case sensitive and map back to the appropriate key in the .ini
file. If in the menu, the core will be restarted to apply the changes.

Example request:

```shell
curl --request PUT --url "http://mister:8182/api/settings/inis/1" --data '{"composite_sync":"1"}'
```

#### Set menu background mode

Set the "background mode" of the menu core. Equivalent to when `F1` is pressed in the menu, but doesn't use keyboard
input.

```plaintext
PUT /settings/core/menu
```

Arguments (JSON):

| Attribute | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `mode`    | number | Yes      | `0` to `7`  |

On success, returns `200`.

Example request:

```shell
curl --request PUT --url "http://mister:8182/api/settings/core/menu" --data '{"mode":0}'
```

#### Restart Remote service

Restart the Remote service. Used for reloading after an update.

```plaintext
POST /settings/remote/restart
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/settings/remote/restart"
```

#### Download Remote log file

Offers the Remote log file for download.

```plaintext
GET /settings/remote/log
```

This method takes no arguments.

On success, returns `200` and raw log file data with appropriate HTTP headers.

Example request:

```shell
curl --request GET --url "http://mister:8182/api/settings/remote/log" > remote.log
```

#### List Remote peers on network

By default, Remote will scan regularly for other instances of Remote (other MiSTers) on the local network. This method
gives access to that list of clients.

```plaintext
GET /settings/remote/peers
```

This method takes no arguments.

On success, returns `200` and object:

| Attribute | Type   | Description                       |
|-----------|--------|-----------------------------------|
| `peers`   | Peer[] | List of Peer objects (see below). |

Peer object:

| Attribute  | Type   | Description                |
|------------|--------|----------------------------|
| `hostname` | string | Hostname of peer.          |
| `version`  | string | Version of Remote on peer. |
| `ip`       | string | IP address of peer.        |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/settings/remote/peers"
```

Example response:

```json
{
  "peers": [
    {
      "hostname": "MiSTuh.local",
      "version": "0.2.4",
      "ip": "10.0.0.107"
    }
  ]
}
```

#### Get custom Remote logo

Download the custom Remote logo file. This is just used for optional customisation in the Remote web UI.

```plaintext
GET /settings/remote/logo
```

This method takes no arguments.

On success, returns `200` and raw logo file data with appropriate HTTP headers.

Example request:

```shell
curl --request GET --url "http://mister:8182/api/settings/remote/logo" > logo.png
```

#### Reboot MiSTer

Reboot the MiSTer.

```plaintext
POST /settings/system/reboot
```

This method takes no arguments.

On success, returns `200`.

Example request:

```shell
curl --request POST --url "http://mister:8182/api/settings/system/reboot"
```

#### Generate a MAC address

Generate a random MAC address. This is a utility function used for the `__ethernetMacAddress` key in the .ini files. It
doesn't actually set the value in the .ini file.

```plaintext
GET /settings/system/generate-mac
```

This method takes no arguments.

On success, returns `200` and object:

| Attribute | Type   | Description               |
|-----------|--------|---------------------------|
| `mac`     | string | Random valid MAC address. |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/settings/system/generate-mac"
```

Example response:

```json
{
  "mac": "32:d4:c2:00:00:6b"
}
```

### Get system information

Get information about the MiSTer system such as network, hostname, last update and disk usage.

```plaintext
GET /sysinfo
```

This method takes no arguments.

On success, returns `200` and object:

| Attribute  | Type     | Description                                      |
|------------|----------|--------------------------------------------------|
| `ips`      | string[] | List of IP addresses.                            |
| `hostname` | string   | Hostname of MiSTer.                              |
| `dns`      | string   | DNS name of MiSTer.                              |
| `version`  | string   | Version of Remote.                               |
| `updated`  | string   | Last update time of MiSTer (through downloader). |
| `disks`    | Disk[]   | List of Disk objects (see below).                |

Disk object:

| Attribute     | Type   | Description                                 |
|---------------|--------|---------------------------------------------|
| `path`        | string | Mount path of disk.                         |
| `total`       | number | Total size of disk in bytes.                |
| `used`        | number | Used size of disk in bytes.                 |
| `free`        | number | Free size of disk in bytes.                 |
| `displayName` | string | Friendly name of disk. Hardcoded in Remote. |

Example request:

```shell
curl --request GET --url "http://mister:8182/api/sysinfo"
```

Example response:

```json
{
  "ips": [
    "10.0.0.107",
    "10.0.0.218"
  ],
  "hostname": "MiSTuh",
  "dns": "MiSTuh.local",
  "version": "0.2.4",
  "updated": "2023-08-30T19:35:02+08:00",
  "disks": [
    {
      "path": "/media/fat",
      "total": 511848742912,
      "used": 449839759360,
      "free": 62008983552,
      "displayName": "SD card"
    }
  ]
}
```

## WebSocket

Remote's WebSocket interface is available on the `/ws` endpoint. It's used for monitoring game and core status of the
MiSTer, search indexing status, and a more granular lower latency keyboard input method.

Multiple WebSocket connections are supported.

### Connection

On initial connection, the client will be sent messages giving current state.

#### Indexing status

Format: `indexStatus:{exists},{inProgress},{totalSteps},{currentStep},{currentStepDescription}`

| Attribute                | Type    | Description                                                               |
|--------------------------|---------|---------------------------------------------------------------------------|
| `exists`                 | boolean | `y` if an index exists on disk, otherwise `n`.                            |
| `inProgress`             | boolean | `y` if an index is currently being generated, otherwise `n`.              |
| `totalSteps`             | number  | Total number of steps in the index generation process. Split by system.   |
| `currentStep`            | number  | Current step in the index generation process. Split by system.            |
| `currentStepDescription` | string  | Description of current step in the index generation process. System name. |

Steps are used for displaying detailed indexing status to the user.

#### Core status

Format: `coreRunning:{name}`

| Attribute | Type   | Description                                                    |
|-----------|--------|----------------------------------------------------------------|
| `name`    | string | `setname` of currently running core, blank if menu is running. |

#### Game status

Format: `gameRunning:{system}/{name}`

| Attribute | Type   | Description |
|-----------|---------|------------------|
| `system`  | string | System ID of currently running core. |
| `name`    | string | Filename of currently running game. |

This field is blank if no game is running.

### Events

As their state changes, all clients connected to the WebSocket will be sent event updates for the messages listed in the
connection section. The format of these is exactly the same as the connection messages.

If the MiSTer exits to menu, the `gameRunning` and `coreRunning` events will be sent with blank values.

### Commands

These commands can be sent from the client to the server to perform actions.

#### Get indexing status

Format: `getIndexStatus`

Returns the index status message as described in the connection section.

#### Send named keyboard key or combo

Format: `kbd:{name}`

| Attribute | Type   | Description                                                                 |
|-----------|--------|-----------------------------------------------------------------------------|
| `name`    | string | Name of keyboard key, as described in the `/controls/keyboard` REST method. |

#### Send raw keyboard key

Format: `kbdRaw:{code}`

| Attribute | Type   | Description                                                                   |
|-----------|--------|-------------------------------------------------------------------------------|
| `code`    | number | uinput code of key, as described in the `/controls/keyboard-raw` REST method. |

#### Send raw keyboard key down

Sends just a key down event, enabling holding keys and key combos.

Format: `kbdRawDown:{code}`

| Attribute | Type   | Description                                                                   |
|-----------|--------|-------------------------------------------------------------------------------|
| `code`    | number | uinput code of key, as described in the `/controls/keyboard-raw` REST method. |

#### Send raw keyboard key up

Sends just a key up event. Don't forget to do this after key down.

Format: `kbdRawUp:{code}`

| Attribute | Type   | Description                                                                   |
|-----------|--------|-------------------------------------------------------------------------------|
| `code`    | number | uinput code of key, as described in the `/controls/keyboard-raw` REST method. |
