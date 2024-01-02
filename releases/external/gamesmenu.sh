#!/usr/bin/env python

import os
import zipfile
import subprocess
import sys
import shutil
import re
import math

GAMES_MENU_PATH = "/media/fat/_Games"
NAMES_FILE = "/media/fat/names.txt"
NAMES_CACHE = {}
MENU_CACHE = {}

# TODO: combined meta folders for HTGDB packs
# TODO: cleanup mgl files with broken links
# TODO: link combined systems to the top level (game gear, mega duck etc.)
# TODO: update screenshots, readme

# (<games folder name>, <rbf>, (<file extensions>[], <delay>, <type>, <index>)[])[]
MGL_MAP = (
    ("Amiga", "_Computer/Minimig", (({".adf"}, 1, "f", 0),)),
    ("Arcadia", "_Console/Arcadia", (({".bin"}, 1, "f", 1),)),
    ("AVision", "_Console/AdventureVision", (({".bin"}, 1, "f", 1),)),
    ("Astrocade", "_Console/Astrocade", (({".bin"}, 1, "f", 1),)),
    ("ATARI2600", "_Console/Atari7800", (({".a78", ".a26", ".bin"}, 1, "f", 1),)),
    (
        "ATARI5200",
        "_Console/Atari5200",
        (({".car", ".a52", ".bin", ".rom"}, 1, "s", 1),),
    ),
    ("AtariLynx", "_Console/AtariLynx", (({".lnx"}, 1, "f", 0),)),
    ("C64", "_Computer/C64", (({".prg", ".crt", ".reu", ".tap"}, 1, "f", 1),)),
    ("ChannelF", "_Console/ChannelF", (({".rom", ".bin"}, 1, "f", 1),)),
    (
        "Coleco",
        "_Console/ColecoVision",
        (({".col", ".bin", ".rom", ".sg"}, 1, "f", 0),),
    ),
    ("CreatiVision", "_Console/CreatiVision", (({".rom", ".bin"}, 1, "f", 1),)),
    ("GAMEBOY2P", "_Console/Gameboy2P", (({".gb", ".gbc"}, 1, "f", 1),)),
    ("GAMEBOY", "_Console/Gameboy", (({".gb"}, 1, "f", 1),)),
    ("GBC", "_Console/Gameboy", (({".gbc"}, 1, "f", 1),)),
    ("Gamate", "_Console/Gamate", (({".bin"}, 1, "f", 1),)),
    ("GameNWatch", "_Console/GnW", (({".bin"}, 1, "f", 1),)),
    ("GBA2P", "_Console/GBA2P", (({".gba"}, 1, "f", 0),)),
    ("GBA", "_Console/GBA", (({".gba"}, 1, "f", 0),)),
    ("GameGear", "_Console/SMS", (({".gg"}, 1, "f", 2),)),
    ("Genesis", "_Console/Genesis", (({".bin", ".gen", ".md"}, 1, "f", 0),)),
    (
        "Intellivision",
        "_Console/Intellivision",
        (({".rom", ".int", ".bin"}, 1, "f", 1),),
    ),
    ("MegaDrive", "_Console/MegaDrive", (({".bin", ".gen", ".md"}, 1, "f", 1),)),
    ("MegaCD", "_Console/MegaCD", (({".cue", ".chd"}, 1, "s", 0),)),
    ("N64", "_Console/N64", (({".n64", ".z64"}, 1, "f", 1),)),
    ("NeoGeo-CD", "_Console/NeoGeo", (({".cue", ".chd"}, 1, "s", 1),),),
    ("NeoGeo", "_Console/NeoGeo", (({".neo"}, 1, "f", 1),),),
    ("NES", "_Console/NES", (({".nes", ".fds", ".nsf"}, 1, "f", 0),)),
    ("ODYSSEY2", "_Console/Odyssey2", (({".bin"}, 1, "f", 1),)),
    ("PSX", "_Console/PSX", (({".cue", ".chd"}, 1, "s", 1),)),
    ("PocketChallengeV2", "_Console/WonderSwan", (({".pc2"}, 1, "f", 1),)),
    ("PokemonMini", "_Console/PokemonMini", (({".min"}, 1, "f", 1),)),
    ("S32X", "_Console/S32X", (({".32x"}, 1, "f", 0),)),
    ("Saturn", "_Console/Saturn", (({".cue", ".chd"}, 1, "s", 0),)),
    ("SG1000", "_Console/ColecoVision", ({".sg"}, 1, "f", 2),),
    ("SGB", "_Console/SGB", (({".gb", ".gbc"}, 1, "f", 1),)),
    ("SMS", "_Console/SMS", (({".sms", ".sg"}, 1, "f", 1),)),
    ("SNES", "_Console/SNES", (({".sfc", ".smc"}, 2, "f", 0),)),
    ("SuperVision", "_Console/SuperVision", (({".bin", ".sv"}, 1, "s", 1),)),
    (
        "TGFX16-CD",
        "_Console/TurboGrafx16",
        (({".cue", ".chd"}, 1, "s", 0),),
    ),
    (
        "TGFX16",
        "_Console/TurboGrafx16",
        (
            ({".pce", ".bin"}, 1, "f", 0),
            ({".sgx"}, 1, "f", 1),
        ),
    ),
    ("VC4000", "_Console/VC4000", (({".bin"}, 1, "f", 1),)),
    ("VECTREX", "_Console/Vectrex", (({".ovr", ".vec", ".bin", ".rom"}, 1, "f", 1),)),
    ("WonderSwan", "_Console/WonderSwan", (({".ws"}, 1, "f", 1),)),
    ("WonderSwanColor", "_Console/WonderSwan", (({".wsc"}, 1, "f", 1),)),
)

