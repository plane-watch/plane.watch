FROM golang:1.21 AS build

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOAMD64=v3

ARG IMAGE_VERSION=dev

WORKDIR /app

# prefill the build steps
COPY go.mod /app
COPY go.sum /app
RUN go mod verify && go mod download

# Copy the code into the container
COPY . /app/

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Build the application
RUN set -x && go build -buildvcs=false -ldflags="-X main.version=${IMAGE_VERSION//\"/}" plane.watch/cmd/pw_discord_bot

# test binary (make sure it will run on this arch)
RUN ./pw_discord_bot --help

FROM alpine:latest as compressor
WORKDIR /app
COPY --from=build /app/pw_discord_bot /app

RUN apk --no-cache add upx=4.2.1-r0
RUN upx --brute -9 /app/pw_discord_bot

# test binary (make sure it runs after compressing)
RUN ./pw_discord_bot --help

FROM alpine:latest as dev
LABEL org.opencontainers.image.description="Plane.Watch ATC API"

WORKDIR /app
COPY --from=build /app/pw_discord_bot /app

HEALTHCHECK --interval=15s --timeout=3s CMD wget --spider http://localhost:9602/status

# Command to run
CMD [ "/app/pw_atc_api", "daemon" ]

FROM alpine:latest
LABEL org.opencontainers.image.description="Plane.Watch Discord Bot"

WORKDIR /app
COPY --from=compressor /app/pw_discord_bot /app

# Command to run
ENTRYPOINT [ "/app/pw_discord_bot" ]

# default args for it
CMD [ "daemon" ]
