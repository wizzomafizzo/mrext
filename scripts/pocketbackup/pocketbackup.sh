#!/usr/bin/env python3
"""Backup Analogue Pocket saves to MiSTer."""

import configparser
import datetime
import os.path
import shutil
import zipfile
from enum import Enum
from typing import Dict, TypedDict

# TODO: arcade high scores? dunno if that's a thing on AP
# TODO: cleanup files deleted on AP
# TODO: check for max storage limit for snapshots
# TODO: restore backup to pocket
# TODO: optionally copy backups to other locations (cifs)
# TODO: check disk usage on backup locations
# TODO: check for AP firmware updates?
# TODO: pre-sync saves snapshot?

INI_FILENAME: str = "pocketbackup.ini"
# backup root location and working directory for operations
BACKUP_FOLDER: str = "/media/fat/pocket"
# storage for previous backups
SNAPSHOTS_FOLDER: str = os.path.join(BACKUP_FOLDER, "snapshots")
# total number of snapshots to keep
SNAPSHOTS_MAX: int = 50

# potential USB mount locations on MiSTer
USB_MOUNTS: list[str] = [
    "/media/usb0",
    "/media/usb1",
    "/media/usb2",
    "/media/usb3",
    "/media/usb4",
    "/media/usb5",
    "/media/usb6",
    "/media/usb7",
]

# special file telling us it's an AP storage device
POCKET_JSON: str = "Analogue_Pocket.json"
# which folders on the AP to backup
POCKET_BACKUP_FOLDERS: list[str] = [
    "Memories",
    "Saves",
    "Settings",
]

# mapping for AP platform IDs to MiSTer core folder names
# NOTE: many items on this list may not have save files to sync
POCKET_CORES_MAP: Dict[str, list[str]] = {
    "2600": ["Atari2600", "ATARI7800"],
    "7800": ["ATARI7800"],
    "amiga": ["Amiga"],
    "arcadia": ["Arcadia"],
    "arduboy": ["Arduboy"],
    "avision": ["AVision"],
    "channel_f": ["ChannelF"],
    "coleco": ["Coleco"],
    "creativision": ["CreatiVision"],
    "gamate": ["Gamate"],
    "gameandwatch": ["GameNWatch", "Game and Watch"],
    "gb": ["GAMEBOY"],
    "gba": ["GBA"],
    "gbc": ["GBC", "GAMEBOY"],
    "genesis": ["MegaDrive", "Genesis"],
    "gg": ["GameGear", "SMS"],
    "intv": ["Intellivision"],
    "mega_duck": ["MegaDuck"],
    "nes": ["NES"],
    "ng": ["NEOGEO"],
    "odyssey2": ["ODYSSEY2"],
    "pce": ["TGFX16"],
    "pcecd": ["TGFX16-CD"],
    "pdp1": ["PDP1"],
    "poke_mini": ["PokemonMini"],
    "sg1000": ["SG1000", "Coleco", "SMS"],
    "sgb": ["SGB"],
    "sms": ["SMS"],
    "snes": ["SNES"],
    "supervision": ["SuperVision"],
    "tamagotchi_p1": ["Tamagotchi"],
    "wonderswan": ["WonderSwan", "WonderSwanColor"],
}


def reverse_cores_map() -> Dict[str, str]:
    """Reverse the cores map to get a map of MiSTer core folder names to AP platform IDs."""
    reverse_map: Dict[str, str] = {}
    for platform_id, core_folders in POCKET_CORES_MAP.items():
        for core_folder in core_folders:
            reverse_map[core_folder] = platform_id
    return reverse_map


MISTER_CORES_MAP = reverse_cores_map()


class BackupStatus(Enum):
    """Final status of a file operation during backup job."""

    NEW = 1
    UPDATED = 2
    UNCHANGED = 3


def get_pocket_folder(mounts: list[str]) -> str or None:
    """Search for the Pocket folder in list of mounts and return the first match."""
    for mount in mounts:
        if os.path.exists(os.path.join(mount, POCKET_JSON)):
            return mount
    return None


