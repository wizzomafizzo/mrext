# Remote

> [!IMPORTANT]
> Remote is maintained for legacy users. Before building a new MiSTer remote control with it, consider Zaparoo App/Web UI, Frontend, or the public Core API and CLI. Remote can remain appropriate for stock-menu file management, wallpaper browsing, arbitrary `MiSTer.ini` editing, and other gaps listed in [MIGRATE.md](../MIGRATE.md).

Remote is a web-based interface with a stack of modern features to manage all aspects of your MiSTer. Can be used from your phone, tablet or computer.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/remote.sh"><img src="images/download.svg" alt="Download Remote" title="Download Remote" width="140"></a>

[API Documentation](remote-api.md)

## Features

* Control MiSTer directly with a virtual remote control interface
  * Includes all common media keys and hotkeys
  * Full on-screen keyboard and keypad
  * Virtual mouse trackpad with clicks, drag and desktop mouse capture
* Launch cores and game shortcuts with an in-app version of the MiSTer menu
  * Move, rename and delete anything in the menu
* Search and launch your entire game collection
  * Create shortcuts in the menu from results
* Browse and launch all cores installed on your MiSTer
* View, browse, download and take new screenshots
* Control [BGM](https://github.com/wizzomafizzo/mrext#bgm) music playback
* Browse and activate wallpapers
* Launch scripts from the Scripts menu
* Change all MiSTer.ini file settings
  * Set the current active .ini file
  * Set hostname and MAC address settings
* Auto-discover and connect to other MiSTers on your network running Remote
* Quickly view system information (disk usage, network settings, last update, etc.)

## Install

Download [Remote](https://github.com/wizzomafizzo/mrext/releases/latest/download/remote.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add this database to `downloader.ini` in SD card root, then run `downloader` or `update` to receive updates:

```ini
[mrext/remote]
db_url = https://raw.githubusercontent.com/wizzomafizzo/mrext/main/releases/remote/remote.json
```

Once installed, run `remote` from the MiSTer `Scripts` menu, and a prompt will offer to enable Remote as a startup service.

This service must be running to use Remote's web UI or API.

## Usage

From a web browser, navigate to `http://<mister_ip>:8182` to access Remote. The `remote` app in the `Scripts` menu will display the exact address to use if you're not sure.

## Rebuilding the search index

Use the regenerate button in Search after moving, renaming, or deleting games. A successful rebuild replaces all names and system metadata in `/media/fat/Scripts/.config/mrext/games.db`, shared with Search and LaunchSync. The legacy root-level `search.db` is not the current index.

Progress and failures appear in the existing indexing status flow. If regeneration fails, the previous index remains usable and the error stays visible until another attempt. Rebuilds stream game matches into bounded database batches on SD rather than keeping a whole-library list in RAM. Directory/ZIP listings and directory-cycle tracking still use memory.

A rebuild needs space beside `games.db` for a replacement index. Do not delete the adjacent `games.db.lock` while any app is indexing; it serializes writers across index replacement. Failed attempts normally clean up their temporary `.games-index-*` files. After a hard interruption, leftover files with that prefix can be removed only when no indexer is running. Mount the libraries you want included before regenerating: absent libraries are omitted from a successful full rebuild.

## Screenshot compatibility notes

The Screenshots-page camera and `POST /screenshots` use the same normal screenshot shortcut as Control's Screenshot button. This avoids the separate command-interface capture path reported to stretch PSX images. Raw Screenshot remains a separate Control action.

Both normal screenshot actions depend on MiSTer accepting `Alt+Scroll Lock`; PS/2 keyboard mode can disable that shortcut. The API reports keyboard-send failures, but its one-second delay does not confirm a screenshot was saved. Check the screenshot list afterward.

## MiSTer SAM compatibility

After successful game, core, file, or launch-token requests, Remote signals user activity through SAM's existing `/tmp/.SAM_tmp/SAM_Joy_Activity` file. Current SAM MCP recognizes the `zaparoo` message as external activity: in normal mode it resets idle and exits attract mode while keeping the current game, rather than returning to Menu. This supports stock MiSTer without requiring Zaparoo Core.

Remote never creates the activity file or starts SAM. A missing file is ignored; notification errors are logged without failing the launch. SAM must be running with its activity handling enabled (`listenjoy`); M82/kiosk mode retains its own behavior. Delivery is best-effort, not acknowledged. Older SAM versions using a different path are not targeted.

The protocol is implemented in [SAM MCP's activity poller and action handler](https://github.com/mrchrisster/MiSTer_SAM/blob/45af68dd7a7e1b15337c2b04c79f196e7dd6da47/.MiSTer_SAM/MiSTer_SAM_MCP.py).

## Arcade tracking

Remote's shared tracker now supplies a resolved `.mra` path for arcade games recognized through Arcade Database. It searches nested installed arcade folders and confirms the active set name against MRA XML, rather than relying on display names. Ambiguous or unreadable matches remain unresolved. Existing event fields are unchanged, and `/tmp/ACTIVEGAME` retains its legacy arcade set-name value.

## Startup diagnostics

Remote retries tracker setup when required MiSTer state files or directories are missing: at most six attempts, with one-second gaps. Each retry is logged. Permission errors and other failures stop immediately. This does not retry a missing `remote.sh` before the executable starts, or missing input devices.

The HTTP listener is bound before device/tracker setup. A port conflict therefore follows normal startup-error logging and PID cleanup rather than terminating from the background HTTP goroutine. A service-start command still launches a background process; its return is not an HTTP-readiness acknowledgment.

Remote verifies that a PID belongs to its copied executable running the service subcommand before treating it as running or sending it a stop signal. Numeric PID-file format is unchanged. This prevents an unrelated process using a stale PID from being mistaken for Remote.

If startup fails, capture `/tmp/remote.log`, `/tmp/remote.pid` (if present), and `/media/fat/linux/user-startup.sh` before restarting or re-enabling Remote. Also note executable location, mounted storage, and whether the recorded PID exists. These checks harden startup but do not establish the cause of intermittent reboot failures.

## Main configuration

Settings always offers Main, even if `MiSTer.ini` is absent. Loading it uses blank defaults without writing a file. Save creates `/media/fat/MiSTer.ini` with a `[MiSTer]` section and the selected settings; the Save screen explains this. `MiSTer_example.ini` is never offered as an active configuration or edited. Saving Main leaves alternate INIs unchanged, and the menu is relaunched only after saving succeeds.

## INI slot identity

Remote follows current MiSTer `cfg_get_name` ordering: take the first three `MiSTer_*.ini` names in filesystem directory order, then sort that subset case-insensitively. Main remains slot 1. The example file can consume a MiSTer slot but is not editable in Remote, so displayed IDs can have gaps. Custom names are shown in full rather than truncated to four characters.

Each settings page shows the file being edited. Save targets that loaded filename even if the active slot changes elsewhere. Switching files asks before discarding edits; missing values in the new file return to defaults. Failed loads disable Save and report the error.

Remote retains its first observed slot layout and refuses further access if it detects a change. **After adding, removing, or renaming alternate INIs, restart both MiSTer and Remote before editing.** MiSTer caches its mapping internally; Remote cannot reconstruct a different mapping cached before Remote started. Older firmware with different ordering is not verified. No INI files are renamed to force an order.

The `announce_game_url` webhook is sent in the background. Earlier versions posted it inline while the tracker was locked, so an endpoint that had gone offline froze core and game tracking for the full 15-second timeout on every core change. Requests are queued; if the endpoint cannot keep up, events are dropped and logged rather than delaying tracking.

Adding or removing a startup entry preserves whatever `#!` line `user-startup.sh` already has, so a `#!/bin/bash` script is not rewritten to `#!/bin/sh`. Uninstalling when Remote is the only entry now succeeds; it previously reported "no startup entries to save" and left the entry in place, so the service returned on the next boot.

`MiSTer.ini` and `u-boot.txt` are written by staging the replacement beside the original and renaming over it, so losing power mid-save leaves the previous file intact rather than a truncated one. Each keeps a single `.backup` of its contents from before mrext first changed it; that backup is no longer overwritten on every save.

Changing the MAC address preserves the rest of `u-boot.txt` exactly, including comments and line order, and the value must be a valid MAC. Saving settings now reports which ones could not be applied instead of returning success regardless.

`-service start`, `stop` and `restart` print why they failed on the console as well as to `/tmp/remote.log`; they previously exited 1 with no output at all. `restart` no longer waits forever for a wedged daemon: after 20 seconds it escalates to `SIGKILL` and starts the new one.

## Uninstall

Remove Remote's Downloader subscription if configured, so updates do not reinstall it.

### Built-in uninstall

The Scripts-menu interface offers **Uninstall**. The console equivalent is `/media/fat/Scripts/remote.sh -uninstall`.

This requests service shutdown, removes the `mrext/remote` startup entry, deletes legacy `/media/fat/search.db` if present, and removes `menu.jpg`/`menu.png` **when they are symlinks**. It does not check who created those links. Back up the legacy database if older apps still use it; choose manual uninstall below if you want to retain the active wallpaper links. Original wallpaper images are not removed.

The command leaves `Scripts/remote.sh`, `Scripts/remote.ini`, and the current shared `Scripts/.config/mrext/games.db` in place. Check reported errors and confirm the service has stopped, then remove the binary yourself and optionally preserve or delete its INI. Keep any INI selected through `MREXT_CONFIG` if another app uses it.

### Manual uninstall

1. Run `/media/fat/Scripts/remote.sh -service stop` and wait for shutdown. Close Remote's browser interface.
2. Back up `/media/fat/linux/user-startup.sh`, then remove only this block (paths may be quoted or customized):

   ```sh
   # mrext/remote
   [[ -e /media/fat/Scripts/remote.sh ]] && /media/fat/Scripts/remote.sh -service $1
   ```

3. Delete `/media/fat/Scripts/remote.sh`. Optionally back up and remove `/media/fat/Scripts/remote.ini` to discard settings, respecting any shared configuration override.
4. Keep `/media/fat/Scripts/.config/mrext/games.db` for Search or LaunchSync. The legacy `/media/fat/search.db` is optional cleanup only when no older app needs it. Do not delete the shared `.config/mrext` directory.
5. Keep wallpaper images, screenshots, and menu shortcuts you created through Remote. If you no longer want the active wallpaper, remove only verified `menu.png`/`menu.jpg` symlinks, not their targets or regular files at those paths. Check `cores` links before removing any generated MRA shortcuts; other menu entries may still need them.

After shutdown, `/tmp/remote.pid`, `/tmp/remote.log`, and `/tmp/remote.sh` are optional cleanup. Uninstall does not reset MiSTer settings, network settings, or SSH authorization; keep shared INIs and `authorized_keys` files.
