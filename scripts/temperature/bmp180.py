# The MIT License (MIT)
# Copyright (c) 2015-2016 MrTijn/Tijndagamer
# Source: https://github.com/m-rtijn/bmp180

import smbus
import math
import time

CONTROL_REG = 0xF4
DATA_REG = 0xF6
CAL_AC5_REG = 0xB2
CAL_AC6_REG = 0xB4
CAL_MC_REG = 0xBC
CAL_MD_REG = 0xBE

class bmp180:
    address = None
    bus = smbus.SMBus(2)

    cal_ac5 = 0
    cal_ac6 = 0
    cal_mc = 0
    cal_md = 0

    def __init__(self, address):
        self.address = address
        self.cal_ac5 = self.read_u16(CAL_AC5_REG)
        self.cal_ac6 = self.read_u16(CAL_AC6_REG)
        self.cal_mc = self.read_s16(CAL_MC_REG)
        self.cal_md = self.read_s16(CAL_MD_REG)

    def read_s16(self, register):
        msb = self.bus.read_byte_data(self.address, register)
        lsb = self.bus.read_byte_data(self.address, register + 1)
        if msb > 127:
            msb -= 256
        return (msb << 8) + lsb

    def read_u16(self, register):
        msb = self.bus.read_byte_data(self.address, register)
        lsb = self.bus.read_byte_data(self.address, register + 1)
        return (msb << 8) + lsb

    def get_temp(self):
        self.bus.write_byte_data(self.address, CONTROL_REG, 0x2E)

        time.sleep(0.0045)

        ut = self.read_u16(DATA_REG)

        x1 = 0
        x2 = 0
        b5 = 0
        actual_temp = 0.0

        x1 = ((ut - self.cal_ac6) * self.cal_ac5) / math.pow(2, 15)
        x2 = (self.cal_mc * math.pow(2, 11)) / (x1 + self.cal_md)
        b5 = x1 + x2
        actual_temp = ((b5 + 8) / math.pow(2, 4)) / 10

        return actual_temp

bmp = bmp180(0x77)
while True:
    print("%.2fC" % bmp.get_temp())
    time.sleep(1)