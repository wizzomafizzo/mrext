# LaunchSeq

LaunchSeq is a tool to launch games in alphabetical or random order, and automatically swap to the next game after a set amount of time. Instead of regression testing an entire library by manually loading and selecting each game, let LaunchSeq handle that for you and enjoy the show!

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/launchseq.sh"><img src="images/download.svg" alt="Download LaunchSeq" title="Download LaunchSeq" width="140"></a>

## Install

Download [LaunchSeq](https://github.com/wizzomafizzo/mrext/releases/latest/download/launchseq.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `downloader` or `update_all` scripts:
```
[mrext/launchseq]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/launchseq/launchseq.json
```

## Usage

Launching the script over an SSH session is preferred so you can see output, but it can also be launched from MiSTer's Scripts menu.

### From Scripts menu

* Launch as normal from the Scripts menu for N64 with default delay of 40 seconds
* Restart the MiSTer to stop
* Use a [custom launcher](#custom-launchers) to change the options

You can also use [Remote](https://github.com/wizzomafizzo/mrext#remote) to see what the currently loaded game is. Tap the play/pause icon in the top right of the page to see it.

### Over SSH

* Run `launchseq.sh` with the desired options
* Press `Ctrl+C` to stop
* Run `launchseq.sh -h` to see all options

```
Usage of launchseq.sh:
  -delay int
        number of seconds between loading each game (default 40)
  -offset int
        offset of games list to start at (not used for random)
  -path string
        custom additional path to scan for games
  -random
        randomize the order of games
  -system string
        system to load games from (default "n64")
```

## Custom launchers

LaunchSeq can be customized by creating your own shell scripts which call `launchseq.sh` with specific arguments.

To create, for example, a launcher which only loads N64 games and swaps games every 60s:

1. Create a new file in `Scripts` called `launchseq_N64.sh` (or anything with `.sh` on the end)
2. Set the contents of the file to:

```bash
#!/bin/bash
/media/fat/launchseq.sh -system n64 -delay 60 -offset 0
```

And that's it, you'll have a new entry in your `Scripts` menu to launch.