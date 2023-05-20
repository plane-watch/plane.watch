# ATC API

This is a nats bus connected api that efficiently gets information for various parts of the stack.

The file at https://github.com/plane-watch/pw-pipeline/blob/main/lib/export/nats_api.go describes the API parts.

Currently Supported sections are:

* Feeders
* Enrichment
* Search

## Feeders

The Feeders API consists of the following APIs

* v1.feeder.list
* v1.feeder.update-stats

### v1.feeder.list

Get the current list of feeders

#### Request
This request has no payload

```
nats request v1.feeder.list -
```

#### Response
```json
[
    {
        "Id": "int",
        "User": "string",
        "Latitude": "float64",
        "Longitude": "float64",
        "Altitude": "float64",
        "ApiKey": "uuid",
        "FeedDirection": "string",
        "FeedProtocol": "string",
        "Label": "string",
        "MlatEnabled": "bool",
        "Mux": "string"
    }
]
```

### v1.feeder.update-stats

This request is used to update some statistics about a given feeder.

#### Request
Currently it accepts a JSON payload in the form of

```json
{"ApiKey":"xxxx", "LastSeen": "RFC3339 Compliant"}
```

#### Response
This Response has no response payload

## Search
This is so the web page can search for various things, it has the following APIs

* v1.search.airport
* v1.search.route

### v1.search.airport
Search for an airport via Name, Icao code, or IATA code

#### Request

```
nats request v1.search.airport ypph
```

The airport you are searching for

#### Response

```json
[
  {
    "Id": 3351,
    "Name": "Perth International Airport",
    "City": "Perth",
    "Country": "Australia",
    "IataCode": "PER",
    "IcaoCode": "YPPH",
    "Latitude": -31.9403,
    "Longitude": 115.967003,
    "Altitude": 67,
    "Timezone": 8,
    "DstType": "N",
    "CreatedAt": "2021-03-01T09:52:06.114401Z",
    "UpdatedAt": "2021-03-01T09:52:06.114401Z"
  }
]
```

### v1.search.route
Not Yet Implemented
#### Request
#### Response

## Enrichment
This API mimics the original HTTP Api. We have the following apis

* v1.enrich.aircraft
* v1.enrich.route

### v1.enrich.aircraft
Gets additional aircraft information
#### Request
```

```
#### Response

### v1.enrich.route
#### Request
#### Response
