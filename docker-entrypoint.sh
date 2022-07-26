#!/bin/sh -eu

cd /app

echo "deb http://deb.debian.org/debian/ bullseye-backports main contrib non-free" > /etc/apt/sources.list.d/backports.list
apt update
apt upgrade -y
apt install -y yt-dlp fuse

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

# Build credentials file
mkdir -p /root/.aws/ /mnt/s3
chmod 700 /root/.aws
cat - > /root/.aws/credentials <<CREDENTIALS
[default]
aws_access_key_id = ${S3_ACCESS}
aws_secret_access_key = ${S3_SECRET}
CREDENTIALS

./bin/goofys -o allow_other -f --endpoint=${S3_ENDPOINT} ${S3_BUCKET} /mnt/s3 &

go run . &

child=$!
wait "$child"
