#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build -o amber-login.linux.bin
upx amber-login.linux.bin
