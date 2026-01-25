#!/bin/bash

echo -n "QUIT" | /usr/bin/nc -u -w1 127.0.0.1 55355
killall dosbox
killall redream.aarch64.elf
killall redream.aarch32.elf
killall aethersx2
