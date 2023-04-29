package node

import (
	"fmt"
	"github.com/yplam/lwm2m/encoding"
	"strings"
)

type ResourceInstance struct {
	id      uint16
	resType ResourceType
	data    encoding.Valuer
	path    Path
}

func (r *ResourceInstance) Value() interface{} {
	switch r.resType {
	case R_NONE:
		return r.data
	case R_STRING:
		return r.data.StringVal()
	case R_INTEGER:
		if d, err := r.data.Integer(); err == nil {
			return d
		}
	case R_FLOAT:
		if d, err := r.data.Float(); err == nil {
			return d
		}
	case R_BOOLEAN:
		if d, err := r.data.Boolean(); err == nil {
			return d
		}
	case R_OPAQUE:
		return r.data.Opaque()
	case R_TIME:
		if d, err := r.data.Time(); err == nil {
			return d
		}
	case R_OBJLNK:
		if d0, d1, err := r.data.ObjectLink(); err == nil {
			return [2]uint16{d0, d1}
		}
	}
	return nil
}

func (r *ResourceInstance) ID() uint16 {
	return r.id
}

func (r *ResourceInstance) String() string {
	return fmt.Sprintf("%v", r.Value())
}

func (r *ResourceInstance) Data() encoding.Valuer {
	return r.data
}

func NewResourceInstance(p Path, data encoding.Valuer) (r *ResourceInstance, err error) {
	id, err := p.ResourceInstanceId()
	if err != nil {
		id = 0
		p.resourceInstanceId = 0
	}
	resType, err := GetRegistry().DetectResourceType(p)
	if err != nil {
		return
	}
	r = &ResourceInstance{
		id:      id,
		resType: resType,
		data:    data,
		path:    p,
	}
	return
}

type Resource struct {
	id         uint16
	isMultiple bool
	instances  map[uint16]*ResourceInstance
	path       Path
}

func (r *Resource) SetInstance(val *ResourceInstance) error {
	if val.path.IsResourceInstance() {
		riid, _ := val.path.ResourceInstanceId()
		r.instances[riid] = val
		if riid > 0 {
			r.isMultiple = true
		}
		return nil
	}
	return ErrPathNotMatch
}

func (r *Resource) InstanceCount() int {
	return len(r.instances)
}
func (r *Resource) GetInstance(index uint16) (*ResourceInstance, error) {
	if ins, ok := r.instances[index]; ok {
		return ins, nil
	}
	return nil, ErrNotFound
}

func (r *Resource) ID() uint16 {
	return r.id
}

func (r *Resource) String() string {
	var b strings.Builder
	b.WriteString("Resource { ")
	b.WriteString(fmt.Sprintf("id: %v, ", r.id))
	b.WriteString("instances: [ ")
	for k, v := range r.instances {
		b.WriteString(fmt.Sprintf("%v:", k))
		b.WriteString(v.String())
	}
	b.WriteString("] ")
	b.WriteString("}")
	return b.String()
}

func (r *Resource) Data() encoding.Valuer {
	if i, err := r.GetInstance(0); err == nil {
		return i.Data()
	}
	return nil
}

func NewResource(p Path, isMultiple bool) (r *Resource, err error) {
	resID, err := p.ResourceId()
	if err != nil {
		return
	}
	r = &Resource{
		id:         resID,
		isMultiple: isMultiple,
		instances:  make(map[uint16]*ResourceInstance),
		path:       p,
	}
	return
}

func NewSingleResource(p Path, val encoding.Valuer) (r *Resource, err error) {
	r, err = NewResource(p, false)
	if err != nil {
		return
	}
	p.SetResourceInstanceId(0)
	ri, err := NewResourceInstance(p, val)
	if err != nil {
		return
	}
	err = r.SetInstance(ri)
	return
}
