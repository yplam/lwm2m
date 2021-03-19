package model

type Object struct {
	Id        uint16
	Instances map[uint16]*ObjectInstance
}

func (o *Object) ID() uint16 {
	return o.Id
}

type ObjectInstance struct {
	Id        uint16
	Resources map[uint16]*Resource
}

func (i ObjectInstance) ID() uint16 {
	return i.Id
}

func NewObject(id uint16) *Object {
	return &Object{
		Id:        id,
		Instances: make(map[uint16]*ObjectInstance),
	}
}

func NewObjectInstance(id uint16) *ObjectInstance {
	return &ObjectInstance{
		Id:        id,
		Resources: make(map[uint16]*Resource),
	}
}

type ObjectDefinition struct {
	Id           uint16
	Name         string
	Description  string
	Multiple     bool
	Mandatory    bool
	Version      string
	LWM2MVersion string
	URN          string
	Resources    []*ResourceDefinition
}
