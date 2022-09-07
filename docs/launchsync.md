# Launch Sync

Launch Sync allows people to create, share and maintain live-updating game playlists for the MiSTer.

You create a sync file with a list of games, someone copies the file to their MiSTer, and Launch Sync will use it to generate a working list of game shortcuts in the MiSTer main menu. Sync files are subscriptable, so you can publish changes to your playlist and people will see your updates on their own system.

[![Download Launch Sync](download.png "Download Launch Sync")](https://github.com/wizzomafizzo/mrext/raw/main/releases/launchsync/launchsync.sh)

## Install

Download Launch Sync and copy it to the `Scripts` folder on your MiSTer.

Add the following text to the `downloader.ini` file on your MiSTer to receive updates with the `update` script:
```
[mrext/launchsync]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/launchsync/launchsync.json
```