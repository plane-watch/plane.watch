FROM golang:1.18-alpine AS build

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOAMD64=v3

WORKDIR /app

# prefill the build steps
COPY go.mod /app
COPY go.sum /app
RUN go mod verify && go mod download

# Copy the code into the container
COPY . /app/

# Build the application
RUN go build -buildvcs=false plane.watch/cmd/pw_discord_bot

# test binary (make sure it will run on this arch)
RUN ./pw_discord_bot --help

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/pw_discord_bot /app
COPY docker/pw_discord_bot/entrypoint.sh /app

# Command to run
CMD [ "/app/entrypoint.sh" ]
