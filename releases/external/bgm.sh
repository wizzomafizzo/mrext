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

DEFAULT_PLAYBACK = "random"
DEFAULT_PLAYLIST = None
MUSIC_FOLDER = "/media/fat/music"
BOOT_FOLDER = os.path.join(MUSIC_FOLDER, "boot")
CORE_BOOT_DELAY = 0
ENABLE_STARTUP = True
PLAY_IN_CORE = False
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
DEBUG = False


# TODO: separate remote control http server
# TODO: option to play music after inactivity period
# TODO: option to adjust adjust volume on menu launch


# read ini file
INI_FILE = os.path.join(MUSIC_FOLDER, INI_FILENAME)
if os.path.exists(INI_FILE):
    ini = configparser.ConfigParser()
    ini.read(INI_FILE)
    DEFAULT_PLAYBACK = ini.get("bgm", "playback", fallback=DEFAULT_PLAYBACK)
    DEBUG = ini.getboolean("bgm", "debug", fallback=DEBUG)
    ENABLE_STARTUP = ini.getboolean("bgm", "startup", fallback=ENABLE_STARTUP)
    PLAY_IN_CORE = ini.getboolean("bgm", "playincore", fallback=PLAY_IN_CORE)
    CORE_BOOT_DELAY = ini.getfloat("bgm", "corebootdelay", fallback=CORE_BOOT_DELAY)
    DEFAULT_PLAYLIST = ini.get("bgm", "playlist", fallback=DEFAULT_PLAYLIST)
    if DEFAULT_PLAYLIST == "none":
        DEFAULT_PLAYLIST = None
else:
    # create a default ini
    if os.path.exists(MUSIC_FOLDER):
        with open(INI_FILE, "w") as f:
            f.write(
                "[bgm]\nplayback = random\nplaylist = none\nstartup = yes\nplayincore = no\ncorebootdelay = 0\ndebug = no\n"
            )


def log(msg: str, always_print=False):
    if msg == "":
        return
    if always_print or DEBUG:
        print(msg)
    if DEBUG:
        with open(LOG_FILE, "a") as f:
            f.write(
                "[{}] {}\n".format(
                    datetime.datetime.isoformat(datetime.datetime.now()), msg
                )
            )


def random_index(list):
    return random.randint(0, len(list) - 1)


def get_core():
    if not os.path.exists(CORENAME_FILE):
        return None

    with open(CORENAME_FILE) as f:
        return str(f.read())


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


class Player:
    mutex = threading.Lock()
    player = None
    playing = None
    playback = DEFAULT_PLAYBACK
    playlist = DEFAULT_PLAYLIST
    playlist_thread = None
    end_playlist = threading.Event()
    history = []

    def is_mp3(self, filename: str):
        return filename.lower().endswith(".mp3")

    def is_pls(self, filename: str):
        return filename.lower().endswith(".pls")

    def is_ogg(self, filename: str):
        return filename.lower().endswith(".ogg")

    def is_wav(self, filename: str):
        return filename.lower().endswith(".wav")

    def is_mid(self, filename: str):
        return filename.lower().endswith(".mid")

    def is_vgm(self, filename: str):
        match = re.search(".*\.(vgm|vgz|vgm\.gz)$", filename.lower())
        return match is not None

    def is_valid_file(self, filename: str):
        return (
            self.is_mp3(filename)
            or self.is_ogg(filename)
            or self.is_wav(filename)
            or self.is_mid(filename)
            or self.is_vgm(filename)
            or self.is_pls(filename)
        )

    def get_loop_amount(self, filename: str):
        loop_match = re.search("^X(\d\d)\_", os.path.basename(filename))
        if loop_match is not None:
            return int(loop_match.group(1))
        else:
            return 1

    def get_pls_url(self, filename: str):
        with open(filename, "r") as f:
            contents = f.read()
            match = re.search("https?:.+", contents, re.MULTILINE)
            if match is not None:
                return match[0]
            else:
                log("Playlist URL not found")
                return ""

    def play_mp3(self, filename: str):
        # get url from playlist files
        if self.is_pls(filename):
            filename = self.get_pls_url(filename)

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
                return None
            else:
                return folder

    def filter_tracks(self, files, include_boot=False):
        tracks = []
        for track in files:
            if include_boot and self.is_valid_file(track):
                tracks.append(track)
            else:
                if self.is_valid_file(track) and not track.startswith("_"):
                    if self.playlist == "all" and not self.is_pls(track):
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

        if self.is_valid_file(filename):
            self.playing = filename
            self.add_history(filename)
            log("Now playing: {}".format(filename))
        else:
            return

        loop = self.get_loop_amount(filename)

        while loop > 0 and self.playing is not None:
            log("Loop #{}".format(loop))
            if self.is_mp3(filename):
                self.play_mp3(filename)
            elif self.is_pls(filename):
                self.play_mp3(filename)
            elif self.is_ogg(filename):
                self.play_ogg(filename)
            elif self.is_wav(filename):
                self.play_wav(filename)
            elif self.is_mid(filename):
                self.play_mid(filename)
            elif self.is_vgm(filename):
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
            log("Changed playist: {}".format(name))
            self.playlist = name
            if self.in_playlist():
                self.start_playlist()

    def start_remote(self):
        s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        s.bind(SOCKET_FILE)

        def handler(cmd: str):
            global PLAY_IN_CORE

            self.mutex.acquire()

            log("Received command: {}".format(cmd))
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
                PLAY_IN_CORE = True
            elif cmd == "set playincore no":
                PLAY_IN_CORE = False
            elif cmd.startswith("get"):
                args = cmd.split(" ", 1)
                self.mutex.release()
                if len(args) > 1:
                    if args[1] == "playlist":
                        return self.playlist
                    elif args[1] == "playback":
                        return self.playback
                    elif args[1] == "playincore":
                        if PLAY_IN_CORE:
                            return "yes"
                        else:
                            return "no"
                    else:
                        return ""
            else:
                log("Unknown command")

            self.mutex.release()

        def listener():
            while True:
                log("Waiting for command...")
                s.listen()
                conn, addr = s.accept()
                data = conn.recv(MESSAGE_SIZE).decode()
                if data == "quit":
                    break
                log("Got command, sending so handler...")
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
            if name.startswith("_") and self.is_valid_file(name):
                boot_tracks.append(os.path.join(self.get_playlist_path(), name))

        # include tracks from global boot folder
        if os.path.exists(BOOT_FOLDER):
            for name in os.listdir(BOOT_FOLDER):
                if self.is_valid_file(name):
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
                        if self.is_valid_file(filename):
                            tracks.append(filename)

                    if len(tracks) == 0:
                        return

                    time.sleep(CORE_BOOT_DELAY)
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


