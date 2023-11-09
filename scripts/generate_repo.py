#!/usr/bin/env python3

import os
import json
import hashlib
import sys
import time
from zipfile import ZipFile
from typing import TypedDict, Union, Optional

APPS = ["lastplayed", "launchseq", "launchsync", "nfc", "playlog", "random", "remote", "search"]
FILES = {
    "lastplayed": ["lastplayed.sh"],
    "launchseq": ["launchseq.sh"],
    "launchsync": ["launchsync.sh"],
    "nfc": ["nfc.sh", "scripts/nfcui/nfcui.sh"],
    "playlog": ["playlog.sh"],
    "random": ["random.sh"],
    "remote": ["remote.sh"],
    "search": ["search.sh"],
}
REBOOT = ["remote"]
EXTERNAL_FILES = [
    "releases/external/bgm.sh",
    "releases/external/favorites.sh",
    "releases/external/gamesmenu.sh",
]

DB_ID = "mrext/{}"
RELEASES_FOLDER = "releases"
DL_FOLDER = "_bin/releases"
DL_URL = "https://github.com/wizzomafizzo/mrext/releases/download/{}"
EXTERNAL_URL = "https://github.com/wizzomafizzo/mrext/raw/main/releases/external/{}"


class RepoDbFilesItem(TypedDict):
    hash: str
    size: int
    url: Optional[str]
    overwrite: Optional[bool]
    reboot: Optional[bool]


RepoDbFiles = dict[str, RepoDbFilesItem]


class RepoDbFoldersItem(TypedDict):
    tags: Optional[list[Union[str, int]]]


RepoDbFolders = dict[str, RepoDbFoldersItem]


class RepoDb(TypedDict):
    db_id: str
    timestamp: int
    files: RepoDbFiles
    folders: RepoDbFolders
    base_files_url: Optional[str]


def create_app_db(app: str, tag: str) -> RepoDb:
    if app not in APPS:
        raise ValueError("Invalid app name")

    folders: RepoDbFolders = {
        "Scripts/": RepoDbFoldersItem(tags=None),
    }

    dl_files = FILES[app]
    reboot = app in REBOOT

    files: RepoDbFiles = {}
    for file in dl_files:
        local_path = file
        if "/" not in file:
            local_path = os.path.join(DL_FOLDER, file)

        key = "Scripts/{}".format(os.path.basename(local_path))
        size = os.stat(local_path).st_size
        md5 = hashlib.md5(open(local_path, "rb").read()).hexdigest()
        url = "{}/{}".format(DL_URL.format(tag), os.path.basename(local_path))

        file_entry = RepoDbFilesItem(
            hash=md5, size=size, url=url, overwrite=None, reboot=reboot
        )

        files[key] = file_entry

    return RepoDb(
        db_id=DB_ID.format(app),
        timestamp=int(time.time()),
        files=files,
        folders=folders,
        base_files_url=None,
    )


def create_all_db(tag: str) -> RepoDb:
    folders: RepoDbFolders = {
        "Scripts/": RepoDbFoldersItem(tags=None),
    }
    files: RepoDbFiles = {}

    for app in APPS:
        dl_files = FILES[app]
        reboot = app in REBOOT

        for file in dl_files:
            local_path = file
            if "/" not in file:
                local_path = os.path.join(DL_FOLDER, file)

            key = "Scripts/{}".format(os.path.basename(local_path))
            size = os.stat(local_path).st_size
            md5 = hashlib.md5(open(local_path, "rb").read()).hexdigest()
            url = "{}/{}".format(DL_URL.format(tag), os.path.basename(local_path))

            file_entry = RepoDbFilesItem(
                hash=md5, size=size, url=url, overwrite=None, reboot=reboot
            )

            files[key] = file_entry

    for file in EXTERNAL_FILES:
        local_path = file
        key = "Scripts/{}".format(os.path.basename(local_path))
        size = os.stat(local_path).st_size
        md5 = hashlib.md5(open(local_path, "rb").read()).hexdigest()
        url = EXTERNAL_URL.format(os.path.basename(local_path))

        file_entry = RepoDbFilesItem(
            hash=md5, size=size, url=url, overwrite=None, reboot=False
        )

        files[key] = file_entry

    return RepoDb(
        db_id=DB_ID.format("all"),
        timestamp=int(time.time()),
        files=files,
        folders=folders,
        base_files_url=None,
    )


def remove_nulls(v: any) -> any:
    if isinstance(v, dict):
        return {key: remove_nulls(val) for key, val in v.items() if val is not None}
    else:
        return v


def generate_json(repo_db: RepoDb) -> str:
    return json.dumps(remove_nulls(repo_db), indent=4)


def main():
    tag = sys.argv[1]

    for app in APPS:
        repo_db = create_app_db(app, tag)
        with open("{}/{}/{}.json".format(RELEASES_FOLDER, app, app), "w") as f:
            f.write(generate_json(repo_db))

    repo_db = create_all_db(tag)
    with open("{}/all.json".format(RELEASES_FOLDER), "w") as f:
        f.write(generate_json(repo_db))


if __name__ == "__main__":
    main()
