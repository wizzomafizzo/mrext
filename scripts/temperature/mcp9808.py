#!/usr/bin/python

import time
from smbus import SMBus

bus = SMBus(2)


def lsb_msb(word):
    return word[0:8], word[12:16]


while True:
    temp_binary = format(bus.read_word_data(0x18, 0x05), "016b")
    lsb, msb = lsb_msb(temp_binary)
    print("%.2fC" % (float(int(msb + lsb, 2)) / 16))
    time.sleep(1)
