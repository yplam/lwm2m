package lwm2m

type Object struct {
	ID        uint16
	Def       *ObjectDefinition
	Instances map[uint16]bool
}

func NewObject(id uint16, def *ObjectDefinition) *Object {
	return &Object{
		ID:        id,
		Def:       def,
		Instances: make(map[uint16]bool),
	}
}

type ObjectDefinition struct {
	ID           uint16
	Name         string
	Description  string
	Multiple     bool
	Mandatory    bool
	Version      string
	LWM2MVersion string
	URN          string
	Resources    []*ResourceDefinition
}
