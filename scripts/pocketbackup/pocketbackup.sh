#!/usr/bin/env python3

import datetime
import os.path
import shutil
from enum import Enum
from typing import Dict
import zipfile

# TODO: arcade high scores? dunno if that's a thing on AP
# TODO: cleanup files deleted on AP
# TODO: check for max storage limit for snapshots
# TODO: restore backup to pocket
# TODO: optionally copy backups to other locations (cifs)
# TODO: check disk usage on backup locations

# backup root location and working directory for operations
BACKUP_FOLDER: str = "/media/fat/pocket"
# storage for previous backups
SNAPSHOTS_FOLDER: str = os.path.join(BACKUP_FOLDER, "snapshots")
# total number of snapshots to keep
SNAPSHOTS_MAX: int = 25

# potential USB mount locations on MiSTer
USB_MOUNTS: tuple[str] = (
    "/media/usb0",
    "/media/usb1",
    "/media/usb2",
    "/media/usb3",
    "/media/usb4",
    "/media/usb5",
    "/media/usb6",
    "/media/usb7",
    "/run/media/callan/AP",
)

# special file telling us it's an AP storage device
POCKET_JSON: str = "Analogue_Pocket.json"
# which folders on the AP to backup
POCKET_BACKUP_FOLDERS: tuple[str] = (
    "Memories",
    "Saves",
    "Settings",
)

# mapping for AP platform IDs to MiSTer core folder names
# NOTE: many items on this list may not have save files to sync
POCKET_CORES_MAP: Dict[str, tuple[str]] = {
    "2600": ("Atari2600", "ATARI7800"),
    "7800": ("ATARI7800",),
    "amiga": ("Amiga",),
    "arcadia": ("Arcadia",),
    "arduboy": ("Arduboy",),
    "avision": ("AVision",),
    "channel_f": ("ChannelF",),
    "coleco": ("Coleco",),
    "creativision": ("CreatiVision",),
    "gamate": ("Gamate",),
    "gameandwatch": ("GameNWatch", "Game and Watch"),
    "gb": ("GAMEBOY",),
    "gba": ("GBA",),
    "gbc": ("GBC", "GAMEBOY"),
    "genesis": ("MegaDrive", "Genesis"),
    "gg": ("GameGear", "SMS"),
    "intv": ("Intellivision",),
    "mega_duck": ("MegaDuck",),
    "nes": ("NES",),
    "ng": ("NEOGEO",),
    "odyssey2": ("ODYSSEY2",),
    "pce": ("TGFX16",),
    "pcecd": ("TGFX16-CD",),
    "pdp1": ("PDP1",),
    "poke_mini": ("PokemonMini",),
    "sg1000": ("SG1000", "Coleco", "SMS"),
    "sgb": ("SGB",),
    "sms": ("SMS",),
    "snes": ("SNES",),
    "supervision": ("SuperVision",),
    "tamagotchi_p1": ("Tamagotchi",),
    "wonderswan": ("WonderSwan", "WonderSwanColor"),
}

MISTER_CORES_MAP = {v: k for k, v in POCKET_CORES_MAP.items()}


class BackupStatus(Enum):
    NEW = 1
    UPDATED = 2
    UNCHANGED = 3


def get_pocket_folder() -> str or None:
    for mount in USB_MOUNTS:
        if os.path.exists(os.path.join(mount, POCKET_JSON)):
            return mount
    return None


# recursively copy a folder from the AP to the MiSTer, skipping files that are
# already up to date based on modification time. returns a generator that yields
# a dict for each file copied, with the result
def backup_folder(pocket_path: str, folder: str) -> dict:
    backup_path = os.path.join(BACKUP_FOLDER, folder)
    if not os.path.exists(backup_path):
        os.mkdir(backup_path)

    from_path = os.path.join(pocket_path, folder)

    for root, dirs, files in os.walk(from_path):
        for dir in dirs:
            dst = os.path.join(backup_path, os.path.relpath(root, from_path), dir)
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


# create a zip file from all backed up folders and save it to the snapshots
# folder with a timestamp
def zip_backup():
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


def setup():
    if not os.path.exists(BACKUP_FOLDER):
        os.mkdir(BACKUP_FOLDER)

    if not os.path.exists(SNAPSHOTS_FOLDER):
        os.mkdir(SNAPSHOTS_FOLDER)


def main():
    pocket_folder = get_pocket_folder()
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
