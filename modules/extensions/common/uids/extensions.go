package uids

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Preconfigured registry of available UID generators.
var methods = map[string]UID{
	UuidV4.Name(): UuidV4,
	NanoID.Name(): NanoID,
	Cuid2.Name():  Cuid2,
	UuidV7.Name(): UuidV7,
	Ulid.Name():   Ulid,
	XId.Name():    XId,
}

var lock = new(sync.RWMutex)

// Register adds a UID generator to the registry.
func Register(uid UID) {
	cassert.NotNil(uid, "uid is nil")

	lock.Lock()
	defer lock.Unlock()

	methods[uid.Name()] = uid
}

// Lookup retrieves a registered UID generator by name.
func Lookup(name string) (UID, error) {
	cassert.NotEmpty(name, "name is empty")

	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return alg, nil
}

// Supported returns a slice of all registered UID generators.
func Supported() []UID {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]UID, 0, len(methods))
	for _, u := range methods {
		list = append(list, u)
	}

	return list
}
