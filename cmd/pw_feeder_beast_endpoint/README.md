# Plane.watch Feeder Beast Endpoint

This binary is used to receive data from multiple individual sources and put it into the pipeline.

It can read from Beast sources via stunnel (eg: feeder container).

## Testing

To set up a test environment:

* In the `test_data` directory:
  * Edit `stunnel.conf` and add your feeder api key on the `sni = ...` line
  * run `stunnel ./stunnel.conf`
* Run the binary: `go run main.go -cert ./test_data/testcert.pem -key ./test_data/testkey.pem -debug --atcurl "http://<atc_api_base_url>" --atcuser "<atc_api_user>" --atcpass "<atc_api_password>"`
* Wait to see `DBG updating api key cache from atc`
* Run `telnet 127.0.0.1 22345` and send some data!