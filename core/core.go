package core

import (
	"context"
	"errors"
	"github.com/yplam/lwm2m/encoding"
	"github.com/yplam/lwm2m/node"
)

var (
	ErrUnexpectedResponseCode = errors.New("unexpected response code")
)

type Observer interface {
	ObserveSync(p node.Path, onMsg ObserveFunc) error
	Observe(p node.Path, onMsg ObserveFunc) error
	ObserveObject(p node.Path, onMsg ObserveObjectFunc) error
	ObserveResource(p node.Path, onMsg ObserveResourceFunc) error
	CancelObserve(p node.Path) error
}

type DeviceManager interface {
	Read(ctx context.Context, p node.Path) ([]node.Node, error)
	ReadObject(ctx context.Context, p node.Path) (*node.Object, error)
	ReadResource(ctx context.Context, p node.Path) (*node.Resource, error)
	Write(ctx context.Context, p node.Path, val ...node.Node) error
	WriteResource(ctx context.Context, p node.Path, val *node.Resource) error
	WriteObjectInstance(ctx context.Context, p node.Path, val *node.ObjectInstance) error
	Discover(ctx context.Context, p node.Path) ([]*encoding.CoreLink, error)
}
