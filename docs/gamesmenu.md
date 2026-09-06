# GamesMenu

GamesMenu mirrors game libraries into `/media/fat/_Games`, letting you launch games directly from the stock MiSTer menu without opening a core first. It does not require Zaparoo Core.

For new Zaparoo integrations, prefer Frontend library browsing and Core's public interfaces. Frontend does not mirror libraries into stock-menu folders; see [migration guidance](../MIGRATE.md).

## Installation

Download [gamesmenu.sh](https://github.com/wizzomafizzo/mrext/releases/latest/download/gamesmenu.sh) into `/media/fat/Scripts/`, then run it from MiSTer's Scripts menu. Despite its suffix, this is a native Go binary: **keep the `.sh` filename**. Downloader's existing `Scripts/gamesmenu.sh` entry upgrades in place; existing games and menu folders need no changes.

## Main screen

Each row represents an existing games folder supported by the Zaparoo MiSTer catalog's MGL definitions. Shared folders appear once. Names are sorted by folder key and displayed using `names.txt` replacements. The footer shows the selected row's source paths.

- **Toggle** changes `[x]` (include) to `[ ]` (exclude), or back.
- **All/None** selects every folder if any are unselected; otherwise it clears every selection. It changes selection only, not files. Generate still asks for confirmation before deleting deselected folders or the whole menu.
- **Generate** creates missing shortcuts for selected folders.
- **Clean Up** explicitly removes broken shortcuts.
- **Settings** opens staged interface settings.
- **Exit**, or Back/Escape on the main screen, exits without changing the menu.

Use Up/Down to choose rows, Left/Right to choose footer buttons, and Select/Enter to activate the highlighted button. Mouse navigation is optional; every action works with a controller. On first run, all discovered folders are selected and a welcome message appears. Otherwise selection reflects existing `_Games` folders, not a separate saved list.

## Generating the menu

**Large sets can consume substantial SD-card space**, even though each shortcut is small. Generation is additive within selected folders: existing filenames are skipped, including hand-edited shortcuts. It does not run system hooks or launch cores.

Shortcuts mirror source subfolders, prefixing each menu folder with `_`. For example:

```text
/media/fat/games/NES/Packs/set.zip/dir/game.nes
/media/fat/_Games/_NES/_Packs/_dir/game.mgl
```

The ZIP filename adds no folder level. Duplicate game basenames in the same output folder use the first source encountered; subsequent collisions are skipped. Like Python, each source directory's files and ZIPs are processed before descending into subfolders, so a parent ZIP member wins over a colliding loose game below. Regular files, ZIP members, source roots and catalog MGL slots determine which core each shortcut uses. Built-in SD, USB and network roots are scanned, plus configured `games_folder` roots.

**Generate removes every deselected directory under `_Games`, including custom user folders.** Plain files at the `_Games` root remain untouched. Before removal, a confirmation shows the total and lists affected folders; choose No to preserve everything and return to your selections. In long dialogs, use Up/Down to scroll through every folder, Left/Right to choose Yes or No, and Select/Enter to activate that choice. Scrolling never confirms removal. With nothing selected, Generate asks whether to remove the entire Games menu, including its root-level files. Removal returns to the main screen.

Progress shows the current source folder and shortcut count. Completion reports created, skipped and failed shortcuts, removed folders, and up to ten error details. Long summaries and errors also scroll with Up/Down. Leave GamesMenu running until it finishes; progress is not cancellable.

## Clean Up

Clean Up asks for confirmation, then checks `.mgl` targets. It removes shortcuts when the referenced file or ZIP member is missing, and prunes empty nested menu folders. System-level folders remain so toggle state is preserved. Both Go-generated shortcuts and legacy Python shortcuts with absolute paths and raw ampersands are understood.

Malformed, unreadable or ambiguous shortcuts, invalid/unreadable archives, and core-only launchers are kept. Cleanup checks every file target in custom multi-file MGLs; an unknown target prevents deletion. Menu symlinks are not traversed. The summary reports checked and removed shortcuts, pruned folders and unreadable paths.

**Reconnect removable drives and mount network libraries before cleanup.** A disconnected library can look like missing games. Cleanup is never automatic and has no undo; back up custom shortcuts first.

## Settings and configuration

Settings contains **Theme**, **Mouse** and **CRT mode**. Change stages edits; Save writes and applies them. Cancel/Back asks before discarding unsaved edits. CRT mode centers a 75 × 15 interface on larger screens. No on-screen keyboard option is offered because GamesMenu needs no text entry.

On first interactive run, GamesMenu creates `Scripts/gamesmenu.ini` beside the binary only if missing:

```ini
[tui]
theme = default
mouse = true
crt_mode = true

[systems]
; Add custom game roots by repeating this key. Built-in MiSTer roots remain enabled.
; games_folder = /media/network
```

Available themes: `default`, `high_contrast`, `dracula`, `nord`, `gruvbox`, `monogreen`. Settings saves preserve unknown keys, sections and comments using an atomic replacement; existing configuration is never replaced with defaults. Shared `MREXT_CONFIG` and `MREXT_APP_PATH` overrides are supported.

To add libraries, repeat `games_folder` under `[systems]`. Each root and its `games` subfolder are considered. Built-in roots remain enabled on MiSTer. Shared `[systems] set_core` overrides are honored when generating new shortcuts, for example `set_core = NES:_Console/Custom`. Existing shortcuts are not rewritten when overrides change.

## names.txt

`/media/fat/names.txt` supplies display and menu folder names:

```text
SNES: Super Nintendo
NES: Famicom/Disk
```

Keys are trimmed and compared case-insensitively; the first matching line wins. Replacement values are trimmed, `/` becomes ` & `, and other filename-illegal characters become spaces. Every path component is looked up, including nested folders. The example creates `_Super Nintendo` and `_Famicom & Disk` under `_Games`.

Existing menu folder spelling is reused case-insensitively. Changing a replacement after generation changes the desired output folder name: review the removal confirmation carefully before generating again.

## Compatibility notes

Preserved: `_Games` at the SD root, `_`-prefixed system and nested folders, `names.txt` rules, mirrored folder layouts, flattened ZIP contents, never overwriting existing shortcuts, selection derived from folder existence, removal of deselected folders, first-run welcome, and Exit without menu changes. The Python reference remains at `scripts/gamesmenu.sh` but is no longer shipped as the release artifact.

Intentional differences:

- Shared controller-first TUI, `[x]`/`[ ]` markers, staged Settings, removal confirmations and result summaries. Removing the entire menu returns to the main screen instead of exiting.
- Systems and MGL slots come solely from Zaparoo's MiSTer catalog, not the old script's 44-entry table. Every supported MGL games folder can appear; Arcade is excluded. Shared folders choose systems per extension, such as `.gbc` in `GAMEBOY` and `.gg` in `SMS`. Candidate priority is folder position in the catalog system, then system ID; longer suffixes match first.
- Catalog RBF paths, slot indices, delays, extensions, set names and reset flags replace the script's table. Genesis now uses `_Console/MegaDrive`; FDS and Gameboy Color shortcuts carry their catalog set names, and Jaguar carries its reset flag. **The `ATARI2600` folder no longer generates shortcuts for `.a78` or `.bin`, and `VECTREX` no longer generates them for `.ovr`.** Existing shortcuts for these files remain untouched. See [supported systems](systems.md); no separate GamesMenu catalog is maintained.
- New MGLs use XML-escaped, `../../../../..`-prefixed absolute target paths and honor `set_core`. Legacy MGL bytes remain untouched. Filename collisions are compared case-insensitively, matching MiSTer's SD filesystem behavior.
- Source files and subfolders are each scanned in name order rather than Python's filesystem-dependent order; ZIP members retain archive order. Current-directory files still precede all subfolders. If same-level sources collide, the first name wins.
- Existing `_NeoGeo` and `_ATARI2600` folder spelling survives catalog casing differences. Source directory symlinks are followed with loop protection. Additional network and configured roots are supported.
- Dotfiles, AppleDouble files and hidden ZIP members are skipped. Absolute and parent-traversing ZIP paths are rejected, but harmless `./` and repeated separators are accepted. Exact ZIP member names and Python's `_.`/`_` menu-folder quirks are preserved (for example, `./game.nes` creates `_./game.mgl`). Output writes are confined to the SD root; symlinks escaping it fail rather than writing elsewhere. Exclusive file creation protects existing shortcuts against concurrent writers.
- Optional `gamesmenu.ini` and explicit, conservative Clean Up are new. Errors are reported rather than silently discarded (invalid source ZIPs are still silently ignored).

## Local development

Use a disposable filesystem, not a mounted live MiSTer:

```sh
mkdir -p /tmp/gamesmenu-mister/games/NES
# Place fixture ROMs or ZIPs under games/NES.
go run ./cmd/gamesmenu --root /tmp/gamesmenu-mister
```

This relocates `_Games`, `names.txt` and `Scripts/gamesmenu.ini` and disables built-in device root discovery. Explicit extra roots in the fixture INI remain enabled. Tests use temporary fixtures and simulated terminal screens. Device testing requires explicit approval; never point destructive tests at a live installation.
