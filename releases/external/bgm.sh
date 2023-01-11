#!/usr/bin/env python

import os
import sys
import subprocess
import threading
import random
import math
import socket
import configparser
import datetime
import time
import signal
import re

MUSIC_FOLDER = "/media/fat/music"
BOOT_FOLDER = os.path.join(MUSIC_FOLDER, "boot")
HISTORY_SIZE = 0.2  # ratio of total tracks to keep in play history
SOCKET_FILE = "/tmp/bgm.sock"
MESSAGE_SIZE = 4096  # max size of socket payloads
MIDI_PORT = "128:0"
SCRIPTS_FOLDER = "/media/fat/Scripts"
STARTUP_SCRIPT = "/media/fat/linux/user-startup.sh"
CORENAME_FILE = "/tmp/CORENAME"
LOG_FILE = "/tmp/bgm.log"
INI_FILENAME = "bgm.ini"
MENU_CORE = "MENU"
CMD_INTERFACE = "/dev/MiSTer_cmd"
INI_FILE = os.path.join(MUSIC_FOLDER, INI_FILENAME)
CONFIG_DEFAULTS = {
    "playback": "random",
    "playlist": None,
    "startup": True,
    "playincore": False,
    "corebootdelay": 0,
    "menuvolume": -1,
    "defaultvolume": -1,
    "debug": False,
}

# TODO: separate remote control http server
# TODO: option to play music after inactivity period


def write_default_ini():
    if os.path.exists(MUSIC_FOLDER):
        with open(INI_FILE, "w") as f:
            f.write(
                "[bgm]\nplayback = random\nplaylist = none\nstartup = yes\nplayincore = no\n"
                "corebootdelay = 0\nmenuvolume = -1\ndefaultvolume = -1\ndebug = no\n"
            )


def get_ini():
    if not os.path.exists(INI_FILE):
        write_default_ini()

    ini = configparser.ConfigParser()
    ini.read(INI_FILE)

    config = {
        "playback": ini.get("bgm", "playback", fallback=CONFIG_DEFAULTS["playback"]),
        "playlist": ini.get("bgm", "playlist", fallback=CONFIG_DEFAULTS["playlist"]),
        "startup": ini.getboolean("bgm", "startup", fallback=CONFIG_DEFAULTS["startup"]),
        "playincore": ini.getboolean("bgm", "playincore", fallback=CONFIG_DEFAULTS["playincore"]),
        "corebootdelay": ini.getfloat("bgm", "corebootdelay", fallback=CONFIG_DEFAULTS["corebootdelay"]),
        "menuvolume": ini.getint("bgm", "menuvolume", fallback=CONFIG_DEFAULTS["menuvolume"]),
        "defaultvolume": ini.getint("bgm", "defaultvolume", fallback=CONFIG_DEFAULTS["defaultvolume"]),
        "debug": ini.getboolean("bgm", "debug", fallback=CONFIG_DEFAULTS["debug"]),
    }

    if config["playlist"] == "none":
        config["playlist"] = None

    return config


def log(msg: str, always_print=False):
    debug = get_ini()["debug"]
    if msg == "":
        return
    if always_print or debug:
        print(msg)
    if debug:
        with open(LOG_FILE, "a") as f:
            f.write(
                "[{}] {}\n".format(
                    datetime.datetime.isoformat(datetime.datetime.now()), msg
                )
            )


def run_cmd(cmd: str):
    # FIXME: for some reason trying to open the cmd interface while a core is
    #        being loaded can sometimes cause the bgm process to hang waiting
    #        for the handle. this is an attempt to work around that happening
    #        but I've got no idea why it happens in the first place
    return os.system("echo \"{}\" > {}".format(cmd, CMD_INTERFACE))
    # with open(CMD_INTERFACE, "w") as f:
    #     f.write(cmd)


def random_index(xs):
    return random.randint(0, len(xs) - 1)


def get_core():
    if not os.path.exists(CORENAME_FILE):
        return None

    with open(CORENAME_FILE) as f:
        return str(f.read())


def volume_mute():
    run_cmd("volume mute")


def volume_unmute():
    run_cmd("volume unmute")


