#!/usr/sh

echo "Building getanime-arm"
GOOS=linux GOARCH=arm go build -o getanime-arm .

