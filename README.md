# lwm2m

![build](https://github.com/yplam/lwm2m/actions/workflows/go.yml/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/yplam/lwm2m)

A simple and lightweight ( but not full feature ) lwm2m server aim to run on edge router.

## Features

- [ ] Bootstrap Interface
- [x] Client Registration interface.
  * [x] Register
  * [x] Update
  * [x] Deregister
- [ ] Device Management and Service Enablement interface.
  * [x] Read Operation, Read Resource, Read Object
  * [x] Write Operation, Write Resource, Write Object Instance
  * [x] Execute Operation
  * [x] Discover Operation
  * [x] Create Operation
  * [x] Delete Operation
  * [ ] Write-Attributes Operation
  * [ ] Read-Composite Operation
  * [ ] Write-Composite Operation
- [x] Information Reporting interface.
  * [x] Observe Operation, Observe Resource, Observe Object
  * [x] Cancel Observation Operation
  * [ ] Observe-Composite Operation
  * [ ] Cancel Observation-Composite Operation
  * [ ] Send Operation
- [ ] Data formats
  * [ ] Plain Text
  * [ ] Opaque
  * [ ] CBOR 
  * [x] TLV
  * [ ] SenML JSON
  * [ ] SenML CBOR
  * [ ] LwM2M JSON
- [ ] Security
  * [ ] DTLS with Certificates
  * [x] DTLS with PSK, only support DTLS 1.2
- [x] Transport
  * [x] UDP transport support.
  * [ ] TCP transport support.(Some features may not work properly)
- [ ] Tested with clients
  * [x] Leshan client: coap, coaps + psk
  * [x] Anjay client running on ESP32: coap
  * [x] Anjay client running on Linux: coap, coaps + psk
  * [x] Zephyr LWM2M client running on nrf52840 with w5500 ethernet: coap, coaps + psk
  * [ ] Zephyr LWM2M client running on nrf52840 with Openthread

## Installation

You need a working Go environment.

```
go get github.com/yplam/lwm2m
```

## Getting Started

```go
package main

import (
  "github.com/yplam/lwm2m/core"
  "github.com/yplam/lwm2m/registration"
  "github.com/yplam/lwm2m/server"
  "log"
)

func main() {
  r := server.DefaultRouter()
  deviceManager := core.DefaultManager()
  registration.EnableHandler(r, deviceManager)
  err := server.ListenAndServe(r,
    server.EnableUDPListener("udp", ":5683"),
  )
  if err != nil {
    log.Printf("serve lwm2m with err: %v", err)
  }
}

```

## Test Commands

leshan client

```shell
# coap 
java -jar leshan-client-demo.jar --server-url=localhost:5683

# coaps
java -jar leshan-client-demo.jar --server-url=localhost:5684 -i 123 --psk-key=000102030405060708090a0b0c0d0e0f
```
anjay client
```shell
# coap
./output/bin/demo --endpoint-name $(hostname) --server-uri coap://127.0.0.1:5683

# coap + tcp
./output/bin/demo --endpoint-name $(hostname) --server-uri coap+tcp://127.0.0.1:5685

# coaps
./output/bin/demo --endpoint-name $(hostname) --server-uri coaps://127.0.0.1:5684 --security-mode psk --identity 666f6f --key 000102030405060708090a0b0c0d0e0f --ciphersuites 49320 --tls-version TLSv1.2
```

## License

Apache License Version 2.0. See the [LICENSE](LICENSE) file for details.