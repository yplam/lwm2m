package bootstrap

import (
	"context"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/yplam/lwm2m/core"
	"github.com/yplam/lwm2m/encoding"
	"github.com/yplam/lwm2m/node"
	"io"
)

type Client struct {
	Endpoint   string
	conn       mux.Conn
	selectedCt message.MediaType
}

func (c *Client) ContentType() message.MediaType {
	return c.selectedCt
}

func (c *Client) SetContentType(ct message.MediaType) {
	c.selectedCt = ct
}

func (c *Client) finish(ctx context.Context) (err error) {
	res, err := c.conn.Post(ctx, "/bs", c.selectedCt, nil)
	if err != nil {
		return err
	}
	if res.Code() != codes.Changed {
		return fmt.Errorf("unexpected response code: %v", res.Code())
	}
	return nil
}

func (c *Client) Conn() mux.Conn {
	return c.conn
}

func (c *Client) Write(ctx context.Context, path node.Path, val ...node.Node) error {
	msg, err := node.EncodeMessage(c.selectedCt, val)
	if err != nil {
		return err
	}
	resp, err := c.conn.Put(ctx, path.String(), c.selectedCt, msg)
	if err != nil {
		return err
	}
	if resp.Code() != codes.Changed {
		return fmt.Errorf("unexpected response code: %v", resp.Code())
	}
	return nil
}

func (c *Client) Read(ctx context.Context, path node.Path) ([]node.Node, error) {

	// accept
	buf := make([]byte, 2)
	_, _ = message.EncodeUint32(buf, uint32(c.selectedCt))
	acceptOption := message.Option{
		ID:    message.Accept,
		Value: buf[:],
	}

	msg, err := c.conn.Get(ctx, path.String(), acceptOption)
	if err != nil {
		return nil, err
	}
	if msg.Code() != codes.Content {
		return nil, fmt.Errorf("unexpected response code: %v", msg.Code())
	}
	if msg.Body() == nil {
		return nil, core.ErrEmptyBody
	}
	return node.DecodeMessage(path, msg)
}

func (c *Client) Discover(ctx context.Context, path node.Path) ([]*encoding.CoreLink, error) {

	buf := make([]byte, 2)
	l, _ := message.EncodeUint32(buf, uint32(message.AppLinkFormat))
	r, err := c.conn.Get(ctx, path.String(),
		message.Option{
			ID:    message.Accept,
			Value: buf[:l],
		})
	if err != nil {
		return nil, err
	}
	if r.Code() != codes.Content {
		return nil, fmt.Errorf("unexpected response code: %v", r.Code())
	}
	links := make([]*encoding.CoreLink, 0)
	if r.Body() != nil {
		if b, err := io.ReadAll(r.Body()); err == nil {
			links, _ = encoding.CoreLinksFromString(string(b))
		}
	}
	return links, nil
}

func (c *Client) Delete(ctx context.Context, path node.Path) (err error) {
	resp, err := c.conn.Delete(ctx, path.String())
	if err != nil {
		return err
	}
	if resp.Code() != codes.Deleted {
		return fmt.Errorf("unexpected response code: %v", resp.Code())
	}
	return nil
}

type Provider interface {
	HandleBsRequest(ctx context.Context, device *Client) error
}
