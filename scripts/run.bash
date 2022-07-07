#!/bin/bash -eu

cd $(dirname $(realpath $0))/..

echo -en "\033]0;🤖 Running...\a"

. .env

BINDIR=$(realpath .bin)

mkdir -p $BINDIR

while sleep 2s;
do 
    echo -en "\033]0;🤖 Running...\a"
    echo "=================================================="
    echo "🤖 Running..."
    go run . || echo "✖✖✖ failed to start"
    echo -en "\033]0;🤖 Restarting...\a"
done