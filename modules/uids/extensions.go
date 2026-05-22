package uids

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// methods holds the registry of UID generators. The parent module ships
// no preconfigured entries; consumers explicitly call Register at startup
// for the providers they want to look up by name (no init()-based auto
// registration).
var methods = map[string]UID{}

var lock = new(sync.RWMutex)

// Register adds a UID generator to the registry. Intended to be called
// explicitly at application startup for each provider singleton the
// consumer wants to look up by name (e.g. uids.Register(uuid.UuidV4)).
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
