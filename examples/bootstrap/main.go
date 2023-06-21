package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/yplam/lwm2m/bootstrap"
	"github.com/yplam/lwm2m/encoding"
	"github.com/yplam/lwm2m/node"
	"github.com/yplam/lwm2m/server"
	"log"
)

var ct = ""

type BootstrapProvider struct {
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (p *BootstrapProvider) HandleBsRequest(ctx context.Context, client *bootstrap.Client) error {
	fmt.Println("handle bs request for client: ", client.Endpoint)

	switch ct {
	case "":
		// should use preferred by device
	case "text":
		client.SetContentType(message.TextPlain)
	case "opaque":
		client.SetContentType(message.AppOctets)
	case "tlv":
		client.SetContentType(message.AppLwm2mTLV)
	default:
		return errors.New("unknown content type")
	}

	pp, _ := node.NewPathFromString("/0")
	err := client.Delete(ctx, pp)
	check(err)
	pp, _ = node.NewPathFromString("/1")
	err = client.Delete(ctx, pp)
	check(err)

	if client.ContentType() == message.TextPlain {
		// this scope for plaintext. although it is not in specs it should work
		// let's make a try
		// $ ./wakaama-client -b

		{
			// uri resource
			p := "/0/1/0"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue("coap://localhost:5683")
			val, err := node.NewSingleResource(pp, vv)
			check(err)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// is bootstrap server?
			p := "/0/1/1"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue(false)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// security mode
			p := "/0/1/2"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue(3)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// short server id
			p := "/0/1/10"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue(1)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}

		{
			// short server id
			p := "/1/0/0"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue(1)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// lifetime
			p := "/1/0/1"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue(42)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// notification storing when disabled or offline
			p := "/1/0/6"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue(false)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// binding
			p := "/1/0/7"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewPlainTextValue("U")
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
	} else if client.ContentType() == message.AppOctets {

		{
			// uri resource
			p := "/0/1/0"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue("coap://localhost:5683")
			val, err := node.NewSingleResource(pp, vv)
			check(err)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// is bootstrap server?
			p := "/0/1/1"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue(false)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// security mode
			p := "/0/1/2"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue(3)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// short server id
			p := "/0/1/10"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue(1)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}

		{
			// short server id
			p := "/1/0/0"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue(1)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// lifetime
			p := "/1/0/1"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue(42)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// notification storing when disabled or offline
			p := "/1/0/6"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue(false)
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}
		{
			// binding
			p := "/1/0/7"
			pp, _ := node.NewPathFromString(p)
			vv, _ := encoding.NewOpaqueValue("U")
			val, _ := node.NewSingleResource(pp, vv)
			err = client.Write(ctx, pp, val)
			check(err)
		}

	} else if client.ContentType() == message.AppLwm2mTLV {

		oi0 := node.NewObjectInstance(1)
		{
			// uri resource
			p := "/0/1/0"
			id := uint16(0)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, "coap://localhost:5683")
			res, _ := node.NewSingleResource(pp, val)
			oi0.SetResource(id, res)
		}
		{
			// is bootstrap server?
			p := "/0/1/1"
			id := uint16(1)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, false)
			res, _ := node.NewSingleResource(pp, val)
			oi0.SetResource(id, res)
		}
		{
			// security mode
			p := "/0/1/2"
			id := uint16(2)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, uint16(3))
			res, _ := node.NewSingleResource(pp, val)
			oi0.SetResource(id, res)
		}
		{
			// short server id
			p := "/0/1/10"
			id := uint16(10)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, uint16(1))
			res, _ := node.NewSingleResource(pp, val)
			oi0.SetResource(id, res)
		}

		oi1 := node.NewObjectInstance(0)
		{
			// short server id
			p := "/1/0/0"
			id := uint16(0)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, uint16(1))
			res, _ := node.NewSingleResource(pp, val)
			oi1.SetResource(id, res)
		}
		{
			// lifetime
			p := "/1/0/1"
			id := uint16(1)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, int64(42))
			res, _ := node.NewSingleResource(pp, val)
			oi1.SetResource(id, res)
		}
		{
			// notification storing when disabled or offline
			p := "/1/0/6"
			id := uint16(6)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, false)
			res, _ := node.NewSingleResource(pp, val)
			oi1.SetResource(id, res)
		}
		{
			// binding
			p := "/1/0/7"
			id := uint16(7)
			pp, _ := node.NewPathFromString(p)
			val := encoding.NewTlv(encoding.TlvSingleResource, id, "U")
			res, _ := node.NewSingleResource(pp, val)
			oi1.SetResource(id, res)
		}

		pp, _ = node.NewPathFromString("/0/1")
		err = client.Write(ctx, pp, oi0)
		check(err)
		pp, _ = node.NewPathFromString("/1/0")
		err = client.Write(ctx, pp, oi1)
		check(err)
	}

	{
		// read and discover
		pp, _ = node.NewPathFromString("/1/0/1")
		res, err := client.Read(ctx, pp)
		fmt.Println(res, err)

		pp, _ = node.NewPathFromString("/1/0")
		res, err = client.Read(ctx, pp)
		fmt.Println(res, err)

		pp, _ = node.NewPathFromString("/1")
		res, err = client.Read(ctx, pp)
		fmt.Println(res, err)

		pp, _ = node.NewPathFromString("/0")
		res, err = client.Read(ctx, pp)
		fmt.Println(res, err)

		pp, _ = node.NewPathFromString("/0/1")
		res, err = client.Read(ctx, pp)
		fmt.Println(res, err)

		pp, _ = node.NewPathFromString("/0/1/0")
		res, err = client.Read(ctx, pp)
		fmt.Println(res, err)

		pp, err = node.NewPathFromString("")
		check(err)
		links, err := client.Discover(ctx, pp)
		fmt.Println(links, err)

		pp, err = node.NewPathFromString("/")
		check(err)
		links, err = client.Discover(ctx, pp)
		fmt.Println(links, err)

		pp, _ = node.NewPathFromString("/1")
		links, err = client.Discover(ctx, pp)
		fmt.Println(links, err)

		pp, _ = node.NewPathFromString("/1/0")
		links, err = client.Discover(ctx, pp)
		fmt.Println(links, err)

		pp, _ = node.NewPathFromString("/0/1")
		links, err = client.Discover(ctx, pp)
		fmt.Println(links, err)
	}

	fmt.Println("bootstrap finished without errors")
	return nil
}

func main() {
	flag.StringVar(&ct, "format", "", "text/tlv/opaque. if not set then default(tlv) or preferred by client will be used.")
	flag.Parse()
	fmt.Println("bootstrap server demo")

	r := server.DefaultRouter()
	provider := &BootstrapProvider{}
	bootstrap.EnableHandler(r, provider)
	err := server.ListenAndServe(r,
		server.EnableUDPListener("udp", ":5685"),
	)
	if err != nil {
		log.Printf("serve bootstrap with err: %v", err)
	}
}
