# Plane.Watch Websocket Broker

This is the binary that website clients talk to. The clients start a websocket session and request which
tiles they are interested in. The list of tiles can be fetched from the `/tiles` endpoint.

The `--serve-test-web` option serves up the test web page that shows how to use it.