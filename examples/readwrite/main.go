package main

import (
	"context"
	"fmt"
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
		if device.HasObjectWithInstance(3347) {
			go func() {
				ctx := context.Background()
				op, _ := node.NewPathFromString("/3311")
				oip, _ := node.NewPathFromString("/3311/0")
				rp, _ := node.NewPathFromString("/3311/0/5850")
				if n, err := device.Discover(ctx, op); err == nil {
					fmt.Printf("links %v\n", n)
				}

				if v, err := device.Read(ctx, rp); err == nil {
					fmt.Printf("Read: %v\n", v)
				}
				time.Sleep(time.Second * 5)
				if v, err := device.ReadResource(ctx, rp); err == nil {
					fmt.Printf("Resource: %v\n", v)
				}
				time.Sleep(time.Second * 5)
				if v, err := device.ReadObject(ctx, op); err == nil {
					fmt.Printf("Object: %v\n", v)
				}
				time.Sleep(time.Second * 5)
				value := encoding.NewTlv(encoding.TlvSingleResource, 5850, false)
				if value != nil {
					if res, err := node.NewSingleResource(rp, value); err == nil {
						if err = device.WriteResource(ctx, rp, res); err == nil {
							fmt.Printf("Write resource ok\n")
						}
					}
				}
				time.Sleep(time.Second * 5)
				if v, err := device.ReadResource(ctx, rp); err == nil {
					fmt.Printf("Resource: %v\n", v)
				}
				time.Sleep(time.Second * 5)
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
				time.Sleep(time.Second * 5)
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
