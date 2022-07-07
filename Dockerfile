FROM golang:1.18-bullseye

COPY . /app

ENTRYPOINT [ "/app/docker-entrypoint.sh" ]

