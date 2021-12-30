package lwm2m

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrPathNilValue       = errors.New("id is nil")
	ErrPathInvalidValue   = errors.New("invalid path string value")
	ErrPathInvalidIdValue = errors.New("invalid path id value")
)

// Path is LWM2M resource Path Real id value is uint16, -1 means null
type Path struct {
	objectId           int32
	objectInstanceId   int32
	resourceId         int32
	resourceInstanceId int32
}

func NewPathFromString(s string) (Path, error) {
	s = strings.Trim(s, "/")
	sps := strings.Split(s, "/")
	if len(sps) > 4 {
		return Path{}, ErrPathInvalidValue
	}
	p := Path{
		objectId:           -1,
		objectInstanceId:   -1,
		resourceId:         -1,
		resourceInstanceId: -1,
	}
	if len(sps) >= 1 && sps[0] != "" {
		tmpId, err := strconv.Atoi(sps[0])
		if err != nil {
			return Path{}, ErrPathInvalidValue
		}
		p.objectId = int32(tmpId)
		if len(sps) >= 2 {
			tmpId, err = strconv.Atoi(sps[1])
			if err != nil {
				return Path{}, ErrPathInvalidValue
			}
			p.objectInstanceId = int32(tmpId)
			if len(sps) >= 3 {
				tmpId, err = strconv.Atoi(sps[2])
				if err != nil {
					return Path{}, ErrPathInvalidValue
				}
				p.resourceId = int32(tmpId)
				if len(sps) == 4 {
					tmpId, err = strconv.Atoi(sps[3])
					if err != nil {
						return Path{}, ErrPathInvalidValue
					}
					p.resourceInstanceId = int32(tmpId)
				}
			}
		}
	}
	if !p.validate() {
		return Path{}, ErrPathInvalidIdValue
	}
	return p, nil
}

func NewObjectPath(objID uint16) Path {
	return Path{
		objectId:           int32(objID),
		objectInstanceId:   -1,
		resourceId:         -1,
		resourceInstanceId: -1,
	}
}

func NewObjectInstancePath(objID, obiID uint16) Path {
	return Path{
		objectId:           int32(objID),
		objectInstanceId:   int32(obiID),
		resourceId:         -1,
		resourceInstanceId: -1,
	}
}

func NewResourcePath(objID, obiID, resourceID uint16) Path {
	return Path{
		objectId:           int32(objID),
		objectInstanceId:   int32(obiID),
		resourceId:         int32(resourceID),
		resourceInstanceId: -1,
	}
}

func NewResourceInstancePath(objID, obiID, resourceID, resiID uint16) Path {
	return Path{
		objectId:           int32(objID),
		objectInstanceId:   int32(obiID),
		resourceId:         int32(resourceID),
		resourceInstanceId: int32(resiID),
	}
}

func (p Path) ObjectId() (uint16, error) {
	if p.objectId < 0 {
		return 0, ErrPathNilValue
	}
	return uint16(p.objectId), nil
}

func (p Path) ObjectInstanceId() (uint16, error) {
	if p.objectInstanceId < 0 {
		return 0, ErrPathNilValue
	}
	return uint16(p.objectInstanceId), nil
}

func (p Path) ResourceId() (uint16, error) {
	if p.resourceId < 0 {
		return 0, ErrPathNilValue
	}
	return uint16(p.resourceId), nil
}

func (p Path) ResourceInstanceId() (uint16, error) {
	if p.resourceInstanceId < 0 {
		return 0, ErrPathNilValue
	}
	return uint16(p.resourceInstanceId), nil
}

func (p Path) IsRoot() bool {
	return p.objectId == -1 && p.objectInstanceId == -1 &&
		p.resourceId == -1 && p.resourceInstanceId == -1
}

func (p Path) IsObject() bool {
	return p.objectId > -1 && p.objectInstanceId == -1 &&
		p.resourceId == -1 && p.resourceInstanceId == -1
}

func (p Path) IsObjectInstance() bool {
	return p.objectId > -1 && p.objectInstanceId > -1 &&
		p.resourceId == -1 && p.resourceInstanceId == -1
}

func (p Path) IsResource() bool {
	return p.objectId > -1 && p.objectInstanceId > -1 &&
		p.resourceId > -1 && p.resourceInstanceId == -1
}

func (p Path) IsResourceInstance() bool {
	return p.objectId > -1 && p.objectInstanceId > -1 &&
		p.resourceId > -1 && p.resourceInstanceId > -1
}

func (p Path) validate() bool {
	if p.IsObject() {
		return p.objectId >= 0 && p.objectId <= 65535
	} else if p.IsObjectInstance() {
		return p.objectId >= 0 && p.objectId <= 65535 &&
			p.objectInstanceId >= 0 && p.resourceInstanceId <= 65534
		// MAX_ID 65535 is a reserved value and MUST NOT be used for identifying an Object Instance.
	} else if p.IsResource() {
		return p.objectId >= 0 && p.objectId <= 65535 &&
			p.objectInstanceId >= 0 && p.resourceInstanceId <= 65534 &&
			p.resourceId >= 0 && p.resourceId <= 65535
	} else if p.IsResourceInstance() {
		return p.objectId >= 0 && p.objectId <= 65535 &&
			p.objectInstanceId >= 0 && p.resourceInstanceId <= 65534 &&
			p.resourceId >= 0 && p.resourceId <= 65535 &&
			p.resourceInstanceId >= 0 && p.resourceInstanceId <= 65535
	} else {
		return p.IsRoot()
	}
}

func (p Path) String() string {
	var b strings.Builder
	b.WriteString("/")
	if p.objectId > -1 {
		_, _ = fmt.Fprintf(&b, "%d", p.objectId)
		if p.objectInstanceId > -1 {
			_, _ = fmt.Fprintf(&b, "/%d", p.objectInstanceId)
			if p.resourceId > -1 {
				_, _ = fmt.Fprintf(&b, "/%d", p.resourceId)
				if p.resourceInstanceId > -1 {
					_, _ = fmt.Fprintf(&b, "/%d", p.resourceInstanceId)
				}
			}
		}
	}
	return b.String()
}
