package uids

import (
	_ "crypto/sha256"
	_ "crypto/sha3"
	_ "crypto/sha512"
	"sync"

	_ "golang.org/x/crypto/blake2b"
)

var methods = map[string]UID{
	UuidV4.name: *UuidV4,
	NanoID.name: *NanoID,
	Cuid2.name:  *Cuid2,
	UuidV7.name: *UuidV7,
	Ulid.name:   *Ulid,
	XId.name:    *XId,
}

var lock = new(sync.RWMutex)

func Register(uid UID) {
	lock.Lock()
	defer lock.Unlock()

	methods[uid.name] = uid
}

func Get(name string) (*UID, error) {
	lock.Lock()
	defer lock.Unlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}
	return &alg, nil
}

func Supported() []UID {
	lock.Lock()
	defer lock.Unlock()

	var list []UID
	for _, uid := range methods {
		list = append(list, uid)
	}
	return list
}
