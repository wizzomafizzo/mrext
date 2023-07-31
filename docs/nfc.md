# NFC

⚠️ ***Work in progress***

NFC is a service for loading your favourite cores using NFC tags or cards.

## Hardware required

Any reader compatible with [libnfc](https://nfc-tools.github.io/projects/libnfc/) can be plugged straight into the MiSTer. 

This was developed and tested with a PN532 NFC module v3 connected to a FTDI USB to TTL cable.

## Setup

On your MiSTer, create a `/etc/nfc/libnfc.conf` with the following content matching your hardware:

```
device.name = "microBuilder.eu"
device.connstring = "pn532_uart:/dev/ttyUSB0"
```

Now tell MiSTer which tags belong to which games by
creating a csv file: `/media/fat/nfc-mapping.csv` in the format:

```csv
040fa2e2356281,/media/fat/_Arcade/1942 (Revision B).mra
0427d3e2356280,/media/fat/_Arcade/Arkanoid (Japan).mra
```

[Download](https://github.com/symm/mrext/releases/download/v0.1/nfc.sh) and copy

`./nfc.sh` to your MiSTer and run it. When scanning a card you should see something like:

```
root@MiSTer:/tmp>./nfc.sh
2023/07/31 22:07:53 MiSTer NFC Reader (libnfc version1.8.0)
2023/07/31 22:07:53 Loaded 2 NFC mappings from the CSV
2023/07/31 22:07:53 Opened:  microBuilder.eu pn532_uart:/dev/ttyUSB0
2023/07/31 22:07:57 New card UID: 040fa2e2356281
2023/07/31 22:07:57 Loading core: /media/fat/_Arcade/1942 (Revision B).mra
```
