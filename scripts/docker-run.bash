#!/bin/bash -eu

cd $(dirname $(realpath $0))/..

APP_NAME=$(basename $(realpath .))

docker build -t $APP_NAME .

docker run -d \
    --name $APP_NAME \
    --restart=unless-stopped \
    $APP_NAME \

