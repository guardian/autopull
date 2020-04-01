#!/bin/bash

SCRIPT_PATH=$(realpath $0)
echo "$SCRIPT_PATH"

SCRIPT_DIR=$(dirname "$SCRIPT_PATH")
echo "$SCRIPT_DIR"

if [ ! -f autopull.exe ]; then
    echo Windows build not found. Ensure that build.sh has been run first, and that you are running from the root of the source
    exit 1
fi

mkdir -p /tmp/autopull-win/autopull
cp autopull.exe /tmp/autopull-win/autopull
cp autopull.yaml /tmp/autopull-win/autopull
cp install_handler.ps1 /tmp/autopull-win/autopull

cd /tmp/autopull-win

zip -r autopull-win.zip autopull/*
mv autopull-win.zip "$SCRIPT_DIR/.." || echo Could not move zipfile
rm -rf /tmp/autopull-win

if [ ! -f autopull.linux64 ]; then
  echo Linux build not found. Ensure that build.sh has been run first, and that you are running from the root of the source
  exit 1
fi 

mkdir -p /tmp/autopull-lin/autopull
cp autopull.linux64 /tmp/autopull-lin/autopull
cp autopull.yaml /tmp/autopull-lin/autopull

cd /tmp/autopull-lin

zip -r autopull-lin.zip autopull/*
mv autopull-lin.zip "$SCRIPT_DIR/.." || echo Could not move zipfile
rm -rf /tmp/autopull-lin