#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o amber-login.linux.bin
upx --brute amber-login.linux.bin
