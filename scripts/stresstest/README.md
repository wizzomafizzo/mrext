# HPS-Core Stress Testing

## Todo
- [ ] Testing with HDD images on a computer core
- [ ] DDR3 testing
- [ ] Any other potential risks?

## CD Images

- Custom PSX core build which tallies and reports disk read errors
- `Spyro the Dragon (USA).chd` used for all tests (used because it has XA audio tracks)
  - MD5: 916479df840e2a922818f18001fbc2c4
  - Image stored on:
    - SD card
    - External USB mechanical HDD (USB powered)
    - CIFS over ethernet
    - CIFS over Wi-Fi
- Script daemons running:
  - `remote.sh` (nice 1, many inotify watchers, runs regular mdns scan and service register)
  - `lastplayed.sh` (nice 1, some inotify watchers, otherwise does nothing until game or core changes)
  - `playlog.sh` (nice 1, some inotify watchers, makes small write to sqlite db on SD every 5 minutes)
  - `nfc.sh` (nice 1, one inotify watcher, no disk activity but has a constant ~3-4% CPU usage)
  - `bgm.sh` (some inotify watchers, otherwise does nothing unless explicitly made to play music)
  - TODO: SAM w/ controller polling
- Before each test, `hpscore.py` is run to restart the PSX core and load Spyro save state to World 1

### Notes

- Opening OSD can cause (safe) read errors
- Mounting CIFS causes read errors (1e072)
- Browsing mounted CIFS does not cause read errors
- Browsing SD card does not cause read errors
- Wallpapers and screenshots load without error consistently (over 500mb of SD data)
- Connecting a controller causes read errors (fe98c)

### From SD card

| Test                                         | Errors? | Error count | Notes                                                  |
|----------------------------------------------|---------|-------------|--------------------------------------------------------|
| No daemons                                   | No      | 0           |                                                        |
| Idle 6hrs w/ daemons                         | No      | 0           |                                                        |
| scp Spyro CHD to PC (ethernet)               | Yes     | c7ff6       | Some minor audio clipping and slowdown                 |
| Opening remote search page                   | Yes     | d2b30       | Likely due to system list scan, obvious audio clipping |
| Opening cached remote search page            | Yes     | 24239       | No audio clipping                                      |
| Opening all other remote pages               | No      | 0           | Including full load of screenshots & wallpapers        |
| Sending control key presses                  | No      | 0           |                                                        |
| Browsing menu Arcade folder                  | Yes     | 0a662       | Obvious clipping                                       |
| Browsing cached Arcade folder                | No      | 0           |                                                        |
| Browsing other menu folders                  | No      | 0           |                                                        |
| Polling bgm service                          | No      | 0           |                                                        |
| Playing music file                           | Yes     | 1de64       | Intermittently, during some initial track load         |
| Setting active wallpaper                     | No      | 0           |                                                        |
| Changing .ini settings                       | No      | 0           |                                                        |
| Generating search index (SD & CIFS ethernet) | Yes     | fcbff       | Long process, many obvious audio issues                |

### From wifi
- Loading game and save state over wifi causes immediate read errors and these tick up intermittently at idle
- Mounting images over wifi causes so many read errors that the count often wraps

| Test                                     | Errors? | Error count | Notes |
|------------------------------------------|---------|-------------|-------|
| No daemons                               |         |             |       |
| Idle 6hrs w/ daemons                     | Yes     | -           |       |
| scp Spyro CHD to PC (wifi)               | Yes     |             |       |
| Opening remote search page               | Yes     |             |       |
| Opening cached remote search page        | Yes     |             |       |
| Opening all other remote pages           | No      |             |       |
| Sending control key presses              | No      |             |       |
| Browsing menu Arcade folder              | No      |             |       |
| Browsing cached Arcade folder            | No      |             |       |
| Browsing other menu folders              | No      |             |       |
| Polling bgm service                      | No      |             |       |
| Playing music file                       | Yes     |             |       |
| Setting active wallpaper                 |         |             |       |
| Changing .ini settings                   |         |             |       |
| Generating search index (SD & CIFS wifi) | Yes     |             |       |
