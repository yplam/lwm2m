package model

type Resource struct {
	Id       uint16
	Multiple bool
	Values   map[uint16][]byte
}

func (r Resource) ID() uint16 {
	return r.Id
}

func NewResource(id uint16, isMultiple bool) *Resource {
	return &Resource{
		Id:       id,
		Multiple: isMultiple,
		Values:   make(map[uint16][]byte),
	}
}

func (r *Resource) SetValue(v []byte) {
	r.Values[0] = v
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
