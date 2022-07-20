#!/bin/bash -eu

cd $(dirname $(realpath $0))/..

APP_NAME=$(basename $(realpath .))

. .env

docker build -t $APP_NAME .

docker run -d \
    --name $APP_NAME \
    -v /var/almaz-extractor-bot:/var/almaz-extractor-bot \
    -v /root/.acme.sh:/root/.acme.sh \
    --restart=unless-stopped \
    -p $PORT_REDIRECT \
    $APP_NAME \

