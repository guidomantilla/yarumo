package uids

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// methods is the package-level registry of UID generators. It starts
// empty; callers populate it via Register. The canonical generators
// shipped by modules/extension/common/uids/ register themselves via that
// package's init().
var methods = map[string]UID{}

// lock guards every access to the methods map.
var lock = new(sync.RWMutex)

// Register adds a UID generator to the registry. Re-registering under the
// same name overwrites the previous entry.
func Register(uid UID) {
	cassert.NotNil(uid, "uid is nil")

	lock.Lock()
	defer lock.Unlock()

	methods[uid.Name()] = uid
}

// Lookup retrieves a registered UID generator by name. It returns an error
// wrapping ErrAlgorithmNotSupported when no generator is registered under
// name.
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

// Supported returns a slice of all registered UID generators in unspecified order.
func Supported() []UID {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]UID, 0, len(methods))
	for _, u := range methods {
		list = append(list, u)
	}

	return list
}
