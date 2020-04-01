#!/bin/bash

if [ ! -f autopull.exe ]; then
    echo Windows build not found. Ensure that build.sh has been run first, and that you are running from the root of the source
    exit 1
fi

mkdir -p /tmp/autopull-win/autopull
cp autopull.exe /tmp/autopull/autopull-win
cp autopull.yaml /tmp/autopull/autopull-win
cp install_handler.ps1 /tmp/autopull/autopull-win

cd /tmp/autopull-win

zip -r autopull-win.zip autopull/*