def backup_folder(pocket_path: str, folder: str) -> dict:
    """Copy a folder from the Pocket to the backup location, skipping unchanged files.
    Returns a generator that yields a dict for each file copied, with the result.
    """
    backup_path = os.path.join(BACKUP_FOLDER, folder)
    if not os.path.exists(backup_path):
        os.mkdir(backup_path)

    from_path = os.path.join(pocket_path, folder)

    for root, dirs, files in os.walk(from_path):
        for dir_ in dirs:
            dst = os.path.join(backup_path, os.path.relpath(root, from_path), dir_)
            if not os.path.exists(dst):
                os.mkdir(dst)

        for file in files:
            src = os.path.join(root, file)
            dst = os.path.join(backup_path, os.path.relpath(src, from_path))

            if os.path.exists(dst):
                if os.path.getmtime(src) > os.path.getmtime(dst):
                    shutil.copy2(src, dst)
                    yield {
                        "file": file,
                        "status": BackupStatus.UPDATED,
                    }
                else:
                    yield {
                        "file": file,
                        "status": BackupStatus.UNCHANGED,
                    }
            else:
                shutil.copy2(src, dst)
                yield {
                    "file": file,
                    "status": BackupStatus.NEW,
                }


def zip_backup():
    """Create a zip file from all backed up folders and save it to the snapshots folder."""
    path = os.path.join(
        SNAPSHOTS_FOLDER, datetime.datetime.now().strftime("%Y-%m-%d_%H-%M-%S") + ".zip"
    )
    if os.path.exists(path):
        raise Exception("Snapshot already exists: {}".format(path))

    zipf = zipfile.ZipFile(path, "w", zipfile.ZIP_DEFLATED)

    for folder in POCKET_BACKUP_FOLDERS:
        for root, dirs, files in os.walk(os.path.join(BACKUP_FOLDER, folder)):
            for file in files:
                zipf.write(
                    os.path.join(root, file),
                    os.path.relpath(os.path.join(root, file), BACKUP_FOLDER),
                )

    zipf.close()


class Config(TypedDict):
    """User configuration options from .ini file."""

    mounts: list[str]


def get_config() -> Config:
    """Get user configuration options from .ini file."""
    config: Config = {
        "mounts": list(USB_MOUNTS),
    }

    ini_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), INI_FILENAME)
    if not os.path.exists(ini_path):
        return config

    parser = configparser.ConfigParser()
    parser.read(ini_path)

    if not parser.has_section("pocketbackup"):
        return config

    if parser.has_option("pocketbackup", "mounts"):
        mounts = config["mounts"]
        for mount in parser.get("pocketbackup", "mounts").split("|"):
            clean = mount.strip()
            if clean != "" and clean not in mounts:
                mounts.append(clean)
        config["mounts"] = mounts

    return config


def setup():
    """Set up environment for backup jobs."""
    if not os.path.exists(BACKUP_FOLDER):
        os.mkdir(BACKUP_FOLDER)

    if not os.path.exists(SNAPSHOTS_FOLDER):
        os.mkdir(SNAPSHOTS_FOLDER)


def main():
    """Main entry point for the script."""
    config = get_config()

    pocket_folder = get_pocket_folder(config["mounts"])
    if pocket_folder is None:
        print(
            "Pocket not found, check it's plugged in and USB SD Access is enabled in the Developer menu."
        )
        return
    else:
        print("Pocket found at: {}".format(pocket_folder))

    setup()

    print("Starting backup...")

    for folder in POCKET_BACKUP_FOLDERS:
        print("Backing up {}...".format(folder), end="", flush=True)

        for result in backup_folder(pocket_folder, folder):
            if result["status"] == BackupStatus.NEW:
                print("*", end="", flush=True)
            elif result["status"] == BackupStatus.UPDATED:
                print("^", end="", flush=True)
            else:
                print(".", end="", flush=True)

        print("...Done!", flush=True)

    print("Backup complete!", flush=True)

    print("Creating backup snapshot...", end="", flush=True)
    zip_backup()
    print("...Done!", flush=True)

    # count snapshots and delete oldest if we're over the limit
    snapshots = os.listdir(SNAPSHOTS_FOLDER)
    if len(snapshots) > SNAPSHOTS_MAX:
        snapshots.sort()
        os.remove(os.path.join(SNAPSHOTS_FOLDER, snapshots[0]))


if __name__ == "__main__":
    main()
