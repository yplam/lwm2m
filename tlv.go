package lwm2m

type TLVType byte
var (
	ObjectInstance       TLVType = 0 // Object Instance in which case the Value contains one or more Resource TLVs
	MultipleResourceItem TLVType = 1 // Resource Instance with Value for use	within a multiple Resource TLV
	MultipleResource     TLVType = 2 // multiple Resource, in which case the	Value contains one or more Resource Instance TLVs
	SingleResource       TLVType = 3 // Resource with Value
)

type TLV struct {
	Type TLVType
	Identifier uint16
	Value []byte
	Length uint32
}
