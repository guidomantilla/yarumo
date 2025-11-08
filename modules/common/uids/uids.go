package uids

const (
	UuidV4 = "UUIDv4"
	NanoID = "NANOID"
	Cuid2  = "CUID2"
	UuidV7 = "UUIDv7"
	Ulid   = "ULID"
	XId    = "XId"
)

func GetByName(name string) (UIDFn, error) {
	switch name {
	case UuidV4:
		return UUIDv4, nil
	case NanoID:
		return NANOID, nil
	case Cuid2:
		return CUID2, nil
	case UuidV7:
		return UUIDv7, nil
	case Ulid:
		return ULID, nil
	case XId:
		return XID, nil
	default:
		return nil, ErrUIDFunctionNotFound(name)
	}
}
