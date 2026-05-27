package cache

// ResolveKeyPrefix returns the namespace prefix backends should prepend
// to every logical key. When configured is non-empty it wins; otherwise
// the default "<name>:" is returned. Backends that share underlying
// storage (redis sharing a DB, memcached sharing an instance) use this
// to keep caches with different names from colliding; backends with
// per-instance storage (in-memory maps) may still apply the prefix for
// uniformity.
func ResolveKeyPrefix(name, configured string) string {
	if configured != "" {
		return configured
	}

	return name + ":"
}
