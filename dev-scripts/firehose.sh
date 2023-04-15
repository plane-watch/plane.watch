bin/pw_ingest \
  --fetch=beast://crawled.mapwithlove.com:3004 \
  --sink=nats://localhost:4222 \
  --debug \
  simple
