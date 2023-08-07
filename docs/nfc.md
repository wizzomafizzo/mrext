# NFC

NFC is a service for launching games, cores and custom dynamic commands using a USB NFC card reader. All hardware required is inexpensive, easily available and quick to set up.

<a href="https://github.com/wizzomafizzo/mrext/releases/latest/download/nfc.sh"><img src="images/download.svg" alt="Download Remote" title="Download Remote" width="140"></a>

## Hardware

The following hardware is currently known to work. Many other devices may work, but might also require a project update for proper support. Please [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) if you'd like to add a working device to this list, or troubleshoot a device that isn't working.

This project uses the [libnfc](https://nfc-tools.github.io/projects/libnfc/) library, so any device supported by it should work.

### Readers

- **ACR122U USB NFC reader**: plug and play, cheap, littered on Amazon, eBay and AliExpress
- **PN532 NFC module**: really cheap, small, also requires a USB to TTL cable, some (small) manual configuration and possibly some soldering

### Tags

The form factor of the tag is up to you. Can be a card, sticker, keychain, etc.

- **NTAG213**: 144 bytes of storage
- **NTAG215**: 504 bytes of storage
- **NTAG216**: 888 bytes of storage
- **Amiibo**: supported using the `nfc.csv` file describe below

Custom NFC commands can be written to NTAG213 without issue, but keep storage size in mind if you have a large collection of games with deep folders.

## Install

Download [Remote](https://github.com/wizzomafizzo/mrext/releases/latest/download/nfc.sh) and copy it to the `Scripts` folder on your MiSTer's SD card.

Optionally, add the following to the `downloader.ini` file on your MiSTer, to receive updates with the `update` or `downloader` script:
```
[mrext/nfc]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/nfc/nfc.json
```

Once installed, run `nfc` from the MiSTer `Scripts` menu, a prompt will offer to enable NFC as a startup service, then the service will be started in the background.

This service must be running for NFC to work, but it has no impact on your MiSTer's performance.

### Hardware configuration

Your reader may work out of the box with no extra configuration. 

It should work out of the box with a `ACR122U USB NFC reader`

If you are using a PN532 connected to a USB -> TTL cable then the following config may be needed in `/media/fat/Scripts/nfc.ini`:

```
[nfc]
connection_string="pn532_uart:/dev/ttyUSB0"
```

## Method 1: Writing games to a card (Recommended)
Write a single text record to the NFC tag using your favourite writing software e.g. [NFC Tools](https://play.google.com/store/apps/details?id=com.wakdev.wdnfc) on Android. The content should be the name of a mgl or mra relative to `/media/fat/` e.g.

```
_Arcade/1942 (Revision B).mra
```

or

```
_Favourites/Castlevania - Aria of Sorrow.mgl
```

## Method 2: Mapping Card UID to a game (Fallback)
Create a csv file: `/media/fat/nfc-mapping.csv` in the format:

```csv
040fa2e2356281,_Arcade/1942 (Revision B).mra
0427d3e2356280,_Arcade/Arkanoid (Japan).mra
```

[Download](https://github.com/wizzomafizzo/mrext/releases/latest/download/nfc.sh) and copy

`./nfc.sh` to your MiSTer and run it. When scanning a card you should see something like:

```
root@MiSTer:/tmp>./nfc.sh
2023/08/02 20:29:34 MiSTer NFC Reader (libnfc version1.8.0)
2023/08/02 20:29:34 Loaded 5 NFC mappings from the CSV
2023/08/02 20:29:34 Opened:  microBuilder.eu pn532_uart:/dev/ttyUSB0
2023/08/02 20:29:40 New card UID: 04fa7cfa904980
2023/08/02 20:29:40 Card hex: 0103a00c340325d101215402656e4265207375726520746f206472696e6b20796f7572204f76616c74696e65fe00000002656e6d6f72652074657874fe0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
2023/08/02 20:29:40 Decoded text NDEF is: Be sure to drink your Ovaltine
2023/08/02 20:29:40 Core does not exist: /media/fat/Be sure to drink your Ovaltine
2023/08/02 20:29:48 New card UID: 046100e2356284
2023/08/02 20:29:48 Card hex: a500030020e840fc23fb7bbced2e56e876e43495de17f29f43da15d21fc5e5d3f6d047fa92c66bcf04a49a1e21136434f7ab4f840e139e519c1dd79d989e53b411cb6feb01030000024f09020d12c50b9c4b2e6483593d141bb38afdee5f635f4a2b70bdd918404cfda928cd34a9a371ea974fa1579d38b11f5348708df9f96cfc393bde90db8d672c92153224d51e2e
2023/08/02 20:29:48 No text NDEF found, falling back to UID mapping in CSV file
2023/08/02 20:29:48 Loading core: /media/fat/_Arcade/Arkanoid (Japan).mra
2023/08/02 20:29:54 New card UID: 04087cfa904981
2023/08/02 20:29:54 Card hex: 0103a00c340323d1011f5402656e5f4172636164652f41726b616e6f696420284a6170616e292e6d7261fe0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
2023/08/02 20:29:54 Decoded text NDEF is: _Arcade/Arkanoid (Japan).mra
2023/08/02 20:29:54 Loading core: /media/fat/_Arcade/Arkanoid (Japan).mra
```
