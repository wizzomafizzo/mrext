# MiSTer Extensions

Extensions and utilities to make your [MiSTer](https://github.com/MiSTer-devel/Main_MiSTer/wiki) even better.

Make sure to check the linked documentation for each script you use. Most are simple and work out-of-the-box, but some require manual setup before they do anything useful.

[Remote](#remote) • [BGM](#bgm) • [Favorites](#favorites) • [GamesMenu](#gamesmenu) • [LastPlayed](#lastplayed) • [LaunchSync](#launchsync) • [PlayLog](#playlog) • [Random](#random) • [Search](#search)

[Supported Systems](docs/systems.md) • [Developer Guide](docs/dev.md)

## Install

### Update All

Open the [Update All](https://github.com/theypsilon/Update_All_MiSTer) settings menu, the `Unofficial Scripts` submenu, and enable the MiSTer Extensions repository from there.

### Update/Downloader

Add the following to your `downloader.ini` file to install everything at once through the `update` script:

```
[mrext/all]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/all.json
```

Each script also provides its own individual update file if you only want certain ones. Check the script's README.

### Manual

All scripts listed can be installed by downloading the linked file below, placing it in the `Scripts` folder on your SD card, and running it from the `Scripts` menu on your MiSTer.

## Remote

Control the MiSTer from any device on your network. Remote is a web-based interface with a stack of modern features to manage all aspects of your MiSTer.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/remote.sh"><img src="docs/images/download.svg" alt="Download Remote" title="Download Remote" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/blob/main/docs/remote.md"><img src="docs/images/readme.svg" alt="Readme Remote" title="Readme Remote" width="140"></a>

## BGM
Play your own music in the MiSTer menu. BGM is a highly configurable background music player that automatically pauses when you're playing games. Supports many common audio formats including internet radio streams.

<a href="https://github.com/wizzomafizzo/MiSTer_BGM/raw/main/bgm.sh"><img src="docs/images/download.svg" alt="Download BGM" title="Download BGM" width="140"></a>
<a href="https://github.com/wizzomafizzo/MiSTer_BGM"><img src="docs/images/readme.svg" alt="Readme BGM" title="Readme BGM" width="140"></a>

## Favorites
Create and manage shortcuts for your favorite games. Favorites allows you to pick any game or core from your system and automatically generate a shortcut to it in the MiSTer menu.

<a href="https://github.com/wizzomafizzo/MiSTer_Favorites/raw/main/favorites.sh"><img src="docs/images/download.svg" alt="Download Favorites" title="Download Favorites" width="140"></a>
<a href="https://github.com/wizzomafizzo/MiSTer_Favorites"><img src="docs/images/readme.svg" alt="Readme Favorites" title="Readme Favorites" width="140"></a>

## GamesMenu
Browse your entire collection from the main MiSTer menu. GamesMenu indexes all your games and generates a set of shortcuts in the menu mirroring your folder layout.

<a href="https://github.com/wizzomafizzo/MiSTer_GamesMenu/raw/main/games_menu.sh"><img src="docs/images/download.svg" alt="Download GamesMenu" title="Download GamesMenu" width="140"></a>
<a href="https://github.com/wizzomafizzo/MiSTer_GamesMenu"><img src="docs/images/readme.svg" alt="Readme GamesMenu" title="Readme GamesMenu" width="140"></a>

## LaunchSync
Create shareable and subscribable game playlists. LaunchSync automatically generates working menu shortcuts from custom playlist files, with the ability to keep them up-to-date with the author's live version.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/launchsync.sh"><img src="docs/images/download.svg" alt="Download LaunchSync" title="Download LaunchSync" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/tree/main/docs/launchsync.md"><img src="docs/images/readme.svg" alt="Readme LaunchSync" title="Readme LaunchSync" width="140"></a>

## LastPlayed
Automatically generate dynamic shortcuts in the MiSTer menu.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/lastplayed.sh"><img src="docs/images/download.svg" alt="Download LastPlayed" title="Download LastPlayed" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/tree/main/docs/lastplayed.md"><img src="docs/images/readme.svg" alt="Readme LastPlayed" title="Readme LastPlayed" width="140"></a>

## PlayLog
Track and report on what games you've been playing on your MiSTer.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/playlog.sh"><img src="docs/images/download.svg" alt="Download PlayLog" title="Download PlayLog" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/tree/main/docs/playlog.md"><img src="docs/images/readme.svg" alt="Readme PlayLog" title="Readme PlayLog" width="140"></a>

## Random
Instantly launch a random game in your collection from the Scripts menu.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/random.sh"><img src="docs/images/download.svg" alt="Download Random" title="Download Random" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/tree/main/docs/random.md"><img src="docs/images/readme.svg" alt="Readme Random" title="Readme Random" width="140"></a>

## Search
Search for and launch games from your collection. Searching is *fast* and great for discovering games.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/search.sh"><img src="docs/images/download.svg" alt="Download Search" title="Download Search" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/tree/main/docs/search.md"><img src="docs/images/readme.svg" alt="Readme Search" title="Readme Search" width="140"></a>

## NFC
Launch games, cores and dynamic commands using NFC tags or cards. Uses easily available and cheap USB NFC card readers.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/nfc.sh"><img src="docs/images/download.svg" alt="Download NFC" title="Download NFC" width="140"></a>
<a href="https://github.com/wizzomafizzo/mrext/tree/main/docs/nfc.md"><img src="docs/images/readme.svg" alt="Readme NFC" title="Readme NFC" width="140"></a>

## Other Projects

Great projects by other people that add heaps of functionality to your MiSTer.

Please [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) if you'd like to suggest something for this list. Anything is welcome, though the focus is on software projects that work without custom hardware.

### Cores & Games

- [AMMiSTer](https://github.com/city41/AMMiSTer)

  A slick PC application for managing your arcade game collection. Includes updates, bulk management, favorites and game metadata.

- [MGL Core Setnames](https://github.com/RGarciaLago/MGL_Core_Setnames)

  A preset pack of modified core shortcuts which let you have automatic alternate core configs. Useful for cores which support multiple systems or input devices.

- [mister-boot-roms](https://github.com/uberyoji/mister-boot-roms)

  Adds high quality MiSTer-themed boot screens to cores which support loadable boot roms.

- [mistercon](https://github.com/tatsutron/mistercon)
    
  A MiSTer frontend for Android. Browse your collection and launch games from your phone.

- [VIDEO PRESETS by Robby](https://github.com/RGarciaLago/VIDEO_PRESETS_by_Robby)

  A curated and extensive set of video presets for the MiSTer cores.

### Frontend

- [Insert Coin](https://github.com/funkycochise/Insert-Coin)

  An alternative layout for browsing the Arcade folder.

- [MiSTer Super Attract Mode (SAM)](https://github.com/mrchrisster/MiSTer_SAM)

  Add an attract screen to your MiSTer. When idle, games will start to play at random and rotate after a set period. You can even jump in and start playing if a game looks fun! A mature project and highly configurable.

- [MiSTerWallpapers](https://github.com/RetroDriven/MiSTerWallpapers)

  Automatically download a large collection of high quality wallpapers.

- [MiSTer-CRT-Wallpapers](https://github.com/RetroDriven/MiSTer-CRT-Wallpapers)

  The same, but specifically for 4:3 CRT screens.

- [MiSTress](https://github.com/sigboe/MiSTress)

  An RSS reader for MiSTer. Display the latest core updates right on your wallpaper.

### Ports

- [MiSTer Basilisk II](https://github.com/bbond007/MiSTer_BasiliskII)

  A build of the [Basilisk II](https://basilisk.cebix.net/) project for MiSTer. A 68k Macintosh emulator.

- [MiSTer DOSBox](https://github.com/bbond007/MiSTer_DOSBox)

  A build of the [DOSBox](https://www.dosbox.com/) project for MiSTer. Play a huge range of DOS games.

- [MiSTer PrBoom-Plus](https://github.com/bbond007/MiSTer_PrBoom-Plus)

  A build of the [PrBoom](http://prboom.sourceforge.net/) project for MiSTer. An enhanced Doom engine with a massive number of expansions.

- [MiSTer ScummVM](https://github.com/bbond007/MiSTer_ScummVM)

  A build of the [ScummVM](https://www.scummvm.org/) project for MiSTer. Runs well and even works for games out of reach of the AO486 core.

### System

- [Migrate SD](https://github.com/Natrox/MiSTer_Utils_Natrox)

  A utility to migrate your entire MiSTer SD card to a new one, straight from the MiSTer itself.

- [MiSTer Batch Control](https://github.com/pocomane/MiSTer_Batch_Control)

  A command line utility to perform low-level functions that may not be possible via scripting languages.

- [MiSTer FPGA Overclock Scripts](https://github.com/coolbho3k/MiSTer-Overclock-Scripts)

  An overclocked system can run Munt (MT32 emulator) at full speed and get extra performance out of software like ScummVM.

- [MiSTerArch](https://github.com/MiSTerArch/PKGBUILDs)

  A replacement image for the MiSTer with a full [Arch Linux](https://archlinux.org/) system.

- [MiSTerTools](https://github.com/morfeus77/MiSTerTools/)

  Scripts for custom aspect ratio calculation, modeline to video_mode conversion, video_mode to modeline conversion, ini profile switcher and to parse MRA files.

- [MOnSieurFPGA](https://github.com/MOnSieurFPGA/MOnSieurFPGA-SD_Image_Builds)

  Another replacement image for the MiSTer with a full [Arch Linux](https://archlinux.org/) system.

- [Official Scripts](https://github.com/MiSTer-devel/Scripts_MiSTer)

  The official MiSTer scripts repository. A miscellaneous collection of small scripts for various system tasks and configuration.

- [reMiSTer](https://github.com/sigboe/reMiSTer)

  A tool for using your keyboard on MiSTer over the network.

- [Remote Input Server Daemon](https://github.com/sofakng/risd)

  Server daemon that monitors commands over TCP and emulates keystrokes using TCP.

### Updaters

- [DOS Shareware Updater](https://github.com/flynnsbit/DOS_Shareware_MyMenu)

  Update script for the Flynn's DOS Shareware pack on the AO486 core.

- [Top 300 Updater](https://github.com/flynnsbit/Top300_updates)

  Update script for Flynn's Top 300 pack on the AO486 core.

- [Update All](https://github.com/theypsilon/Update_All_MiSTer)
  
  If you're reading this, you already use it. Don't forget to check the advanced options.

- [Update tty2xxx](https://github.com/ojaksch/MiSTer_update_tty2xxx)

  A unified updater script for the various MiSTer display projects.

- [Yet another random MiSTer utility script (YARMUS)](https://github.com/jayp76/MiSTer_get_optional_installers)

  A script to install a lot of MiSTer scripts at once. It includes many things on this list, plus extra custom installers for software like [DevilutionX](https://github.com/diasurgical/devilutionX), [Cave Story](https://nxengine.sourceforge.io/) and homebrew packs.
