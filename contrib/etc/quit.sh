#!/bin/bash

echo -n "QUIT" | /usr/bin/nc -u -w1 127.0.0.1 55355
