/*
Package entropy provides deterministic entropy providers for systems that need
reproducible pseudo-random streams without coupling to global process state.

[Context]
The package is designed for infrastructure and simulation-style workflows where
repeatability is a hard requirement. A caller explicitly provides a seed and
receives a provider instance that owns its own internal state.

[Design]
The API exposes function fields through EntropyProvider instead of hidden
package-level state, making mutation boundaries explicit and keeping usage
ergonomic in low-level code.

[Side Effects]
Provider methods mutate only provider-local generator state. No global state,
I/O, or synchronization primitives are used by this package.
*/
package entropy
