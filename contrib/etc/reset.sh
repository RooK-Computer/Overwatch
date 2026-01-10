#!/bin/bash

echo -n "RESET" | /usr/bin/nc -u -w1 127.0.0.1 55355
