# lwm2m

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
  * [ ] Execute Operation
  * [x] Discover Operation
  * [ ] Create Operation
  * [ ] Delete Operation
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
  * [ ] DTLS with PSK
- [x] Transport
  * [x] UDP transport support.
  * [x] TCP transport support.
- [ ] Tested with clients
  * [x] Leshan client
  * [x] Anjay client running on ESP32
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
  m := core.DefaultManager()
  registration.EnableHandler(r, m)
  err := server.ListenAndServe(r,
    server.EnableUDPListener("udp", ":5683"))
  if err != nil {
    log.Printf("serve lwm2m with err: %v", err)
  }
}

```

## License

Apache License Version 2.0. See the [LICENSE](LICENSE) file for details.