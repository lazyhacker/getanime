#!/usr/sh

echo "Building sorttorrent-arm"
GOOS=linux GOARCH=arm go build -o sorttorrent-arm sorttorrent.go

echo "Building getanime-arm"
GOOS=linux GOARCH=arm go build -o getanime-arm getanime.go

