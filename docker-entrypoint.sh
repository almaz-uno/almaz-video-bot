#!/bin/sh -eu

cd /app

echo "deb http://deb.debian.org/debian/ bullseye-backports main contrib non-free" > /etc/apt/sources.list.d/backports.list
apt update
apt upgrade -y
apt install -y yt-dlp

_term() { 
  echo "Caught SIGTERM signal!" 
  kill -TERM "$child" 2>/dev/null
}

trap _term TERM INT

# make lint run-intest

. /app/.env

# mkdir -p /app/.bin
# GOBIN=/app/.bin go install
# /app/.bin/almaz-video-bot &

go run . &

child=$!
wait "$child"
