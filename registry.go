package lwm2m

// ref http://openmobilealliance.org/tech/profiles/LWM2M.xsd

import (
	"embed"
	"encoding/xml"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
)

//go:embed definition/*.xml
var regDefaultDir embed.FS

var (
	_reg      *Registry
	_reg_lock sync.Mutex
)

type Registry struct {
	Objs map[uint16]*ObjectDefinition
}

func GetRegistry() *Registry {
	if _reg == nil {
		return UseDefaultRegistry()
	}
	return _reg
}

func UseDefaultRegistry() *Registry {
	_reg_lock.Lock()
	defer _reg_lock.Unlock()
	entities, err := regDefaultDir.ReadDir("definition")
	if err != nil {
		panic("can not load default registry")
	}
	objs := make(map[uint16]*ObjectDefinition, len(entities))
	for _, v := range entities {
		if v.IsDir() {
			continue
		}
		data, err := regDefaultDir.ReadFile("definition/" + v.Name())
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		obj, err := loadObjectDefinition(data)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		objs[uint16(obj.Id)] = obj
	}
	_reg = &Registry{
		Objs: objs,
	}
	return _reg
}

func ConfigRegistry(paths ...string) *Registry {
	_reg_lock.Lock()
	defer _reg_lock.Unlock()
	objs := make(map[uint16]*ObjectDefinition)
	for _, p := range paths {
		p = strings.TrimRight(p, "/")
		files, err := os.ReadDir(p)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			content, err := os.ReadFile(p + "/" + f.Name())
			if err != nil {
				continue
			}
			obj, err := loadObjectDefinition(content)
			if err != nil {
				continue
			}
			objs[uint16(obj.Id)] = obj
		}
	}
	_reg = &Registry{
		Objs: objs,
	}
	return _reg
}

func loadObjectDefinition(x []byte) (*ObjectDefinition, error) {
	var xx xLWM2M
	if err := xml.Unmarshal(x, &xx); err != nil {
		return nil, err
	}
	xo := xx.Object
	if xo.ObjectID < 0 {
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

//<Name>Temperature</Name>
//<Description1>This IPSO object should be used with a temperature sensor to report a temperature measurement.  It also provides resources for minimum/maximum measured values and the minimum/maximum range that can be measured by the temperature sensor. An example measurement unit is degrees Celsius.</Description1>
//<ObjectID>3303</ObjectID>
//<ObjectURN>urn:oma:github.com/yplam/lwm2m:ext:3303:1.1</ObjectURN>
//<LWM2MVersion>1.0</LWM2MVersion>
//<ObjectVersion>1.1</ObjectVersion>
//<MultipleInstances>Multiple</MultipleInstances>
//<Mandatory>Optional</Mandatory>
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

//<Item ID="5700">
//<Name>Sensor Value</Name>
//<Operations>R</Operations>
//<MultipleInstances>Single</MultipleInstances>
//<Mandatory>Mandatory</Mandatory>
//<Type>Float</Type>
//<RangeEnumeration></RangeEnumeration>
//<Units></Units>
//<Description>Last or Current Measured Value from the Sensor.</Description>
//</Item>
type xResourceDefinition struct {
	ID                uint16 `xml:"ID,attr"`
	Name              string
	Description       string
	Operations        string
	MultipleInstances string
	Mandatory         string
	Type              string
}
