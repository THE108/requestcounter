# Request Counter

`Request Counter` application - Go HTTP server that on each request responds with a counter of the total number of requests that it has received during the last time period (could be configurable).
`Request Counter` can store it's data in memory only or persist it to a file.

`Request Counter` stores counts into intervals (buckets). Each of them represent particular time interval.

Intervals are sored in a ring.

First two uint64 values in data are metainfo:
  * Current position in the ring
  * Time of first access of interval on current position

## HTTP interface

To get request count execute next `curl` command:
```
curl http://localhost:8080/requestcount
```

GET `/requestcount` return json:
```
{
    "count":3
}
```

## Installation

To install `Request Counter` application `glide` (https://github.com/Masterminds/glide) package manager must be installed.

To build `Request Counter` you can just execute:
```
make
```

To run tests:
```
make test
```

To run the application:
```
./requestcounter -config config-file.yaml
```

## Configuration

```
# address to listen on
host: 0.0.0.0

# port to listen on
port: 8080

# logging lever, could be: error, warning, info, debug
log-level: error

# count of intervals (buckets)
interval-count: 100

# duration of each interval 
interval-duration: 600ms

# persist data to a file or store in memory
persistent: true

# file name where data will be persisted
filename: /tmp/reqcnt.dat

# flush data to a file time interval
persist-duration: 5s
```

See `example-config.yaml`.

## Limitations
 - `Request Counter` would not return correct values if config values `interval-count` or `interval-duration` were changed before restart.

## TODO
 - Add more unit and functional tests