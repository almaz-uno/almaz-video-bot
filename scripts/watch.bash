#!/bin/bash -eu

cd $(dirname $(realpath $0))/..
echo -en "\033]0;⏱🤖 Watching $(pwd)...\a"
#⌚ 

BINDIR=$(realpath .bin)

while inotifywait -e close_write -r . --exclude '(\.git)|(testdata)|(\.bin)' ; 
do 
    echo -en "\033]0;⏯ Sending stop signal!\a"
    sleep 1s
    killall -TERM almaz-video-bot || true
    echo -en "\033]0;⏱ Watching $(pwd)...\a"
done