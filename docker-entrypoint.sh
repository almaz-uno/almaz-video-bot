#!/bin/sh -eu

cd /app

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