def volume_set(volume: int):
    if volume < 0:
        volume = 0
    if volume > 7:
        volume = 7

    log("Setting volume to {}".format(volume))
    run_cmd("volume {}".format(volume))


def should_change_volume(ini) -> bool:
    return ini["menuvolume"] >= 0 and ini["defaultvolume"] >= 0


def wait_core_change():
    if get_core() is None:
        log("CORENAME file does not exist, retrying...")
        # keep trying to read it for a little while
        attempts = 0
        while get_core() is None and attempts <= 15:
            time.sleep(1)
            attempts += 1
        if get_core() is None:
            log("No CORENAME file found")
            return None

    # FIXME: not a big deal, but this process can be orphaned during service shutdown
    args = ("inotifywait", "-e", "modify", CORENAME_FILE)
    monitor = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    while monitor is not None and monitor.poll() is None:
        line = monitor.stdout.readline()
        log(line.decode().rstrip())

    if monitor.returncode != 0:
        log("Error when running inotify watch")
        return None

    core = get_core()
    log("Core change to: {}".format(core))
    return core


def is_mp3(filename: str):
    return filename.lower().endswith(".mp3")


def is_pls(filename: str):
    return filename.lower().endswith(".pls")


def is_ogg(filename: str):
    return filename.lower().endswith(".ogg")


def is_wav(filename: str):
    return filename.lower().endswith(".wav")


def is_mid(filename: str):
    return filename.lower().endswith(".mid")


def is_vgm(filename: str):
    match = re.search(r".*\.(vgm|vgz|vgm\.gz)$", filename.lower())
    return match is not None


def is_valid_file(filename: str):
    return (
        is_mp3(filename)
        or is_ogg(filename)
        or is_wav(filename)
        or is_mid(filename)
        or is_vgm(filename)
        or is_pls(filename)
    )


def get_loop_amount(filename: str):
    loop_match = re.search(r"^X(\d\d)_", os.path.basename(filename))
    if loop_match is not None:
        return int(loop_match.group(1))
    else:
        return 1


def get_pls_url(filename: str):
    with open(filename, "r") as f:
        contents = f.read()
        match = re.search(r"https?:.+", contents, re.MULTILINE)
        if match is not None:
            return match[0]
        else:
            log("Playlist URL not found")
            return ""


