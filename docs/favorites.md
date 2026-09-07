# Favorites

Favorites creates and manages game and core shortcuts in MiSTer's stock menu. It runs as a standalone Go application and does not require Zaparoo Core.

Zaparoo Frontend also provides favorites for indexed media, but it does not create or repair stock-menu `.mgl` shortcuts. See [MIGRATE.md](../MIGRATE.md) before choosing between them.

## Installation

Install Favorites through the combined mrext Downloader database documented in the [README](../README.md). Downloader installs the ARMv7 binary as:

```text
/media/fat/Scripts/favorites.sh
```

The `.sh` filename is retained for compatibility with existing Downloader entries, startup hooks, and MiSTer Scripts-menu behavior. It is a native executable, not a shell script.

Run `favorites` from MiSTer's Scripts menu.

## Features

- Add game, core, arcade-core, and existing MGL shortcuts.
- Browse SD, USB, CIFS, symlinked game folders, and supported files inside ZIP archives.
- Generate MGL files from Zaparoo's maintained MiSTer launcher catalog.
- Preserve NeoGeo ZIP naming from `romsets.xml` and relative NeoGeo launcher paths.
- Create, rename, move, and remove Favorites folders and entries.
- Refresh broken dated core symlinks after core updates.
- Maintain arcade `cores` links needed by MRA favorites. These link directly to `/media/fat/_Arcade/cores`, so newly downloaded cores are available without refreshing Favorites.
- Operate with a controller through an on-screen keyboard.

## Compatibility notes

The Go version preserves Python Favorites' folder discovery, top-level destinations, nested folder management, startup refresh hook, LLAPI/YC selection, root filtering, external-drive shortcut, ZIP traversal, NeoGeo names and relative paths, dated-core repair, and safe link-based core favorites. Symlinked game directories remain browsable.

Arcade link management creates missing `cores` links when the interactive app opens and when it creates a Favorites folder. Existing files, directories, and symlinks named `cores` are preserved, not repaired or replaced. If an older Favorites folder uses a copied `cores` directory or an incorrect link, back it up and move it aside before reopening Favorites. MRAs that require a custom cores directory rather than `_Arcade/cores` are not automatically supported; Favorites does not merge multiple core directories.

Core-update repair now recognizes the date in the symlink target rather than requiring it in the favorite's name. Renaming a core favorite no longer prevents its repair after an update.

Zaparoo's maintained MiSTer catalog now supplies system aliases, extensions, RBF paths, MGL slots, set names, and reset timing. Canonical definitions intentionally replace stale Python-table behavior: Genesis uses the current MegaDrive core path, Vectrex `.ovr` overlay files are not treated as games, and Atari 7800 images placed in the Atari 2600 folder are no longer accepted. NeoGeo ZIP support remains as an explicit compatibility extension until present in the standalone catalog. ZIP and `.neo` launchers use the same catalog ROM-slot parameters and retain Favorites' relative NeoGeo paths.

## Moving to another SD card

