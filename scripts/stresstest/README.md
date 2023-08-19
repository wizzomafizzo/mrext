# HPS-Core Stress Testing

## CD Images

- Custom PSX core build which tallies and reports disk read errors
- `Spyro the Dragon (USA).chd` used for all tests (used because it has XA audio tracks)
  - MD5: 916479df840e2a922818f18001fbc2c4
  - Image stored on:
    - SD card
    - External USB mechanical HDD (USB powered)
    - CIFS over ethernet
    - CIFS over WiFi
- Script daemons running:
  - `remote.sh` (nice 1)
  - `lastplayed.sh` (nice 1)
  - `playlog.sh` (nice 1)
  - `nfc.sh` (nice 1)
  - `bgm.sh`
  - TODO: SAM w/ controller polling
- Active SSH sessions:
  - `htop`
- Before each test, `hpscore.py` is run to restart the PSX core and load Spyro save state to World 1

### Notes

- Opening OSD can cause (safe) read errors
- Mounting CIFS causes read errors (1e072)
- Browsing mounted CIFS does not cause read errors
- Browsing SD card does not cause read errors
- Wallpapers and screenshots load without error consistently (over 500mb of SD data)
- Connecting a controller causes read errors (fe98c)

### From SD Card

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