# source: https://mister-devel.github.io/MkDocs_MiSTer/cores/paths/#path-priority
GAMES_FOLDERS = (
    "/media/usb0/games",
    "/media/usb0",
    "/media/usb1/games",
    "/media/usb1",
    "/media/usb2/games",
    "/media/usb2",
    "/media/usb3/games",
    "/media/usb3",
    "/media/usb4/games",
    "/media/usb4",
    "/media/usb5/games",
    "/media/usb5",
    "/media/fat/cifs/games",
    "/media/fat/cifs",
    "/media/fat/games",
    "/media/fat",
)


def get_names_replacement(name: str):
    if name in NAMES_CACHE:
        return NAMES_CACHE[name]
    if not os.path.exists(NAMES_FILE):
        return name
    with open(NAMES_FILE, "r") as f:
        for entry in f:
            if ":" in entry:
                system, replacement = entry.split(":", maxsplit=1)
                replacement = replacement.strip()
                if system.strip().lower() == name.lower():
                    # remove illegal filename characters
                    replacement = replacement.replace("/", " & ")
                    for char in '<>:"/\\|?*':
                        if char in replacement:
                            replacement = replacement.replace(char, " ")
                    NAMES_CACHE[name] = replacement
                    return replacement
    return name


def folder_name(system_name):
    return "_" + get_names_replacement(system_name)


# generate XML contents for MGL file
def generate_mgl(rbf, delay, type, index, path):
    mgl = '<mistergamedescription>\n\t<rbf>{}</rbf>\n\t<file delay="{}" type="{}" index="{}" path="{}"/>\n</mistergamedescription>'
    return mgl.format(rbf, delay, type, index, path)


def get_mgl_target(path):
    with open(path, "r") as f:
        match = re.search(r'path="\.\./\.\./\.\./\.\.(.+)"', f.read())
        if match:
            return match.group(1)
        else:
            match = re.search(r'path="(.+)"', f.read())
            if match:
                return match.group(1)
        return ""


def get_system(name: str):
    for system in MGL_MAP:
        if name.lower() == system[0].lower():
            return system
    return False


def match_system_file(system, filename):
    _, ext = os.path.splitext(filename)
    for type in system[2]:
        if ext.lower() in type[0]:
            return type


