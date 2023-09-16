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
    * [WebSocket](#websocket)

<!-- TOC -->

## REST

### Screenshots

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

| Attribute  | Type   | Description                                              |
|------------|--------|----------------------------------------------------------|
| `id`       | string | Remote's internal ID of the system.                      |
| `name`     | string | Friendly name of system. Prefers using `names.txt` file. |
| `category` | string | Name of subfolder core .rbf file is contained in.        |

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

| Attribute | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| `id`      | string | Yes      | System's internal ID. |

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
| `wallpapers`     | wallpaper[] | See below.                                                                                              |

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
| `data`     | result[] | List of result objects (see below).                                                          |
| `total`    | number   | Total number of results.                                                                     |
| `pageSize` | number   | Max number of results per page. *Accurate, but multiple pages aren't currently implemented.* |
| `page`     | number   | Current page number.                                                                         |

Result object:

| Attribute | Type   | Description                              |
|-----------|--------|------------------------------------------|
| `system`  | system | Information of system game is linked to. |
| `name`    | string | Filename of game excluding extension.    |
| `path`    | string | Absolute path to game file.              |

System object:

| Attribute | Type   | Description                   |
|-----------|--------|-------------------------------|
| `id`      | string | Internal ID of linked system. |
| `name`    | string | Friendly name of system.      |

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
the [NFC script](https://github.com/wizzomafizzo/mrext/blob/main/docs/nfc.md#setting-up-tags) which includes cores,
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

## WebSocket
