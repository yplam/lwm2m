package lwm2m

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

type Resource struct {
	id         uint16
	isMultiple bool
	resType    ResourceType
	data       []byte
	instances  map[uint16]Resource
	path       Path
}

func (r Resource) addInstance(val Resource) {
	if !val.path.IsResourceInstance() {
		return
	}
	if r.isMultiple == false {
		r.isMultiple = true
	}
	riid, _ := val.path.ResourceInstanceId()
	r.instances[riid] = val
}

func (r Resource) ID() uint16 {
	return r.id
}

func (r Resource) Data() []byte {
	return r.data
}

func (r Resource) String() string {
	var b strings.Builder
	b.WriteString("Resource { ")
	b.WriteString(fmt.Sprintf("id: %v, ", r.id))
	b.WriteString(fmt.Sprintf("type: %v, ", r.resType))
	b.WriteString(fmt.Sprintf("data: %v, ", r.Value()))
	if r.isMultiple {
		b.WriteString("instances: [ ")
		for k, v := range r.instances {
			b.WriteString(fmt.Sprintf("%v:", k))
			b.WriteString(v.String())
		}
		b.WriteString("] ")
	}
	b.WriteString("}")
	return b.String()
}

func (r Resource) StringValue() string {
	return fmt.Sprintf("%v", r.Value())
}

func (r Resource) Value() interface{} {
	switch r.resType {
	case R_NONE:
		return r.data
	case R_STRING:
		return string(r.data)
	case R_INTEGER:
		return r.Integer()
	case R_FLOAT:
		return r.Float()
	case R_BOOLEAN:
	case R_OPAQUE:
	case R_TIME:
	case R_OBJLNK:
	}
	return nil
}

func (r Resource) Float() float64 {
	if len(r.data) == 4 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(r.data)))
	} else if len(r.data) == 8 {
		return math.Float64frombits(binary.BigEndian.Uint64(r.data))
	} else {
		return 0
	}
}

func (r Resource) Integer() (val int64) {
	buff := bytes.NewBuffer(r.data)
	l := len(r.data)
	if l == 1 {
		var i1 int8
		_ = binary.Read(buff, binary.BigEndian, &i1)
		val = int64(i1)
	} else if l == 2 {
		var i2 int16
		_ = binary.Read(buff, binary.BigEndian, &i2)
		val = int64(i2)
	} else if l == 4 {
		var i4 int32
		_ = binary.Read(buff, binary.BigEndian, &i4)
		val = int64(i4)
	} else if l == 8 {
		var i8 int64
		_ = binary.Read(buff, binary.BigEndian, &i8)
		val = i8
	}
	return
}

func NewResource(p Path, isMultiple bool, data []byte) (r Resource, err error) {
	objID, err := p.ObjectId()
	if err != nil {
		return
	}
	resID, err := p.ResourceId()
	if err != nil {
		return
	}
	reg := GetRegistry()
	objDef, err := reg.GetObjectDefinition(objID)
	if err != nil {
		return
	}
	resDef, ok := objDef.Resources[resID]
	if !ok {
		err = ErrNotFound
		return
	}
	r = Resource{
		id:         resID,
		isMultiple: isMultiple,
		resType:    resDef.Type,
		data:       data,
		instances:  make(map[uint16]Resource),
		path:       p,
	}
	return
}

type ResourceOperations byte

var (
	OP_NONE ResourceOperations = 0
	OP_R    ResourceOperations = 1
	OP_W    ResourceOperations = 2
	OP_RW   ResourceOperations = 3
	OP_E    ResourceOperations = 4
)

type ResourceType byte

func (r ResourceType) String() string {
	switch r {
	case R_NONE:
		return "R_NONE"
	case R_STRING:
		return "R_STRING"
	case R_INTEGER:
		return "R_INTEGER"
	case R_FLOAT:
		return "R_FLOAT"
	case R_BOOLEAN:
		return "R_BOOLEAN"
	case R_OPAQUE:
		return "R_OPAQUE"
	case R_TIME:
		return "R_TIME"
	case R_OBJLNK:
		return "R_OBJLNK"
	default:
		return ""
	}
}

var (
	R_NONE    ResourceType = 0
	R_STRING  ResourceType = 1
	R_INTEGER ResourceType = 2
	R_FLOAT   ResourceType = 3
	R_BOOLEAN ResourceType = 4
	R_OPAQUE  ResourceType = 5
	R_TIME    ResourceType = 6
	R_OBJLNK  ResourceType = 7
)

type ResourceDefinition struct {
	ID          uint16
	Name        string
	Description string
	Operations  ResourceOperations
	Multiple    bool
	Mandatory   bool
	Type        ResourceType
}
