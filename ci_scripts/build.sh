#!/bin/bash

echo Building for Windows...
GOOS=windows GOARCH=amd64 go build -o autopull.exe
ls -lh autopull.exe
echo Building for Mac...
GOOS=darwin GOARCH=amd64 go build -o autopull.macos
ls -lh autopull.macos
echo Bulding for Linux...
GOOS=linux GOARCH=amd64 go build -o autopull.linux64
ls -lh autopull.linux64