FROM golang:1.18-alpine AS build

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOAMD64=v3

WORKDIR /app

# prefill the build steps
COPY go.mod /app
COPY go.sum /app
RUN go mod verify && go mod download

# Copy the code into the container
COPY . /app/

# Build the application
RUN go build plane.watch/cmd/pw_ingest

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/pw_ingest /app

# limit ourselves to just location updates
ENV PUBLISH_TYPES=location-updates

HEALTHCHECK --interval=15s --timeout=3s CMD wget --spider http://localhost:9602/status

# Command to run
CMD /app/pw_ingest daemon
