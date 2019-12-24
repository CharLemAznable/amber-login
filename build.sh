#!/usr/bin/env bash

go build -ldflags="-s -w" -o amber-login.linux.bin
upx --brute amber-login.linux.bin
