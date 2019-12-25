#!/usr/bin/env bash

go build -ldflags="-s -w" -o amber-login.bin
upx --brute amber-login.bin
