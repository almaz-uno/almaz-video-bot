#!/bin/bash -eu

cd $(dirname $(realpath $0))/..

APP_NAME=$(basename $(realpath .))

docker build -t $APP_NAME .

docker run -d \
    --name $APP_NAME \
    -v /var/media:/var/media \
    -v /root/.acme.sh:/root/.acme.sh \
    --restart=unless-stopped \
    -p 0.0.0.0:8443:18443 \
    $APP_NAME \