class Player:
    mutex = None
    player = None
    playing = None
    playback = CONFIG_DEFAULTS["playback"]
    playlist = CONFIG_DEFAULTS["playlist"]
    play_in_core = CONFIG_DEFAULTS["playincore"]
    playlist_thread = None
    end_playlist = None
    history = []

    def __init__(self):
        ini = get_ini()
        self.playback = ini["playback"]
        self.playlist = ini["playlist"]
        self.play_in_core = ini["playincore"]
        self.mutex = threading.Lock()
        self.end_playlist = threading.Event()

    def play_mp3(self, filename: str):
        # get url from playlist files
        if is_pls(filename):
            filename = get_pls_url(filename)

        args = ("mpg123", "--no-control", filename)
        self.player = subprocess.Popen(
            args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT
        )

        # workaround for a strange issue with mpg123 on MiSTer
        # some mp3 files will play but cause mpg123 to hang at the end
        # this may be fixed when MiSTer ships with a newer version
        while self.player is not None:
            line = self.player.stdout.readline()
            output = line.decode().rstrip()
            log(output)
            if (
                "finished." in output
                or self.player is None
                or self.player.poll() is not None
            ):
                self.kill_player()
                break

    def play_file(self, args):
        self.player = subprocess.Popen(
            args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT
        )
        while self.player is not None and self.player.poll() is None:
            line = self.player.stdout.readline()
            log(line.decode().rstrip())
        self.kill_player()

    def play_ogg(self, filename: str):
        args = ("ogg123", filename)
        self.play_file(args)

    def play_wav(self, filename: str):
        args = ("aplay", filename)
        self.play_file(args)

    def play_mid(self, filename: str):
        args = ("aplaymidi", filename, "--port=" + MIDI_PORT)
        self.play_file(args)

    def play_vgm(self, filename: str):
        args = ("vgmplay", filename)
        self.play_file(args)

    def get_playlist_path(self, name=None):
        if name is None:
            name = self.playlist
        if name is None:
            return MUSIC_FOLDER
        else:
            if name == "all":
                folder = MUSIC_FOLDER
            else:
                folder = os.path.join(MUSIC_FOLDER, name)
            if not os.path.exists(folder):
                log("Playlist folder does not exist: {}".format(folder))
                return MUSIC_FOLDER
            else:
                return folder

    def filter_tracks(self, files, include_boot=False):
        tracks = []
        for track in files:
            if include_boot and is_valid_file(track):
                tracks.append(track)
            else:
                if is_valid_file(track) and not track.startswith("_"):
                    if self.playlist == "all" and not is_pls(track):
                        tracks.append(track)
                    elif self.playlist != "all":
                        tracks.append(track)
        return tracks

    def get_tracks(self, playlist=None, include_boot=False):
        if playlist is None:
            playlist = self.playlist
        folder = self.get_playlist_path(playlist)
        tracks = []

        if playlist is None:
            # just the top level folder
            filtered = self.filter_tracks(os.listdir(folder), include_boot)
            for track in filtered:
                tracks.append(os.path.join(folder, track))
        elif folder is not None:
            # otherwise do recursively
            for root, dirs, files in os.walk(folder):
                filtered = self.filter_tracks(files, include_boot)
                for track in filtered:
                    tracks.append(os.path.join(root, track))

        return tracks

    def total_tracks(self, playlist=None, include_boot=False):
        return len(self.get_tracks(playlist, include_boot))

    def add_history(self, filename: str):
        history_size = math.floor(self.total_tracks() * HISTORY_SIZE)
        if history_size < 1:
            return
        while len(self.history) > history_size:
            self.history.pop(0)
        self.history.append(filename)

    def kill_player(self):
        if self.player is not None:
            self.player.kill()
            self.player = None

    def stop(self):
        if self.player is not None:
            self.playing = None
            self.player.kill()
            self.player = None

    def play(self, filename: str):
        self.stop()

        if is_valid_file(filename):
            self.playing = filename
            self.add_history(filename)
            log("Now playing: {}".format(filename))
        else:
            return

        loop = get_loop_amount(filename)

        while loop > 0 and self.playing is not None:
            log("Loop #{}".format(loop))
            if is_mp3(filename):
                self.play_mp3(filename)
            elif is_pls(filename):
                self.play_mp3(filename)
            elif is_ogg(filename):
                self.play_ogg(filename)
            elif is_wav(filename):
                self.play_wav(filename)
            elif is_mid(filename):
                self.play_mid(filename)
            elif is_vgm(filename):
                self.play_vgm(filename)
            loop -= 1

        self.playing = None

    def get_random_track(self):
        tracks = self.get_tracks()
        if len(tracks) == 0:
            return

        index = random_index(tracks)
        # avoid replaying recent tracks
        while tracks[index] in self.history:
            index = random_index(tracks)

        return tracks[index]

    def start_random_playlist(self):
        log("Starting random playlist...")
        self.end_playlist.clear()

        def playlist_loop():
            while not self.end_playlist.is_set():
                track = self.get_random_track()
                if track is None:
                    break
                self.play(track)
            log("Random playlist ended")

        self.playlist_thread = threading.Thread(target=playlist_loop)
        self.playlist_thread.start()

    def start_loop_playlist(self):
        log("Starting loop playlist...")
        self.end_playlist.clear()

        track = self.get_random_track()

        def playlist_loop():
            while not self.end_playlist.is_set() and track is not None:
                self.play(track)
            log("Loop playlist ended")

        self.playlist_thread = threading.Thread(target=playlist_loop)
        self.playlist_thread.start()

    def start_playlist(self, playback=None):
        if playback is None:
            playback = self.playback

        self.stop_playlist()
        if self.playlist_thread is not None and self.playlist_thread.is_alive():
            self.playlist_thread.join()
            self.playlist_thread = None

        if playback == "random":
            self.playback = "random"
            self.start_random_playlist()
        elif playback == "loop":
            self.playback = "loop"
            self.start_loop_playlist()
        elif playback == "disabled":
            self.playback = "disabled"
            return
        else:
            # random playlist is fallback
            self.start_random_playlist()

    def stop_playlist(self):
        self.end_playlist.set()
        self.stop()

    def in_playlist(self):
        return not self.end_playlist.is_set()

    def change_playlist(self, name: str):
        if name == "none":
            name = None
        folder = self.get_playlist_path(name)
        if folder is not None and self.total_tracks(name, include_boot=True) > 0:
            log("Changed playlist: {}".format(name))
            self.playlist = name
            if self.in_playlist():
                self.start_playlist()

    def start_remote(self):
        s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        s.bind(SOCKET_FILE)

        def handler(cmd: str):
            self.mutex.acquire()

            if cmd == "stop":
                self.stop_playlist()
            elif cmd == "play":
                self.start_playlist()
            elif cmd == "skip":
                self.stop()
            elif cmd == "pid":
                self.mutex.release()
                return os.getpid()
            elif cmd == "status":
                if self.playing is not None:
                    is_playing = "yes"
                else:
                    is_playing = "no"
                if self.playlist is None:
                    playlist = "none"
                else:
                    playlist = self.playlist
                if self.playing is not None:
                    filename = os.path.basename(self.playing)
                else:
                    filename = ""
                self.mutex.release()
                return "{}\t{}\t{}\t{}".format(
                    is_playing, self.playback, playlist, filename
                )
            elif cmd.startswith("set playlist"):
                args = cmd.split(" ", 2)
                if len(args) > 2:
                    name = args[2]
                    self.change_playlist(name)
            elif cmd.startswith("set playback"):
                args = cmd.split(" ", 2)
                if len(args) > 2:
                    self.playback = args[2]
                    if self.in_playlist():
                        self.start_playlist()
            elif cmd == "set playincore yes":
                self.play_in_core = True
            elif cmd == "set playincore no":
                self.play_in_core = False
            elif cmd.startswith("get"):
                args = cmd.split(" ", 1)
                self.mutex.release()
                if len(args) > 1:
                    if args[1] == "playlist":
                        return self.playlist
                    elif args[1] == "playback":
                        return self.playback
                    elif args[1] == "playincore":
                        if self.play_in_core:
                            return "yes"
                        else:
                            return "no"
                    else:
                        return ""
            else:
                log("Unknown command: {}".format(cmd))

            self.mutex.release()

        def listener():
            while True:
                s.listen()
                conn, addr = s.accept()
                data = conn.recv(MESSAGE_SIZE).decode()
                if data == "quit":
                    break
                response = handler(data)
                if response is not None:
                    conn.send(str(response).encode())
                conn.close()
            s.close()
            log("Remote stopped")

        log("Starting remote...")
        remote = threading.Thread(target=listener)
        remote.start()

    def get_boot_track(self):
        boot_tracks = []

        for name in os.listdir(self.get_playlist_path()):
            if name.startswith("_") and is_valid_file(name):
                boot_tracks.append(os.path.join(self.get_playlist_path(), name))

        # include tracks from global boot folder
        if os.path.exists(BOOT_FOLDER):
            for name in os.listdir(BOOT_FOLDER):
                if is_valid_file(name):
                    boot_tracks.append(os.path.join(BOOT_FOLDER, name))

        if len(boot_tracks) > 0:
            return boot_tracks[random_index(boot_tracks)]
        else:
            return None

    def play_boot(self):
        track = self.get_boot_track()
        if track is not None:
            log("Selected boot track: {}".format(track))
            self.play(track)

    def play_core_boot(self, core):
        if not os.path.exists(BOOT_FOLDER):
            return

        for core_folder in os.listdir(BOOT_FOLDER):
            if core_folder.lower() == core.lower():
                if os.path.isdir(os.path.join(BOOT_FOLDER, core_folder)):
                    tracks = []

                    for f in os.listdir(os.path.join(BOOT_FOLDER, core_folder)):
                        filename = os.path.join(BOOT_FOLDER, core_folder, f)
                        if is_valid_file(filename):
                            tracks.append(filename)

                    if len(tracks) == 0:
                        return

                    time.sleep(get_ini()["corebootdelay"])
                    log("Playing core boot track...")
                    self.play(tracks[random_index(tracks)])


