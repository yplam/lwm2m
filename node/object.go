package node

import (
	"fmt"
	"strings"
)

type Object struct {
	Id        uint16
	Instances map[uint16]ObjectInstance
}

func (o Object) ID() uint16 {
	return o.Id
}

func (o Object) String() string {
	var b strings.Builder
	b.WriteString("Object {")
	b.WriteString(fmt.Sprintf("id: %v, ", o.Id))
	b.WriteString("instance: [")
	for k, v := range o.Instances {
		b.WriteString(fmt.Sprintf("%v:", k))
		b.WriteString(v.String())
	}
	b.WriteString("]")
	b.WriteString("}")
	return b.String()
}

type ObjectInstance struct {
	Id        uint16
	Resources map[uint16]Resource
}

func (i ObjectInstance) ID() uint16 {
	return i.Id
}

func (i ObjectInstance) String() string {
	var b strings.Builder
	b.WriteString("OBI {")
	b.WriteString(fmt.Sprintf("id: %v, ", i.Id))
	b.WriteString("Resources: [")
	for k, v := range i.Resources {
		b.WriteString(fmt.Sprintf("%v:", k))
		b.WriteString(v.String())
	}
	b.WriteString("]")
	b.WriteString("}")
	return b.String()
}

func NewObject(id uint16) Object {
	return Object{
		Id:        id,
		Instances: make(map[uint16]ObjectInstance),
	}
}

func NewObjectInstance(id uint16) ObjectInstance {
	return ObjectInstance{
		Id:        id,
		Resources: make(map[uint16]Resource),
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
	Resources    map[uint16]*ResourceDefinition
}