def cleanup(player: Player):
    if player is not None:
        player.stop_playlist()
        send_socket("quit")
    if os.path.exists(SOCKET_FILE):
        os.remove(SOCKET_FILE)


def start_service(player: Player):
    log("Starting service...")
    log("Playlist folder: {}".format(player.get_playlist_path()))

    player.start_remote()
    # FIXME: make this non-blocking so it can be cut off during core launch
    #        this only affects people with really long boot sounds
    player.play_boot()

    core = get_core()
    # don't start playing if the boot track ran into a core launch
    # do start playing for a bit if the CORENAME file is still being created
    if core == MENU_CORE or core is None or PLAY_IN_CORE:
        player.start_playlist(DEFAULT_PLAYBACK)

    while True:
        new_core = wait_core_change()

        if new_core is None:
            log("CORENAME file is missing, exiting...")
            break

        if core == new_core:
            pass
        elif PLAY_IN_CORE:
            pass
        elif new_core == MENU_CORE:
            log("Switched to menu core, starting playlist...")
            player.mutex.acquire()
            player.start_playlist()
            player.mutex.release()
        elif new_core != MENU_CORE:
            log("Exited menu core, stopping playlist...")
            player.mutex.acquire()
            player.stop_playlist()
            player.play_core_boot(new_core)
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


def display_gui():
    def get_menu_output(output):
        try:
            return int(output)
        except ValueError:
            return None

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

    def active(condition):
        if condition:
            return " [ACTIVE]"
        else:
            return ""

    def menu(status, playlists, config, last_item):
        if status[0] == "yes":
            play_text = "Stop playing"
        else:
            play_text = "Start playing"

        if status[3] == "":
            now_playing = "---"
        else:
            now_playing = status[3]

        if config.getboolean("bgm", "startup", fallback=ENABLE_STARTUP):
            startup = "Disable startup on boot"
        else:
            startup = "Enable startup on boot"

        if config.getboolean("bgm", "playincore", fallback=PLAY_IN_CORE):
            playincore = "Disable music in cores"
        else:
            playincore = "Enable music in cores"

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
            "PLAYLIST > None (just top level files)" + active(status[2] == "none"),
            "9",
            "PLAYLIST > All (all playlists combined)" + active(status[2] == "all"),
        ]

        number = 10
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
            startup = config.getboolean("bgm", "startup", fallback=ENABLE_STARTUP)
            if startup:
                config["bgm"]["startup"] = "no"
            else:
                config["bgm"]["startup"] = "yes"
        elif selection == 7:
            playincore = config.getboolean("bgm", "playincore", fallback=PLAY_IN_CORE)
            if playincore:
                config["bgm"]["playincore"] = "no"
                send_socket("set playincore no")
            else:
                config["bgm"]["playincore"] = "yes"
                send_socket("set playincore yes")
        elif selection == 8:
            send_socket("set playlist none")
            config["bgm"]["playlist"] = "none"
        elif selection == 9:
            send_socket("set playlist all")
            config["bgm"]["playlist"] = "all"
        elif selection > 9:
            name = playlists[selection - 10]
            send_socket("set playlist {}".format(name))
            config["bgm"]["playlist"] = name


if __name__ == "__main__":
    if len(sys.argv) == 2:
        if sys.argv[1] == "exec":
            if os.path.exists(SOCKET_FILE):
                log("BGM service is already running, exiting...", True)
                sys.exit()

            def stop(sn=0, f=0):
                log("Stopping service ({})".format(sn))
                cleanup(player)
                sys.exit()

            signal.signal(signal.SIGINT, stop)
            signal.signal(signal.SIGTERM, stop)
            player = Player()
            start_service(player)
            stop()
        elif sys.argv[1] == "start":
            if not ENABLE_STARTUP:
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
