#!/usr/bin/env python3

import time
import requests
import os

MISTER_ADDR = "10.0.0.218"
REMOTE_URL = "http://{}:8182/api".format(MISTER_ADDR)
SPYRO_PATH = "/media/fat/Spyro_cifs.mgl"


def launch_menu():
    print("Launching menu... ", end="")
    r = requests.post(REMOTE_URL + "/launch/menu")
    if r.status_code == 200:
        print("OK")
    else:
        print("ERROR")


def launch_spyro():
    print("Launching Spyro... ", end="")
    r = requests.post(REMOTE_URL + "/launch", json={"path": SPYRO_PATH})
    if r.status_code == 200:
        print("OK")
    else:
        print("ERROR")


def load_savestate():
    print("Loading save state... ", end="")
    r = requests.post(REMOTE_URL + "/controls/keyboard-raw/59")  # F1
    if r.status_code == 200:
        print("OK")
    else:
        print("ERROR")


def generate_test_file():
    print("Generating test file... ", end="")
    with open("random_file", "wb") as f:
        f.write(os.urandom(1024 * 1024 * 100))
    print("OK")


def scp_test_file():
    print("Copying test file... ", end="")
    ret = os.system("scp random_file root@{}:/media/fat".format(MISTER_ADDR))
    if ret == 0:
        print("OK")
    else:
        print("ERROR")


def main():
    launch_menu()
    time.sleep(3)
    launch_spyro()
    time.sleep(5)
    load_savestate()
    # generate_test_file()
    # scp_test_file()


if __name__ == "__main__":
    main()