# {<system name> -> <full games path>[]}
def get_system_paths():
    systems = {}

    def add_system(name, folder):
        if name in systems:
            systems[name].append(folder)
        else:
            systems[name] = [folder]

    def check_folder(path):
        if not os.path.exists(path) or not os.path.isdir(path):
            return False
        return True

    for system in MGL_MAP:
        for games_path in GAMES_FOLDERS:
            games_path = os.path.join(games_path, system[0])
            if check_folder(games_path):
                add_system(system[0], games_path)

    return systems


def to_mgl_args(system, match, full_path):
    return (
        system[1],
        match[1],
        match[2],
        match[3],
        full_path,
    )


# return a generator for all valid system roms
# (<system>, <full path>, <relative folder>, <mgl filename>, match)[]
def get_system_files(name, folder):
    system = get_system(name)
    for root, _, files in os.walk(folder):
        for filename in files:
            path = os.path.join(root, filename)
            if filename.lower().endswith(".zip") and zipfile.is_zipfile(path):
                # zip files
                for zip_path in zipfile.ZipFile(path).namelist():
                    match = match_system_file(system, zip_path)
                    if match:
                        full_path = os.path.join(path, zip_path)
                        rel_path = (
                            os.path.join(root, os.path.dirname(zip_path))
                            .replace(folder, "")
                            .lstrip("/")
                        )
                        yield (
                            system,
                            full_path,
                            rel_path,
                            os.path.basename(zip_path),
                            match,
                        )
            else:
                # regular files
                match = match_system_file(system, filename)
                if match is not None:
                    rel_path = root.replace(folder, "").lstrip("/")
                    yield (system, path, rel_path, filename, match)


# format menu folder names to show in menu core
def menu_format(sub_path):
    if sub_path in MENU_CACHE:
        return MENU_CACHE[sub_path]
    folders = sub_path.split(os.path.sep)
    path = "/".join([folder_name(x) for x in folders])
    MENU_CACHE[sub_path] = path
    return path


def mgl_name(filename):
    name, _ = os.path.splitext(filename)
    return name + ".mgl"


def create_mgl_file(system_name, filename, mgl_args, sub_path):
    rel_path = os.path.join(system_name, sub_path).rstrip("/")
    mgl_folder = os.path.join(GAMES_MENU_PATH, menu_format(rel_path))
    mgl_path = os.path.join(mgl_folder, mgl_name(filename))
    if not os.path.exists(mgl_folder):
        os.makedirs(mgl_folder)
    if not os.path.exists(mgl_path):
        try:
            with open(mgl_path, "w") as f:
                mgl = generate_mgl(*mgl_args)
                f.write(mgl)
            return True
        except OSError:
            return False
    else:
        return False


def dialog_env():
    return dict(os.environ, DIALOGRC="/media/fat/Scripts/.dialogrc")


def display_message(msg, info=False, height=5, title="Games Menu"):
    if info:
        type = "--infobox"
    else:
        type = "--msgbox"

    args = [
        "dialog",
        "--title",
        title,
        "--ok-label",
        "Ok",
        type,
        msg,
        str(height),
        "75",
    ]

    subprocess.run(args, env=dialog_env())


def display_generate_mgls(system_names):
    system_paths = get_system_paths()
    system_names = [x for x in system_names if not x.startswith("ZOPT")]

    def display_progress(msg, pct):
        args = [
            "dialog",
            "--title",
            "Creating MGL files...",
            "--gauge",
            msg,
            "6",
            "75",
            str(pct),
        ]
        progress = subprocess.Popen(args, env=dialog_env(), stdin=subprocess.PIPE)
        progress.communicate("".encode())

    for i, system_name in enumerate(system_names):
        for folder in system_paths[system_name]:
            pct = math.ceil(i / len(systems) * 100)
            display_progress(f"Scanning {get_names_replacement(system_name)} ({folder})", pct)
            for system, path, parent, filename, match in get_system_files(
                    system_name, folder
            ):
                mgl_args = to_mgl_args(system, match, path)
                created = create_mgl_file(system_name, filename, mgl_args, parent)
    display_progress(f"Scanning {get_names_replacement(system_name)} ({folder})", 100)


