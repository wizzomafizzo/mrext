# BGM

BGM plays background music in MiSTer's menu, stops it while a core runs, and plays boot sounds when MiSTer starts or a core launches. It runs as a standalone Go application and does not require Zaparoo Core.

Zaparoo Core also offers core audio playback. See [MIGRATE.md](../MIGRATE.md) before choosing between them; internet radio and BGM's folder-based configuration have no confirmed equivalent there.

## Installation

Install BGM through the combined mrext Downloader database documented in the [README](../README.md). Downloader installs the ARMv7 binary as:

```text
/media/fat/Scripts/bgm.sh
```

The `.sh` filename is retained for compatibility with existing Downloader entries, the `user-startup.sh` hook, and MiSTer Scripts-menu behavior. It is a native executable, not a shell script, but it accepts exactly the same arguments as the Python script it replaces.

Run `bgm` from MiSTer's Scripts menu. The first run creates the `music` folder on the SD card, adds the startup hook, and asks for music files. Once files are present, running `bgm` again starts the service and opens the control screen.

## Features

- Play music in the main MiSTer menu and stop automatically when a core launches.
- Play boot sounds when MiSTer starts and when specific cores launch.
- Play `.mp3`, `.ogg`, `.wav`, `.mid`, `.vgm`, `.vgz`, `.vgm.gz` files and `.pls` internet radio streams.
- Organise tracks into playlist folders and switch between them.
- Random or single-track-loop playback, and per-track loop counts.
- Optionally keep music playing inside cores.
- Automatically switch MiSTer's volume between the menu and cores.
- Remote control through `/tmp/bgm.sock`, used by mrext Remote's music page.

## Compatibility notes

The Go version preserves the Python script's music folder layout, `bgm.ini` keys and file format, boot-sound and core-boot-sound rules, `_` and `X##_` filename prefixes, playlist semantics (`none`, `all`, folders, missing folders), play history, socket protocol and replies, `start`/`stop`/`restart`/`exec` commands, the `# Startup BGM` line in `user-startup.sh`, the `/tmp/bgm.log` debug log and every volume rule. Existing installations need no changes.

Intentional differences:

- The control screen uses the shared mrext TUI with themes and a controller-friendly Settings page. Playback and playlist choices are written to `bgm.ini` immediately; other settings are written when `Save` is selected.
- Running `bgm.sh` with no arguments while the service is stopped starts the service and then opens the control screen. The Python script exited after starting the service.
- Invalid values in `bgm.ini` fall back to their defaults instead of stopping the service. A `%` in a value no longer causes an error.
- Core changes are detected with a native file watch instead of an `inotifywait` child process, with a short settle so CORENAME is never read while MiSTer is still writing it.
- The service runs from a copy of the binary in `/tmp` so `bgm.sh` can be updated while music plays. The copy is removed when the service stops.
- A socket file left behind by a crashed service is removed automatically instead of blocking BGM until reboot. `restart` waits for the old service to exit before starting the new one.
- Stopping a playlist waits for the running track to be killed, so a stopped playlist can never start one more track. Unknown command-line arguments print usage instead of opening the menu.
- Playlist folders are listed alphabetically in the menu.
- Optional `music/boot/default/` tracks provide a generic core boot sound when no core-specific boot directory exists. Existing empty core-specific directories remain silent.
- Optional `bootinplaylist = yes` includes boot sounds in normal playback without changing the selected playlist. It defaults to `no`, preserving existing behavior. When enabled, random playback avoids immediately repeating a track whenever another candidate exists, including small playlists.

## Music folder

```text
/media/fat/music/
  bgm.ini              configuration
  *.mp3 …              tracks for the "none" playlist (top level only)
  _startup.mp3         boot sound for the active playlist (underscore prefix)
  X05_short_loop.mp3   plays five times in a row
  Chiptunes/           a playlist folder; subfolders are included
  Radio/station.pls    an internet radio playlist
  boot/                global boot sounds, no prefix needed
  boot/SNES/           boot sounds played when the SNES core launches
  boot/default/        optional fallback for cores without their own boot folder
```

Core boot folders are matched case-insensitively. BGM chooses a random supported track from the matching folder's top level and applies `corebootdelay`. If no core-specific directory exists, it uses `boot/default/` with the same selection and delay rules. A core-specific folder with no supported tracks suppresses the fallback. Missing or empty fallback folders remain silent; startup tracks directly inside `boot/` are not reused for core launches. No INI changes are needed to enable this: add tracks to `boot/default/`.

