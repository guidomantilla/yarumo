package uids

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Preconfigured registry of available UID generators.
var methods = map[string]UID{
	"UUIDv4": UuidV4,
	"NanoID": NanoID,
	"CUID2":  Cuid2,
	"UUIDv7": UuidV7,
	"ULID":   Ulid,
	"XID":    XId,
}

var (
	lock    = new(sync.RWMutex)
	current = UuidV7
)

// Register adds a UID generator to the registry.
func Register(uid UID) {
	cassert.NotNil(uid, "uid is nil")

	lock.Lock()
	defer lock.Unlock()

	methods[uid.Name()] = uid
}

// Get retrieves a registered UID generator by name.
func Get(name string) (UID, error) {
	cassert.NotEmpty(name, "name is empty")

	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return alg, nil
}

// Use selects the default UID generator from the registry by name.
func Use(name string) error {
	cassert.NotEmpty(name, "name is empty")

	lock.Lock()
	defer lock.Unlock()

	alg, ok := methods[name]
	if !ok {
		return ErrAlgorithmNotSupported(name)
	}

	current = alg

	return nil
}

// Generate delegates to the current default UID generator.
func Generate() string {
	lock.RLock()
	defer lock.RUnlock()

	return current.Generate()
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