def display_menu(system_paths):
    systems = {}
    max_name_len = 0
    last_item = ""

    for name in system_paths.keys():
        display_name = get_names_replacement(name)

        if os.path.exists(
                os.path.join(GAMES_MENU_PATH, folder_name(display_name))
        ) or not os.path.exists(GAMES_MENU_PATH):
            selected = True
        else:
            selected = False

        systems[name] = {
            "display_name": display_name,
            "selected": selected,
        }

    for system in systems.values():
        name_len = len(system["display_name"])
        if name_len > max_name_len:
            max_name_len = name_len

    def menu():
        args = [
            "dialog",
            "--title",
            "Games Menu",
            "--ok-label",
            "Toggle",
            "--cancel-label",
            "Exit",
            "--extra-button",
            "--extra-label",
            "Generate Menu",
            "--default-item",
            str(last_item),
            "--menu",
            "Select systems to show in Games menu:",
            "20",
            "75",
            "20",
        ]

        for name in sorted(systems.keys(), key=str.lower):
            args.append(str(name))
            display_str = systems[name]["display_name"].ljust(max_name_len + 2)
            if systems[name]["selected"]:
                display_str = display_str + "[YES]"
            else:
                display_str = display_str + " [NO]"
            args.append(str(display_str))

        result = subprocess.run(args, env=dialog_env(), stderr=subprocess.PIPE)

        button = result.returncode
        selection = result.stderr.decode()

        return button, selection

    button, selection = menu()
    while button == 0:
        systems[selection]["selected"] = not systems[selection]["selected"]
        last_item = selection
        button, selection = menu()

    if button == 3:
        selected = []
        for k, v in systems.items():
            if v["selected"]:
                selected.append(k)
        return selected
    else:
        return None


def display_yesno(msg, title="Games Menu"):
    args = [
        "dialog",
        "--title",
        title,
        "--ok-label",
        "Yes",
        "--cancel-label",
        "No",
        "--yesno",
        msg,
        "5",
        "75",
    ]

    result = subprocess.run(args, env=dialog_env(), stderr=subprocess.PIPE)
    return result.returncode == 0


def display_welcome():
    msg = """Games Menu generates a set of direct shortcuts to games in your MiSTer menu. Select the systems you want to be included on the next screen, and then select Generate Menu. A new Games menu will appear in the main MiSTer menu. Shortcuts will reflect the layout of games on disk.

WARNING: Shortcuts can take up much more disk space than expected, depending on how many games you have. For example, a full set of all selectable systems can take up to 10GB of disk space. This is based on the number of games, not game file size."""

    display_message(msg, height=13)


if __name__ == "__main__":
    if not os.path.exists(GAMES_MENU_PATH):
        display_welcome()

    system_paths = get_system_paths()
    systems = display_menu(system_paths)
    print("")

    if systems is not None:
        if len(systems) == 0 or systems[0] == "":
            do_delete = display_yesno("Remove the Games menu from your system?")

            if do_delete:
                print("")
                print("Deleting Games menu folder...", end="", flush=True)
                if os.path.exists(GAMES_MENU_PATH):
                    shutil.rmtree(GAMES_MENU_PATH)
                print("Done!", flush=True)

            sys.exit(0)

        if not os.path.exists(GAMES_MENU_PATH):
            os.mkdir(GAMES_MENU_PATH)

        # delete systems
        all_folders = [folder_name(x) for x in systems if not x.startswith("ZOPT")]
        for folder in os.listdir(GAMES_MENU_PATH):
            path = os.path.join(GAMES_MENU_PATH, folder)
            if os.path.isdir(path) and folder not in all_folders:
                print(f"Removing {folder} ...", end="", flush=True)
                shutil.rmtree(path)
                print(" Done!", flush=True)

        # add/update systems
        display_generate_mgls(systems)

    sys.exit(0)
