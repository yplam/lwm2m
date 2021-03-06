package lwm2m

type ResourceId uint16

type Resource struct {
	ID ResourceId
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
	ID          ResourceId
	Name        string
	Description string
	Operations  ResourceOperations
	Multiple    bool
	Mandatory   bool
	Type        ResourceType
}
