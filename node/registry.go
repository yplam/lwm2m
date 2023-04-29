package node

// ref http://openmobilealliance.org/tech/profiles/LWM2M.xsd

import (
	"embed"
	"encoding/xml"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var (
	_registryOnce   sync.Once
	_globalRegistry *Registry
)

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

type Registry struct {
	objs map[uint16]*ObjectDefinition
	mux  sync.RWMutex
}

func GetRegistry() *Registry {
	_registryOnce.Do(func() {
		_globalRegistry = newRegistry()
		_globalRegistry.loadFromFS(regDefaultDir, "definition")
	})
	return _globalRegistry
}

func ConfigRegistry(paths ...string) *Registry {
	_registryOnce.Do(func() {
		_globalRegistry = newRegistry()
		_globalRegistry.loadFromFS(_wrapFS{}, paths...)
	})
	return _globalRegistry
}

func newRegistry() *Registry {
	return &Registry{
		objs: make(map[uint16]*ObjectDefinition),
	}
}

func (r *Registry) Append(paths ...string) {
	r.loadFromFS(_wrapFS{}, paths...)
}

func (r *Registry) GetObjectDefinition(id uint16) (*ObjectDefinition, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if obj, ok := r.objs[id]; ok {
		return obj, nil
	}
	return nil, ErrNotFound
}

func (r *Registry) loadFromFS(fsw _FS, paths ...string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	for _, p := range paths {
		ft, err := fsw.Open(p)
		if err != nil {
			continue
		}
		defer ft.Close()
		fi, err := ft.Stat()
		if err != nil {
			continue
		}
		if !fi.IsDir() {
			content, err := fsw.ReadFile(p)
			if err != nil {
				continue
			}
			obj, err := loadObjectDefinition(content)
			if err != nil {
				continue
			}
			r.objs[obj.Id] = obj
			continue
		}
		files, err := fsw.ReadDir(p)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			content, err := fsw.ReadFile(filepath.Join(p, f.Name()))
			if err != nil {
				continue
			}
			obj, err := loadObjectDefinition(content)
			if err != nil {
				continue
			}
			r.objs[obj.Id] = obj
		}
	}
}

func (r *Registry) DetectResourceType(p Path) (ResourceType, error) {
	if !(p.IsResource() || p.IsResourceInstance()) {
		return R_NONE, ErrPathInvalidValue
	}
	if o, ok := r.objs[uint16(p.objectId)]; ok {
		if r, ok := o.Resources[uint16(p.resourceId)]; ok {
			return r.Type, nil
		}
	}
	return R_NONE, ErrNotFound
}

func loadObjectDefinition(x []byte) (*ObjectDefinition, error) {
	var xx xLWM2M
	if err := xml.Unmarshal(x, &xx); err != nil {
		return nil, err
	}
	xo := xx.Object
	if xo.ObjectID <= 0 {
		return nil, errors.New("no object definition found")
	}
	var res = make(map[uint16]*ResourceDefinition)
	for _, v := range xo.Resources.Item {
		//rv, _ := json.MarshalIndent(v, "", "\t")
		//logrus.Warn(string(rv))
		res[v.ID] = &ResourceDefinition{
			ID:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			Operations:  strToResourceOperations(v.Operations),
			Multiple:    v.MultipleInstances == "Multiple",
			Mandatory:   v.Mandatory == "Mandatory",
			Type:        strToResourceType(v.Type),
		}
	}
	return &ObjectDefinition{
		Id:           xo.ObjectID,
		Name:         xo.Name,
		Description:  xo.Description1,
		Multiple:     xo.MultipleInstances == "Multiple",
		Mandatory:    xo.Mandatory == "Mandatory",
		Version:      xo.ObjectVersion,
		LWM2MVersion: xo.LWM2MVersion,
		URN:          xo.ObjectURN,
		Resources:    res,
	}, nil
}

func strToResourceType(str string) ResourceType {
	switch str {
	case "String":
		return R_STRING
	case "Integer":
		return R_INTEGER
	case "Float":
		return R_FLOAT
	case "Boolean":
		return R_BOOLEAN
	case "Opaque":
		return R_OPAQUE
	case "Time":
		return R_TIME
	case "Objlnk":
		return R_OBJLNK
	default:
		return R_NONE

	}
}

func strToResourceOperations(str string) ResourceOperations {
	switch str {
	case "R":
		return OP_R
	case "W":
		return OP_W
	case "RW":
		return OP_RW
	case "E":
		return OP_E
	default:
		return OP_NONE
	}
}

type xLWM2M struct {
	Object xObjectDefinition
}

// <Name>Temperature</Name>
// <Description1>This IPSO object should be used with a temperature sensor to report a temperature measurement.  It also provides resources for minimum/maximum measured values and the minimum/maximum range that can be measured by the temperature sensor. An example measurement unit is degrees Celsius.</Description1>
// <ObjectID>3303</ObjectID>
// <ObjectURN>urn:oma:github.com/yplam/lwm2m:ext:3303:1.1</ObjectURN>
// <LWM2MVersion>1.0</LWM2MVersion>
// <ObjectVersion>1.1</ObjectVersion>
// <MultipleInstances>Multiple</MultipleInstances>
// <Mandatory>Optional</Mandatory>
type xObjectDefinition struct {
	ObjectID          uint16
	Name              string
	Description1      string
	MultipleInstances string
	Mandatory         string
	Resources         xResourcesDefinition
	ObjectVersion     string
	LWM2MVersion      string
	ObjectURN         string
}

type xResourcesDefinition struct {
	Item []xResourceDefinition
}

// <Item ID="5700">
// <Name>Sensor Value</Name>
// <Operations>R</Operations>
// <MultipleInstances>Single</MultipleInstances>
// <Mandatory>Mandatory</Mandatory>
// <EventType>Float</EventType>
// <RangeEnumeration></RangeEnumeration>
// <Units></Units>
// <Description>Last or Current Measured Value from the Sensor.</Description>
// </Item>
type xResourceDefinition struct {
	ID                uint16 `xml:"ID,attr"`
	Name              string
	Description       string
	Operations        string
	MultipleInstances string
	Mandatory         string
	Type              string
}

//go:embed definition/*.xml
var regDefaultDir embed.FS

type _FS interface {
	fs.ReadFileFS
	fs.ReadDirFS
}

type _wrapFS struct{}

func (_ _wrapFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (_ _wrapFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (_ _wrapFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

var _ _FS = (*_wrapFS)(nil)