def send_socket(msg: str):
    if not os.path.exists(SOCKET_FILE):
        return
    s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    s.connect(SOCKET_FILE)
    s.send(msg.encode())
    response = s.recv(MESSAGE_SIZE)
    s.close()
    if len(response) > 0:
        return response.decode()


def cleanup(p: Player):
    if p is not None:
        p.stop_playlist()
        send_socket("quit")
    if os.path.exists(SOCKET_FILE):
        os.remove(SOCKET_FILE)


def start_service(player: Player):
    log("Starting service...")
    log("Playlist folder: {}".format(player.get_playlist_path()))

    ini = get_ini()

    player.start_remote()
    # FIXME: this is a hack to make sure the boot sound doesn't play too loud
    if should_change_volume(ini):
        volume_set(ini["menuvolume"])
        # FIXME: make this non-blocking so it can be cut off during core launch
        #        this only affects people with really long boot sounds
        player.play_boot()
        volume_set(ini["defaultvolume"])
    else:
        player.play_boot()

    core = get_core()
    # don't start playing if the boot track ran into a core launch
    # do start playing for a bit if the CORENAME file is still being created
    if core == MENU_CORE or core is None or player.play_in_core:
        if should_change_volume(ini) and core == MENU_CORE:
            volume_set(ini["menuvolume"])
        player.start_playlist(ini["playback"])

    while True:
        new_core = wait_core_change()
        ini = get_ini()

        if new_core is None:
            log("CORENAME file is missing, exiting...")
            break

        if core == new_core:
            log("CORENAME file changed, but core is the same")
            pass
        elif player.play_in_core:
            log("playincore is enabled")
            pass
        elif new_core == MENU_CORE:
            log("Switched to menu core, starting playlist...")
            log("Grabbing mutex")
            player.mutex.acquire()
            if should_change_volume(ini):
                log("Changing volume to menu volume")
                volume_set(ini["menuvolume"])
            log("Starting playlist")
            player.start_playlist()
            log("Releasing mutex")
            player.mutex.release()
        elif new_core != MENU_CORE:
            log("Exited menu core, stopping playlist...")
            log("Grabbing mutex")
            player.mutex.acquire()
            log("Stopping playlist")
            player.stop_playlist()
            log("Playing core boot")
            player.play_core_boot(new_core)
            if should_change_volume(ini):
                log("Changing volume to default volume")
                volume_set(ini["defaultvolume"])
            log("Releasing mutex")
            player.mutex.release()

        core = new_core