### Supported files

BGM plays `.mp3`, `.ogg`, `.wav`, `.mid`, `.vgm`, `.vgz` and `.vgm.gz` files through `mpg123`, `ogg123`, `aplay`, `aplaymidi` and `vgmplay` from the MiSTer Linux image. Files can be mixed freely within a playlist.

### Internet radio

Internet radio stations are played from `.pls` files containing a stream URL. The best way to manage these is a playlist folder with a single `.pls` file. Multiple `.pls` files in a folder are only moved on by skipping. The `all` playlist excludes `.pls` files.

### Playback types

- `random` (default): repeatedly pick a random track, avoiding the most recent 20% of the playlist.
- `loop`: pick one random track and repeat it until a reboot, core change or playlist change.
- `disabled`: play nothing except boot sounds.

### Playlists

Create a subfolder in `music` and fill it with tracks; it appears in the control screen as a playlist and may contain any depth of subfolders. The `none` playlist (default) uses only files in the top level of `music`. The `all` playlist combines every folder, including `boot`. The chosen playlist is saved to `bgm.ini` and restored on boot. Boot sounds are taken from the active playlist.

### Boot sounds

Prefix a file with `_` to mark it as a boot sound (for example `_Startup.mp3`). One boot sound is picked at random from the active playlist's top level when MiSTer starts, and `_` files are excluded from normal playback. Files placed directly in `music/boot` are global boot sounds that apply to every playlist and need no prefix.

To keep boot sounds in normal rotation, enable **Boot sounds in rotation** in Settings and Save, or set `bootinplaylist = yes` in `[bgm]` and restart BGM. This includes underscore-prefixed tracks within the selected playlist and files directly inside `music/boot`; no copies or renaming are needed. It does not add unrelated playlists or core-specific boot folders (unless already included by `all`). Startup selection remains unchanged. Global tracks already present in `all` are not added twice.

Settings changes apply to subsequent track selections without interrupting the current track. Random playback avoids an immediate repeat when another candidate exists; a one-track playlist can repeat. Loop playback still repeats its chosen track until the playlist restarts. Disabled playback still plays only boot sounds.

### Core boot sounds

Create a folder inside `music/boot` named after a core, such as `music/boot/SNES`, and add tracks. One is played when that core launches. The match is case-insensitive and also triggers for `.mgl` launches of the core. `corebootdelay` in `bgm.ini` waits the given number of seconds (decimals allowed) before playing, to allow a display to sync.

### Per-track looping

Rename a file with `X##_` in front, where `##` is a two-digit count with a leading zero. `X05_My File.mp3` plays five times before the next track; `X23_My File.mp3` plays twenty-three times.

### Volume control

Set both `Menu volume` and `Default volume` in Settings, or in `bgm.ini`:

```ini
menuvolume = 7
defaultvolume = 3
```

Both values must be set for the feature to work. `0` mutes, `7` is maximum, and `-1` (the default) disables the feature. MiSTer's volume is set to `menuvolume` while the menu is open and to `defaultvolume` while a core runs.

## Configuration

BGM reads `/media/fat/music/bgm.ini`. The file is created with these defaults when missing:

```ini
[bgm]
playback = random
playlist = none
startup = yes
playincore = no
bootinplaylist = no
corebootdelay = 0
menuvolume = -1
defaultvolume = -1
debug = no
```

- `playback`: `random`, `loop` or `disabled`.
- `playlist`: `none`, `all` or a folder name inside `music`.
- `startup`: start the service from `user-startup.sh` on boot.
- `playincore`: keep playing music while a core runs.
- `bootinplaylist`: include boot sounds in normal playlist playback; defaults to `no`.
- `corebootdelay`: seconds to wait before core boot sounds.
- `menuvolume`, `defaultvolume`: `-1` to `7`, see Volume control.
- `debug`: print service output and append it to `/tmp/bgm.log`. Re-read on every log line, so it can be toggled while the service runs.

Booleans accept `yes`/`no`, `true`/`false`, `on`/`off` and `1`/`0`. Keys are case-insensitive; unknown keys and sections are kept when the control screen writes the file.

### TUI settings

An optional `[tui]` section controls the control screen and is written when settings are saved:

```ini
[tui]
theme = default
mouse = yes
crt_mode = yes
on_screen_keyboard = yes
```

