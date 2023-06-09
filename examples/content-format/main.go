package main

import (
	"context"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/yplam/lwm2m/core"
	"github.com/yplam/lwm2m/encoding"
	"github.com/yplam/lwm2m/node"
	"github.com/yplam/lwm2m/registration"
	"github.com/yplam/lwm2m/server"
	"log"
	"time"
)

func onDeviceStateChange(e core.DeviceEvent, m core.Manager) {
	device := e.Device
	switch e.EventType {
	case core.DevicePostRegister:
		device.SetMediaTypes(message.TextPlain, message.TextPlain)
		if device.HasObjectWithInstance(1) {
			go func() {
				ctx := context.Background()
				op, _ := node.NewPathFromString("/1")
				oip, _ := node.NewPathFromString("/1/0")
				rp, _ := node.NewPathFromString("/1/0/1")
				if n, err := device.Discover(ctx, op); err == nil {
					fmt.Printf("links %v\n", n)
				}

				if v, err := device.Read(ctx, rp); err == nil {
					fmt.Printf("Read: %v\n", v)
				}
				time.Sleep(time.Second * 1)
				if v, err := device.ReadResource(ctx, rp); err == nil {
					fmt.Printf("Resource: %v\n", v)
				}
				time.Sleep(time.Second * 1)
				if v, err := device.ReadObject(ctx, op); err == nil {
					fmt.Printf("Object: %v\n", v)
				}
				time.Sleep(time.Second * 1)
				value := encoding.NewTlv(encoding.TlvSingleResource, 1, uint16(123))
				if value != nil {
					if res, err := node.NewSingleResource(rp, value); err == nil {
						if err = device.WriteResource(ctx, rp, res); err == nil {
							fmt.Printf("Write resource ok\n")
						} else {
							fmt.Println("error writing resource: ", err)
						}
					} else {
						fmt.Println("error creating resource", err)
					}
				}
				time.Sleep(time.Second * 1)
				if v, err := device.ReadResource(ctx, rp); err == nil {
					fmt.Printf("Resource: %v\n", v)
				}
				time.Sleep(time.Second * 1)
				value = encoding.NewTlv(encoding.TlvSingleResource, 5850, true)
				if value != nil {
					if res, err := node.NewSingleResource(rp, value); err == nil {
						obi := node.NewObjectInstance(0)
						obi.Resources[5850] = res
						if err = device.WriteObjectInstance(ctx, oip, obi); err == nil {
							fmt.Printf("Write object instance ok\n")
						}
					}

				}
				time.Sleep(time.Second * 1)
				if v, err := device.ReadResource(ctx, rp); err == nil {
					fmt.Printf("Resource: %v\n", v)
				}
			}()

		}
	default:

	}
}

func main() {
	r := server.DefaultRouter()
	manager := core.DefaultManager()
	manager.OnDeviceStateChange(onDeviceStateChange)
	registration.EnableHandler(r, manager)

	err := server.ListenAndServe(r,
		server.EnableUDPListener("udp", ":5683"))
	if err != nil {
		log.Printf("Serve lwm2m with err: %v\n", err)
	}
}