def try_add_to_startup():
    if not os.path.exists(STARTUP_SCRIPT):
        # create a new startup script
        with open(STARTUP_SCRIPT, "w") as f:
            f.write("#!/bin/sh\n")

    with open(STARTUP_SCRIPT, "r") as f:
        if "Startup BGM" in f.read():
            return

    with open(STARTUP_SCRIPT, "a") as f:
        bgm = os.path.join(SCRIPTS_FOLDER, "bgm.sh")
        f.write("\n# Startup BGM\n[[ -e {} ]] && {} $1\n".format(bgm, bgm))
        log("Added service to startup script.", True)


def get_menu_output(output):
    try:
        return int(output)
    except ValueError:
        return None


def active(condition):
    if condition:
        return " [ACTIVE]"
    else:
        return ""


def volume_select_dialog(title: str, blurb: str, current: int, set_live=False) -> (int, int):
    args = [
        "dialog",
        "--title",
        title,
        "--ok-label",
        "Set",
        "--cancel-label",
        "Cancel",
        "--default-item",
        str(current),
        "--menu",
        blurb,
        "20",
        "75",
        "20",
        "-1",
        "Disabled (make no changes to volume)" + active(current == -1),
        "0",
        "Mute" + active(current == 0),
        "1",
        "Level 1" + active(current == 1),
        "2",
        "Level 2" + active(current == 2),
        "3",
        "Level 3" + active(current == 3),
        "4",
        "Level 4" + active(current == 4),
        "5",
        "Level 5" + active(current == 5),
        "6",
        "Level 6" + active(current == 6),
        "7",
        "Level 7 (max)" + active(current == 7),
    ]

    result = subprocess.run(args, stderr=subprocess.PIPE)

    selection = get_menu_output(result.stderr.decode())
    button = get_menu_output(result.returncode)

    if button == 0 and selection is not None and selection >= 0 and set_live:
        volume_set(selection)

    return selection, button