Available themes: `default`, `high_contrast`, `dracula`, `nord`, `gruvbox`, `monogreen`. `crt_mode` uses MiSTer's 75×15 layout. `on_screen_keyboard` enables controller-friendly text entry for the core boot delay.

## Control screen

The main screen shows the current track, playback type and playlist, followed by `Skip current track`, `Start playing`/`Stop playing`, the playback types and the playlists. The active playback type and playlist are marked `[ACTIVE]`. Up and Down select a row, Left and Right select `Select`, `Settings` or `Exit`, and Enter activates the selected button. The screen refreshes itself while music plays.

`Settings` opens a staged page for start on boot, play music in cores, core boot delay, menu and default volume, debug logging, theme, mouse, CRT mode and the on-screen keyboard. Changing the menu volume applies it to MiSTer immediately so it can be heard. Nothing is written until `Save`; `Cancel` or Escape asks before discarding changes.

## Commands

```text
bgm.sh            interactive: prepare folders, start the service if needed, open the control screen
bgm.sh start      start the service in the background if startup = yes
bgm.sh stop       stop the running service
bgm.sh restart    stop, wait for the service to exit, then start
bgm.sh exec       run the service in the foreground (used internally)
```

The first interactive run appends this hook to `/media/fat/linux/user-startup.sh`, creating the file when missing:

```sh
# Startup BGM
[[ -e /media/fat/Scripts/bgm.sh ]] && /media/fat/Scripts/bgm.sh $1
```

MiSTer passes `start` and `stop` through this hook on boot and shutdown.

## Socket protocol

The service listens on the unix socket `/tmp/bgm.sock`. Each connection carries one command of at most 4096 bytes; the service replies where noted and closes the connection.

| Command | Reply |
|---|---|
| `status` | `yes` or `no`, playback type, playlist (`none` when unset) and the current track filename, separated by tabs |
| `play`, `stop`, `skip` | none |
| `set playback <type>` | none; unknown types play random tracks but are reported verbatim |
| `set playlist <name>` | none; `none` clears the playlist, names may contain spaces |
| `set playincore yes` / `set playincore no` | none |
| `set bootinplaylist yes` / `set bootinplaylist no` | none; applies to subsequent track selections |
| `get playback`, `get playincore` | the value |
| `get playlist` | the name; nothing when no playlist is set |
| `pid` | the service process id |
| `quit` | none; stops the socket listener |

Commands are processed one at a time and wait while a core boot sound plays, exactly as before.

## Logging

With `debug = yes`, every service message is printed and appended to `/tmp/bgm.log` with an ISO 8601 timestamp, including the output of the audio players.

## Uninstall

Remove BGM's Downloader subscription if configured, so updates do not reinstall it.

1. Close the control screen, then run `/media/fat/Scripts/bgm.sh stop` and wait for the service and audio player to exit. Do not reopen the screen afterward: it can start the service again.
2. Remove the `# Startup BGM` comment and its `bgm.sh $1` command from `/media/fat/linux/user-startup.sh`. Preserve all other startup entries.
3. Delete `/media/fat/Scripts/bgm.sh`. Optionally back up and remove `/media/fat/music/bgm.ini` to discard BGM settings.
4. Keep `/media/fat/music/`: tracks, playlists, `.pls` subscriptions, and `boot/` sounds are user content. BGM has no separate persistent database or generated menu tree to remove.

After shutdown, leftover `/tmp/bgm.sock`, `/tmp/bgm.log`, and `/tmp/bgm.sh` can be removed or left until reboot. Keep MiSTer's audio players installed.

## Local development

Run BGM against a temporary MiSTer root without touching `/media/fat`:

```bash
go run ./cmd/bgm --root /tmp/bgm-mister
```

The root gains `music`, `tmp`, `dev` and `linux` folders. `tmp/CORENAME` stands in for MiSTer's core name file, `dev/MiSTer_cmd` is a regular file that records volume commands, and the service socket is `tmp/bgm.sock`. Put stub `mpg123`, `aplay` and similar scripts first on `PATH` to test without audio hardware. All commands accept `--root`:

```bash
echo MENU > /tmp/bgm-mister/tmp/CORENAME
go run ./cmd/bgm --root /tmp/bgm-mister start
printf status | nc -U /tmp/bgm-mister/tmp/bgm.sock
go run ./cmd/bgm --root /tmp/bgm-mister stop
```

Keep the root path short: unix socket paths are limited to about 100 characters.
