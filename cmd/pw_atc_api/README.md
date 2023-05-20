# ATC API

This is a nats bus connected api that efficiently gets information for various parts of the stack.

The file at https://github.com/plane-watch/pw-pipeline/blob/main/lib/export/nats_api.go describes the API parts.

Currently Supported sections are:

* Feeders
* Enrichment
* Search

## Bad Subjects
If you request something that pw_atc_api does not understand you will get a generic error

```nats request v1.enrich.donkey asd```

```json
{"error":"Unsupported Request","Type":"v1.enrich.donkey"}
```

if we encountered an error while serving your request, you will get the following error
```json
{"error":"Something went wrong with the request","Type":"<some error info here>"}
```

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
nats request v1.enrich.aircraft 7C33A6
```
The body of the request is the ICAO we want to enrich

#### Response
Successful response example

```json
{
  "Aircraft": {
    "Icao": "7C33A2",
    "Country": "Australia",
    "Registration": "VH-KHG",
    "TypeCode": "PA44",
    "TypeCodeLong": "Piper PA44-180",
    "Serial": "44-8095019",
    "RegisteredOwner": "Aviation Management Services Pty. Ltd.",
    "COFAOwner": "",
    "EngineType": "",
    "FlagCode": "PIPER"
  }
}
```
Failed Response example
```json
{
  "Aircraft": {
    "Icao": null,
    "Country": null,
    "Registration": null,
    "TypeCode": null,
    "TypeCodeLong": null,
    "Serial": null,
    "RegisteredOwner": null,
    "COFAOwner": null,
    "EngineType": null,
    "FlagCode": null
  }
}
```

### v1.enrich.routes
#### Request
```
nats request v1.enrich.routes JSA561
```
The request body is the callsign

#### Response

Successful response example
```json
{
  "Route": {
    "CallSign": "JSA561",
    "Operator": "Jetstar Asia Airways",
    "RouteCode": "WSSS-RPLL",
    "Segments": [
      {
        "Name": "Singapore Changi Airport",
        "ICAOCode": "WSSS"
      },
      {
        "Name": "Ninoy Aquino International Airport",
        "ICAOCode": "RPLL"
      }
    ]
  }
}
```
Failed Response example
```json
{
  "CallSign": null,
  "Operator": null,
  "RouteCode": null,
  "Segments": null
}
```
