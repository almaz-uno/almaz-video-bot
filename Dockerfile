FROM golang:1.20-bullseye

COPY . /app

ENTRYPOINT [ "/app/docker-entrypoint.sh" ]