def display_gui():
    def get_status():
        # FIXME: this is a hack to give service time to update itself, would be
        #        better if it could check an update was successful before
        #        getting the status
        time.sleep(0.2)
        return send_socket("status").split("\t")

    def get_playlists():
        excluded_folders = {"boot"}
        playlists = []
        for item in os.listdir(MUSIC_FOLDER):
            if item in excluded_folders:
                continue
            if os.path.isdir(os.path.join(MUSIC_FOLDER, item)):
                playlists.append(item)
        return playlists

    def get_config():
        ini = configparser.ConfigParser()
        ini.read(INI_FILE)
        return ini

    def write_config(config):
        with open(INI_FILE, "w") as f:
            config.write(f)

    def menu(status, playlists, config, last_item):
        if status[0] == "yes":
            play_text = "Stop playing"
        else:
            play_text = "Start playing"

        if status[3] == "":
            now_playing = "---"
        else:
            now_playing = status[3]

        if config.getboolean("bgm", "startup", fallback=CONFIG_DEFAULTS["startup"]):
            startup = "Disable startup on boot"
        else:
            startup = "Enable startup on boot"

        if config.getboolean("bgm", "playincore", fallback=CONFIG_DEFAULTS["playincore"]):
            playincore = "Disable music in cores"
        else:
            playincore = "Enable music in cores"

        menu_volume = config.getint("bgm", "menuvolume", fallback=CONFIG_DEFAULTS["menuvolume"])
        if menu_volume < 0:
            menu_volume_status = "(disabled)"
        else:
            menu_volume_status = "({})".format(menu_volume)

        default_volume = config.getint("bgm", "defaultvolume", fallback=CONFIG_DEFAULTS["defaultvolume"])
        if default_volume < 0:
            default_volume_status = "(disabled)"
        else:
            default_volume_status = "({})".format(default_volume)

        args = [
            "dialog",
            "--title",
            "Background Music",
            "--ok-label",
            "Select",
            "--cancel-label",
            "Exit",
            "--default-item",
            str(last_item),
            "--menu",
            "Now playing: {}\nPlayback: {}\nPlaylist: {}".format(
                now_playing, status[1].title(), status[2]
            ),
            "20",
            "75",
            "20",
            "1",
            "Skip current track",
            "2",
            play_text,
            "3",
            "PLAYBACK > Play random tracks (random)" + active(status[1] == "random"),
            "4",
            "PLAYBACK > Play a single random track on repeat (loop)"
            + active(status[1] == "loop"),
            "5",
            "PLAYBACK > Disable all playback (disabled)"
            + active(status[1] == "disabled"),
            "6",
            "CONFIG   > {}".format(startup),
            "7",
            "CONFIG   > {}".format(playincore),
            "8",
            "CONFIG   > Auto-change menu volume " + menu_volume_status,
            "9",
            "CONFIG   > Auto-change default volume " + default_volume_status,
            "10",
            "PLAYLIST > None (just top level files)" + active(status[2] == "none"),
            "11",
            "PLAYLIST > All (all playlists combined)" + active(status[2] == "all"),
        ]

        number = 12
        for playlist in playlists:
            args.append(str(number))
            args.append(
                "PLAYLIST > {}".format(playlist) + active(status[2] == playlist)
            )
            number += 1

        result = subprocess.run(args, stderr=subprocess.PIPE)

        selection = get_menu_output(result.stderr.decode())
        button = get_menu_output(result.returncode)

        return selection, button

    last_item = ""
    config = get_config()
    button = 0
    while button == 0:
        status = get_status()
        playlists = get_playlists()

        selection, button = menu(status, playlists, config, last_item)

        if selection is None:
            write_config(config)
            break

        last_item = str(selection)

        if selection == 1:
            send_socket("skip")
        elif selection == 2:
            if status[0] == "yes":
                send_socket("stop")
            else:
                send_socket("play")
        elif selection == 3:
            send_socket("set playback random")
            config["bgm"]["playback"] = "random"
        elif selection == 4:
            send_socket("set playback loop")
            config["bgm"]["playback"] = "loop"
        elif selection == 5:
            send_socket("set playback disabled")
            config["bgm"]["playback"] = "disabled"
        elif selection == 6:
            startup = config.getboolean("bgm", "startup", fallback=CONFIG_DEFAULTS["startup"])
            if startup:
                config["bgm"]["startup"] = "no"
            else:
                config["bgm"]["startup"] = "yes"
        elif selection == 7:
            playincore = config.getboolean("bgm", "playincore", fallback=CONFIG_DEFAULTS["playincore"])
            if playincore:
                config["bgm"]["playincore"] = "no"
                send_socket("set playincore no")
            else:
                config["bgm"]["playincore"] = "yes"
                send_socket("set playincore yes")
        elif selection == 8:
            menu_volume = config.getint("bgm", "menuvolume", fallback=CONFIG_DEFAULTS["menuvolume"])
            vol_selection, vol_button = volume_select_dialog(
                "Menu volume", "Automatically change MiSTer's volume in the menu when is open, to adjust music "
                               "volume. The \"default volume\" setting must also be enabled for this feature to work.",
                menu_volume, True
            )
            if vol_button == 0:
                config["bgm"]["menuvolume"] = str(vol_selection)
        elif selection == 9:
            default_volume = config.getint("bgm", "defaultvolume", fallback=CONFIG_DEFAULTS["defaultvolume"])
            vol_selection, vol_button = volume_select_dialog(
                "Default volume", "Automatically revert MiSTer's volume when the menu is closed, to set volume back "
                                  "to normal. The \"menu volume\" setting must also be enabled for this feature to "
                                  "work.",
                default_volume, False
            )
            if vol_button == 0:
                config["bgm"]["defaultvolume"] = str(vol_selection)
        elif selection == 10:
            send_socket("set playlist none")
            config["bgm"]["playlist"] = "none"
        elif selection == 11:
            send_socket("set playlist all")
            config["bgm"]["playlist"] = "all"
        elif selection > 11:
            name = playlists[selection - 12]
            send_socket("set playlist {}".format(name))
            config["bgm"]["playlist"] = name


