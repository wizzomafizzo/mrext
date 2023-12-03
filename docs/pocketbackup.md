# PocketBackup

PocketBackup is a simple utility to backup the following files from an Analogue Pocket to your MiSTer via USB:

- Saves
- Save states
- Screenshots
- Settings

These files will be backed up to a folder called `pocket` on your MiSTer's SD card.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/pocketbackup.sh"><img src="images/download.svg" alt="Download PocketBackup" title="Download PocketBackup" width="140"></a>

## Install

Download [PocketBackup](https://github.com/wizzomafizzo/mrext/releases/latest/download/pocketbackup.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` or `update_all` scripts:
```
[mrext/pocketbackup]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/pocketbackup/pocketbackup.json
```

## Usage

Before using for the first time, enable the `USB SD Access` option in the `Tools > Developer` menu on your Analogue Pocket. This will let you plug the Pocket directly into your MiSTer via USB.

1. Plug in your Analogue Pocket via USB on the MiSTer
2. Run `pocketbackup` from the MiSTer `Scripts` menu