Favorites are stored directly in menu folders, not in a separate database. The default folder is `/media/fat/_@Favorites` (shown as `@Favorites` in MiSTer's menu). Also copy any custom top-level Favorites folders recognized by your `default_folder` and `folder_name_contains` settings, including their subfolders. Keep `/media/fat/Scripts/favorites.ini` to preserve those settings.

These folders contain generated `.mgl` files and symbolic links to existing cores or launchers. Arcade favorites can also depend on a `cores` link. Copy links as links, not as copies of their targets, and preserve the referenced games, cores, and folder layout. A copied target can still appear in MiSTer's menu without behaving like the original favorite.

Keep the original card or a backup until the new card is verified. Use a copy tool's preserve-symbolic-links option, not its follow-links option. On systems where both filesystems expose and support symbolic links, `cp -a` preserves them; this is not a guarantee for every macOS/Windows SD-card copy workflow. Verify the result on MiSTer rather than relying only on the desktop file browser.

After migration:

- Check the actual folder names and `favorites.ini` if entries appear in MiSTer's menu but not in Favorites.
- Inspect a known link with `ls -l` or `readlink` on MiSTer and compare its target with the original. Not every entry is a link: generated `.mgl` files are regular files.
- Confirm linked targets exist at the same MiSTer paths, including any USB or network storage, then test representative game, core, and arcade shortcuts.
- If links became regular copies, restore them from the original card with link-preserving copying or recreate the affected favorites. Back up the migrated folders before removing duplicates.

The `refresh` command repairs supported broken core links; it does not reconstruct links that a copy tool replaced with regular files or rewrite every game path after storage moves.

The new opt-in `ra` alternate-core mode adds RetroAchievements game shortcuts. Standard, LLAPI, and YC behavior remains unchanged; no existing favorites are migrated automatically.

## Settings screen

Choose `Settings` from the main screen's bottom action bar. Up and Down select a setting; Left and Right select `Change`, `Save`, or `Cancel`. The footer shows one sentence explaining the selected setting. `Change` toggles boolean values, opens an alternate-core or theme picker, or opens the appropriate editor. Pickers highlight the current value and do not change it when canceled.

Changes remain staged until `Save` is selected. `Cancel` leaves the file untouched, and Back or Escape asks before discarding unsaved changes. Saving atomically updates managed Favorites, core, and TUI keys while retaining unknown INI sections and comments. Notes attached to repeated or removed keys may move to section or file comments. Theme, mouse, layout, and other runtime settings apply after saving.

`Folder name matches` and `Additional game folders` are marked as lists. Their editor shows one entry per row with `Add`, `Edit`, and `Remove` controls. `Done` keeps edits in staged settings; `Cancel` discards the list edits. Enter one name fragment or full folder path at a time, without comma separators. Additional game folders are still saved as repeated `games_folder` keys for compatibility with manual configuration.

`USB shortcut folder` is the filesystem target of the browser's USB shortcut, normally `/media/usb0`. If that target contains a `games` folder, the shortcut opens it instead. This setting does not mount storage or change the destination of saved favorites; the INI key remains `external_folder`.

The on-screen keyboard uses Zaparoo's compact 41×8 layout. Where an input has a prompt explaining what to enter, that line is shown above the keyboard; controller users could not see it otherwise, and they are the ones it is written for. Setting explanations stay in the settings page footer.

### Shared interface settings

Theme, mouse, CRT mode and on-screen keyboard can be set once for every mrext app instead of app by app.

Each app reads these in order: its own `[tui]` section, then `/media/fat/Scripts/.config/mrext/tui.ini`, then the built-in defaults. An app's own INI still wins for any key it declares, so nothing changes until you ask for it.

The settings screen has a **Use in all apps** row. It writes the current interface settings to the shared file and removes those four keys from this app's INI, so this app follows the shared file from then on. Other apps pick it up the next time they start, unless their own INI still sets those keys.

## Configuration

Favorites reads `favorites.ini` beside the executable:

```text
/media/fat/Scripts/favorites.ini
```

First interactive launch creates this file when missing. Existing files are never replaced. Downloader does not install or overwrite it.

```ini
[favorites]
default_folder = _@Favorites
folder_name_contains = fav
create_default_folder = true
manage_arcade_core_links = true
hide_root_files = true
external_folder = /media/usb0
core_prefix =
; games_folder = /media/network

[cores]
all =

[tui]
theme = default
mouse = true
crt_mode = true
on_screen_keyboard = true
```

### Favorites settings

- `default_folder`: folder created when no matching Favorites folder exists. Must start with `_` and is always recognized as a Favorites root, regardless of `folder_name_contains`.
- `folder_name_contains`: comma-separated, case-insensitive fragments used to recognize top-level Favorites folders. Folder names must also start with `_`.
- `create_default_folder`: create `default_folder` when needed.
- `manage_arcade_core_links`: create required `cores` symlinks without replacing existing files or links.
- `hide_root_files`: show only standard MiSTer menu/game folders at SD-card root while browsing.
- `external_folder`: target of root-browser USB shortcut. Its `games` child is preferred when present.
- `core_prefix`: optional prefix applied to standard catalog RBF paths.
- `games_folder`: optional additional game root. Repeat key for multiple roots. Built-in MiSTer roots remain enabled.

### Alternate cores

`cores.all` preserves existing Favorites configuration:

- blank: standard cores
- `llapi`: use matching LLAPI variants when installed
- `yc`: use matching YC variants when installed
- `ra`: use installed RetroAchievements variants from `_RA_Cores/Cores`

Favorites falls back to standard catalog core when requested variant is unavailable. Select **RetroAchievements** in Settings, or use `[cores]` with `all = ra`. This affects newly created game favorites; existing shortcuts are not rewritten.

RA paths and setnames follow Zaparoo Core's MiSTer launcher mappings. Generated shortcuts include the appropriate `RA_*` setname and `same_dir` behavior, including separate FDS/GBC/Game Gear/Super Game Boy/NeoGeo CD/TurboGrafx-CD names. Atari 2600 uses the RA Atari7800 core's loading slot. Standard catalog data continues to supply media formats and loading parameters.

Install and configure a compatible RetroAchievements MiSTer binary and RA cores first. Favorites does not install binaries, edit `MiSTer.ini`, or handle credentials. An installation that selects `MiSTer_RA` using `[RA_*]` must already have that rule configured. Availability of a core file does not prove that achievements or hardcore mode work on the device.

The RA adaptation is sourced from Zaparoo Core `pkg/platforms/mister/launchers.go` (revision `7cae7f1f`); it can move to a shared library when Core exposes these variant mappings. Ordinary core-file favorites remain direct links to the file selected by the user.

### TUI settings

Available themes:

- `default`
- `high_contrast`
- `dracula`
- `nord`
- `gruvbox`
- `monogreen`

`crt_mode` uses MiSTer's 75×15 layout. `on_screen_keyboard` enables controller-friendly text input. `mouse` enables pointer input where available.

List screens preserve Favorites' original controller model: Up and Down always change rows, Left and Right always change the selected bottom button, and Enter activates that button. Rows and button selection remain visibly active together; no focus-switching step is required.

When adding or moving a favorite file, the destination picker also offers `Create Folder`. Creating or canceling a folder returns to the pending destination selection without losing the selected favorite.

The main screen retains Add Favorite as first row so existing controller flow stays unchanged, but presents it as a styled `+ Add favorite` action followed by an `Existing favorites` section instead of the legacy `<ADD NEW FAVORITE>` placeholder. Folder and navigation rows use text markers that remain understandable without color.

Default styling uses MiSTer-safe named terminal colors: dark blue and blue backgrounds, white primary text, gray secondary text, yellow borders/actions/selections, and red errors. Bold labels and textual markers carry the same meaning when a display reduces the palette. Alternate themes remain optional and never provide the only indication of state.

## Local development

Run Favorites against a temporary MiSTer root without creating `/media/fat`:

```bash
go run ./cmd/favorites --root /tmp/favorites-mister
```

Favorites creates the root and `Scripts` directory when missing. Add fixture games under paths such as `/tmp/favorites-mister/games/SNES`. Local configuration is stored at `/tmp/favorites-mister/Scripts/favorites.ini`; all menu changes stay inside the selected root.

The non-interactive refresh command also accepts this mode:

```bash
go run ./cmd/favorites --root /tmp/favorites-mister refresh
```

## Refresh command

Favorites preserves non-interactive startup behavior:

```bash
/media/fat/Scripts/favorites.sh refresh
```

Refresh repairs shortcuts whose target is a dated core (such as `NES_20260101.rbf`), even when the favorite has a custom name such as `Nintendo.rbf`. Custom names are preserved; dated favorite names follow the replacement core's date. Both absolute and relative symlink targets are supported. Broken shortcuts without a replacement are removed, unless the target lives on a removable or network drive that is not attached: refresh also runs at boot, before USB and network mounts settle, so favorites pointing at an unplugged drive or a powered-off NAS are left alone rather than deleted. Interactive launch adds the same command to an existing `linux/user-startup.sh` only when no Favorites startup entry exists.

## Uninstall

Remove Favorites' Downloader subscription if configured, so updates do not reinstall it.

1. Exit Favorites and let any refresh finish. There is no persistent Favorites service to stop.
2. Remove the `# Startup favorites` comment and its `favorites.sh refresh` command from `/media/fat/linux/user-startup.sh`.
3. Delete `/media/fat/Scripts/favorites.sh`. Optionally back up and remove `/media/fat/Scripts/favorites.ini` to discard settings, respecting any shared `MREXT_CONFIG` override.
4. Keep `/media/fat/_@Favorites` and other favorite folders unless you want to remove your curated shortcuts. The folder can be renamed through `default_folder`, and favorites may also be at the SD root or in other menu folders. There is no Favorites database to delete.

If discarding generated entries, remove only selected `.mgl` files and symlinks after inspecting their contents and ownership. Favorites can create `cores` symlinks in the SD root and favorite subfolders. Leave these if other MRA shortcuts use them; otherwise remove only the links, never their `_Arcade/cores` target. Remove symlinks themselves, not their destinations. Keep real core files and custom content.

## Safety

Favorites never recursively deletes a folder containing user files. Folder deletion succeeds only when folder is empty or contains only managed `cores` symlink. Existing configuration files, non-symlink `cores` entries, and unrelated menu content are left untouched.
