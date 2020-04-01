#!/bin/bash

GOOS=windows GOARCH=amd64 go build -o autopull.exe
GOOS=darwin GOARCH=amd64 go build -o autopull.macos
GOOS=linux GOARCH=amd64 go build -o autopull.linux64
