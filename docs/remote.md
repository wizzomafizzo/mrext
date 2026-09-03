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
* Control [BGM](https://github.com/wizzomafizzo/MiSTer_BGM) music playback
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

## Uninstall

After opening `remote` from the `Scripts` menu, there is an option available to uninstall Remote called `Uninstall`. You can also run `remote.sh -uninstall` from the console or via SSH.

### Manual

To manually uninstall Remote from your MiSTer, delete these files from the SD card:

* `Scripts/remote.sh`
* `Scripts/remote.ini` (if present)
* `search.db` (this file is also used by Search if you have it installed)

If you have an active wallpaper set by Remote, you will also need to remove `menu.png` or `menu.jpg`. These are just links to the actual file in the `wallpapers` folder.

Finally, remove the following lines from `linux/user-startup.sh`:
```
# mrext/remote
[[ -e /media/fat/Scripts/remote.sh ]] && /media/fat/Scripts/remote.sh -service $1
```
