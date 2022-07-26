#!/bin/bash -eu

cd $(dirname $(realpath $0))/..

. .env

docker run -d \
    --name minio \
    --restart=unless-stopped \
    -v /var/minio:/data \
    -p 10.8.5.1:9010:9000 \
    -p 9011:9011 \
    -e "MINIO_ROOT_USER=$MINIO_ACCESS" \
    -e "MINIO_ROOT_PASSWORD=$MINIO_SECRET" \
    quay.io/minio/minio server /data --console-address ":9011" \