if __name__ == "__main__":
    ini = get_ini()

    if len(sys.argv) == 2:
        if sys.argv[1] == "exec":
            if os.path.exists(SOCKET_FILE):
                log("BGM service is already running, exiting...", True)
                sys.exit()

            def stop(sn=0, f=None):
                log("Stopping service ({})".format(sn))
                cleanup(player)
                sys.exit()

            signal.signal(signal.SIGINT, stop)
            signal.signal(signal.SIGTERM, stop)
            player = Player()
            start_service(player)
            stop()
        elif sys.argv[1] == "start":
            if not ini["startup"]:
                log("Auto-start is disabled in configuration", True)
                sys.exit()
            os.system("{} exec &".format(os.path.join(SCRIPTS_FOLDER, "bgm.sh")))
            sys.exit()
        elif sys.argv[1] == "stop":
            if not os.path.exists(SOCKET_FILE):
                log("BGM service is not running", True)
                sys.exit()
            pid = send_socket("pid")
            if pid is not None:
                os.system("kill {}".format(pid))
            sys.exit()
        elif sys.argv[1] == "restart":
            script = os.path.join(SCRIPTS_FOLDER, "bgm.sh")
            os.system("{} stop".format(script))
            os.system("{} start".format(script))
            sys.exit()

    if not os.path.exists(MUSIC_FOLDER):
        os.mkdir(MUSIC_FOLDER)
        log("Created music folder.", True)
    try_add_to_startup()

    player = Player()
    if player.total_tracks(include_boot=True) == 0:
        log(
            "Add music files to {} and re-run this script to start.".format(
                MUSIC_FOLDER
            ),
            True,
        )
        sys.exit()
    else:
        if not os.path.exists(SOCKET_FILE):
            log("Starting BGM service...", True)
            os.system("{} exec &".format(os.path.join(SCRIPTS_FOLDER, "bgm.sh")))
            sys.exit()
        else:
            display_gui()
            print("")
            sys.exit()
