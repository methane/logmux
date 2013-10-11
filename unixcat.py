#!/usr/bin/env python
import socket
import sys

sock = socket.socket(socket.AF_UNIX)
sock.connect(sys.argv[1])

for L in iter(sys.stdin.readline, ''):
    sock.sendall(L)
