package model

// A Node is the base type of lwm2m message, can be one of Object, ObjectInstance, Resource
// One lwm2m message package may contain one or more Node
type Node interface {
	ID() uint16
}
