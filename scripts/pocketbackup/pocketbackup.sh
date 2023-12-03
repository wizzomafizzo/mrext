#!/usr/bin/env python3

import os.path
import shutil

BACKUP_FOLDER = "/media/fat/pocket"
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

POCKET_JSON = "Analogue_Pocket.json"
POCKET_FOLDERS = (
    "Memories",
    "Saves",
    "Settings",
)


def get_pocket_folder():
    for mount in USB_MOUNTS:
        if os.path.exists(os.path.join(mount, POCKET_JSON)):
            return mount
    return None


def main():
    pocket_folder = get_pocket_folder()
    if pocket_folder is None:
        print("Pocket not found, check it's plugged in and USB SD Access is enabled in the Developer menu.")
        return
    else:
        print("Pocket found at: {}".format(pocket_folder))

    print("Starting backup...")
    if not os.path.exists(BACKUP_FOLDER):
        os.mkdir(BACKUP_FOLDER)

    for folder in POCKET_FOLDERS:
        print("Backing up {}...".format(folder), end="", flush=True)

        if not os.path.exists(os.path.join(BACKUP_FOLDER, folder)):
            os.mkdir(os.path.join(BACKUP_FOLDER, folder))

        for root, dirs, files in os.walk(os.path.join(pocket_folder, folder)):
            for file in files:
                src = os.path.join(root, file)
                dst = os.path.join(BACKUP_FOLDER, folder, file)
                if os.path.exists(dst):
                    if os.path.getmtime(src) > os.path.getmtime(dst):
                        print("U", end="", flush=True)
                        shutil.copy2(src, dst)
                    else:
                        print(".", end="", flush=True)
                else:
                    print("N", end="", flush=True)
                    shutil.copy2(src, dst)

        print("...Done!", flush=True)

    print("Backup complete!", flush=True)


if __name__ == "__main__":
    main()
