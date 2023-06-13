package encoding

import "errors"

var (
	ErrInvalidLength = errors.New("invalid data length")
	ErrNotEnoughData = errors.New("not enough data")
	ErrNotBoolean    = errors.New("not a boolean representation")
	ErrUnknownType   = errors.New("unknown type")
)

type Valuer interface {
	StringVal() string
	Integer() (val int64, err error)
	Float() (float64, error)
	Boolean() (val bool, err error)
	Opaque() []byte
	Time() (int64, error)
	ObjectLink() (uint16, uint16, error)
	Raw() []byte
}
