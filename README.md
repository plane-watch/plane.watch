# Plane.Watch Pipeline

<<<<<<< HEAD
Development is now continued over at https://github.com/plane-watch/pw-pipeline/

This repo has a few of the things I have done around plane.watch.
=======
This repo contains several tools used to decode, understand and process ADSB information.
>>>>>>> 3581c34592d3b204d2437af529a6d74b3c54daa9

We typically develop against the latest version of Golang.

## Info
 The `pw_ingest`, `pw_router` and `pw_ws_broker` commands are connected together with a message queue. We currently
support:
* RabbitMQ
* Redis PubSub
* Nats.io

## Components

### Commands
You can find commands in the `cmd/` directory

#### Filtering and Finding
There are some small programs to help find examples of ADSB messages

* plane.filter
* df_example_finder
* recorder

#### Displaying

* website_decode

You can find it running at http://jasonplayne.com:8080/. Throw in your ADSB message and it'll show you want it can about
the message.

#### Processing

* pw_ingest
* pw_router
* pw_ws_broker

These three components are used to take incoming ADSB messages (beast, avr, sbs1) decode them, turn them into plane
tracking json blobs and make them available via websocket to a website.

#### Integration

* pw_discord_bot

Allows for basic integration with discord and alerting.

### Libraries

Reusable bits!

#### Decoding and Tracking

* tile_grid
* tracker
* tracker/beast
* tracker/mode_s
* tracker/sbs1
* export

These libs form the basis of the whole decoding part

#### Helpers

The other libs in the `lib/` folder are common shared parts of the larger whole.

## Further Reading

Some Links for More Information around ADSB

* http://airmetar.main.jp/radio/ADS-B%20Decoding%20Guide.pdf
* https://mode-s.org/decode/book-the_1090mhz_riddle-junzi_sun.pdf
* https://pypi.org/project/pyModeS/
* https://mode-s.org/decode/content/mode-s/6-els.html
* https://www.eurocontrol.int/sites/default/files/content/documents/nm/asterix/archives/asterix-cat062-system-track-data-part9-v1.10-122009.pdf

## Building

### Development

    make

That's it. It runs the tests and builds the binaries and puts them into `bin/`

If you want to build a specific binary

    go build plane.watch/cmd/pw_ingest

or you can run it with

    go run plane.watch/cmd/pw_ingest

### Building Docker Containers

<<<<<<< HEAD
### pw_ws_broker
This is our plane.watch websocket broker. Connect to it and speak its language to get your location information
=======
    docker build -t plane.watch/pw_ws_broker:latest -f docker/pw_ws_broker/Dockerfile .
    docker build -t plane.watch/pw_router:latest -f docker/pw_router/Dockerfile .
    docker build -t plane.watch/pw_ingest:latest -f docker/pw_ingest/Dockerfile .
>>>>>>> 3581c34592d3b204d2437af529a6d74b3c54daa9
