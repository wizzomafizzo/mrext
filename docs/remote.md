# Remote

Remote is a web-based interface with a stack of modern features to manage all aspects of your MiSTer. Can be used from your phone, tablet or computer.

*NOTE: Remote is in active development. It's totally safe to try out and use, but some sections of the app may be marked as not working or only have basic functionality. Features are also designed for mobile first, with responsive design for tablet and computer added later.*

<a href="https://github.com/wizzomafizzo/mrext/raw/main/releases/remote/remote.sh"><img src="images/download.svg" alt="Download Remote" title="Download Remote" width="140"></a>

## Install

Download [Remote](https://github.com/wizzomafizzo/mrext/raw/main/releases/remote/remote.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` script:
```
[mrext/remote]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/remote/remote.json
```

Once installed, run `remote` from the MiSTer `Scripts` menu, and a prompt will offer to enable Remote as a startup service.

This service must be running for remote to work, but it has no impact on your MiSTer's performance.

## Usage

From a web browser, navigate to `http://<mister_ip>:8182` to access Remote. The `remote` app in the `Scripts` menu will display the exact address to use if you're not sure.

