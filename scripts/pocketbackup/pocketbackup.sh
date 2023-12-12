#!/usr/bin/env python3

import os.path
import shutil

# TODO: arcade high scores? dunno if that's a thing on AP

# backup root location and working directory for operations
BACKUP_FOLDER = "/media/fat/pocket"
# potential USB mount locations on MiSTer
USB_MOUNTS = (
    "/media/usb0",
    "/media/usb1",
    "/media/usb2",
    "/media/usb3",
    "/media/usb4",
    "/media/usb5",
    "/media/usb6",
    "/media/usb7",
)

# special file telling us it's an AP storage device
POCKET_JSON = "Analogue_Pocket.json"
# which folders on the AP to backup
POCKET_BACKUP_FOLDERS = (
    "Memories",
    "Saves",
    "Settings",
)

# mapping for AP platform IDs to MiSTer core folder names
# NOTE: many items on this list may not have save files to sync
POCKET_CORES_MAP = {
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

MISTER_CORES_MAP = reverse_dict(CORES_MAP)


def reverse_dict(d):
    return {v: k for k, v in d.items()}


def get_pocket_folder():
    for mount in USB_MOUNTS:
        if os.path.exists(os.path.join(mount, POCKET_JSON)):
            return mount
    return None


# recursively copy a folder from the AP to the MiSTer, skipping files that are
# already up to date based on modification time. returns a generator that yields
# a dict for each file copied, with the result
def backup_folder(folder):
    if not os.path.exists(os.path.join(BACKUP_FOLDER, folder)):
        os.mkdir(os.path.join(BACKUP_FOLDER, folder))

    for root, dirs, files in os.walk(os.path.join(pocket_folder, folder)):
        for file in files:
            src = os.path.join(root, file)
            dst = os.path.join(BACKUP_FOLDER, folder, file)
            if os.path.exists(dst):
                if os.path.getmtime(src) > os.path.getmtime(dst):
                    shutil.copy2(src, dst)
                    yield {
                        "file": file,
                        "updated": True,
                        "new": False,
                    }
                else:
                    yield {
                        "file": file,
                        "updated": False,
                        "new": False,
                    }
            else:
                shutil.copy2(src, dst)
                yield {
                    "file": file,
                    "updated": False,
                    "new": True,
                }


def setup():
    if not os.path.exists(BACKUP_FOLDER):
        os.mkdir(BACKUP_FOLDER)


def main():
    pocket_folder = get_pocket_folder()
    if pocket_folder is None:
        print("Pocket not found, check it's plugged in and USB SD Access is enabled in the Developer menu.")
        return
    else:
        print("Pocket found at: {}".format(pocket_folder))

    print("Starting backup...")
    setup()

    for folder in POCKET_BACKUP_FOLDERS:
        print("Backing up {}...".format(folder), end="", flush=True)

        for result in backup_folder(folder):
            if result["new"]:
                print("N", end="", flush=True)
            elif result["updated"]:
                print("U", end="", flush=True)
            else:
                print(".", end="", flush=True)

        print("...Done!", flush=True)

    print("Backup complete!", flush=True)


if __name__ == "__main__":
    main()
